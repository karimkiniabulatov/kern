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
	flagDisk     = flag.Bool("d", false, "Show disk information")
	flagCPU      = flag.Bool("c", false, "Show CPU information")
	flagMem      = flag.Bool("m", false, "Show memory information")
	flagNet      = flag.Bool("n", false, "Show network information")
	flagAll      = flag.Bool("a", false, "Show all information")
	flagRefresh  = flag.Int("refresh", 2, "Refresh interval in seconds")
	flagLang     = flag.String("l", "", "Language code (e.g., 'ru' for Russian)")
	flagRemote   = flag.Int("r", 0, "Start remote API on specified port")
	flagVersion  = flag.Bool("v", false, "Show version")
	flagHelp     = flag.Bool("h", false, "Show help")
	flagDetailed = flag.Bool("detailed", false, "Show detailed CPU core information")
)

const version = "1.0.0"

func main() {
	// Добавляем альтернативные имена флагов
	flag.BoolVar(flagDisk, "disk", false, "Show disk information")
	flag.BoolVar(flagCPU, "cpu", false, "Show CPU information")
	flag.BoolVar(flagMem, "mem", false, "Show memory information")
	flag.BoolVar(flagNet, "net", false, "Show network information")
	flag.BoolVar(flagAll, "all", false, "Show all information")
	flag.BoolVar(flagHelp, "help", false, "Show help")

	flag.Usage = func() {
		fmt.Printf("kern v%s - System Monitoring Tool\n\n", version)
		fmt.Println("Usage: kern [OPTIONS]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  kern                    # Show all information")
		fmt.Println("  kern --cpu --mem        # Show only CPU and memory")
		fmt.Println("  kern -d -l ru           # Disk info with Russian interface")
		fmt.Println("  kern --refresh=5        # Update every 5 seconds")
		fmt.Println("  kern --detailed         # Show detailed CPU core info")
	}

	flag.Parse()

	if *flagHelp {
		flag.Usage()
		return
	}

	if *flagVersion {
		fmt.Printf("kern v%s\n", version)
		return
	}

	// Load configuration and localization
	cfg, err := config.Load(*flagLang)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Если указан порт для remote, запускаем сервер
	if *flagRemote != 0 {
		startRemoteServer(cfg, *flagRemote)
		return
	}

	// Определяем какие модули показывать
	showDisk := *flagDisk || *flagAll || (!*flagCPU && !*flagMem && !*flagNet && !*flagDisk && !*flagNet)
	showCPU := *flagCPU || *flagAll || (!*flagDisk && !*flagMem && !*flagNet && !*flagCPU && !*flagNet)
	showMem := *flagMem || *flagAll || (!*flagDisk && !*flagCPU && !*flagNet && !*flagMem && !*flagNet)
	showNet := *flagNet || *flagAll || (!*flagDisk && !*flagCPU && !*flagMem && !*flagNet && !*flagDisk)

	// Передаем флаги в конфиг
	cfg.ShowDisk = showDisk
	cfg.ShowCPU = showCPU
	cfg.ShowMem = showMem
	cfg.ShowNet = showNet
	cfg.DetailedCPU = *flagDetailed

	// Запускаем мониторинг
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

	ticker := time.NewTicker(time.Duration(cfg.RefreshRate) * time.Second)
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

	resultChan := make(chan result, 4)

	// Launch goroutines only for enabled modules
	if cfg.ShowDisk {
		go func() {
			data, err := disk.Summary()
			resultChan <- result{"disk", data, err}
		}()
	}

	if cfg.ShowCPU {
		go func() {
			data, err := cpu.Summary()
			resultChan <- result{"cpu", data, err}
		}()
	}

	if cfg.ShowMem {
		go func() {
			data, err := mem.Summary()
			resultChan <- result{"mem", data, err}
		}()
	}

	if cfg.ShowNet {
		go func() {
			data, err := net.Summary()
			resultChan <- result{"net", data, err}
		}()
	}

	// Collect results
	results := make(map[string]interface{})
	moduleCount := 0
	if cfg.ShowDisk { moduleCount++ }
	if cfg.ShowCPU { moduleCount++ }
	if cfg.ShowMem { moduleCount++ }
	if cfg.ShowNet { moduleCount++ }

	for i := 0; i < moduleCount; i++ {
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