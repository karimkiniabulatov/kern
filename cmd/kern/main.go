package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karimkiniabulatov/kern/internal/config"
	"github.com/karimkiniabulatov/kern/internal/cpu"
	"github.com/karimkiniabulatov/kern/internal/disk"
	"github.com/karimkiniabulatov/kern/internal/mem"
	"github.com/karimkiniabulatov/kern/internal/net"
	"github.com/karimkiniabulatov/kern/internal/ui"
	"golang.org/x/term"
)

var (
	flagDisk    = flag.Bool("d", false, "Show disk information")
	flagCPU     = flag.Bool("c", false, "Show CPU information")
	flagMem     = flag.Bool("m", false, "Show memory information")
	flagNet     = flag.Bool("n", false, "Show network information")
	flagAll     = flag.Bool("a", true, "Show all information")
	flagRefresh = flag.Int("refresh", 2, "Refresh interval in seconds")
	flagLang    = flag.String("l", "", "Language code (e.g., 'ru' for Russian)")
	flagRemote  = flag.Int("r", 0, "Start remote API on specified port")
	flagVersion = flag.Bool("v", false, "Show version")
)

const version = "1.0.0"

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Printf("kern v%s\n", version)
		return
	}

	// Load configuration and localization
	cfg, err := config.Load(*flagLang)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// If no specific module is selected, show all
	if !*flagDisk && !*flagCPU && !*flagMem && !*flagNet {
		*flagAll = true
	}

	// If remote mode is enabled, start the remote server and exit
	if *flagRemote != 0 {
		startRemoteServer(cfg, *flagRemote)
		return
	}

	// Start the main monitoring loop
	runMonitor(cfg)
}

func runMonitor(cfg *config.Config) {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize UI
	renderer := ui.NewRenderer(cfg)

	// Set up non-blocking input reading
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err == nil {
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	// Channel for keyboard input
	keyChan := make(chan rune, 1)
	go readKeys(keyChan)

	ticker := time.NewTicker(time.Duration(*flagRefresh) * time.Second)
	defer ticker.Stop()

	// Initial render
	results := collectData(cfg)
	renderer.Render(results)

	for {
		select {
		case <-ticker.C:
			results := collectData(cfg)
			renderer.Render(results)
			
		case key := <-keyChan:
			if key == 'q' || key == 'Q' {
				renderer.Cleanup()
				fmt.Println("\nMonitoring stopped.")
				return
			}
			
		case <-sigChan:
			renderer.Cleanup()
			fmt.Println("\nMonitoring stopped.")
			return
		}
	}
}

func readKeys(keyChan chan rune) {
	reader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := reader.ReadRune()
		if err == nil {
			keyChan <- r
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func collectData(cfg *config.Config) map[string]interface{} {
	type result struct {
		module string
		data   interface{}
		err    error
	}

	modules := make(map[string]bool)
	if *flagAll || *flagDisk {
		modules["disk"] = true
	}
	if *flagAll || *flagCPU {
		modules["cpu"] = true
	}
	if *flagAll || *flagMem {
		modules["mem"] = true
	}
	if *flagAll || *flagNet {
		modules["net"] = true
	}

	resultChan := make(chan result, len(modules))

	// Launch goroutines for each enabled module
	if modules["disk"] {
		go func() {
			data, err := disk.Summary()
			resultChan <- result{"disk", data, err}
		}()
	}

	if modules["cpu"] {
		go func() {
			data, err := cpu.Summary()
			resultChan <- result{"cpu", data, err}
		}()
	}

	if modules["mem"] {
		go func() {
			data, err := mem.Summary()
			resultChan <- result{"mem", data, err}
		}()
	}

	if modules["net"] {
		go func() {
			data, err := net.Summary()
			resultChan <- result{"net", data, err}
		}()
	}

	// Collect results
	results := make(map[string]interface{})
	for i := 0; i < len(modules); i++ {
		res := <-resultChan
		if res.err != nil {
			results[res.module] = map[string]string{"error": res.err.Error()}
		} else {
			results[res.module] = res.data
		}
	}

	return results
}

func startRemoteServer(cfg *config.Config, port int) {
	// Remote server implementation would go here
	log.Printf("Remote API server starting on port %d...", port)
	// This would start the HTTP/gRPC server for remote monitoring
	select {} // Block forever
}