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
	config   *config.Config
	showLogo bool
}

func NewRenderer(cfg *config.Config, showLogo bool) *Renderer {
	return &Renderer{
		config:   cfg,
		showLogo: showLogo,
	}
}

func (r *Renderer) Render(data map[string]interface{}) {
	// Move cursor to top and clear screen
	fmt.Print("\033[H\033[2J")

	// Show logo if needed
	if r.showLogo {
		r.renderLogo()
	}

	// Render only enabled modules
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

func (r *Renderer) renderLogo() {
	logo := `
 ██╗  ██╗███████╗██████╗ ███╗   ██╗
 ██║ ██╔╝██╔════╝██╔══██╗████╗  ██║
 █████╔╝ █████╗  ██████╔╝██╔██╗ ██║
 ██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║
 ██║  ██╗███████╗██║  ██║██║ ╚████║
 ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝
 kern v1.1.0 - System Monitoring Tool
`
	fmt.Print("\033[1;36m" + logo + "\033[0m\n\n")
}

func (r *Renderer) renderCPU(data interface{}) {
	r.printHeader(r.config.T("cpu.title"))

	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("cpu.model"), cpuInfo.Model))
		r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %d %s, %d %s", 
			r.config.T("cpu.cores"), cpuInfo.Cores, r.config.T("cpu.cores"), 
			cpuInfo.Threads, r.config.T("cpu.threads")))

		graph := r.createSimpleGraph(cpuInfo.Usage, 25)
		r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s \033[38;5;215m%.1f%%\033[0m",
			r.config.T("cpu.usage"), graph, cpuInfo.Usage))

		r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("cpu.frequency"), cpuInfo.Frequency))
		r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %.2f, %.2f, %.2f",
			r.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15))

		if r.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
			r.printLine(fmt.Sprintf("\033[36m%s:\033[0m", r.config.T("cpu.core_usage")))
			for i, usage := range cpuInfo.CoreUsage {
				coreGraph := r.createSimpleGraph(usage, 15)
				r.printLine(fmt.Sprintf("  %s %d: %s \033[38;5;215m%.1f%%\033[0m",
					r.config.T("cpu.core"), i+1, coreGraph, usage))
			}
		}
	}
	r.printEmptyLine()
}

func (r *Renderer) renderMemory(data interface{}) {
	r.printHeader(r.config.T("memory.title"))

	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		ramGraph := r.createSimpleGraph(memInfo.UsagePercent, 25)
		r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s / %s %s \033[38;5;154m%.1f%%\033[0m",
			r.config.T("memory.ram"), memInfo.Used, memInfo.Total, ramGraph, memInfo.UsagePercent))

		r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("common.available"), memInfo.Available))

		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapGraph := r.createSimpleGraph(memInfo.SwapUsagePercent, 25)
			r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s / %s %s \033[38;5;154m%.1f%%\033[0m",
				r.config.T("memory.swap"), memInfo.SwapUsed, memInfo.SwapTotal, swapGraph, memInfo.SwapUsagePercent))
		}
	}
	r.printEmptyLine()
}

func (r *Renderer) renderDisk(data interface{}) {
	r.printHeader(r.config.T("disk.title"))

	if disks, ok := data.([]disk.DiskInfo); ok {
		count := 0
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") && count < 3 {
				devType := r.getDeviceType(d.Filesystem, d.MountedOn)
				mountPoint := d.MountedOn
				if mountPoint == "/" {
					mountPoint = "ROOT"
				}
				
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s (%s)", 
					r.config.T("disk.filesystem"), d.Filesystem, devType))
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("disk.mounted"), mountPoint))
				
				diskGraph := r.createSimpleGraph(d.UsePercent, 25)
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s / %s %s \033[38;5;216m%.1f%%\033[0m",
					r.config.T("disk.usage"), d.Used, d.Size, diskGraph, d.UsePercent))
				
				count++
				if count < 3 {
					r.printEmptyLine()
				}
			}
		}
	}
}

func (r *Renderer) renderNetwork(data interface{}) {
	r.printHeader(r.config.T("network.title"))

	if networks, ok := data.([]net.NetworkInfo); ok {
		count := 0
		for _, net := range networks {
			if net.Status == "UP" && net.Interface != "lo" && count < 2 {
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("network.interface"), net.Interface))
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s", "IP Address", net.IPAddress))
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s", "MAC Address", net.MACAddress))
				
				activityGraph := r.createSimpleGraph(net.ActivityPercent, 25)
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s \033[38;5;165m%.1f%%\033[0m",
					"Activity", activityGraph, net.ActivityPercent))
				
				r.printLine(fmt.Sprintf("\033[36m%s:\033[0m %s↓ / %s↑",
					"Speed", net.RXSpeed, net.TXSpeed))
				
				count++
				if count < 2 {
					r.printEmptyLine()
				}
			}
		}
	}
}

func (r *Renderer) renderFooter() {
	r.printLine(fmt.Sprintf("\033[90mPress 'q' to quit | Auto-refresh every %d seconds\033[0m", 
		r.config.RefreshRate))
}

func (r *Renderer) printHeader(text string) {
	fmt.Printf("\033[1;33m%s\033[0m\n", text)
}

func (r *Renderer) printLine(text string) {
	fmt.Println(text)
}

func (r *Renderer) printEmptyLine() {
	fmt.Println()
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