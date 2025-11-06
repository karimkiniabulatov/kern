package ui

import (
	"fmt"
	"strings"

	"github.com/karimkiniabulatov/kern/internal/config"
	"github.com/karimkiniabulatov/kern/internal/cpu"
	"github.com/karimkiniabulatov/kern/internal/disk"
	"github.com/karimkiniabulatov/kern/internal/mem"
	"github.com/karimkiniabulatov/kern/internal/net"
)

type Renderer struct {
	config *config.Config
}

func NewRenderer(cfg *config.Config) *Renderer {
	return &Renderer{
		config: cfg,
	}
}

func (r *Renderer) Render(data map[string]interface{}) {
	// Полная очистка экрана и курсор в начало
	fmt.Print("\033[2J\033[H")

	// Рендерим все модули
	r.renderHeader()

	if r.config.ShowDisk {
		if diskData, exists := data["disk"]; exists {
			r.renderDisk(diskData)
		}
	}

	if r.config.ShowMem {
		if memData, exists := data["mem"]; exists {
			r.renderMemory(memData)
		}
	}

	if r.config.ShowNet {
		if netData, exists := data["net"]; exists {
			r.renderNetwork(netData)
		}
	}

	if r.config.ShowCPU {
		if cpuData, exists := data["cpu"]; exists {
			r.renderCPU(cpuData)
		}
	}

	r.renderFooter()
}

func (r *Renderer) renderHeader() {
	title := r.config.T("common.title")
	fmt.Printf("\033[1;33m%s\033[0m\n", title) // Ярко-желтый
	r.renderTopBorder()
}

func (r *Renderer) renderTopBorder() {
	fmt.Printf("\033[34m%s\033[0m\n", strings.Repeat(" ", 10))
}

func (r *Renderer) renderBottomBorder() {
	fmt.Printf("\033[34m%s\033[0m\n", strings.Repeat(" ", 10))
}

func (r *Renderer) renderSeparator() {
	fmt.Printf("\033[34m%s\033[0m\n", strings.Repeat(" ", 5))
}

func (r *Renderer) renderCPU(data interface{}) {
	fmt.Printf("\033[1;33m%s\033[0m\n", r.config.T("cpu.title"))
	r.renderTopBorder()

	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		r.renderSeparator()
		fmt.Printf("  \033[36m%s:\033[0m %s\n", r.config.T("cpu.model"), cpuInfo.Model)

		r.renderSeparator()
		fmt.Printf("  \033[36m%s:\033[0m %d %s, %d %s\n", 
			r.config.T("cpu.cores"), cpuInfo.Cores, r.config.T("cpu.cores"), 
			cpuInfo.Threads, r.config.T("cpu.threads"))

		r.renderSeparator()
		graph := r.createSimpleGraph(cpuInfo.Usage, 25)
		fmt.Printf("  \033[36m%s:\033[0m %s \033[38;5;215m%.1f%%\033[0m\n",
			r.config.T("cpu.usage"), graph, cpuInfo.Usage)

		r.renderSeparator()
		fmt.Printf("  \033[36m%s:\033[0m %s\n", r.config.T("cpu.frequency"), cpuInfo.Frequency)

		r.renderSeparator()
		fmt.Printf("  \033[36m%s:\033[0m %.2f, %.2f, %.2f\n",
			r.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15)

		if r.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
			r.renderSeparator()
			fmt.Printf("  \033[36m%s:\033[0m\n", r.config.T("cpu.core_usage"))
			for i, usage := range cpuInfo.CoreUsage {
				coreGraph := r.createSimpleGraph(usage, 15)
				fmt.Printf("    %s %d: %s \033[38;5;215m%.1f%%\033[0m\n",
					r.config.T("cpu.core"), i+1, coreGraph, usage)
			}
		}
	}
	r.renderBottomBorder()
	fmt.Println()
}

func (r *Renderer) renderMemory(data interface{}) {
	fmt.Printf("\033[1;33m%s\033[0m\n", r.config.T("memory.title"))
	r.renderTopBorder()

	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		r.renderSeparator()
		ramGraph := r.createSimpleGraph(memInfo.UsagePercent, 25)
		fmt.Printf("  \033[36m%s:\033[0m %s / %s %s \033[38;5;154m%.1f%%\033[0m\n",
			r.config.T("memory.ram"), memInfo.Used, memInfo.Total, ramGraph, memInfo.UsagePercent)

		r.renderSeparator()
		fmt.Printf("  \033[36m%s:\033[0m %s\n", r.config.T("common.available"), memInfo.Available)

		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			r.renderSeparator()
			swapGraph := r.createSimpleGraph(memInfo.SwapUsagePercent, 25)
			fmt.Printf("  \033[36m%s:\033[0m %s / %s %s \033[38;5;154m%.1f%%\033[0m\n",
				r.config.T("memory.swap"), memInfo.SwapUsed, memInfo.SwapTotal, swapGraph, memInfo.SwapUsagePercent)
		}
	}
	r.renderBottomBorder()
	fmt.Println()
}

func (r *Renderer) renderDisk(data interface{}) {
	fmt.Printf("\033[1;33m%s\033[0m\n", r.config.T("disk.title"))
	r.renderTopBorder()

	if disks, ok := data.([]disk.DiskInfo); ok {
		count := 0
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") && count < 3 {
				r.renderSeparator()
				
				devType := r.getDeviceType(d.Filesystem, d.MountedOn)
				mountPoint := d.MountedOn
				if mountPoint == "/" {
					mountPoint = "ROOT"
				}
				
				fmt.Printf("  \033[36m%s:\033[0m %s (%s)\n", 
					r.config.T("disk.filesystem"), d.Filesystem, devType)
				
				r.renderSeparator()
				fmt.Printf("  \033[36m%s:\033[0m %s\n", r.config.T("disk.mounted"), mountPoint)
				
				r.renderSeparator()
				diskGraph := r.createSimpleGraph(d.UsePercent, 25)
				fmt.Printf("  \033[36m%s:\033[0m %s / %s %s \033[38;5;216m%.1f%%\033[0m\n",
					r.config.T("disk.usage"), d.Used, d.Size, diskGraph, d.UsePercent)
				
				count++
			}
		}
	}
	r.renderBottomBorder()
	fmt.Println()
}

func (r *Renderer) renderNetwork(data interface{}) {
	fmt.Printf("\033[1;33m%s\033[0m\n", r.config.T("network.title"))
	r.renderTopBorder()

	if networks, ok := data.([]net.NetworkInfo); ok {
		for _, net := range networks {
			if net.Status == "UP" && net.Interface != "lo" {
				r.renderSeparator()
				fmt.Printf("  \033[36m%s:\033[0m %s\n", r.config.T("network.interface"), net.Interface)
				
				r.renderSeparator()
				fmt.Printf("  \033[36m%s:\033[0m %s\n", "IP Address", net.IPAddress)
				
				r.renderSeparator()
				fmt.Printf("  \033[36m%s:\033[0m %s\n", "MAC Address", net.MACAddress)
				
				r.renderSeparator()
				activityGraph := r.createSimpleGraph(net.ActivityPercent, 25)
				fmt.Printf("  \033[36m%s:\033[0m %s \033[38;5;165m%.1f%%\033[0m\n",
					"Activity", activityGraph, net.ActivityPercent)
				
				r.renderSeparator()
				fmt.Printf("  \033[36m%s:\033[0m %s↓ / %s↑\n",
					"Speed", net.RXSpeed, net.TXSpeed)
			}
		}
	}
	r.renderBottomBorder()
	fmt.Println()
}

func (r *Renderer) renderFooter() {
	fmt.Printf("\033[90m%s | %s %d %s\033[0m\n", 
		r.config.T("ui.press_quit"), 
		r.config.T("ui.refresh_every"), 
		r.config.RefreshRate, 
		r.config.T("ui.seconds"))
}

func (r *Renderer) createSimpleGraph(percent float64, width int) string {
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
	fmt.Print("\033[0m\033[2J\033[H")
}