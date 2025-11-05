package ui

import (
	"fmt"
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
	lineCount   int
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
	// Сохраняем последние данные
	for module, moduleData := range data {
		r.lastData[module] = moduleData
	}

	// Рассчитываем общее количество строк для фиксированного вывода
	totalLines := r.calculateTotalLines()
	
	// Очищаем только необходимое пространство
	if !r.initialized {
		fmt.Print("\033[2J") // Полная очистка только при первом запуске
		r.initialized = true
	}
	
	// Перемещаем курсор в начало и очищаем нужное количество строк
	fmt.Printf("\033[H\033[%dS", totalLines)
	fmt.Print("\033[H")

	// Рендерим все модули в фиксированном порядке
	r.renderHeader()
	
	if diskData, exists := r.lastData["disk"]; exists {
		r.renderDisk(diskData)
	}
	
	if memData, exists := r.lastData["mem"]; exists {
		r.renderMemory(memData)
	}
	
	if netData, exists := r.lastData["net"]; exists {
		r.renderNetwork(netData)
	}
	
	if cpuData, exists := r.lastData["cpu"]; exists {
		r.renderCPU(cpuData)
	}

	// Подсказка для выхода
	fmt.Printf("\n\033[90mPress 'q' to quit | Auto-refresh every %d seconds\033[0m\n", r.config.RefreshRate)
}

func (r *Renderer) calculateTotalLines() int {
	lines := 4 // header + footer
	
	if _, exists := r.lastData["disk"]; exists {
		if disks, ok := r.lastData["disk"].([]disk.DiskInfo); ok {
			lines += 3 + len(disks) // header + separator + lines
		}
	}
	
	if _, exists := r.lastData["mem"]; exists {
		lines += 6 // header + separator + RAM + Swap
	}
	
	if _, exists := r.lastData["net"]; exists {
		if networks, ok := r.lastData["net"].([]net.NetworkInfo); ok {
			lines += 3 + len(networks) // header + separator + lines
		}
	}
	
	if _, exists := r.lastData["cpu"]; exists {
		if cpuInfo, ok := r.lastData["cpu"].(*cpu.CPUInfo); ok {
			lines += 3 // header + separator
			if r.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
				// Для детального режима - одна строка на каждые 2 потока
				lines += (len(cpuInfo.CoreUsage) + 1) / 2
			} else {
				lines += 2 // основная строка + общая загрузка
			}
		}
	}
	
	return lines
}

func (r *Renderer) renderHeader() {
	title := "kern - System Monitor"
	width := runewidth.StringWidth(title)
	padding := (r.termWidth - width) / 2
	
	fmt.Printf("\033[1;36m%s%s\033[0m\n", strings.Repeat(" ", padding), title)
	fmt.Println(strings.Repeat("=", r.termWidth))
	fmt.Println()
}

func (r *Renderer) renderCPU(data interface{}) {
	fmt.Println("\033[1;34mCPU Information:\033[0m")
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		// Основная информация о процессоре
		fmt.Printf("  \033[36m%-20s\033[0m  %-18s  \033[38;5;215m%5.1f%%\033[0m     %-12s  %-12s\n", 
			truncateString(cpuInfo.Model, 20),
			fmt.Sprintf("%d cores, %d threads", cpuInfo.Cores, cpuInfo.Threads),
			cpuInfo.Usage,
			cpuInfo.Vendor,
			cpuInfo.Architecture)

		if r.config.DetailedCPU {
			// Детальный режим - показываем все ядра/потоки
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
						graph := r.createCPUGraph(usage1, usage2)
						
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
		} else {
			// Компактный режим - одна общая гистограмма
			graph := r.createSimpleGraph(cpuInfo.Usage, 20)
			fmt.Printf("  Overall Usage        -                  %s   \033[38;5;215m%5.1f%%\033[0m\n", graph, cpuInfo.Usage)
		}
	}
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
	fmt.Println()
}

func (r *Renderer) renderMemory(data interface{}) {
	fmt.Println("\033[1;34mMemory Information:\033[0m")
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		// RAM
		ramGraph := r.createSimpleGraph(memInfo.UsagePercent, 20)
		fmt.Printf("  %-17s  %-8s  %-8s  %-8s  %s   \033[38;5;154m%5.1f%%\033[0m\n",
			"RAM", memInfo.Total, memInfo.Used, memInfo.Free, ramGraph, memInfo.UsagePercent)
			
		// Swap
		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapGraph := r.createSimpleGraph(memInfo.SwapUsagePercent, 20)
			fmt.Printf("  %-17s  %-8s  %-8s  %-8s  %s   \033[38;5;154m%5.1f%%\033[0m\n",
				"Swap", memInfo.SwapTotal, memInfo.SwapUsed, memInfo.SwapFree, swapGraph, memInfo.SwapUsagePercent)
		}
	}
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────\033[0m")
	fmt.Println()
}

func (r *Renderer) renderDisk(data interface{}) {
	fmt.Println("\033[1;34mDisk Information:\033[0m")
	fmt.Println("\033[34m──────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if disks, ok := data.([]disk.DiskInfo); ok {
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") {
				// Определяем тип устройства
				devType := r.getDeviceType(d.Filesystem, d.MountedOn)
				
				// Строим гистограмму
				graph := r.createDiskGraph(d.UsePercent)
				
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
	fmt.Println()
}

func (r *Renderer) renderNetwork(data interface{}) {
	fmt.Println("\033[1;34mNetwork Information:\033[0m")
	fmt.Println("\033[34m────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\033[0m")
	
	if networks, ok := data.([]net.NetworkInfo); ok {
		for _, net := range networks {
			if net.Status == "UP" && net.Interface != "lo" {
				// Определяем активность
				activity := net.ActivityPercent
				if activity > 100 {
					activity = 100
				}
				
				// Строим гистограмму
				graph := r.createSimpleGraph(activity, 20)
				
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
	fmt.Println()
}

// Вспомогательные методы для создания графиков
func (r *Renderer) createSimpleGraph(percent float64, width int) string {
	graph := ""
	usedSegments := int(percent / (100.0 / float64(width)))
	
	for i := 0; i < width; i++ {
		if i < usedSegments {
			graph += "█"
		} else {
			graph += "░"
		}
	}
	return graph
}

func (r *Renderer) createDiskGraph(percent float64) string {
	graph := ""
	usedSegments := int(percent / 4) // 25 segments for 100%
	
	for i := 0; i < 25; i++ {
		if i < usedSegments {
			graph += "█"
		} else {
			graph += "░"
		}
	}
	return graph
}

func (r *Renderer) createCPUGraph(usage1, usage2 float64) string {
	graph := ""
	for j := 0; j < 20; j++ {
		pos := float64(j) * 5.0 // 5% per segment
		if pos < usage1 && pos < usage2 {
			graph += "█"
		} else if pos < usage1 {
			graph += "▀"
		} else if pos < usage2 {
			graph += "▄"
		} else {
			graph += "░"
		}
	}
	return graph
}

func (r *Renderer) getDeviceType(filesystem, mountPoint string) string {
	if strings.Contains(filesystem, "md") {
		return "RAID"
	} else if strings.Contains(filesystem, "nvme") {
		return "NVMe"
	} else if strings.Contains(filesystem, "sd") {
		return "SSD/HDD"
	} else if mountPoint == "/boot/efi" {
		return "UEFI"
	} else if strings.Contains(filesystem, "vd") {
		return "Virtual"
	}
	return "Disk"
}

func (r *Renderer) Cleanup() {
	fmt.Print("\033[0m") // Сброс цветов
	fmt.Print("\033[2J") // Очистка экрана
	fmt.Print("\033[H")  // Курсор в начало
}

// Вспомогательная функция для обрезки строк
func truncateString(s string, length int) string {
	if runewidth.StringWidth(s) <= length {
		return s
	}
	
	// Обрезаем с учетом ширины символов
	result := ""
	width := 0
	for _, r := range s {
		charWidth := runewidth.RuneWidth(r)
		if width+charWidth > length {
			break
		}
		result += string(r)
		width += charWidth
	}
	return result
}