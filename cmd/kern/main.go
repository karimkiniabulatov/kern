package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/karimkiniabulatov/kern/internal/config"
	"github.com/karimkiniabulatov/kern/internal/cpu"
	"github.com/karimkiniabulatov/kern/internal/disk"
	"github.com/karimkiniabulatov/kern/internal/i18n"
	"github.com/karimkiniabulatov/kern/internal/mem"
	"github.com/karimkiniabulatov/kern/internal/net"
	"github.com/karimkiniabulatov/kern/internal/gpu"
	"github.com/karimkiniabulatov/kern/internal/ai"
	"github.com/karimkiniabulatov/kern/internal/mining"
	"github.com/karimkiniabulatov/kern/internal/audio"
	"github.com/karimkiniabulatov/kern/internal/video"
	"github.com/karimkiniabulatov/kern/internal/ui"
	"github.com/karimkiniabulatov/kern/internal/service"
	"github.com/mattn/go-colorable"
)

var (
	flagDisk      = flag.Bool("d", false, "Show disk information")
	flagCPU       = flag.Bool("c", false, "Show CPU information")
	flagMem       = flag.Bool("m", false, "Show memory information")
	flagNet       = flag.Bool("n", false, "Show network information")
	flagGPU       = flag.Bool("g", false, "Show GPU information")
	flagAI        = flag.Bool("ai", false, "Show AI training information")
	flagMining    = flag.Bool("mining", false, "Show mining information")
	flagAudio     = flag.Bool("audio", false, "Show audio stream information")
	flagVideo     = flag.Bool("video", false, "Show video stream information")
	flagAll       = flag.Bool("a", false, "Show all information")
	flagRefresh   = flag.Int("refresh", 2, "Refresh interval in seconds")
	flagLang      = flag.String("l", "", "Language code (e.g., 'ru' for Russian)")
	flagRemote    = flag.Bool("r", false, "Start remote API server (default port: 28126)")
	flagRemotePort = flag.Int("remote-port", 28126, "Port for remote API server")
	flagVersion   = flag.Bool("v", false, "Show version")
	flagHelp      = flag.Bool("h", false, "Show help")
	flagDetailed  = flag.Bool("detailed", false, "Show detailed CPU core information")
	flagDownload  = flag.String("download-lang", "", "Download language pack (e.g., 'fr' for French)")
	flagListLangs = flag.Bool("list-languages", false, "List all supported languages")
	flagLogo      = flag.Bool("logo", false, "Show logo during monitoring")
	flagAPI       = flag.String("api", "", "Monitor remote server via API (http://host:port)")
	flagSSH       = flag.String("ssh", "", "Monitor remote server via SSH (user@host)")
	flagDaemon    = flag.Bool("daemon", false, "Start kern as a daemon service")
	flagStart     = flag.Bool("start-service", false, "Start the kern daemon")
	flagStop      = flag.Bool("stop-service", false, "Stop the kern daemon")
	flagRestart   = flag.Bool("restart-service", false, "Restart the kern daemon")
	flagStatus    = flag.Bool("service-status", false, "Show daemon status")
	flagEnable    = flag.Bool("enable-service", false, "Enable auto-start on boot")
	flagDisable   = flag.Bool("disable-service", false, "Disable auto-start on boot")
	flagEnsureRunning = flag.Bool("ensure-running", false, "Ensure daemon is running")
)

const version = "1.2.1"

func init() {
	// Для Windows: настраиваем цветной вывод
	if runtime.GOOS == "windows" {
		colorable.EnableColorsStdout(nil)
	}
}

func main() {
	// Добавляем альтернативные имена флагов
	flag.BoolVar(flagDisk, "disk", false, "Show disk information")
	flag.BoolVar(flagCPU, "cpu", false, "Show CPU information")
	flag.BoolVar(flagMem, "mem", false, "Show memory information")
	flag.BoolVar(flagNet, "net", false, "Show network information")
	flag.BoolVar(flagGPU, "gpu", false, "Show GPU information")
	flag.BoolVar(flagAudio, "a", false, "Show audio stream information")
	flag.BoolVar(flagVideo, "v", false, "Show video stream information")
	flag.BoolVar(flagAll, "all", false, "Show all information")
	flag.BoolVar(flagHelp, "help", false, "Show help")
	flag.BoolVar(flagLogo, "show-logo", false, "Show logo during monitoring")
	flag.BoolVar(flagRemote, "remote", false, "Start remote API server (default port: 28126)")
	flag.BoolVar(flagDaemon, "dmn", false, "Start kern as a daemon service")
	flag.BoolVar(flagStart, "start", false, "Start the kern daemon")
	flag.BoolVar(flagStop, "stop", false, "Stop the kern daemon")
	flag.BoolVar(flagRestart, "restart", false, "Restart the kern daemon")
	flag.BoolVar(flagStatus, "status", false, "Show daemon status")

	flag.Usage = func() {
		showLogo()
		fmt.Printf("kern v%s - System Monitoring Tool\n\n", version)
		fmt.Println("Usage: kern [OPTIONS]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nRemote Monitoring:")
		fmt.Println("  --api URL          Monitor remote server via HTTP/HTTPS API")
		fmt.Println("  --ssh HOST         Monitor remote server via SSH")
		fmt.Println("  -r, --remote       Start API server on default port 28126")
		fmt.Println("  --remote-port PORT Start API server on custom port")
		fmt.Println("\nService Management:")
		fmt.Println("  --daemon           Start kern as a daemon service")
		fmt.Println("  --start-service    Start the kern daemon")
		fmt.Println("  --stop-service     Stop the kern daemon")
		fmt.Println("  --restart-service  Restart the kern daemon")
		fmt.Println("  --service-status   Show daemon status")
		fmt.Println("  --enable-service   Enable auto-start on boot")
		fmt.Println("  --disable-service  Disable auto-start on boot")
		fmt.Println("  --ensure-running   Ensure daemon is running")
		fmt.Println("\nExamples:")
		fmt.Println("  kern                       # Show all system information")
		fmt.Println("  kern --cpu --mem           # Show only CPU and memory")
		fmt.Println("  kern --gpu --ai           # Show GPU and AI training info")
		fmt.Println("  kern --mining             # Show mining information")
		fmt.Println("  kern --audio --video      # Show audio and video streams")
		fmt.Println("  kern -d -l ru              # Disk info with Russian interface")
		fmt.Println("  kern --refresh=5           # Update every 5 seconds")
		fmt.Println("  kern --detailed            # Show detailed CPU core info")
		fmt.Println("  kern -r                    # Start API server on port 28126")
		fmt.Println("  kern --remote-port 26001   # Start API server on custom port")
		fmt.Println("  kern --api http://192.168.1.100:28126 # Monitor remote via HTTP")
		fmt.Println("  kern --api https://example.com:28126 # Monitor remote via HTTPS")
		fmt.Println("  kern --ssh user@host       # Monitor remote via SSH")
		fmt.Println("  kern --download-lang fr    # Download French language pack")
		fmt.Println("  kern --logo                # Show logo during monitoring")
		fmt.Println("  kern --start-service       # Start kern daemon")
		fmt.Println("  kern --service-status      # Check daemon status")
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

	// Обработка сервисных команд ДО всего остального
	if *flagStart || *flagStop || *flagRestart || *flagStatus || *flagEnable || *flagDisable {
		handleServiceCommands()
		return
	}

	if *flagEnsureRunning {
		ensureDaemonRunning()
		return
	}

	if *flagDaemon {
		startAsDaemon()
		return
	}

	// Remote API monitoring
	if *flagAPI != "" {
		monitorRemoteAPI(*flagAPI)
		return
	}

	// SSH monitoring
	if *flagSSH != "" {
		monitorRemoteSSH(*flagSSH)
		return
	}

	// NEW: Check for remote server mode first - simple and clean
	if *flagRemote || *flagRemotePort != 28126 {
		port := *flagRemotePort
		if port <= 0 || port > 65535 {
			port = 28126
		}
		
		// Load minimal config for API server
		cfg, err := config.Load("")
		if err != nil {
			cfg = config.GetDefaultConfig("")
		}
		
		startRemoteServer(cfg, port)
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

	// Определяем какие модули показывать на основе флагов
	showDisk := *flagDisk || *flagAll
	showCPU := *flagCPU || *flagAll
	showMem := *flagMem || *flagAll
	showNet := *flagNet || *flagAll
	showGPU := *flagGPU || *flagAll
	showAI := *flagAI || *flagAll
	showMining := *flagMining || *flagAll
	showAudio := *flagAudio || *flagAll
	showVideo := *flagVideo || *flagAll

	// NEW: Smart default behavior - use last used modules if no flags provided
	noFlagsProvided := !*flagDisk && !*flagCPU && !*flagMem && !*flagNet && 
	                  !*flagGPU && !*flagAI && !*flagMining && !*flagAudio && !*flagVideo && !*flagAll

	if noFlagsProvided {
		// Check if we have last used modules saved
		if cfg.LastUsedModules != nil {
			// Use last used modules
			showDisk = cfg.LastUsedModules.ShowDisk
			showCPU = cfg.LastUsedModules.ShowCPU
			showMem = cfg.LastUsedModules.ShowMem
			showNet = cfg.LastUsedModules.ShowNet
			showGPU = cfg.LastUsedModules.ShowGPU
			showAI = cfg.LastUsedModules.ShowAI
			showMining = cfg.LastUsedModules.ShowMining
			showAudio = cfg.LastUsedModules.ShowAudio
			showVideo = cfg.LastUsedModules.ShowVideo
			
			// If no modules were selected in last usage, use default modules
			if !showDisk && !showCPU && !showMem && !showNet && !showGPU && !showAI && !showMining && !showAudio && !showVideo {
				showDisk = true
				showCPU = true
				showMem = true
				showNet = true
			}
		} else {
			// First run or no saved preferences - use default modules
			showDisk = true
			showCPU = true
			showMem = true
			showNet = true
		}
	} else {
		// Flags were provided - save these as last used modules
		cfg.UpdateLastUsedModules(showDisk, showCPU, showMem, showNet, showGPU, showAI, showMining, showAudio, showVideo)
	}

	// Передаем флаги в конфиг
	cfg.ShowDisk = showDisk
	cfg.ShowCPU = showCPU
	cfg.ShowMem = showMem
	cfg.ShowNet = showNet
	cfg.ShowGPU = showGPU
	cfg.ShowAI = showAI
	cfg.ShowMining = showMining
	cfg.ShowAudio = showAudio
	cfg.ShowVideo = showVideo
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
	// Initialize TUI
	tui, err := ui.NewTUI(cfg, showLogo)
	if err != nil {
		log.Fatalf("Failed to initialize TUI: %v", err)
	}
	defer tui.Fini()

	// Обработка сигналов для гарантированного выхода
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Канал для выхода
	quitChan := make(chan bool, 1)

	// Запускаем сбор данных в отдельной горутине
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.RefreshRate) * time.Second)
		defer ticker.Stop()

		// Первое обновление сразу
		data := collectData(cfg)
		tui.Render(data)

		for {
			select {
			case <-ticker.C:
				data := collectData(cfg)
				tui.Render(data)
			case <-quitChan:
				return
			}
		}
	}()

	// Основной цикл обработки событий
	for {
		ev := tui.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *tcell.EventKey:
			if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC ||
				(e.Key() == tcell.KeyRune && (e.Rune() == 'q' || e.Rune() == 'Q')) {
				quitChan <- true
				return
			}
		case *tcell.EventResize:
			tui.ForceRedraw()
		}

		// Проверяем сигналы без блокировки
		select {
		case <-sigChan:
			quitChan <- true
			return
		default:
			// продолжаем
		}
	}
}

func collectData(cfg *config.Config) map[string]interface{} {
	type result struct {
		module string
		data   interface{}
		err    error
	}

	resultChan := make(chan result, 9) // Increased buffer size for new modules

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

	// NEW: GPU monitoring
	if cfg.ShowGPU {
		go func() {
			data, err := gpu.Summary()
			resultChan <- result{"gpu", data, err}
		}()
	}

	// NEW: AI training monitoring
	if cfg.ShowAI {
		go func() {
			data, err := ai.Summary()
			resultChan <- result{"ai", data, err}
		}()
	}

	// NEW: Mining monitoring
	if cfg.ShowMining {
		go func() {
			data, err := mining.Summary()
			resultChan <- result{"mining", data, err}
		}()
	}

	// NEW: Audio monitoring
	if cfg.ShowAudio {
		go func() {
			data, err := audio.Summary()
			resultChan <- result{"audio", data, err}
		}()
	}

	// NEW: Video monitoring
	if cfg.ShowVideo {
		go func() {
			data, err := video.Summary()
			resultChan <- result{"video", data, err}
		}()
	}

	// Collect results
	results := make(map[string]interface{})
	moduleCount := 0
	if cfg.ShowDisk {
		moduleCount++
	}
	if cfg.ShowCPU {
		moduleCount++
	}
	if cfg.ShowMem {
		moduleCount++
	}
	if cfg.ShowNet {
		moduleCount++
	}
	if cfg.ShowGPU {
		moduleCount++
	}
	if cfg.ShowAI {
		moduleCount++
	}
	if cfg.ShowMining {
		moduleCount++
	}
	if cfg.ShowAudio {
		moduleCount++
	}
	if cfg.ShowVideo {
		moduleCount++
	}

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
	if port <= 0 || port > 65535 {
		port = 28126 // порт по умолчанию для API
	}

	showLogo()
	log.Printf("Starting remote API server on port %d...", port)

	// Create a new mux to avoid global http.HandleFunc conflicts
	mux := http.NewServeMux()

	// CPU endpoint
	mux.HandleFunc("/api/cpu", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := cpu.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// Memory endpoint
	mux.HandleFunc("/api/mem", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := mem.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// Disk endpoint
	mux.HandleFunc("/api/disk", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := disk.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// Network endpoint
	mux.HandleFunc("/api/net", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := net.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// NEW: GPU endpoint
	mux.HandleFunc("/api/gpu", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := gpu.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// NEW: AI endpoint
	mux.HandleFunc("/api/ai", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := ai.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// NEW: Mining endpoint
	mux.HandleFunc("/api/mining", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := mining.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// NEW: Audio endpoint
	mux.HandleFunc("/api/audio", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := audio.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// NEW: Video endpoint
	mux.HandleFunc("/api/video", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		data, err := video.Summary()
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// System info endpoint
	mux.HandleFunc("/api/system", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		systemInfo := map[string]interface{}{
			"version": version,
			"time":    time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(systemInfo)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Root endpoint with API info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		apiInfo := map[string]interface{}{
			"name":    "kern API",
			"version": version,
			"endpoints": []string{
				"/api/cpu",
				"/api/mem",
				"/api/disk",
				"/api/net",
				"/api/gpu",
				"/api/ai",
				"/api/mining",
				"/api/audio",
				"/api/video",
				"/api/system",
				"/health",
			},
			"protocols": []string{
				"HTTP",
				"HTTPS (with TLS)",
			},
			"access": "Local network and global internet (if port forwarded)",
		}
		json.NewEncoder(w).Encode(apiInfo)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("API server running on http://localhost:%d", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET /api/cpu    - CPU information")
	log.Printf("  GET /api/mem    - Memory information")
	log.Printf("  GET /api/disk   - Disk information")
	log.Printf("  GET /api/net    - Network information")
	log.Printf("  GET /api/gpu    - GPU information")
	log.Printf("  GET /api/ai     - AI training information")
	log.Printf("  GET /api/mining - Mining information")
	log.Printf("  GET /api/audio  - Audio stream information")
	log.Printf("  GET /api/video  - Video stream information")
	log.Printf("  GET /api/system - System information")
	log.Printf("  GET /health     - Health check")
	log.Printf("")
	log.Printf("Access methods:")
	log.Printf("  Local:  http://localhost:%d/api/cpu", port)
	log.Printf("  Remote: http://your-ip:%d/api/cpu", port)
	log.Printf("  HTTPS:  Configure reverse proxy with TLS")
	log.Printf("  SSH:    Use SSH tunneling: ssh -L %d:localhost:%d user@host", port, port)
	log.Printf("")
	log.Printf("Press Ctrl+C to stop the server")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("Shutting down API server...")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start API server: %v", err)
	}

	log.Printf("API server stopped")
}

// NEW: Remote API monitoring function
func monitorRemoteAPI(apiURL string) {
	showLogo()
	fmt.Printf("Monitoring remote server via API: %s\n", apiURL)
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case <-ticker.C:
			data, err := fetchRemoteData(apiURL)
			if err != nil {
				fmt.Printf("Error fetching data: %v\n", err)
				continue
			}
			displayRemoteData(data)
			fmt.Println("\033[2J\033[H") // Clear screen and move cursor to top
		case <-sigChan:
			fmt.Println("\nExiting remote monitoring...")
			return
		}
	}
}

// NEW: Fetch data from remote API
func fetchRemoteData(baseURL string) (map[string]interface{}, error) {
	endpoints := []string{"cpu", "mem", "disk", "net", "gpu", "ai", "mining", "audio", "video"}
	results := make(map[string]interface{})

	for _, endpoint := range endpoints {
		url := fmt.Sprintf("%s/api/%s", baseURL, endpoint)
		resp, err := http.Get(url)
		if err != nil {
			results[endpoint] = map[string]string{"error": err.Error()}
			continue
		}
		defer resp.Body.Close()

		var data interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			results[endpoint] = map[string]string{"error": err.Error()}
			continue
		}

		results[endpoint] = data
	}

	return results, nil
}

// NEW: Display remote data
func displayRemoteData(data map[string]interface{}) {
	fmt.Println("=== Remote System Monitoring ===")

	for module, moduleData := range data {
		fmt.Printf("\n%s:\n", strings.ToUpper(module))
		if errorData, isError := moduleData.(map[string]interface{}); isError {
			if errorMsg, exists := errorData["error"]; exists {
				fmt.Printf("  Error: %v\n", errorMsg)
			}
		} else {
			// Simple display of remote data
			jsonData, _ := json.MarshalIndent(moduleData, "  ", "  ")
			fmt.Printf("  %s\n", string(jsonData))
		}
	}
}

// NEW: SSH monitoring function (placeholder)
func monitorRemoteSSH(sshHost string) {
	showLogo()
	fmt.Printf("SSH monitoring for %s\n", sshHost)
	fmt.Println("Note: SSH monitoring requires kern installed on remote host")
	fmt.Println("Example: ssh %s 'kern --all --refresh=2'", sshHost)
	fmt.Println("")
	fmt.Println("For full SSH integration, use:")
	fmt.Println("  kern --ssh user@host --api http://localhost:28126")
	fmt.Println("")
	fmt.Println("Or manually:")
	fmt.Println("  1. On remote: kern -r 28126")
	fmt.Println("  2. On local:  kern --api http://remote-host:28126")
	fmt.Println("  3. Or use SSH tunnel: ssh -L 28126:localhost:28126 user@host")
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

// NEW: Service management functions
func handleServiceCommands() {
	daemonManager := service.NewDaemonManager()

	switch {
	case *flagStart:
		if err := daemonManager.StartDaemon(); err != nil {
			log.Fatalf("Failed to start daemon: %v", err)
		}
		fmt.Println("kern daemon started successfully")
		
	case *flagStop:
		if err := daemonManager.StopDaemon(); err != nil {
			log.Fatalf("Failed to stop daemon: %v", err)
		}
		fmt.Println("kern daemon stopped successfully")
		
	case *flagRestart:
		if err := daemonManager.RestartDaemon(); err != nil {
			log.Fatalf("Failed to restart daemon: %v", err)
		}
		fmt.Println("kern daemon restarted successfully")
		
	case *flagStatus:
		status := daemonManager.Status()
		fmt.Println("kern Daemon Status:")
		for key, value := range status {
			fmt.Printf("  %s: %v\n", key, value)
		}
		
	case *flagEnable:
		if err := daemonManager.EnableAutoStart(); err != nil {
			log.Fatalf("Failed to enable auto-start: %v", err)
		}
		fmt.Println("Auto-start enabled for kern daemon")
		
	case *flagDisable:
		if err := daemonManager.DisableAutoStart(); err != nil {
			log.Fatalf("Failed to disable auto-start: %v", err)
		}
		fmt.Println("Auto-start disabled for kern daemon")
	}
}

func ensureDaemonRunning() {
	daemonManager := service.NewDaemonManager()
	
	if err := daemonManager.EnsureRunning(); err != nil {
		log.Fatalf("Failed to ensure daemon is running: %v", err)
	}
	
	status := daemonManager.Status()
	fmt.Printf("kern daemon is running on port %d\n", status["port"].(int))
	fmt.Printf("API URL: http://localhost:%d\n", status["port"].(int))
}

func startAsDaemon() {
	daemonManager := service.NewDaemonManager()
	
	// Убедимся, что демон включен в конфиге
	daemonCfg := daemonManager.GetConfig()
	if !daemonCfg.Enabled {
		daemonCfg.Enabled = true
		daemonManager.UpdateConfig(daemonCfg)
	}

	fmt.Printf("Starting kern daemon on port %d...\n", daemonCfg.Port)
	fmt.Printf("Log file: %s\n", daemonCfg.LogFile)
	fmt.Printf("PID file: %s\n", daemonCfg.PIDFile)
	fmt.Println("Press Ctrl+C to stop the daemon")

	// Запускаем API сервер
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.GetDefaultConfig("")
	}
	startRemoteServer(cfg, daemonCfg.Port)
}