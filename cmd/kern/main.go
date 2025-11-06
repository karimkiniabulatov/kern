package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karimkiniabulatov/kern/internal/config"
	"github.com/karimkiniabulatov/kern/internal/cpu"
	"github.com/karimkiniabulatov/kern/internal/disk"
	"github.com/karimkiniabulatov/kern/internal/i18n"
	"github.com/karimkiniabulatov/kern/internal/mem"
	"github.com/karimkiniabulatov/kern/internal/net"
	"github.com/karimkiniabulatov/kern/internal/ui"
	"golang.org/x/term"
)

var (
	flagDisk      = flag.Bool("d", false, "Show disk information")
	flagCPU       = flag.Bool("c", false, "Show CPU information")
	flagMem       = flag.Bool("m", false, "Show memory information")
	flagNet       = flag.Bool("n", false, "Show network information")
	flagAll       = flag.Bool("a", false, "Show all information")
	flagRefresh   = flag.Int("refresh", 2, "Refresh interval in seconds")
	flagLang      = flag.String("l", "", "Language code (e.g., 'ru' for Russian)")
	flagRemote    = flag.Int("r", 0, "Start remote API on specified port (default: 28126)")
	flagVersion   = flag.Bool("v", false, "Show version")
	flagHelp      = flag.Bool("h", false, "Show help")
	flagDetailed  = flag.Bool("detailed", false, "Show detailed CPU core information")
	flagDownload  = flag.String("download-lang", "", "Download language pack (e.g., 'fr' for French)")
	flagListLangs = flag.Bool("list-languages", false, "List all supported languages")
	flagLogo      = flag.Bool("logo", false, "Show logo during monitoring")
)

const version = "1.1.0"

func main() {
	// Добавляем альтернативные имена флагов
	flag.BoolVar(flagDisk, "disk", false, "Show disk information")
	flag.BoolVar(flagCPU, "cpu", false, "Show CPU information")
	flag.BoolVar(flagMem, "mem", false, "Show memory information")
	flag.BoolVar(flagNet, "net", false, "Show network information")
	flag.BoolVar(flagAll, "all", false, "Show all information")
	flag.BoolVar(flagHelp, "help", false, "Show help")
	flag.BoolVar(flagLogo, "show-logo", false, "Show logo during monitoring")

	flag.Usage = func() {
		showLogo()
		fmt.Printf("kern v%s - System Monitoring Tool\n\n", version)
		fmt.Println("Usage: kern [OPTIONS]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  kern                       # Show memory and network (default)")
		fmt.Println("  kern --cpu --mem           # Show only CPU and memory")
		fmt.Println("  kern -d -l ru              # Disk info with Russian interface")
		fmt.Println("  kern --refresh=5           # Update every 5 seconds")
		fmt.Println("  kern --detailed            # Show detailed CPU core info")
		fmt.Println("  kern -r 28126              # Start API server on port 28126")
		fmt.Println("  kern --download-lang fr    # Download French language pack")
		fmt.Println("  kern --logo                # Show logo during monitoring")
	}

	flag.Parse()

	if *flagHelp {
		flag.Usage()
		return
	}

	if *flagVersion {
		showLogo()
		return
	}

	if *flagListLangs {
		listSupportedLanguages()
		return
	}

	if *flagDownload != "" {
		downloadLanguagePack(*flagDownload)
		return
	}

	// Load configuration and localization
	cfg, err := config.Load(*flagLang)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Проверяем поддержку языка
	if *flagLang != "" && !i18n.IsLanguageSupported(*flagLang) {
		fmt.Printf("Language '%s' is not supported. Using English.\n", *flagLang)
		fmt.Printf("Use 'kern --download-lang %s' to download language pack.\n", *flagLang)
		*flagLang = "en"
	}

	// Сохраняем язык в конфиг если указан
	if *flagLang != "" {
		cfg.Language = *flagLang
		cfg.Save()
	}

	// Если указан порт для remote, запускаем сервер
	if *flagRemote != 0 {
		port := *flagRemote
		if port == 0 {
			port = 28126 // порт по умолчанию
		}
		startRemoteServer(cfg, port)
		return
	}

	// Определяем какие модули показывать
	// По умолчанию показываем только память и сеть
	showDisk := *flagDisk || *flagAll
	showCPU := *flagCPU || *flagAll
	showMem := *flagMem || *flagAll
	showNet := *flagNet || *flagAll

	// Если не указано никаких модулей, показываем память и сеть по умолчанию
	if !*flagDisk && !*flagCPU && !*flagMem && !*flagNet && !*flagAll {
		showMem = true
		showNet = true
	}

	// Передаем флаги в конфиг
	cfg.ShowDisk = showDisk
	cfg.ShowCPU = showCPU
	cfg.ShowMem = showMem
	cfg.ShowNet = showNet
	cfg.DetailedCPU = *flagDetailed
	if *flagRefresh > 0 {
		cfg.RefreshRate = *flagRefresh
	}

	// Запускаем мониторинг
	runMonitor(cfg, *flagLogo)
}

func showLogo() {
	logo := `
 ██╗  ██╗███████╗██████╗ ███╗   ██╗
 ██║ ██╔╝██╔════╝██╔══██╗████╗  ██║
 █████╔╝ █████╗  ██████╔╝██╔██╗ ██║
 ██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║
 ██║  ██╗███████╗██║  ██║██║ ╚████║
 ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝
 kern v` + version + " - System Monitoring Tool\n"
	fmt.Print("\033[1;36m" + logo + "\033[0m")
}

func runMonitor(cfg *config.Config, showLogo bool) {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize UI
	renderer := ui.NewRenderer(cfg, showLogo)

	// Set up terminal for raw input to make 'q' work properly
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Printf("Failed to set terminal raw mode: %v", err)
	} else {
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	// Channel for keyboard input with buffer
	keyChan := make(chan rune, 10)
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
				fmt.Println("Monitoring stopped.")
				return
			}
			
		case <-sigChan:
			renderer.Cleanup()
			fmt.Println("Monitoring stopped.")
			return
		}
	}
}

func readKeys(keyChan chan rune) {
	reader := bufio.NewReader(os.Stdin)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return
		}
		select {
		case keyChan <- char:
		default:
			// Если канал полный, пропускаем ввод (редкий случай)
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
	if port == 0 {
		port = 28126
	}
	
	showLogo()
	log.Printf("Starting remote API server on port %d...", port)

	// CPU endpoint
	http.HandleFunc("/api/cpu", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := cpu.Summary()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// Memory endpoint
	http.HandleFunc("/api/mem", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := mem.Summary()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// Disk endpoint
	http.HandleFunc("/api/disk", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := disk.Summary()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// Network endpoint
	http.HandleFunc("/api/net", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := net.Summary()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// System info endpoint
	http.HandleFunc("/api/system", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		systemInfo := map[string]interface{}{
			"version": version,
			"time":    time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(systemInfo)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Root endpoint with API info
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		apiInfo := map[string]interface{}{
			"name":    "kern API",
			"version": version,
			"endpoints": []string{
				"/api/cpu",
				"/api/mem", 
				"/api/disk",
				"/api/net",
				"/api/system",
				"/health",
			},
		}
		json.NewEncoder(w).Encode(apiInfo)
	})

	log.Printf("API server running on http://localhost:%d", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET /api/cpu    - CPU information")
	log.Printf("  GET /api/mem    - Memory information")
	log.Printf("  GET /api/disk   - Disk information")
	log.Printf("  GET /api/net    - Network information")
	log.Printf("  GET /api/system - System information")
	log.Printf("  GET /health     - Health check")
	log.Printf("  GET /           - API information")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

func listSupportedLanguages() {
	fmt.Println("Supported languages:")
	languages := i18n.GetSupportedLanguages()
	for i, lang := range languages {
		fmt.Printf("  %s", lang)
		if (i+1)%10 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func downloadLanguagePack(lang string) {
	fmt.Printf("Downloading language pack for '%s'...\n", lang)
	if err := i18n.DownloadLanguage(lang); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Language pack '%s' downloaded successfully!\n", lang)
	}
}