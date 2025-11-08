package ui

import (
	"fmt"
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
)

type TUI struct {
	screen   tcell.Screen
	config   *config.Config
	showLogo bool
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

	return &TUI{
		screen:   screen,
		config:   cfg,
		showLogo: showLogo,
	}, nil
}

func (t *TUI) Render(data map[string]interface{}) {
	t.screen.Clear()
	width, height := t.screen.Size()

	row := 0

	// Show logo if needed
	if t.showLogo {
		row = t.renderLogo(row, width)
	}

	// Render only enabled modules
	if t.config.ShowDisk {
		if diskData, exists := data["disk"]; exists {
			row = t.renderDisk(row, width, diskData)
		}
	}

	if t.config.ShowMem {
		if memData, exists := data["mem"]; exists {
			row = t.renderMemory(row, width, memData)
		}
	}

	if t.config.ShowNet {
		if netData, exists := data["net"]; exists {
			row = t.renderNetwork(row, width, netData)
		}
	}

	if t.config.ShowCPU {
		if cpuData, exists := data["cpu"]; exists {
			row = t.renderCPU(row, width, cpuData)
		}
	}

	// NEW: GPU monitoring
	if t.config.ShowGPU {
		if gpuData, exists := data["gpu"]; exists {
			row = t.renderGPU(row, width, gpuData)
		}
	}

	// NEW: AI training monitoring  
	if t.config.ShowAI {
		if aiData, exists := data["ai"]; exists {
			row = t.renderAI(row, width, aiData)
		}
	}

	// NEW: Mining monitoring
	if t.config.ShowMining {
		if miningData, exists := data["mining"]; exists {
			row = t.renderMining(row, width, miningData)
		}
	}

	t.renderFooter(row, width, height)
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

func (t *TUI) renderLogo(startRow int, width int) int {
	logo := []string{
		" ██╗  ██╗███████╗██████╗ ███╗   ██╗",
		" ██║ ██╔╝██╔════╝██╔══██╗████╗  ██║",
		" █████╔╝ █████╗  ██████╔╝██╔██╗ ██║",
		" ██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║",
		" ██║  ██╗███████╗██║  ██║██║ ╚████║",
		" ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝",
		" kern v1.2.0 - System Monitoring Tool",
	}

	cyan := tcell.StyleDefault.Foreground(tcell.ColorTeal).Bold(true)
	for i, line := range logo {
		t.printCentered(startRow+i, line, cyan, width)
	}

	return startRow + len(logo) + 1
}

func (t *TUI) renderCPU(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("cpu.title"), width)

	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("cpu.model"), cpuInfo.Model), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %d %s, %d %s",
			t.config.T("cpu.cores"), cpuInfo.Cores, t.config.T("cpu.cores"),
			cpuInfo.Threads, t.config.T("cpu.threads")), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

		graph := t.createCompactGraph(cpuInfo.Usage, 25)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f%%",
			t.config.T("cpu.usage"), graph, cpuInfo.Usage), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)

		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("cpu.frequency"), cpuInfo.Frequency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %.2f, %.2f, %.2f",
			t.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

		if t.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
			row = t.printLine(row, 0, fmt.Sprintf("%s:", t.config.T("cpu.core_usage")), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			for i, usage := range cpuInfo.CoreUsage {
				coreGraph := t.createCompactGraph(usage, 15)
				row = t.printLine(row, 2, fmt.Sprintf("%s %d: %s %.1f%%",
					t.config.T("cpu.core"), i+1, coreGraph, usage), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}
		}
	}
	return row + 1
}

func (t *TUI) renderMemory(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("memory.title"), width)

	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		ramGraph := t.createCompactGraph(memInfo.UsagePercent, 25)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s %s %.1f%%",
			t.config.T("memory.ram"), memInfo.Used, memInfo.Total, ramGraph, memInfo.UsagePercent), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)

		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("common.available"), memInfo.Available), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapGraph := t.createCompactGraph(memInfo.SwapUsagePercent, 25)
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s %s %.1f%%",
				t.config.T("memory.swap"), memInfo.SwapUsed, memInfo.SwapTotal, swapGraph, memInfo.SwapUsagePercent), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
		}
	}
	return row + 1
}

func (t *TUI) renderDisk(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("disk.title"), width)

	if disks, ok := data.([]disk.DiskInfo); ok {
		count := 0
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") && count < 3 {
				devType := t.getDeviceType(d.Filesystem, d.MountedOn)
				mountPoint := d.MountedOn
				if mountPoint == "/" {
					mountPoint = "ROOT"
				}

				row = t.printLine(row, 0, fmt.Sprintf("%s: %s (%s)",
					t.config.T("disk.filesystem"), d.Filesystem, devType), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("disk.mounted"), mountPoint), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

				diskGraph := t.createCompactGraph(d.UsePercent, 25)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s %s %.1f%%",
					t.config.T("disk.usage"), d.Used, d.Size, diskGraph, d.UsePercent), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)

				count++
				if count < 3 {
					row++
				}
			}
		}
	}
	return row + 1
}

func (t *TUI) renderNetwork(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("network.title"), width)

	if networks, ok := data.([]net.NetworkInfo); ok {
		count := 0
		for _, netInfo := range networks {
			if netInfo.Status == "UP" && netInfo.Interface != "lo" && count < 2 {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("network.interface"), netInfo.Interface), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", "IP Address", netInfo.IPAddress), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", "MAC Address", netInfo.MACAddress), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

				activityGraph := t.createCompactGraph(netInfo.ActivityPercent, 25)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f%%",
					"Activity", activityGraph, netInfo.ActivityPercent), tcell.StyleDefault.Foreground(tcell.ColorFuchsia), width)

				row = t.printLine(row, 0, fmt.Sprintf("%s: %s↓ / %s↑",
					"Speed", netInfo.RXSpeed, netInfo.TXSpeed), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

				count++
				if count < 2 {
					row++
				}
			}
		}
	}
	return row + 1
}

// NEW: GPU rendering function
func (t *TUI) renderGPU(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("gpu.title"), width)

	// Handle both *gpu.GPUInfo and error cases
	switch gpuData := data.(type) {
	case *gpu.GPUInfo:
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("gpu.model"), gpuData.Model), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		
		if gpuData.DriverVersion != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("gpu.driver"), gpuData.DriverVersion), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		}
		
		if gpuData.GPUTemp > 0 {
			tempGraph := t.createCompactGraph(gpuData.GPUTemp, 25)
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f°C", 
				t.config.T("gpu.temperature"), tempGraph, gpuData.GPUTemp), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}

		if gpuData.Utilization > 0 {
			utilGraph := t.createCompactGraph(gpuData.Utilization, 25)
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f%%", 
				t.config.T("gpu.utilization"), utilGraph, gpuData.Utilization), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
		}

		if gpuData.MemoryUsed != "" && gpuData.MemoryTotal != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s", 
				t.config.T("gpu.memory"), gpuData.MemoryUsed, gpuData.MemoryTotal), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		}
		
		if gpuData.PowerDraw != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
				t.config.T("gpu.power"), gpuData.PowerDraw), tcell.StyleDefault.Foreground(tcell.ColorYellow), width)
		}
			
		if gpuData.ClockCore != "" && gpuData.ClockMemory != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s | %s: %s", 
				t.config.T("gpu.clock_core"), gpuData.ClockCore, 
				t.config.T("gpu.clock_memory"), gpuData.ClockMemory), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		}
		
	case map[string]string:
		// Handle error case
		if errorMsg, exists := gpuData["error"]; exists {
			row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorMsg), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}
	default:
		row = t.printLine(row, 0, "No GPU data available", tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

// NEW: AI training rendering function
func (t *TUI) renderAI(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("ai.title"), width)

	switch aiData := data.(type) {
	case *ai.AIInfo:
		if aiData.ProcessCount > 0 {
			if aiData.Framework != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("ai.framework"), aiData.Framework), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
			row = t.printLine(row, 0, fmt.Sprintf("%s: %d", t.config.T("ai.processes"), aiData.ProcessCount), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			
			if aiData.VRAMUsage != "" && aiData.VRAMTotal != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s", 
					t.config.T("ai.vram"), aiData.VRAMUsage, aiData.VRAMTotal), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}

			if aiData.ModelName != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("ai.model"), aiData.ModelName), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}

			if aiData.BatchSize > 0 {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %d | %s: %.1f samples/sec", 
					t.config.T("ai.batch_size"), aiData.BatchSize,
					t.config.T("ai.throughput"), aiData.Throughput), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}

			if aiData.Epoch > 0 {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %d | %s: %.3f | %s: %.1f%%", 
					t.config.T("ai.epoch"), aiData.Epoch,
					t.config.T("ai.loss"), aiData.Loss,
					t.config.T("ai.accuracy"), aiData.Accuracy*100), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}

			if aiData.TrainingTime != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("ai.training_time"), aiData.TrainingTime), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		} else {
			row = t.printLine(row, 0, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
		}
	case map[string]string:
		// Handle error case
		if errorMsg, exists := aiData["error"]; exists {
			row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorMsg), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}
	default:
		row = t.printLine(row, 0, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

// NEW: Mining rendering function
func (t *TUI) renderMining(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("mining.title"), width)

	switch miningData := data.(type) {
	case *mining.MiningInfo:
		if miningData.Algorithm != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s (%s)", 
				t.config.T("mining.algorithm"), miningData.Algorithm, miningData.Currency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			
			if miningData.Hashrate != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.hashrate"), miningData.Hashrate), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}

			if miningData.SharesValid > 0 {
				totalShares := miningData.SharesValid + miningData.SharesInvalid
				successRate := float64(miningData.SharesValid) / float64(totalShares) * 100
				row = t.printLine(row, 0, fmt.Sprintf("%s: %d/%d (%.1f%%)", 
					t.config.T("mining.shares"), miningData.SharesValid, totalShares, successRate),
					tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}

			if miningData.Temperature > 0 {
				tempGraph := t.createCompactGraph(miningData.Temperature, 25)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f°C", 
					t.config.T("mining.temperature"), tempGraph, miningData.Temperature), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}

			if miningData.PowerConsumption != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.power"), miningData.PowerConsumption), tcell.StyleDefault.Foreground(tcell.ColorYellow), width)
			}

			if miningData.Efficiency != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.efficiency"), miningData.Efficiency), tcell.StyleDefault.Foreground(tcell.ColorYellow), width)
			}

			if miningData.Uptime != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.uptime"), miningData.Uptime), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}

			if miningData.Revenue24h != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.revenue_24h"), miningData.Revenue24h), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}

			if miningData.Pool != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.pool"), miningData.Pool), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		} else {
			row = t.printLine(row, 0, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
		}
	case map[string]string:
		// Handle error case
		if errorMsg, exists := miningData["error"]; exists {
			row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorMsg), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}
	default:
		row = t.printLine(row, 0, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

func (t *TUI) renderFooter(row int, width int, height int) {
	if row >= height-1 {
		row = height - 1
	}
	footer := fmt.Sprintf("Press 'q' to quit | Auto-refresh every %d seconds", t.config.RefreshRate)
	t.printCentered(row, footer, tcell.StyleDefault.Foreground(tcell.ColorGray), width)
}

func (t *TUI) renderHeader(row int, text string, width int) int {
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	t.printLine(row, 0, text, style, width)
	return row + 1
}

func (t *TUI) printLine(row int, indent int, text string, style tcell.Style, width int) int {
	_, screenHeight := t.screen.Size()
	if row >= screenHeight {
		return row
	}

	// Apply indentation
	indentSpaces := ""
	for i := 0; i < indent; i++ {
		indentSpaces += " "
	}
	fullText := indentSpaces + text

	// Truncate if too long for screen
	if len(fullText) > width {
		fullText = fullText[:width]
	}

	for col, ch := range fullText {
		if col >= width {
			break
		}
		t.screen.SetContent(col, row, ch, nil, style)
	}
	return row + 1
}

func (t *TUI) printCentered(row int, text string, style tcell.Style, width int) {
	_, screenHeight := t.screen.Size()
	if row >= screenHeight {
		return
	}

	col := (width - len(text)) / 2
	if col < 0 {
		col = 0
	}

	for i, ch := range text {
		if col+i >= width {
			break
		}
		t.screen.SetContent(col+i, row, ch, nil, style)
	}
}

func (t *TUI) createCompactGraph(percent float64, width int) string {
	if percent > 100 {
		percent = 100
	}

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

func (t *TUI) getDeviceType(filesystem, mountPoint string) string {
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