package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/karimkiniabulatov/kern/internal/config"
	"github.com/karimkiniabulatov/kern/internal/cpu"
	"github.com/karimkiniabulatov/kern/internal/disk"
	"github.com/karimkiniabulatov/kern/internal/mem"
	"github.com/karimkiniabulatov/kern/internal/net"
	"github.com/karimkiniabulatov/kern/internal/gpu"
	"github.com/karimkiniabulatov/kern/internal/ai"
	"github.com/karimkiniabulatov/kern/internal/mining"
	"github.com/mattn/go-runewidth"
)

type TUI struct {
	screen   tcell.Screen
	config   *config.Config
	showLogo bool
	width    int
	height   int
}

func NewTUI(cfg *config.Config, showLogo bool) (*TUI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.Clear()
	width, height := screen.Size()

	return &TUI{
		screen:   screen,
		config:   cfg,
		showLogo: showLogo,
		width:    width,
		height:   height,
	}, nil
}

func (t *TUI) Render(data map[string]interface{}) {
	t.screen.Clear()
	t.width, t.height = t.screen.Size()

	row := 0

	// Show logo if needed
	if t.showLogo {
		row = t.renderLogo(row)
	}

	// Render only enabled modules
	if t.config.ShowDisk {
		if diskData, exists := data["disk"]; exists {
			row = t.renderDisk(row, diskData)
		}
	}

	if t.config.ShowMem {
		if memData, exists := data["mem"]; exists {
			row = t.renderMemory(row, memData)
		}
	}

	if t.config.ShowNet {
		if netData, exists := data["net"]; exists {
			row = t.renderNetwork(row, netData)
		}
	}

	if t.config.ShowCPU {
		if cpuData, exists := data["cpu"]; exists {
			row = t.renderCPU(row, cpuData)
		}
	}

	// GPU monitoring
	if t.config.ShowGPU {
		if gpuData, exists := data["gpu"]; exists {
			row = t.renderGPU(row, gpuData)
		}
	}

	// AI training monitoring  
	if t.config.ShowAI {
		if aiData, exists := data["ai"]; exists {
			row = t.renderAI(row, aiData)
		}
	}

	// Mining monitoring
	if t.config.ShowMining {
		if miningData, exists := data["mining"]; exists {
			row = t.renderMining(row, miningData)
		}
	}

	t.renderFooter(row)
	t.screen.Show()
}

// ForceRedraw перерисовывает экран с последними данными
func (t *TUI) ForceRedraw() {
	t.screen.Sync()
}

func (t *TUI) PollEvent() tcell.Event {
	return t.screen.PollEvent()
}

func (t *TUI) Fini() {
	t.screen.Fini()
}

func (t *TUI) renderLogo(startRow int) int {
	logo := []string{
		" ██╗  ██╗███████╗██████╗ ███╗   ██╗",
		" ██║ ██╔╝██╔════╝██╔══██╗████╗  ██║",
		" █████╔╝ █████╗  ██████╔╝██╔██╗ ██║",
		" ██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║",
		" ██║  ██╗███████╗██║  ██║██║ ╚████║",
		" ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝",
		" kern v1.2.1 - System Monitoring Tool",
	}

	cyan := tcell.StyleDefault.Foreground(tcell.ColorTeal).Bold(true)
	for i, line := range logo {
		t.printCentered(startRow+i, line, cyan)
	}

	return startRow + len(logo) + 1
}

func (t *TUI) renderCPU(startRow int, data interface{}) int {
	row := startRow

	if systemInfo, ok := data.(*cpu.SystemCPUInfo); ok {
		// Общая информация о системе
		if systemInfo.TotalSockets > 1 {
			row = t.renderHeader(row, fmt.Sprintf("%s (%d sockets, %d cores, %d threads)", 
				t.config.T("cpu.title"), systemInfo.TotalSockets, systemInfo.TotalCores, systemInfo.TotalThreads))
		} else {
			row = t.renderHeader(row, t.config.T("cpu.title"))
		}

		// Отображаем каждый процессор
		for i, cpuInfo := range systemInfo.CPUs {
			// Заголовок для каждого процессора, если их несколько
			if systemInfo.TotalSockets > 1 {
				header := fmt.Sprintf("CPU #%d", i+1)
				if cpuInfo.NUMANode >= 0 {
					header += fmt.Sprintf(" (NUMA node %d)", cpuInfo.NUMANode)
				}
				row = t.renderSubHeader(row, header)
			}

			// Основная информация о процессоре
			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("cpu.model"), cpuInfo.Model), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("cpu.vendor"), cpuInfo.Vendor), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			
			if systemInfo.TotalSockets > 1 {
				// Для многопроцессорных систем показываем детали по каждому процессору
				row = t.printSimple(row, fmt.Sprintf("%s: %d %s, %d %s",
					t.config.T("cpu.cores"), cpuInfo.Cores, t.config.T("cpu.cores"),
					cpuInfo.Threads, t.config.T("cpu.threads")), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			// Usage with graph
			usageGraph := t.createSolidGraph(cpuInfo.Usage)
			row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", t.config.T("cpu.usage"), cpuInfo.Usage, usageGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))

			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("cpu.frequency"), cpuInfo.Frequency), tcell.StyleDefault.Foreground(tcell.ColorAqua))

			// Средняя загрузка показываем только для первого процессора (она системная)
			if i == 0 {
				row = t.printSimple(row, fmt.Sprintf("%s: %.2f, %.2f, %.2f",
					t.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			// Детальная информация по ядрам - ТОЛЬКО ЕСЛИ ВКЛЮЧЕН ФЛАГ
			if t.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
				if systemInfo.TotalSockets > 1 {
					row = t.printSimple(row, fmt.Sprintf("%s (Socket %d):", t.config.T("cpu.core_usage"), i+1), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				} else {
					row = t.printSimple(row, fmt.Sprintf("%s:", t.config.T("cpu.core_usage")), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				}
				
				// Группируем ядра для компактного отображения (по 3 в строку)
				coresPerLine := 3
				for i := 0; i < len(cpuInfo.CoreUsage); i += coresPerLine {
					line := "  "
					for j := 0; j < coresPerLine && i+j < len(cpuInfo.CoreUsage); j++ {
						coreIdx := i + j
						usage := cpuInfo.CoreUsage[coreIdx]
						
						// Глобальный номер ядра или локальный для процессора
						coreNumber := coreIdx
						if systemInfo.TotalSockets > 1 {
							// Показываем локальный номер ядра в пределах процессора
							coreNumber = j
						}
						
						coreGraph := t.createSolidGraph(usage)
						line += fmt.Sprintf(" %02d: %05.1f%%%s", coreNumber+1, usage, coreGraph)
						
						if j < coresPerLine-1 && i+j+1 < len(cpuInfo.CoreUsage) {
							line += " |"
						}
					}
					row = t.printSimple(row, line, tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
				}
			}

			// Добавляем отступ между процессорами, если их несколько
			if i < len(systemInfo.CPUs)-1 {
				row++
			}
		}
	} else {
		// Фолбэк для старого формата данных
		row = t.renderHeader(row, t.config.T("cpu.title"))
		row = t.printSimple(row, "No CPU data available", tcell.StyleDefault.Foreground(tcell.ColorGray))
	}

	return row + 1
}

// Новая функция для подзаголовков процессоров
func (t *TUI) renderSubHeader(startRow int, title string) int {
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	
	// Create sub-header line with padding
	header := fmt.Sprintf(" %s ", title)
	t.printSimple(startRow, header, style)
	
	return startRow + 1
}

func (t *TUI) renderMemory(startRow int, data interface{}) int {
    row := t.renderHeader(startRow, t.config.T("memory.title"))

    if memInfo, ok := data.(*mem.MemoryInfo); ok {
        // Информация об архитектуре памяти
        archInfo := mem.DetectMemoryArchitecture()
        row = t.printSimple(row, fmt.Sprintf("Architecture: %s", archInfo), 
            tcell.StyleDefault.Foreground(tcell.ColorYellow))

        // ОСНОВНАЯ ИНФОРМАЦИЯ О ПАМЯТИ - ФИКСИРОВАННЫЙ ФОРМАТ ПРОЦЕНТОВ
        ramUsageFormatted := fmt.Sprintf("%05.1f", memInfo.UsagePercent) // Формат 001.3%
        ramGraph := t.createSolidGraph(memInfo.UsagePercent)
        row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%s%%) %s",
            t.config.T("memory.ram"), memInfo.Used, memInfo.Total, 
            ramUsageFormatted, ramGraph), tcell.StyleDefault.Foreground(tcell.ColorGreen))

        // ДЕТАЛЬНАЯ ИНФОРМАЦИЯ О ИСПОЛЬЗОВАНИИ
        row = t.printSimple(row, fmt.Sprintf("Processes: %s | Cached: %s | Buffers: %s", 
            memInfo.UsedByProcesses, memInfo.Cached, memInfo.Buffers), 
            tcell.StyleDefault.Foreground(tcell.ColorAqua))

        row = t.printSimple(row, fmt.Sprintf("Active: %s | Inactive: %s | Shared: %s", 
            memInfo.Active, memInfo.Inactive, memInfo.Shared), 
            tcell.StyleDefault.Foreground(tcell.ColorAqua))

        row = t.printSimple(row, fmt.Sprintf("%s: %s | %s: %s", 
            t.config.T("common.available"), memInfo.Available,
            t.config.T("common.free"), memInfo.Free), tcell.StyleDefault.Foreground(tcell.ColorAqua))

        if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" && memInfo.SwapTotal != "0" {
            swapUsageFormatted := fmt.Sprintf("%05.1f", memInfo.SwapUsagePercent)
            swapGraph := t.createSolidGraph(memInfo.SwapUsagePercent)
            row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%s%%) %s",
                t.config.T("memory.swap"), memInfo.SwapUsed, memInfo.SwapTotal, 
                swapUsageFormatted, swapGraph), tcell.StyleDefault.Foreground(tcell.ColorFuchsia))
        }

        // ИНФОРМАЦИЯ О МОДУЛЯХ ПАМЯТИ - ВЫРОВНЕННЫЕ ГИСТОГРАММЫ
        if len(memInfo.Modules) > 0 {
            row = t.printSimple(row, fmt.Sprintf("%s (%d modules):", 
                t.config.T("memory.modules"), len(memInfo.Modules)), 
                tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
            
            // Находим максимальную длину для выравнивания
            maxLineLength := 0
            var moduleLines []string
            
            for _, module := range memInfo.Modules {
                var moduleInfo string
                
                if module.Size == "Unknown" || module.SizeBytes == 0 {
                    // НЕИЗВЕСТНЫЙ модуль - показываем только слот и используем среднюю загруженность
                    moduleInfo = fmt.Sprintf("  %s: нет информации", module.Slot)
                } else {
                    // ИЗВЕСТНЫЙ модуль с данными
                    moduleInfo = fmt.Sprintf("  %s: %s %s @ %s", 
                        module.Slot, module.Size, module.Type, module.Speed)
                    
                    // Добавляем производителя и серийный номер
                    if module.Manufacturer != "" && module.Manufacturer != "Unknown" {
                        moduleInfo += fmt.Sprintf(" (%s", module.Manufacturer)
                        if module.PartNumber != "" && module.PartNumber != "Unknown" {
                            moduleInfo += fmt.Sprintf(" - %s", module.PartNumber)
                        }
                        if module.SerialNumber != "" && module.SerialNumber != "Unknown" {
                            moduleInfo += fmt.Sprintf(" [SN:%s]", module.SerialNumber)
                        }
                        moduleInfo += ")"
                    }
                }
                
                // Для ВСЕХ модулей (известных и неизвестных) вычисляем длину строки
                currentLength := len(moduleInfo)
                if currentLength > maxLineLength {
                    maxLineLength = currentLength
                }
                
                moduleLines = append(moduleLines, moduleInfo)
            }
            
            // Выводим все строки с выровненными гистограммами
            for i, line := range moduleLines {
                module := memInfo.Modules[i]
                usageFormatted := fmt.Sprintf("%05.1f", module.UsagePercent)
                moduleGraph := t.createSolidGraph(module.UsagePercent)
                
                // Добавляем пробелы для выравнивания
                padding := maxLineLength - len(line)
                if padding > 0 {
                    line += strings.Repeat(" ", padding)
                }
                
                // Добавляем проценты и гистограмму для ВСЕХ модулей
                line += fmt.Sprintf(" %s%% %s", usageFormatted, moduleGraph)
                
                row = t.printSimple(row, line, 
                    tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
            }
        } else {
            row = t.printSimple(row, "Memory modules: Information not available", 
                tcell.StyleDefault.Foreground(tcell.ColorGray))
        }
    }
    return row + 1
}

func (t *TUI) renderDisk(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("disk.title"))

	if disks, ok := data.([]disk.DiskInfo); ok {
		count := 0
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") && count < 3 {
				// Определяем тип устройства для отображения
				devType := t.getDeviceType(d.Filesystem, d.MountedOn)
				
				// Добавляем информацию о физическом/логическом типе и модели
				deviceInfo := devType
				if d.DiskType != "Unknown" && d.DiskType != "" {
					deviceInfo = d.DiskType
				}
				
				// Показываем физический/логический статус
				physStatus := "Logical"
				if d.Physical {
					physStatus = "Physical"
				}
				deviceInfo = fmt.Sprintf("%s (%s)", deviceInfo, physStatus)

				mountPoint := d.MountedOn
				if mountPoint == "/" {
					mountPoint = "ROOT"
				}

				// Основная информация о файловой системе
				row = t.printSimple(row, fmt.Sprintf("%s: %s (%s)", t.config.T("disk.filesystem"), d.Filesystem, deviceInfo), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("disk.mounted"), mountPoint), tcell.StyleDefault.Foreground(tcell.ColorAqua))

				// Информация о модели и серийном номере, если доступна
				if d.Model != "Unknown" && d.Model != "" {
					modelInfo := fmt.Sprintf("Model: %s", d.Model)
					if d.Serial != "Unknown" && d.Serial != "" {
						modelInfo += fmt.Sprintf(" [%s]", d.Serial)
					}
					row = t.printSimple(row, modelInfo, tcell.StyleDefault.Foreground(tcell.ColorGray))
				}

				// Disk usage with graph
				diskGraph := t.createSolidGraph(d.UsePercent)
				row = t.printSimple(row, fmt.Sprintf("%s: %s / %s %.1f%% %s",
					t.config.T("disk.usage"), d.Used, d.Size, d.UsePercent, diskGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))

				count++
				if count < 3 {
					row++
				}
			}
		}
	}
	return row + 1
}

func (t *TUI) renderNetwork(startRow int, data interface{}) int {
	row := startRow

	if networks, ok := data.([]net.NetworkInfo); ok {
		// Main header with device count
		deviceLabel := t.config.T("network.title")
		if len(networks) > 1 {
			deviceLabel = fmt.Sprintf("%s (%d interfaces)", deviceLabel, len(networks))
		}
		row = t.renderHeader(row, deviceLabel)

		// Render each network interface with numbering
		for i, netInfo := range networks {
			// Add sub-header for each interface if multiple devices
			if len(networks) > 1 {
				interfaceHeader := fmt.Sprintf("Network Interface #%d", i+1)
				row = t.renderSubHeader(row, interfaceHeader)
			}

			// Основная информация об интерфейсе
			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("network.interface"), netInfo.Interface), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			
			// Физический/виртуальный статус
			physStatus := "Virtual"
			if netInfo.IsPhysical {
				physStatus = "Physical"
			}
			row = t.printSimple(row, fmt.Sprintf("Type: %s (%s)", physStatus, netInfo.Driver), tcell.StyleDefault.Foreground(tcell.ColorGray))

			row = t.printSimple(row, fmt.Sprintf("%s: %s", "IP Address", netInfo.IPAddress), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			row = t.printSimple(row, fmt.Sprintf("%s: %s", "MAC Address", netInfo.MACAddress), tcell.StyleDefault.Foreground(tcell.ColorAqua))

			// Статус интерфейса
			statusColor := tcell.ColorRed
			if netInfo.Status == "UP" {
				statusColor = tcell.ColorGreen
			} else if netInfo.Status == "UNKNOWN" {
				statusColor = tcell.ColorYellow
			}
			row = t.printSimple(row, fmt.Sprintf("Status: %s", netInfo.Status), tcell.StyleDefault.Foreground(statusColor))

			// MTU information
			if netInfo.MTU > 0 {
				row = t.printSimple(row, fmt.Sprintf("MTU: %d", netInfo.MTU), tcell.StyleDefault.Foreground(tcell.ColorGray))
			}

			// Тип соединения с ASCII обозначением
			connectionLabel := t.getConnectionLabel(netInfo.ConnectionType)
			row = t.printSimple(row, fmt.Sprintf("%s: %s", 
				t.config.T("network.connection_type"), connectionLabel), 
				tcell.StyleDefault.Foreground(t.getConnectionColor(netInfo.ConnectionType)))

			// Технология и максимальная скорость
			row = t.printSimple(row, fmt.Sprintf("%s: %s (%s)", 
				t.config.T("network.technology"), netInfo.Technology, netInfo.MaxSpeed), 
				tcell.StyleDefault.Foreground(tcell.ColorAqua))

			// Сила сигнала для беспроводных соединений
			if netInfo.ConnectionType == "Wi-Fi" && netInfo.SignalStrength > 0 {
				signalGraph := t.createSolidGraph(netInfo.SignalStrength)
				row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", 
					t.config.T("network.signal_strength"), netInfo.SignalStrength, signalGraph), 
					tcell.StyleDefault.Foreground(tcell.ColorGreen))
			}

			// Активность сети с графиком
			activityGraph := t.createSolidGraph(netInfo.ActivityPercent)
			row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", "Activity", netInfo.ActivityPercent, activityGraph), tcell.StyleDefault.Foreground(tcell.ColorFuchsia))

			// Скорость передачи данных
			row = t.printSimple(row, fmt.Sprintf("%s: %s / %s", "Speed", netInfo.RXSpeed, netInfo.TXSpeed), tcell.StyleDefault.Foreground(tcell.ColorAqua))

			// Добавляем отступ между интерфейсами, если их несколько
			if i < len(networks)-1 {
				row++
			}
		}
	} else {
		// Fallback если данные не соответствуют ожидаемому формату
		row = t.renderHeader(row, t.config.T("network.title"))
		row = t.printSimple(row, "No network data available", tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

// getConnectionLabel возвращает ASCII обозначение для типа соединения
func (t *TUI) getConnectionLabel(connectionType string) string {
	switch connectionType {
	case "Ethernet":
		return "[ETH] Ethernet"
	case "Wi-Fi":
		return "[WLAN] Wi-Fi" 
	case "Bluetooth":
		return "[BT] Bluetooth"
	case "Cellular":
		return "[MOB] Cellular"
	case "VPN":
		return "[VPN] VPN"
	case "Bridge":
		return "[BRG] Bridge"
	case "Virtual":
		return "[VIRT] Virtual"
	case "Loopback":
		return "[LO] Loopback"
	default:
		return "[NET] " + connectionType
	}
}

// getConnectionColor возвращает цвет для типа соединения
func (t *TUI) getConnectionColor(connectionType string) tcell.Color {
	switch connectionType {
	case "Ethernet":
		return tcell.ColorBlue
	case "Wi-Fi":
		return tcell.ColorGreen
	case "Bluetooth":
		return tcell.ColorLightBlue
	case "Cellular":
		return tcell.ColorYellow
	case "VPN":
		return tcell.ColorPurple
	case "Bridge", "Virtual":
		return tcell.ColorGray
	default:
		return tcell.ColorWhite
	}
}

// GPU rendering function with multiple GPU support
func (t *TUI) renderGPU(startRow int, data interface{}) int {
	row := startRow

	// Handle both []*gpu.GPUInfo and *gpu.GPUInfo cases
	switch gpuData := data.(type) {
	case []*gpu.GPUInfo:
		if len(gpuData) == 0 {
			row = t.renderHeader(row, t.config.T("gpu.title"))
			row = t.printSimple(row, "No GPU devices found", tcell.StyleDefault.Foreground(tcell.ColorGray))
			return row + 1
		}

		// Main header with device count
		deviceLabel := "GPU"
		if len(gpuData) > 1 {
			deviceLabel = fmt.Sprintf("GPU (%d devices)", len(gpuData))
		}
		row = t.renderHeader(row, deviceLabel)

		// Render each GPU
		for i, gpuInfo := range gpuData {
			// Sub-header for each GPU if multiple devices
			if len(gpuData) > 1 {
				gpuHeader := fmt.Sprintf("GPU #%d", i+1)
				row = t.renderSubHeader(row, gpuHeader)
			}

			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("gpu.model"), gpuInfo.Model), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			
			if gpuInfo.DriverVersion != "" && gpuInfo.DriverVersion != "Unknown" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("gpu.driver"), gpuInfo.DriverVersion), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
			
			if gpuInfo.GPUTemp > 0 {
				// Temperature histogram (normalized to 100°C)
				tempPercent := gpuInfo.GPUTemp
				if tempPercent > 100 {
					tempPercent = 100
				}
				tempGraph := t.createSolidGraph(tempPercent)
				row = t.printSimple(row, fmt.Sprintf("%s: %.1f°C %s", 
					t.config.T("gpu.temperature"), gpuInfo.GPUTemp, tempGraph), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}

			if gpuInfo.Utilization > 0 {
				// Utilization histogram
				utilGraph := t.createSolidGraph(gpuInfo.Utilization)
				row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", 
					t.config.T("gpu.utilization"), gpuInfo.Utilization, utilGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}

			if gpuInfo.MemoryUsed != "" && gpuInfo.MemoryTotal != "" && gpuInfo.MemoryUsed != "0 MB" {
				// Calculate memory usage percentage for histogram
				memUsedMB := extractMemoryMB(gpuInfo.MemoryUsed)
				memTotalMB := extractMemoryMB(gpuInfo.MemoryTotal)
				if memTotalMB > 0 {
					memUsagePercent := float64(memUsedMB) / float64(memTotalMB) * 100
					memGraph := t.createSolidGraph(memUsagePercent)
					row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%.1f%%) %s", 
						t.config.T("gpu.memory"), gpuInfo.MemoryUsed, gpuInfo.MemoryTotal, 
						memUsagePercent, memGraph), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				} else {
					row = t.printSimple(row, fmt.Sprintf("%s: %s / %s", 
						t.config.T("gpu.memory"), gpuInfo.MemoryUsed, gpuInfo.MemoryTotal), 
						tcell.StyleDefault.Foreground(tcell.ColorAqua))
				}
			}
			
			if gpuInfo.PowerDraw != "" && gpuInfo.PowerDraw != "0 W" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("gpu.power"), gpuInfo.PowerDraw), tcell.StyleDefault.Foreground(tcell.ColorYellow))
			}
				
			if gpuInfo.ClockCore != "" && gpuInfo.ClockCore != "0 MHz" && gpuInfo.ClockMemory != "" && gpuInfo.ClockMemory != "0 MHz" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s | %s: %s", 
					t.config.T("gpu.clock_core"), gpuInfo.ClockCore, 
					t.config.T("gpu.clock_memory"), gpuInfo.ClockMemory), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			// Add spacing between GPUs if multiple devices
			if i < len(gpuData)-1 {
				row++
			}
		}

	case *gpu.GPUInfo:
		// Fallback for single GPU (legacy format)
		row = t.renderHeader(row, t.config.T("gpu.title"))
		row = t.renderSingleGPUInfo(row, gpuData) // ИСПРАВЛЕНО: было gpuInfo
		
	case map[string]interface{}:
		// Handle error case
		row = t.renderHeader(row, t.config.T("gpu.title"))
		if errorMsg, exists := gpuData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.renderHeader(row, t.config.T("gpu.title"))
		row = t.printSimple(row, "No GPU data available", tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

// Helper function to render single GPU info (for legacy format)
func (t *TUI) renderSingleGPUInfo(startRow int, gpuInfo *gpu.GPUInfo) int {
	row := startRow
	
	row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("gpu.model"), gpuInfo.Model), tcell.StyleDefault.Foreground(tcell.ColorAqua))
	
	if gpuInfo.DriverVersion != "" && gpuInfo.DriverVersion != "Unknown" {
		row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("gpu.driver"), gpuInfo.DriverVersion), tcell.StyleDefault.Foreground(tcell.ColorAqua))
	}
	
	if gpuInfo.GPUTemp > 0 {
		tempPercent := gpuInfo.GPUTemp
		if tempPercent > 100 {
			tempPercent = 100
		}
		tempGraph := t.createSolidGraph(tempPercent)
		row = t.printSimple(row, fmt.Sprintf("%s: %.1f°C %s", 
			t.config.T("gpu.temperature"), gpuInfo.GPUTemp, tempGraph), tcell.StyleDefault.Foreground(tcell.ColorRed))
	}

	if gpuInfo.Utilization > 0 {
		utilGraph := t.createSolidGraph(gpuInfo.Utilization)
		row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", 
			t.config.T("gpu.utilization"), gpuInfo.Utilization, utilGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
	}

	if gpuInfo.MemoryUsed != "" && gpuInfo.MemoryTotal != "" && gpuInfo.MemoryUsed != "0 MB" {
		memUsedMB := extractMemoryMB(gpuInfo.MemoryUsed)
		memTotalMB := extractMemoryMB(gpuInfo.MemoryTotal)
		if memTotalMB > 0 {
			memUsagePercent := float64(memUsedMB) / float64(memTotalMB) * 100
			memGraph := t.createSolidGraph(memUsagePercent)
			row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%.1f%%) %s", 
				t.config.T("gpu.memory"), gpuInfo.MemoryUsed, gpuInfo.MemoryTotal, 
				memUsagePercent, memGraph), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		} else {
			row = t.printSimple(row, fmt.Sprintf("%s: %s / %s", 
				t.config.T("gpu.memory"), gpuInfo.MemoryUsed, gpuInfo.MemoryTotal), 
				tcell.StyleDefault.Foreground(tcell.ColorAqua))
		}
	}
	
	if gpuInfo.PowerDraw != "" && gpuInfo.PowerDraw != "0 W" {
		row = t.printSimple(row, fmt.Sprintf("%s: %s", 
			t.config.T("gpu.power"), gpuInfo.PowerDraw), tcell.StyleDefault.Foreground(tcell.ColorYellow))
	}
		
	if gpuInfo.ClockCore != "" && gpuInfo.ClockCore != "0 MHz" && gpuInfo.ClockMemory != "" && gpuInfo.ClockMemory != "0 MHz" {
		row = t.printSimple(row, fmt.Sprintf("%s: %s | %s: %s", 
			t.config.T("gpu.clock_core"), gpuInfo.ClockCore, 
			t.config.T("gpu.clock_memory"), gpuInfo.ClockMemory), tcell.StyleDefault.Foreground(tcell.ColorAqua))
	}
	
	return row
}

// AI training rendering function
func (t *TUI) renderAI(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("ai.title"))

	switch aiData := data.(type) {
	case *ai.AIInfo:
		if aiData.ProcessCount > 0 {
			if aiData.Framework != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("ai.framework"), aiData.Framework), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
			row = t.printSimple(row, fmt.Sprintf("%s: %d", t.config.T("ai.processes"), aiData.ProcessCount), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			
			if aiData.VRAMUsage != "" && aiData.VRAMTotal != "" {
				// Вычисляем процент использования VRAM для гистограммы
				usedMB := extractMemoryMB(aiData.VRAMUsage)
				totalMB := extractMemoryMB(aiData.VRAMTotal)
				if totalMB > 0 {
					vramPercent := float64(usedMB) / float64(totalMB) * 100
					vramGraph := t.createSolidGraph(vramPercent)
					row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%.1f%%) %s", 
						t.config.T("ai.vram"), aiData.VRAMUsage, aiData.VRAMTotal, 
						vramPercent, vramGraph), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				} else {
					row = t.printSimple(row, fmt.Sprintf("%s: %s / %s", 
						t.config.T("ai.vram"), aiData.VRAMUsage, aiData.VRAMTotal), 
						tcell.StyleDefault.Foreground(tcell.ColorAqua))
				}
			}

			if aiData.ModelName != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("ai.model"), aiData.ModelName), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if aiData.BatchSize > 0 {
				// Гистограмма для throughput
				throughputGraph := t.createSolidGraph(aiData.Throughput / 100.0) // Нормализуем для отображения
				row = t.printSimple(row, fmt.Sprintf("%s: %d | %s: %.1f samples/sec %s", 
					t.config.T("ai.batch_size"), aiData.BatchSize,
					t.config.T("ai.throughput"), aiData.Throughput, throughputGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}

			if aiData.Epoch > 0 {
				// Гистограмма для точности
				accuracyGraph := t.createSolidGraph(aiData.Accuracy * 100)
				row = t.printSimple(row, fmt.Sprintf("%s: %d | %s: %.3f | %s: %.1f%% %s", 
					t.config.T("ai.epoch"), aiData.Epoch,
					t.config.T("ai.loss"), aiData.Loss,
					t.config.T("ai.accuracy"), aiData.Accuracy*100, accuracyGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}

			if aiData.TrainingTime != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("ai.training_time"), aiData.TrainingTime), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
		} else {
			row = t.printSimple(row, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	case map[string]interface{}:
		if errorMsg, exists := aiData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.printSimple(row, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

// Mining rendering function
func (t *TUI) renderMining(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("mining.title"))

	switch miningData := data.(type) {
	case *mining.MiningInfo:
		if miningData.Algorithm != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("mining.algorithm"), miningData.Algorithm), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			
			if miningData.Hashrate != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("mining.hashrate"), miningData.Hashrate), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}

			if miningData.SharesValid > 0 {
				row = t.printSimple(row, fmt.Sprintf("%s: %d valid, %d invalid", 
					t.config.T("mining.shares"), miningData.SharesValid, miningData.SharesInvalid), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			}

			if miningData.Temperature > 0 {
				// Temperature with graph
				tempGraph := t.createSolidGraph(miningData.Temperature)
				row = t.printSimple(row, fmt.Sprintf("%s: %.1fC %s", 
					t.config.T("mining.temperature"), miningData.Temperature, tempGraph), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}

			if miningData.PowerConsumption != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.power"), miningData.PowerConsumption), tcell.StyleDefault.Foreground(tcell.ColorYellow))
			}

			if miningData.Efficiency != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.efficiency"), miningData.Efficiency), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if miningData.Uptime != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.uptime"), miningData.Uptime), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if miningData.Pool != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.pool"), miningData.Pool), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if miningData.Revenue24h != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.revenue_24h"), miningData.Revenue24h), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			}
		} else {
			row = t.printSimple(row, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	case map[string]interface{}:
		if errorMsg, exists := miningData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.printSimple(row, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

func (t *TUI) renderHeader(startRow int, title string) int {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true).Background(tcell.ColorDarkBlue)
	
	// Create header line with padding
	header := fmt.Sprintf(" %s ", title)
	t.printSimple(startRow, header, style)
	
	return startRow + 1
}

func (t *TUI) renderFooter(startRow int) {
	if startRow >= t.height-1 {
		return
	}

	footerText := fmt.Sprintf("Press 'q' or Ctrl+C to exit | Refresh: %ds | kern v1.2.1", t.config.RefreshRate)
	style := tcell.StyleDefault.Foreground(tcell.ColorGray)
	
	// Обеспечиваем что футер внизу и выровнен по левому краю
	footerRow := t.height - 1
	
	// Обрезаем текст если он слишком длинный
	if len(footerText) > t.width {
		footerText = footerText[:t.width-3] + "..."
	}
	
	// Выводим с левого краю
	t.printSimple(footerRow, footerText, style)
}

// Упрощенная функция вывода - всегда с начала строки
func (t *TUI) printSimple(row int, text string, style tcell.Style) int {
	if row < 0 || row >= t.height {
		return row
	}

	// Обрабатываем текст с учетом ширины символов
	x := 0
	for _, ch := range text {
		if x >= t.width {
			break
		}
		t.screen.SetContent(x, row, ch, nil, style)
		x++
	}
	return row + 1
}

// Функция для центрированного вывода
func (t *TUI) printCentered(row int, text string, style tcell.Style) {
	if row < 0 {
		return
	}

	if row >= t.height {
		return
	}

	x := (t.width - len(text)) / 2
	if x < 0 {
		x = 0
	}

	for i, ch := range text {
		if x+i >= t.width {
			break
		}
		t.screen.SetContent(x+i, row, ch, nil, style)
	}
}

// ОБНОВЛЕННАЯ ФУНКЦИЯ: Создание гистограммы (5% на сегмент)
func (t *TUI) createSolidGraph(percent float64) string {
    if percent < 0 {
        percent = 0
    }
    if percent > 100 {
        percent = 100
    }

    segments := 20
    filled := int((percent / 100) * float64(segments))
    
    // ГАРАНТИРУЕМ ОБНОВЛЕНИЕ: даже при малых изменениях процента
    // Если процент > 0, показываем хотя бы 1 сегмент
    if percent > 0 && filled == 0 {
        filled = 1
    }
    // Если процент < 100, но близок к полному, показываем почти полную гистограмму
    if percent > 95 && filled < segments {
        filled = segments
    }
    if filled > segments {
        filled = segments
    }

    // Используем блоки Unicode для гистограммы (5% на сегмент)
    graph := strings.Repeat("█", filled)
    empty := strings.Repeat("░", segments-filled)
    
    return graph + empty
}

// Старая функция для совместимости
func (t *TUI) createCompactGraph(percent float64) string {
	return t.createSolidGraph(percent)
}

func (t *TUI) getDeviceType(filesystem string, mountPoint string) string {
	if strings.Contains(filesystem, "nvme") || strings.Contains(filesystem, "ssd") {
		return "SSD"
	} else if strings.Contains(filesystem, "sd") {
		return "HDD"
	} else if mountPoint == "/" {
		return "ROOT"
	} else if strings.Contains(mountPoint, "home") {
		return "HOME"
	} else if strings.Contains(mountPoint, "boot") {
		return "BOOT"
	}
	return "STORAGE"
}

// Вспомогательная функция для извлечения числового значения памяти из строки
func extractMemoryMB(memoryStr string) int {
	// Пример: "8192 MB" -> 8192
	parts := strings.Fields(memoryStr)
	if len(parts) >= 2 {
		if value, err := strconv.Atoi(parts[0]); err == nil {
			return value
		}
	}
	return 0
}

// drawText правильно обрабатывает символы разной ширины
func (t *TUI) drawText(x, y int, style tcell.Style, text string) {
	for _, r := range text {
		if x >= t.width {
			break
		}
		t.screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}