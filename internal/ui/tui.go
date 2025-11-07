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
)

type TUI struct {
	screen    tcell.Screen
	config    *config.Config
	showLogo  bool
	lastData  map[string]interface{}
}

func NewTUI(cfg *config.Config, showLogo bool) (*TUI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	return &TUI{
		screen:   screen,
		config:   cfg,
		showLogo: showLogo,
		lastData: make(map[string]interface{}),
	}, nil
}

func (t *TUI) Render(data map[string]interface{}) {
	t.screen.Clear()
	width, height := t.screen.Size()

	// Сохраняем данные для перерисовки при resize
	t.lastData = data

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

	t.renderFooter(row, width, height)
	t.screen.Show()
}

// ForceRedraw перерисовывает экран с последними данными
func (t *TUI) ForceRedraw() {
	if len(t.lastData) > 0 {
		t.Render(t.lastData)
	}
}

func (t *TUI) renderLogo(startRow int, width int) int {
	logo := []string{
		" ██╗  ██╗███████╗██████╗ ███╗   ██╗",
		" ██║ ██╔╝██╔════╝██╔══██╗████╗  ██║",
		" █████╔╝ █████╗  ██████╔╝██╔██╗ ██║",
		" ██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║",
		" ██║  ██╗███████╗██║  ██║██║ ╚████║",
		" ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝",
		" kern v1.1.0 - System Monitoring Tool",
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

		graph := t.createCompactGraph(cpuInfo.Usage, 15)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f%%",
			t.config.T("cpu.usage"), graph, cpuInfo.Usage), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)

		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("cpu.frequency"), cpuInfo.Frequency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %.2f, %.2f, %.2f",
			t.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

		if t.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
			row = t.printLine(row, 0, fmt.Sprintf("%s:", t.config.T("cpu.core_usage")), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			for i, usage := range cpuInfo.CoreUsage {
				coreGraph := t.createCompactGraph(usage, 10)
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
		ramGraph := t.createCompactGraph(memInfo.UsagePercent, 15)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s %s %.1f%%",
			t.config.T("memory.ram"), memInfo.Used, memInfo.Total, ramGraph, memInfo.UsagePercent), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)

		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("common.available"), memInfo.Available), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapGraph := t.createCompactGraph(memInfo.SwapUsagePercent, 15)
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
				
				diskGraph := t.createCompactGraph(d.UsePercent, 15)
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
				
				activityGraph := t.createCompactGraph(netInfo.ActivityPercent, 15)
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

func (t *TUI) PollEvent() tcell.Event {
	return t.screen.PollEvent()
}

func (t *TUI) Fini() {
	t.screen.Fini()
}

func (t *TUI) Clear() {
    t.screen.Clear()
}