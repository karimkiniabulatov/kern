package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/karimkiniabulatov/kern/internal/config"
	"github.com/karimkiniabulatov/kern/internal/cpu"
	"github.com/karimkiniabulatov/kern/internal/disk"
	"github.com/karimkiniabulatov/kern/internal/mem"
	"github.com/karimkiniabulatov/kern/internal/net"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

type Renderer struct {
	config      *config.Config
	termWidth   int
	termHeight  int
	lastData    map[string]interface{}
	initialized bool
}

func NewRenderer(cfg *config.Config) *Renderer {
	width, height, err := term.GetSize(0)
	if err != nil {
		width = 80
		height = 24
	}

	return &Renderer{
		config:     cfg,
		termWidth:  width,
		termHeight: height,
		lastData:   make(map[string]interface{}),
	}
}

func (r *Renderer) Render(data map[string]interface{}) {
	// Сохраняем последние данные для каждого модуля
	for module, moduleData := range data {
		r.lastData[module] = moduleData
	}

	// Не очищаем экран полностью, только перемещаем курсор в начало
	if !r.initialized {
		fmt.Print("\033[2J") // Полная очистка только при первом запуске
		r.initialized = true
	}
	fmt.Print("\033[H") // Курсор в начало

	// Рендерим все модули в фиксированном порядке
	r.renderHeader()
	
	// Disk всегда первый
	if diskData, exists := r.lastData["disk"]; exists {
		r.renderDisk(diskData)
		fmt.Println()
	}
	
	// Memory второй
	if memData, exists := r.lastData["mem"]; exists {
		r.renderMemory(memData)
		fmt.Println()
	}
	
	// Network третий
	if netData, exists := r.lastData["net"]; exists {
		r.renderNetwork(netData)
		fmt.Println()
	}
	
	// CPU всегда последний
	if cpuData, exists := r.lastData["cpu"]; exists {
		r.renderCPU(cpuData)
	}

	// Подсказка для выхода
	fmt.Printf("\n\033[90mPress 'q' to quit | Auto-refresh every %d seconds\033[0m\n", r.config.RefreshRate)
}

func (r *Renderer) renderHeader() {
	title := "kern - System Monitor"
	width := runewidth.StringWidth(title)
	padding := (r.termWidth - width) / 2
	
	fmt.Printf("\033[1;36m%s%s\033[0m\n", strings.Repeat(" ", padding), title)
	fmt.Println(strings.Repeat("=", r.termWidth))
}

func (r *Renderer) renderCPU(data interface{}) {
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("  Processor           Cores/Threads        Usage %     Vendor          Architecture")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		// Основная информация о процессоре
		fmt.Printf("  \033[36m%-20s\033[0m  %-18s  \033[38;5;215m%5.1f%%\033[0m     %-12s  %-12s\n", 
			truncateString(cpuInfo.Model, 20),
			fmt.Sprintf("%d cores, %d threads", cpuInfo.Cores, cpuInfo.Threads),
			cpuInfo.Usage,
			cpuInfo.Vendor,
			cpuInfo.Architecture)

		// Информация по ядрам
		if len(cpuInfo.CoreUsage) > 0 {
			for i, usage := range cpuInfo.CoreUsage {
				// Группируем логические ядра по физическим (2 потока на ядро)
				if i%2 == 0 {
					coreNum := i / 2
					var usage1, usage2 float64
					usage1 = usage
					if i+1 < len(cpuInfo.CoreUsage) {
						usage2 = cpuInfo.CoreUsage[i+1]
					}
					
					// Гистограмма для пары потоков
					graph := ""
					for j := 0; j < 20; j++ {
						if j < int(usage1/5) && j < int(usage2/5) {
							graph += "\033[38;5;215m█\033[0m"
						} else if j < int(usage1/5) {
							graph += "\033[38;5;215m▀\033[0m"
						} else if j < int(usage2/5) {
							graph += "\033[38;5;215m▄\033[0m"
						} else {
							graph += "\033[38;5;153m█\033[0m"
						}
					}
					
					if i+1 < len(cpuInfo.CoreUsage) {
						fmt.Printf("  Core %-2d              -                  %s   \033[38;5;215m%.1f%%/%.1f%%\033[0m\n", 
							coreNum, graph, usage1, usage2)
					} else {
						fmt.Printf("  Core %-2d              -                  %s   \033[38;5;215m%.1f%%\033[0m\n", 
							coreNum, graph, usage1)
					}
				}
			}
		}
	}
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
}

func (r *Renderer) renderMemory(data interface{}) {
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("  Memory Type        Total     Used      Free      Usage Graph                Usage %")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		// Рассчитываем процент использования
		percent := memInfo.UsagePercent
		
		// Строим гистограмму
		memGraph := ""
		usedSegments := int(percent / 5)
		for i := 0; i < 20; i++ {
			if i < usedSegments {
				memGraph += "\033[38;5;154m█\033[0m" // Насыщенный желто-зеленый
			} else {
				memGraph += "\033[38;5;194m█\033[0m" // Бледный зеленый
			}
		}
		
		fmt.Printf("  %-17s  %-8s  %-8s  %-8s  %s   \033[38;5;154m%5.1f%%\033[0m\n",
			"RAM", memInfo.Total, memInfo.Used, memInfo.Free, memGraph, percent)
			
		// Swap информация
		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapPercent := memInfo.SwapUsagePercent
			swapGraph := ""
			swapSegments := int(swapPercent / 5)
			for i := 0; i < 20; i++ {
				if i < swapSegments {
					swapGraph += "\033[38;5;154m█\033[0m"
				} else {
					swapGraph += "\033[38;5;194m█\033[0m"
				}
			}
			
			fmt.Printf("  %-17s  %-8s  %-8s  %-8s  %s   \033[38;5;154m%5.1f%%\033[0m\n",
				"Swap", memInfo.SwapTotal, memInfo.SwapUsed, memInfo.SwapFree, swapGraph, swapPercent)
		}
	}
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────\033[0m")
}

func (r *Renderer) renderDisk(data interface{}) {
	fmt.Println("\033[34m──────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("   Device         Mount Point        Size      Usage Graph                Usage %     Type")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if disks, ok := data.([]disk.DiskInfo); ok {
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") {
				// Определяем тип устройства
				devType := "Disk"
				if strings.Contains(d.Filesystem, "md") {
					devType = "RAID"
				} else if strings.Contains(d.Filesystem, "nvme") {
					devType = "NVMe"
				} else if strings.Contains(d.Filesystem, "sd") {
					devType = "SSD/HDD"
				} else if d.MountedOn == "/boot/efi" {
					devType = "UEFI"
				} else if strings.Contains(d.Filesystem, "vd") {
					devType = "Virtual"
				}
				
				// Строим гистограмму
				graph := ""
				for i := 0; i < 25; i++ {
					if i < int(d.UsePercent/4) {
						graph += "\033[48;5;223m \033[0m" // Бледно-морковный
					} else {
						graph += "\033[48;5;120m \033[0m" // Зеленый
					}
				}
				
				mountPoint := d.MountedOn
				if mountPoint == "/" {
					mountPoint = "ROOT"
				}
				
				fmt.Printf("  \033[36m%-14s\033[0m  %-18s  %-8s  %s   \033[38;5;216m%5.1f%%\033[0m     %-10s\n",
					truncateString(d.Filesystem, 14),
					truncateString(mountPoint, 18),
					d.Size,
					graph,
					d.UsePercent,
					devType)
			}
		}
	}
	fmt.Println("\033[34m──────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
}

func (r *Renderer) renderNetwork(data interface{}) {
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("  Interface        IP Address          MAC Address          RX/TX Speed           Usage Graph                Activity %")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if networks, ok := data.([]net.NetworkInfo); ok {
		for _, net := range networks {
			if net.Status == "UP" && net.Interface != "lo" {
				// Определяем активность (максимум из RX/TX)
				activity := net.ActivityPercent
				if activity > 100 {
					activity = 100
				}
				
				// Строим гистограмму
				graph := ""
				for i := 0; i < 20; i++ {
					if i < int(activity/5) {
						graph += "\033[38;5;165m█\033[0m" // Насыщенный пурпурный
					} else {
						graph += "\033[38;5;225m█\033[0m" // Бледный пурпурный
					}
				}
				
				speedInfo := fmt.Sprintf("%s↓/%s↑", net.RXSpeed, net.TXSpeed)
				
				fmt.Printf("  %-14s  %-18s  %-18s  %-18s  %s   \033[38;5;165m%5.1f%%\033[0m\n",
					truncateString(net.Interface, 14),
					truncateString(net.IPAddress, 18),
					truncateString(net.MACAddress, 18),
					truncateString(speedInfo, 18),
					graph,
					activity)
			}
		}
	}
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
}

func (r *Renderer) Cleanup() {
	fmt.Print("\033[0m") // Сброс цветов
}

// Вспомогательная функция для обрезки строк
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}