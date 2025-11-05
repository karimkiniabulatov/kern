package ui

import (
	"fmt"
	"strings"
	"github.com/karimkiniabulatov/kern/internal/cpu"    // Добавлено
	"github.com/karimkiniabulatov/kern/internal/disk"   // Добавлено 
	"github.com/karimkiniabulatov/kern/internal/mem"    // Добавлено
	"github.com/karimkiniabulatov/kern/internal/net"    // Добавлено
	"github.com/karimkiniabulatov/kern/internal/config"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

type Renderer struct {
	config   *config.Config
	termWidth int
}

func NewRenderer(cfg *config.Config) *Renderer {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80 // Default width
	}

	return &Renderer{
		config:    cfg,
		termWidth: width,
	}
}

func (r *Renderer) Render(data map[string]interface{}) {
	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	// Render header
	r.renderHeader()

	// Render each module
	for module, moduleData := range data {
		switch module {
		case "cpu":
			r.renderCPU(moduleData)
		case "mem":
			r.renderMemory(moduleData)
		case "disk":
			r.renderDisk(moduleData)
		case "net":
			r.renderNetwork(moduleData)
		}
		fmt.Println()
	}
}

func (r *Renderer) renderHeader() {
	title := "kern - System Monitor"
	width := runewidth.StringWidth(title)
	padding := (r.termWidth - width) / 2
	
	fmt.Printf("\033[1;36m%s%s\033[0m\n", strings.Repeat(" ", padding), title)
	fmt.Println(strings.Repeat("=", r.termWidth))
}

func (r *Renderer) renderCPU(data interface{}) {
	fmt.Println("\033[1;34mCPU Information:\033[0m")
	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		fmt.Printf("  Model:    %s\n", cpuInfo.Model)
		fmt.Printf("  Cores:    %d\n", cpuInfo.Cores)
		fmt.Printf("  Frequency: %s\n", cpuInfo.Frequency)
		fmt.Printf("  Load:     %.2f, %.2f, %.2f (1, 5, 15 min)\n", 
			cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15)
		
		// Simple load bar
		r.renderBar("CPU Load", cpuInfo.Load1/float64(cpuInfo.Cores)*100, 50)
	}
}

func (r *Renderer) renderMemory(data interface{}) {
	fmt.Println("\033[1;34mMemory Information:\033[0m")
	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		fmt.Printf("  Total:     %s\n", memInfo.Total)
		fmt.Printf("  Used:      %s\n", memInfo.Used)
		fmt.Printf("  Free:      %s\n", memInfo.Free)
		fmt.Printf("  Available: %s\n", memInfo.Available)
		fmt.Printf("  Swap:      %s / %s\n", memInfo.SwapUsed, memInfo.SwapTotal)
	}
}

func (r *Renderer) renderDisk(data interface{}) {
	fmt.Println("\033[1;34mDisk Information:\033[0m")
	if disks, ok := data.([]disk.DiskInfo); ok {
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") {
				fmt.Printf("  %s: %s used of %s (%.1f%%) on %s\n",
					d.Filesystem, d.Used, d.Size, d.UsePercent, d.MountedOn)
				r.renderBar("Usage", d.UsePercent, 30)
			}
		}
	}
}

func (r *Renderer) renderNetwork(data interface{}) {
	fmt.Println("\033[1;34mNetwork Information:\033[0m")
	if networks, ok := data.([]net.NetworkInfo); ok {
		for _, net := range networks {
			if net.Status == "UP" && net.Interface != "lo" {
				statusColor := "\033[32m"
				if net.Status == "DOWN" {
					statusColor = "\033[31m"
				}
				fmt.Printf("  %s: %s%s\033[0m, RX: %s, TX: %s\n",
					net.Interface, statusColor, net.Status, net.RXBytes, net.TXBytes)
			}
		}
	}
}

func (r *Renderer) renderBar(label string, percent float64, width int) {
	if percent > 100 {
		percent = 100
	}
	
	barWidth := width - len(label) - 10
	filled := int((percent / 100) * float64(barWidth))
	empty := barWidth - filled
	
	color := "\033[32m" // Green
	if percent > 80 {
		color = "\033[31m" // Red
	} else if percent > 60 {
		color = "\033[33m" // Yellow
	}
	
	bar := color + strings.Repeat("█", filled) + 
		"\033[37m" + strings.Repeat("░", empty) + "\033[0m"
	
	fmt.Printf("  %s: [%s] %.1f%%\n", label, bar, percent)
}

func (r *Renderer) Cleanup() {
	// Reset terminal attributes
	fmt.Print("\033[0m")
}