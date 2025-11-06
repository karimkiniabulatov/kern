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

	// Перемещаем курсор в начало
	if !r.initialized {
		fmt.Print("\033[2J") // Полная очистка только при первом запуске
		r.initialized = true
		r.lineCount = 0
	}
	
	fmt.Print("\033[H") // Курсор в начало
	
	// Сбрасываем счетчик строк
	currentLineCount := 0

	// Рендерим все модули в фиксированном порядке
	currentLineCount += r.renderHeader()
	
	// Рендерим только включенные модули
	if r.config.ShowDisk {
		if diskData, exists := r.lastData["disk"]; exists {
			currentLineCount += r.renderDisk(diskData)
			currentLineCount += r.renderEmptyLine()
		}
	}
	
	if r.config.ShowMem {
		if memData, exists := r.lastData["mem"]; exists {
			currentLineCount += r.renderMemory(memData)
			currentLineCount += r.renderEmptyLine()
		}
	}
	
	if r.config.ShowNet {
		if netData, exists := r.lastData["net"]; exists {
			currentLineCount += r.renderNetwork(netData)
			currentLineCount += r.renderEmptyLine()
		}
	}
	
	if r.config.ShowCPU {
		if cpuData, exists := r.lastData["cpu"]; exists {
			currentLineCount += r.renderCPU(cpuData)
			currentLineCount += r.renderEmptyLine()
		}
	}

	// Подсказка для выхода
	currentLineCount += r.renderFooter()

	// Очищаем оставшиеся строки от предыдущего рендера
	if currentLineCount < r.lineCount {
		for i := currentLineCount; i < r.lineCount; i++ {
			fmt.Print("\033[K\n") // Очищаем строку и переходим на следующую
		}
		fmt.Printf("\033[%dA", r.lineCount-currentLineCount) // Возвращаем курсор
	}
	
	r.lineCount = currentLineCount
}

func (r *Renderer) renderHeader() int {
	title := r.config.T("title")
	width := runewidth.StringWidth(title)
	padding := (r.termWidth - width) / 2
	if padding < 0 {
		padding = 0
	}
	
	fmt.Printf("\033[1;36m%s%s\033[0m\n", strings.Repeat(" ", padding), title)
	return 1 + r.renderSeparator()
}

func (r *Renderer) renderSeparator() int {
	separator := strings.Repeat("─", 60)
	fmt.Printf("\033[34m%s\033[0m\n", separator)
	return 1
}

func (r *Renderer) renderEmptyLine() int {
	fmt.Println()
	return 1
}

func (r *Renderer) renderCPU(data interface{}) int {
	lines := 0
	fmt.Println("\033[1;34m" + r.config.T("cpu.title") + "\033[0m")
	lines++
	lines += r.renderSeparator()
	
	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		// Основная информация о процессоре
		model := truncateString(cpuInfo.Model, 20)
		coresInfo := fmt.Sprintf("%d %s, %d %s", 
			cpuInfo.Cores, r.config.T("cpu.cores"), 
			cpuInfo.Threads, r.config.T("cpu.threads"))
		
		fmt.Printf("  \033[36m%-20s\033[0m  %-18s  \033[38;5;215m%5.1f%%\033[0m     %-12s  %-12s\n", 
			model, coresInfo, cpuInfo.Usage, cpuInfo.Vendor, cpuInfo.Architecture)
		lines++
		
		if r.config.DetailedCPU {
			// Детальный режим - показываем все ядра/потоки
			if len(cpuInfo.CoreUsage) > 0 {
				for i, usage := range cpuInfo.CoreUsage {
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
							fmt.Printf("  %s %-2d              -                  %s   \033[38;5;215m%.1f%%/%.1f%%\033[0m\n", 
								r.config.T("cpu.core"), coreNum, graph, usage1, usage2)
						} else {
							fmt.Printf("  %s %-2d              -                  %s   \033[38;5;215m%.1f%%\033[0m\n", 
								r.config.T("cpu.core"), coreNum, graph, usage1)
						}
						lines++
					}
				}
			}
		} else {
			// Компактный режим - одна общая гистограмма
			graph := r.createSimpleGraph(cpuInfo.Usage, 20)
			fmt.Printf("  %s        -                  %s   \033[38;5;215m%5.1f%%\033[0m\n", 
				r.config.T("cpu.overall_usage"), graph, cpuInfo.Usage)
			lines++
		}
	}
	lines += r.renderSeparator()
	return lines
}

func (r *Renderer) renderMemory(data interface{}) int {
	lines := 0
	fmt.Println("\033[1;34m" + r.config.T("memory.title") + "\033[0m")
	lines++
	lines += r.renderSeparator()
	
	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		// RAM
		ramGraph := r.createSimpleGraph(memInfo.UsagePercent, 20)
		fmt.Printf("  %-17s  %-8s  %-8s  %-8s  %s   \033[38;5;154m%5.1f%%\033[0m\n",
			r.config.T("memory.ram"), memInfo.Total, memInfo.Used, memInfo.Free, ramGraph, memInfo.UsagePercent)
		lines++
		
		// Swap
		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapGraph := r.createSimpleGraph(memInfo.SwapUsagePercent, 20)
			fmt.Printf("  %-17s  %-8s  %-8s  %-8s  %s   \033[38;5;154m%5.1f%%\033[0m\n",
				r.config.T("memory.swap"), memInfo.SwapTotal, memInfo.SwapUsed, memInfo.SwapFree, swapGraph, memInfo.SwapUsagePercent)
			lines++
		}
	}
	lines += r.renderSeparator()
	return lines
}

func (r *Renderer) renderDisk(data interface{}) int {
	lines := 0
	fmt.Println("\033[1;34m" + r.config.T("disk.title") + "\033[0m")
	lines++
	lines += r.renderSeparator()
	
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
				
				fs := truncateString(d.Filesystem, 14)
				mp := truncateString(mountPoint, 18)
				
				fmt.Printf("  \033[36m%-14s\033[0m  %-18s  %-8s  %s   \033[38;5;216m%5.1f%%\033[0m     %-10s\n",
					fs, mp, d.Size, graph, d.UsePercent, devType)
				lines++
			}
		}
	}
	lines += r.renderSeparator()
	return lines
}

func (r *Renderer) renderNetwork(data interface{}) int {
	lines := 0
	fmt.Println("\033[1;34m" + r.config.T("network.title") + "\033[0m")
	lines++
	lines += r.renderSeparator()
	
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
				
				iface := truncateString(net.Interface, 14)
				ip := truncateString(net.IPAddress, 18)
				mac := truncateString(net.MACAddress, 18)
				speed := truncateString(speedInfo, 18)
				
				fmt.Printf("  %-14s  %-18s  %-18s  %-18s  %s   \033[38;5;165m%5.1f%%\033[0m\n",
					iface, ip, mac, speed, graph, activity)
				lines++
			}
		}
	}
	lines += r.renderSeparator()
	return lines
}

func (r *Renderer) renderFooter() int {
	fmt.Printf("\033[90m%s | %s %d %s\033[0m\n", 
		r.config.T("ui.press_quit"), 
		r.config.T("ui.refresh_every"), 
		r.config.RefreshRate, 
		r.config.T("ui.seconds"))
	return 1
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