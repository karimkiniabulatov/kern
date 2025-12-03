package main

import (
	"os"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"runtime"
	"strings"
	"sync"
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
	"github.com/karimkiniabulatov/kern/internal/ui"
	"github.com/karimkiniabulatov/kern/internal/service"
	"github.com/mattn/go-colorable"
)

var (
	// Основные флаги мониторинга
	flagDisk      = flag.Bool("d", false, "Show disk information")
	flagCPU       = flag.Bool("c", false, "Show CPU information")
	flagMem       = flag.Bool("m", false, "Show memory information")
	flagNet       = flag.Bool("n", false, "Show network information")
	flagGPU       = flag.Bool("g", false, "Show GPU information")
	flagAI        = flag.Bool("ai", false, "Show AI training information")
	flagMining    = flag.Bool("mining", false, "Show mining information")
	flagDetailedMem = flag.Bool("dm", false, "Show detailed memory module information")
	flagDetailedMemLong = flag.Bool("detailed-mem", false, "Show detailed memory module information")
	flagDetailedNet      = flag.Bool("dn", false, "Show detailed network interface information")
	flagDetailedNetLong  = flag.Bool("detailed-net", false, "Show detailed network interface information")
	flagDetailedDisk      = flag.Bool("dd", false, "Show detailed disk information (including all storage devices)")
	flagDetailedDiskLong  = flag.Bool("detailed-disk", false, "Show detailed disk information (including all storage devices)")
	
	// Общие флаги
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
	
	// Сервисные флаги
	flagDaemon    = flag.Bool("daemon", false, "Start kern as a daemon service")
	flagStart     = flag.Bool("start-service", false, "Start the kern daemon")
	flagStop      = flag.Bool("stop-service", false, "Stop the kern daemon")
	flagRestart   = flag.Bool("restart-service", false, "Restart the kern daemon")
	flagStatus    = flag.Bool("service-status", false, "Show daemon status")
	flagEnable    = flag.Bool("enable-service", false, "Enable auto-start on boot")
	flagDisable   = flag.Bool("disable-service", false, "Disable auto-start on boot")
	flagEnsureRunning = flag.Bool("ensure-running", false, "Ensure daemon is running")
	
	// Короткая версия для detailed
	flagDetailedShort = flag.Bool("de", false, "Show detailed CPU core information (short)")
	
	// Флаги управления приложением
	flagAppPause    = flag.Bool("pause", false, "Pause kern application")
	flagAppResume   = flag.Bool("resume", false, "Resume kern application") 
	flagAppStop     = flag.Bool("stop-app", false, "Stop kern application")
	flagAppRestart  = flag.Bool("restart-app", false, "Restart kern application")
)

// Глобальные переменные для кэширования данных
var (
	// Кэш для данных о дисках
	lastDiskData []disk.DiskInfo = []disk.DiskInfo{}
	diskDataMutex sync.RWMutex
	
	// Кэш для данных о CPU
	lastCPUData interface{}
	cpuDataMutex sync.RWMutex
	
	// Кэш для данных о памяти
	lastMemData interface{}
	memDataMutex sync.RWMutex
	
	// Кэш для данных о сети
	lastNetData interface{}
	netDataMutex sync.RWMutex
	
	// Кэш для данных о GPU
	lastGPUData interface{}
	gpuDataMutex sync.RWMutex
	
	// Кэш для данных об AI
	lastAIData interface{}
	aiDataMutex sync.RWMutex
	
	// Кэш для данных о майнинге
	lastMiningData interface{}
	miningDataMutex sync.RWMutex
)

const version = "1.2.3"

func init() {
	// Для Windows: настраиваем цветной вывод
	if runtime.GOOS == "windows" {
		colorable.EnableColorsStdout(nil)
	}

	// Регистрируем альтернативные имена только для тех флагов, у которых их еще нет
	flag.BoolVar(flagDisk, "disk", false, "Show disk information")
	flag.BoolVar(flagCPU, "cpu", false, "Show CPU information")
	flag.BoolVar(flagMem, "mem", false, "Show memory information")
	flag.BoolVar(flagNet, "net", false, "Show network information")
	flag.BoolVar(flagGPU, "gpu", false, "Show GPU information")
	
	// Общие флаги с альтернативными именами
	flag.BoolVar(flagAll, "all", false, "Show all information")
	flag.BoolVar(flagHelp, "help", false, "Show help")
	flag.BoolVar(flagLogo, "show-logo", false, "Show logo during monitoring")
	flag.BoolVar(flagRemote, "remote", false, "Start remote API server (default port: 28126)")
	flag.BoolVar(flagDetailedMem, "detailed-memory", false, "Show detailed memory module information")
	flag.BoolVar(flagDetailedNet, "detailed-network", false, "Show detailed network interface information")
	flag.BoolVar(flagDetailedDisk, "detailed-storage", false, "Show detailed storage information")
	
	// Сервисные флаги с альтернативными именами
	flag.BoolVar(flagDaemon, "dmn", false, "Start kern as a daemon service")
	flag.BoolVar(flagStart, "start-daemon", false, "Start the kern daemon")
	flag.BoolVar(flagStop, "stop-daemon", false, "Stop the kern daemon")
	flag.BoolVar(flagRestart, "restart-daemon", false, "Restart the kern daemon")
	flag.BoolVar(flagStatus, "daemon-status", false, "Show daemon status")
	flag.BoolVar(flagVersion, "version", false, "Show version")
}

func main() {
	flag.Usage = func() {
		showLogo()
		fmt.Printf("kern v%s - System Monitoring Tool\n\n", version)
		fmt.Println("Usage: kern [OPTIONS]")
		fmt.Println("\nMonitoring Options:")
		fmt.Println("  -d, --disk           Show disk information")
		fmt.Println("  -c, --cpu            Show CPU information") 
		fmt.Println("  -m, --mem            Show memory information")
		fmt.Println("  -n, --net            Show network information")
		fmt.Println("  -g, --gpu            Show GPU information")
		fmt.Println("  --ai                 Show AI training information")
		fmt.Println("  --mining             Show mining information")
		fmt.Println("  -a, --all            Show all information")
		fmt.Println("  --detailed, -de      Show detailed CPU core information")
		fmt.Println("  --refresh SECONDS    Refresh interval in seconds (default: 2)")
		fmt.Println("  --logo, --show-logo  Show logo during monitoring")
		fmt.Println("\nDetailed Information Options:")
		fmt.Println("  --detailed, -deflag.BoolVar(flagDaemon,      Show detailed CPU core information")
		fmt.Println("  -dm, --detailed-mem  Show detailed memory module information")
		fmt.Println("  -dn, --detailed-net  Show detailed network interface information")
		fmt.Println("  -dd, --detailed-disk  Show detailed storage information (all devices)")
		
		fmt.Println("\nLanguage Options:")
		fmt.Println("  -l LANG              Language code (e.g., 'ru' for Russian)")
		fmt.Println("  --download-lang LANG Download language pack")
		fmt.Println("  --list-languages     List all supported languages")
		
		fmt.Println("\nRemote Monitoring:")
		fmt.Println("  --api URL            Monitor remote server via HTTP/HTTPS API")
		fmt.Println("  --ssh HOST           Monitor remote server via SSH")
		fmt.Println("  -r, --remote         Start API server on default port 28126")
		fmt.Println("  --remote-port PORT   Start API server on custom port")
		
		fmt.Println("\nApplication Management:")
		fmt.Println("  --pause               Pause kern application")
		fmt.Println("  --resume              Resume kern application") 
		fmt.Println("  --stop                Stop kern application")
		fmt.Println("  --restart-app         Restart kern application")
		
		fmt.Println("\nService Management:")
		fmt.Println("  --daemon, --dmn      Start kern as a daemon service")
		fmt.Println("  --start-service      Start the kern daemon")
		fmt.Println("  --stop-service       Stop the kern daemon")
		fmt.Println("  --restart-service    Restart the kern daemon")
		fmt.Println("  --service-status     Show daemon status")
		fmt.Println("  --enable-service     Enable auto-start on boot")
		fmt.Println("  --disable-service    Disable auto-start on boot")
		fmt.Println("  --ensure-running     Ensure daemon is running")
		
		fmt.Println("\nOther Options:")
		fmt.Println("  -v, --version        Show version")
		fmt.Println("  -h, --help           Show this help message")
		
		fmt.Println("\nExamples:")
		fmt.Println("  kern --cpu --mem              # Show only CPU and memory")
		fmt.Println("  kern --gpu --ai               # Show GPU and AI training info")
		fmt.Println("  kern --mining                 # Show mining information")
		fmt.Println("  kern --disk -l ru             # Disk info with Russian interface")
		fmt.Println("  kern --refresh=5 --detailed   # Update every 5 sec with detailed CPU")
		fmt.Println("  kern --remote                 # Start API server on port 28126")
		fmt.Println("  kern --api http://192.168.1.100:28126 # Monitor remote via HTTP")
		fmt.Println("  kern --ssh user@host          # Monitor remote via SSH")
		fmt.Println("  kern --download-lang fr       # Download French language pack")
		fmt.Println("  kern --start-service          # Start kern daemon")
		fmt.Println("  kern --status                 # Check daemon status")
	}

	flag.Parse()

	// ДОБАВЛЕНО: Обработка флагов детального отображения диска
	if *flagDetailedDisk {
		detailedDisk = true
	}
	if *flagDetailedDiskLong {
		detailedDisk = true
	}
	
	// Убедитесь, что это делается ДО использования detailedDisk:
	detailedDisk := *flagDetailedDisk || *flagDetailedDiskLong
	
	// Также добавьте проверку в init() или после flag.Parse():
	fmt.Printf("DEBUG: flagDetailedDisk=%v, flagDetailedDiskLong=%v, detailedDisk=%v\n", 
		*flagDetailedDisk, *flagDetailedDiskLong, detailedDisk)

	// Обрабатываем короткую версию флага detailed
	if *flagDetailedShort {
		*flagDetailed = true
	}
	
	// Обработка управления приложением
	if *flagAppPause || *flagAppResume || *flagAppStop || *flagAppRestart {
		handleAppManagement()
		return
	}

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
		// Проверяем валидность URL
		if !strings.HasPrefix(*flagAPI, "http://") && !strings.HasPrefix(*flagAPI, "https://") {
			// Автоматически добавляем http:// если не указан протокол
			*flagAPI = "http://" + *flagAPI
		}
		
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

	// Убираем сохранение детальных флагов в конфиг
	// Вместо этого используем временные переменные
	detailedCPU := *flagDetailed || *flagDetailedShort
	detailedMem := *flagDetailedMem || *flagDetailedMemLong
	detailedNet := *flagDetailedNet || *flagDetailedNetLong
	// detailedDisk уже объявлена выше

	// Load configuration and localization
	cfg, err := config.Load(*flagLang)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Устанавливаем детальные флаги только для этой сессии
	cfg.DetailedCPU = detailedCPU
	cfg.DetailedMem = detailedMem
	cfg.DetailedNet = detailedNet
	cfg.DetailedDisk = detailedDisk

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
	// Объединяем короткие и длинные версии флагов
	showDisk := *flagDisk || *flagAll
	showCPU := *flagCPU || *flagAll
	showMem := *flagMem || *flagAll
	showNet := *flagNet || *flagAll
	showGPU := *flagGPU || *flagAll
	showAI := *flagAI || *flagAll
	showMining := *flagMining || *flagAll

	// NEW: Smart default behavior - use last used modules if no flags provided
	noFlagsProvided := !*flagDisk && !*flagCPU && !*flagMem && !*flagNet && 
	                  !*flagGPU && !*flagAI && !*flagMining && !*flagAll

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
			
			// If no modules were selected in last usage, use default modules
			if !showDisk && !showCPU && !showMem && !showNet && !showGPU && !showAI && !showMining {
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
		cfg.UpdateLastUsedModules(showDisk, showCPU, showMem, showNet, showGPU, showAI, showMining, detailedDisk)
	}

	// Передаем флаги в конфиг
	cfg.ShowDisk = showDisk
	cfg.ShowCPU = showCPU
	cfg.ShowMem = showMem
	cfg.ShowNet = showNet
	cfg.ShowGPU = showGPU
	cfg.ShowAI = showAI
	cfg.ShowMining = showMining

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
 ██╔═██╗ ██╔══╗  ██╔══██╗██║╚██╗██║
 ██║  ██╗███████╗██║  ██║██║ ╚████║
 ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝
 kern v` + version + " - System Monitoring Tool\n"
	
	// Кроссплатформенный цветной вывод
	if runtime.GOOS == "windows" {
		// Для Windows используем colorable
		fmt.Fprint(colorable.NewColorableStdout(), "\033[1;36m"+logo+"\033[0m")
	} else {
		// Для Unix-систем используем обычные ANSI коды
		fmt.Print("\033[1;36m" + logo + "\033[0m")
	}
}

func handleAppManagement() {
	daemonManager := service.NewDaemonManager()
	appManager := daemonManager.AppManagement()
	
	switch {
	case *flagAppPause:
		if err := appManager["pause"](); err != nil {
			log.Fatalf("Failed to pause app: %v", err)
		}
		fmt.Println("kern application paused")
	case *flagAppResume:
		if err := appManager["resume"](); err != nil {
			log.Fatalf("Failed to resume app: %v", err)
		}
		fmt.Println("kern application resumed")
	case *flagAppStop:
		if err := appManager["stop"](); err != nil {
			log.Fatalf("Failed to stop app: %v", err)
		}
		fmt.Println("kern application stopped")
	case *flagAppRestart:
		if err := appManager["restart"](); err != nil {
			log.Fatalf("Failed to restart app: %v", err)
		}
		fmt.Println("kern application restarted")
	}
}

// Структура для результата сбора данных
type result struct {
	module string
	data   interface{}
	err    error
}

// Функция сбора данных с поддержкой кэширования
func collectData(cfg *config.Config) map[string]interface{} {
    resultChan := make(chan result, 7) // Adjusted buffer size after removing audio/video

    // Launch goroutines only for enabled modules
    if cfg.ShowDisk {
        go func() {
            data, err := disk.Summary(cfg.DetailedDisk)
            if err != nil {
                // При ошибке возвращаем предыдущие данные
                diskDataMutex.RLock()
                data = lastDiskData
                diskDataMutex.RUnlock()
                log.Printf("Error getting disk data: %v. Using cached data.", err)
            } else {
                // Обновляем предыдущие данные
                diskDataMutex.Lock()
                lastDiskData = data
                diskDataMutex.Unlock()
            }
            resultChan <- result{"disk", data, nil}
        }()
    }

    if cfg.ShowCPU {
        go func() {
            data, err := cpu.Summary()
            if err != nil {
                // При ошибке возвращаем предыдущие данные
                cpuDataMutex.RLock()
                data = lastCPUData
                cpuDataMutex.RUnlock()
                log.Printf("Error getting CPU data: %v. Using cached data.", err)
            } else {
                // Обновляем предыдущие данные
                cpuDataMutex.Lock()
                lastCPUData = data
                cpuDataMutex.Unlock()
            }
            resultChan <- result{"cpu", data, err}
        }()
    }

    if cfg.ShowMem {
        go func() {
            data, err := mem.Summary()
            if err != nil {
                // При ошибке возвращаем предыдущие данные
                memDataMutex.RLock()
                data = lastMemData
                memDataMutex.RUnlock()
                log.Printf("Error getting memory data: %v. Using cached data.", err)
            } else {
                // Обновляем предыдущие данные
                memDataMutex.Lock()
                lastMemData = data
                memDataMutex.Unlock()
            }
            resultChan <- result{"mem", data, err}
        }()
    }

    if cfg.ShowNet {
        go func() {
            data, err := net.Summary()
            if err != nil {
                // При ошибке возвращаем предыдущие данные
                netDataMutex.RLock()
                data = lastNetData
                netDataMutex.RUnlock()
                log.Printf("Error getting network data: %v. Using cached data.", err)
            } else {
                // Обновляем предыдущие данные
                netDataMutex.Lock()
                lastNetData = data
                netDataMutex.Unlock()
            }
            resultChan <- result{"net", data, err}
        }()
    }

    // GPU monitoring
    if cfg.ShowGPU {
        go func() {
            data, err := gpu.Summary()
            if err != nil {
                // При ошибке возвращаем предыдущие данные
                gpuDataMutex.RLock()
                data = lastGPUData
                gpuDataMutex.RUnlock()
                if data == nil {
                    // Если нет кэшированных данных, создаем fallback
                    fallbackData := []*gpu.GPUInfo{{
                        Model:           "GPU не обнаружена",
                        DriverVersion:   "N/A",
                        GPUTemp:         0.0,
                        MemoryTotal:     "0 MB",
                        MemoryUsed:      "0 MB",
                        MemoryFree:      "0 MB",
                        Utilization:     0.0,
                        PowerDraw:       "0 W",
                        PowerLimit:      "0 W",
                        FanSpeed:        0.0,
                        ClockCore:       "0 MHz",
                        ClockMemory:     "0 MHz",
                        PerformanceState: "N/A",
                    }}
                    data = fallbackData
                }
                log.Printf("Error getting GPU data: %v. Using cached data.", err)
            } else {
                // Обновляем предыдущие данные
                gpuDataMutex.Lock()
                lastGPUData = data
                gpuDataMutex.Unlock()
            }
            resultChan <- result{"gpu", data, nil}
        }()
    }

    // AI training monitoring
    if cfg.ShowAI {
        go func() {
            data, err := ai.Summary()
            if err != nil {
                // При ошибке возвращаем предыдущие данные
                aiDataMutex.RLock()
                data = lastAIData
                aiDataMutex.RUnlock()
                log.Printf("Error getting AI data: %v. Using cached data.", err)
            } else {
                // Обновляем предыдущие данные
                aiDataMutex.Lock()
                lastAIData = data
                aiDataMutex.Unlock()
            }
            resultChan <- result{"ai", data, err}
        }()
    }

    // Mining monitoring
    if cfg.ShowMining {
        go func() {
            data, err := mining.Summary()
            if err != nil {
                // При ошибке возвращаем предыдущие данные
                miningDataMutex.RLock()
                data = lastMiningData
                miningDataMutex.RUnlock()
                log.Printf("Error getting mining data: %v. Using cached data.", err)
            } else {
                // Обновляем предыдущие данные
                miningDataMutex.Lock()
                lastMiningData = data
                miningDataMutex.Unlock()
            }
            resultChan <- result{"mining", data, err}
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

    for i := 0; i < moduleCount; i++ {
        res := <-resultChan
        if res.err != nil {
            results[res.module] = map[string]string{"error": res.err.Error()}
        } else {
            results[res.module] = res.data
        }
    }

    // Гарантируем, что все запрошенные модули возвращают данные
    // Только для модулей, которые были запрошены через cfg
    requestedModules := make([]string, 0)
    if cfg.ShowDisk {
        requestedModules = append(requestedModules, "disk")
    }
    if cfg.ShowCPU {
        requestedModules = append(requestedModules, "cpu")
    }
    if cfg.ShowMem {
        requestedModules = append(requestedModules, "mem")
    }
    if cfg.ShowNet {
        requestedModules = append(requestedModules, "net")
    }
    if cfg.ShowGPU {
        requestedModules = append(requestedModules, "gpu")
    }
    if cfg.ShowAI {
        requestedModules = append(requestedModules, "ai")
    }
    if cfg.ShowMining {
        requestedModules = append(requestedModules, "mining")
    }

    for _, module := range requestedModules {
        if _, exists := results[module]; !exists {
            // Создаем минимальные данные для отображения гистограмм
            switch module {
            case "gpu":
                results[module] = []*gpu.GPUInfo{{
                    Model:           "GPU не обнаружена",
                    DriverVersion:   "N/A",
                    GPUTemp:         0.0,
                    MemoryTotal:     "0 MB",
                    MemoryUsed:      "0 MB",
                    MemoryFree:      "0 MB",
                    Utilization:     0.0,
                    PowerDraw:       "0 W",
                    PowerLimit:      "0 W",
                    FanSpeed:        0.0,
                    ClockCore:       "0 MHz",
                    ClockMemory:     "0 MHz",
                    PerformanceState: "N/A",
                }}
            default:
                results[module] = map[string]string{"status": "no data"}
            }
        }
    }
    
    return results
}

func runMonitor(cfg *config.Config, showLogo bool) {
    // Initialize TUI
    tui, err := ui.NewTUI(cfg, showLogo)
    if err != nil {
        log.Fatalf("Failed to initialize TUI: %v", err)
    }
    defer tui.Fini()

    // УЛУЧШЕННАЯ обработка сигналов для гарантированного выхода
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Для Windows добавляем дополнительные сигналы
    if runtime.GOOS == "windows" {
        signal.Notify(sigChan, syscall.SIGQUIT)
    }

    // Канал для выхода
    quitChan := make(chan bool, 1)

    // Мьютекс для защиты данных от гонки
    var dataMutex sync.RWMutex
    var currentData map[string]interface{}

    // Запускаем сбор данных в отдельной горутине
    go func() {
        ticker := time.NewTicker(time.Duration(cfg.RefreshRate) * time.Second)
        defer ticker.Stop()

        // Первое обновление сразу
        data := collectData(cfg)
        dataMutex.Lock()
        currentData = data
        dataMutex.Unlock()
        tui.Render(data)

        for {
            select {
            case <-ticker.C:
                data := collectData(cfg)
                dataMutex.Lock()
                currentData = data
                dataMutex.Unlock()
                tui.Render(data)
            case <-quitChan:
                return
            }
        }
    }()

    // Основной цикл обработки событий
    for {
        // Используем неблокирующий PollEvent с таймаутом
        ev := tui.PollEvent()
        if ev == nil {
            // Проверяем сигналы без блокировки
            select {
            case <-sigChan:
                quitChan <- true
                return
            default:
                // Небольшая пауза чтобы не грузить CPU
                time.Sleep(50 * time.Millisecond)
            }
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
            // При изменении размера перерисовываем с текущими данными
            dataMutex.RLock()
            if currentData != nil {
                tui.Render(currentData)
            }
            dataMutex.RUnlock()
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

func startRemoteServer(cfg *config.Config, port int) {
	
	log.Printf("WARNING: API server is accessible from all network interfaces on port %d", port)
	log.Printf("For production use, consider configuring firewall rules")
	
    if port <= 0 || port > 65535 {
        port = 28126 // порт по умолчанию для API
    }

    showLogo()
    log.Printf("Starting remote API server on port %d...", port)

    // Create a new mux to avoid global http.HandleFunc conflicts
    mux := http.NewServeMux()

    // Добавляем middleware для логирования и CORS
    mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
        // Логируем запрос
        log.Printf("API Request: %s %s", r.Method, r.URL.Path)
        
        // CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            return
        }
        
        // Продолжаем к оригинальному обработчику
        mux.ServeHTTP(w, r)
    })

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
        
        data, err := disk.Summary(cfg.DetailedDisk)
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
            "os":      runtime.GOOS,
            "arch":    runtime.GOARCH,
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
            "os":      runtime.GOOS,
            "arch":    runtime.GOARCH,
            "endpoints": []string{
                "/api/cpu",
                "/api/mem",
                "/api/disk",
                "/api/net",
                "/api/gpu",
                "/api/ai",
                "/api/mining",
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
		Addr:    fmt.Sprintf("0.0.0.0:%d", port), // Принимать соединения со всех интерфейсов
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

    // Для Windows добавляем дополнительные сигналы
    if runtime.GOOS == "windows" {
        signal.Notify(sigChan, syscall.SIGQUIT)
    }

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

// NEW: Improved Remote API monitoring function
func monitorRemoteAPI(apiURL string) {
    showLogo()
    
    // Проверяем, указан ли конкретный эндпоинт
    isSpecificEndpoint := strings.Contains(apiURL, "/api/") || strings.Contains(apiURL, "/health")
    
    if isSpecificEndpoint {
        // Режим единичного запроса к конкретному эндпоинту
        fmt.Printf("Fetching data from: %s\n", apiURL)
        
        data, err := fetchSpecificEndpoint(apiURL)
        if err != nil {
            fmt.Printf("Error fetching data: %v\n", err)
            return
        }
        
        displaySpecificEndpointData(apiURL, data)
        return
    } else {
        // Режим мониторинга всех эндпоинтов
        fmt.Printf("Monitoring remote server via API: %s\n", apiURL)
        fmt.Println("Press Ctrl+C to exit")
        fmt.Println()

        ticker := time.NewTicker(2 * time.Second)
        defer ticker.Stop()

        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

        // Для Windows добавляем дополнительные сигналы
        if runtime.GOOS == "windows" {
            signal.Notify(sigChan, syscall.SIGQUIT)
        }

        for {
            select {
            case <-ticker.C:
                data, err := fetchRemoteData(apiURL)
                if err != nil {
                    fmt.Printf("Error fetching data: %v\n", err)
                    continue
                }
                displayRemoteData(data)
                // Кроссплатформенная очистка экрана
                clearScreen()
            case <-sigChan:
                fmt.Println("\nExiting remote monitoring...")
                return
            }
        }
    }
}

// NEW: Fetch specific endpoint data
func fetchSpecificEndpoint(url string) (interface{}, error) {
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to %s: %v", url, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("server returned status: %s", resp.Status)
    }

    var data interface{}
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, fmt.Errorf("failed to parse response: %v", err)
    }

    return data, nil
}

// NEW: Display specific endpoint data
func displaySpecificEndpointData(url string, data interface{}) {
    endpoint := extractEndpointName(url)
    fmt.Printf("=== %s ===\n", strings.ToUpper(endpoint))
    
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        fmt.Printf("Error formatting data: %v\n", err)
        return
    }
    
    fmt.Println(string(jsonData))
}

// NEW: Extract endpoint name from URL
func extractEndpointName(url string) string {
    if strings.Contains(url, "/api/") {
        parts := strings.Split(url, "/api/")
        if len(parts) > 1 {
            return parts[1]
        }
    } else if strings.Contains(url, "/health") {
        return "health"
    }
    return "unknown"
}



// Кроссплатформенная функция очистки экрана
func clearScreen() {
    switch runtime.GOOS {
    case "windows":
        // Для Windows
        fmt.Print("\033[2J\033[H")
    default:
        // Для Unix-систем
        fmt.Print("\033[2J\033[H")
    }
}

// NEW: Improved fetchRemoteData to handle errors better
func fetchRemoteData(baseURL string) (map[string]interface{}, error) {
    endpoints := []string{"cpu", "mem", "disk", "net", "gpu", "ai", "mining", "health"}
    results := make(map[string]interface{})
    
    client := &http.Client{Timeout: 5 * time.Second}

    for _, endpoint := range endpoints {
        url := fmt.Sprintf("%s/api/%s", baseURL, endpoint)
        if endpoint == "health" {
            url = fmt.Sprintf("%s/health", baseURL)
        }
        
        resp, err := client.Get(url)
        if err != nil {
            results[endpoint] = map[string]string{"error": err.Error()}
            continue
        }
        
        if resp.StatusCode != http.StatusOK {
            results[endpoint] = map[string]string{"error": fmt.Sprintf("HTTP %d", resp.StatusCode)}
            resp.Body.Close()
            continue
        }

        var data interface{}
        if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
            results[endpoint] = map[string]string{"error": err.Error()}
            resp.Body.Close()
            continue
        }

        results[endpoint] = data
        resp.Body.Close()
    }

    return results, nil
}



// NEW: Improved display for remote data
func displayRemoteData(data map[string]interface{}) {
    fmt.Println("=== Remote System Monitoring ===")
    fmt.Println()

    // Определяем порядок вывода модулей
    modulesOrder := []string{"cpu", "mem", "disk", "net", "gpu", "ai", "mining", "health"}
    
    for _, module := range modulesOrder {
        if moduleData, exists := data[module]; exists {
            fmt.Printf("%s:\n", strings.ToUpper(module))
            
            if errorData, isError := moduleData.(map[string]interface{}); isError {
                if errorMsg, exists := errorData["error"]; exists {
                    fmt.Printf("  Error: %v\n", errorMsg)
                }
            } else {
                // Форматируем вывод для лучшей читаемости
                displayFormattedModuleData(moduleData, "  ")
            }
            fmt.Println()
        }
    }
}

// NEW: Helper function to display formatted module data
func displayFormattedModuleData(data interface{}, indent string) {
    switch v := data.(type) {
    case map[string]interface{}:
        for key, value := range v {
            fmt.Printf("%s%s: ", indent, key)
            displayFormattedModuleData(value, indent+"  ")
        }
    case []interface{}:
        for i, item := range v {
            fmt.Printf("%s[%d]: ", indent, i)
            displayFormattedModuleData(item, indent+"  ")
        }
    default:
        fmt.Printf("%v\n", v)
    }
}

// NEW: SSH monitoring function (placeholder)
func monitorRemoteSSH(sshHost string) {
	showLogo()
	fmt.Printf("SSH monitoring for %s\n", sshHost)
	fmt.Println("Note: SSH monitoring requires kern installed on remote host")
	fmt.Printf("Example: ssh %s 'kern --all --refresh=2'\n", sshHost)
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