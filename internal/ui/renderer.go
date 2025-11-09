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
	config       *config.Config
	showLogo     bool
	screenBuffer []string
}

func NewRenderer(cfg *config.Config, showLogo bool) *Renderer {
	return &Renderer{
		config:   cfg,
		showLogo: showLogo,
	}
}

func (r *Renderer) Render(data map[string]interface{}) {
	// Clear screen buffer
	r.screenBuffer = []string{}

	// Build screen content in buffer
	if r.showLogo {
		r.renderLogoToBuffer()
	}

	// Render only enabled modules
	if r.config.ShowDisk {
		if diskData, exists := data["disk"]; exists {
			r.renderDiskToBuffer(diskData)
		}
	}

	if r.config.ShowMem {
		if memData, exists := data["mem"]; exists {
			r.renderMemoryToBuffer(memData)
		}
	}

	if r.config.ShowNet {
		if netData, exists := data["net"]; exists {
			r.renderNetworkToBuffer(netData)
		}
	}

	if r.config.ShowCPU {
		if cpuData, exists := data["cpu"]; exists {
			r.renderCPUToBuffer(cpuData)
		}
	}

	// NEW: GPU monitoring
	if r.config.ShowGPU {
		if gpuData, exists := data["gpu"]; exists {
			r.renderGPUToBuffer(gpuData)
		}
	}

	// NEW: AI training monitoring
	if r.config.ShowAI {
		if aiData, exists := data["ai"]; exists {
			r.renderAIToBuffer(aiData)
		}
	}

	// NEW: Mining monitoring
	if r.config.ShowMining {
		if miningData, exists := data["mining"]; exists {
			r.renderMiningToBuffer(miningData)
		}
	}

	r.renderFooterToBuffer()

	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")
	
	// Print entire buffer at once
	for _, line := range r.screenBuffer {
		fmt.Println(line)
	}
}

func (r *Renderer) renderLogoToBuffer() {
	logo := `
 ██╗  ██╗███████╗██████╗ ███╗   ██╗
 ██║ ██╔╝██╔════╝██╔══██╗████╗  ██║
 █████╔╝ █████╗  ██████╔╝██╔██╗ ██║
 ██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║
 ██║  ██╗███████╗██║  ██║██║ ╚████║
 ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝
 kern v1.2.0 - System Monitoring Tool
`
	r.screenBuffer = append(r.screenBuffer, "\033[1;36m"+strings.TrimSpace(logo)+"\033[0m")
	r.screenBuffer = append(r.screenBuffer, "")
}

func (r *Renderer) renderCPUToBuffer(data interface{}) {
	r.addHeaderToBuffer(r.config.T("cpu.title"))

	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("cpu.model"), cpuInfo.Model))
		r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %d %s, %d %s", 
			r.config.T("cpu.cores"), cpuInfo.Cores, r.config.T("cpu.cores"), 
			cpuInfo.Threads, r.config.T("cpu.threads")))

		graph := r.createSimpleGraph(cpuInfo.Usage, 10)
		r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s \033[38;5;215m%.1f%%\033[0m",
			r.config.T("cpu.usage"), graph, cpuInfo.Usage))

		r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("cpu.frequency"), cpuInfo.Frequency))
		r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %.2f, %.2f, %.2f",
			r.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15))

		if r.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
			r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m", r.config.T("cpu.core_usage")))
			for i, usage := range cpuInfo.CoreUsage {
				coreGraph := r.createSimpleGraph(usage, 10)
				// Форматируем номер ядра с ведущим нулем для ядер 0-9
				coreNumber := fmt.Sprintf("%02d", i+1)
				r.addLineToBuffer(fmt.Sprintf("  %s %s: %s \033[38;5;215m%.1f%%\033[0m",
					r.config.T("cpu.core"), coreNumber, coreGraph, usage))
			}
		}
	}
	r.addEmptyLineToBuffer()
}

func (r *Renderer) renderMemoryToBuffer(data interface{}) {
	r.addHeaderToBuffer(r.config.T("memory.title"))

	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		ramGraph := r.createSimpleGraph(memInfo.UsagePercent, 10)
		r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s / %s %s \033[38;5;154m%.1f%%\033[0m",
			r.config.T("memory.ram"), memInfo.Used, memInfo.Total, ramGraph, memInfo.UsagePercent))

		r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("common.available"), memInfo.Available))

		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapGraph := r.createSimpleGraph(memInfo.SwapUsagePercent, 10)
			r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s / %s %s \033[38;5;154m%.1f%%\033[0m",
				r.config.T("memory.swap"), memInfo.SwapUsed, memInfo.SwapTotal, swapGraph, memInfo.SwapUsagePercent))
		}
	}
	r.addEmptyLineToBuffer()
}

func (r *Renderer) renderDiskToBuffer(data interface{}) {
	r.addHeaderToBuffer(r.config.T("disk.title"))

	if disks, ok := data.([]disk.DiskInfo); ok {
		count := 0
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") && count < 3 {
				devType := r.getDeviceType(d.Filesystem, d.MountedOn)
				mountPoint := d.MountedOn
				if mountPoint == "/" {
					mountPoint = "ROOT"
				}
				
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s (%s)", 
					r.config.T("disk.filesystem"), d.Filesystem, devType))
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("disk.mounted"), mountPoint))
				
				diskGraph := r.createSimpleGraph(d.UsePercent, 10)
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s / %s %s \033[38;5;216m%.1f%%\033[0m",
					r.config.T("disk.usage"), d.Used, d.Size, diskGraph, d.UsePercent))
				
				count++
				if count < 3 {
					r.addEmptyLineToBuffer()
				}
			}
		}
	}
}

func (r *Renderer) renderNetworkToBuffer(data interface{}) {
	r.addHeaderToBuffer(r.config.T("network.title"))

	if networks, ok := data.([]net.NetworkInfo); ok {
		count := 0
		for _, netInfo := range networks {
			if netInfo.Status == "UP" && netInfo.Interface != "lo" && count < 2 {
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s", r.config.T("network.interface"), netInfo.Interface))
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s", "IP Address", netInfo.IPAddress))
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s", "MAC Address", netInfo.MACAddress))
				
				activityGraph := r.createSimpleGraph(netInfo.ActivityPercent, 10)
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s \033[38;5;165m%.1f%%\033[0m",
					"Activity", activityGraph, netInfo.ActivityPercent))
				
				r.addLineToBuffer(fmt.Sprintf("\033[36m%s:\033[0m %s↓ / %s↑",
					"Speed", netInfo.RXSpeed, netInfo.TXSpeed))
				
				count++
				if count < 2 {
					r.addEmptyLineToBuffer()
				}
			}
		}
	}
}

func (r *Renderer) renderGPUToBuffer(data interface{}) {
	r.addHeaderToBuffer(r.config.T("gpu.title"))

	switch gpuData := data.(type) {
	case map[string]interface{}:
		if errorMsg, exists := gpuData["error"]; exists {
			r.addLineToBuffer(fmt.Sprintf("Error: %v", errorMsg))
		}
	default:
		r.addLineToBuffer("GPU monitoring not available")
	}
	r.addEmptyLineToBuffer()
}

func (r *Renderer) renderAIToBuffer(data interface{}) {
	r.addHeaderToBuffer(r.config.T("ai.title"))

	switch aiData := data.(type) {
	case map[string]interface{}:
		if errorMsg, exists := aiData["error"]; exists {
			r.addLineToBuffer(fmt.Sprintf("Error: %v", errorMsg))
		}
	default:
		r.addLineToBuffer(r.config.T("ai.no_training"))
	}
	r.addEmptyLineToBuffer()
}

func (r *Renderer) renderMiningToBuffer(data interface{}) {
	r.addHeaderToBuffer(r.config.T("mining.title"))

	switch miningData := data.(type) {
	case map[string]interface{}:
		if errorMsg, exists := miningData["error"]; exists {
			r.addLineToBuffer(fmt.Sprintf("Error: %v", errorMsg))
		}
	default:
		r.addLineToBuffer(r.config.T("mining.not_detected"))
	}
	r.addEmptyLineToBuffer()
}

func (r *Renderer) renderFooterToBuffer() {
	r.addLineToBuffer(fmt.Sprintf("\033[90mPress 'q' to quit | Auto-refresh every %d seconds\033[0m", 
		r.config.RefreshRate))
}

func (r *Renderer) addHeaderToBuffer(text string) {
	r.screenBuffer = append(r.screenBuffer, fmt.Sprintf("\033[1;33m%s\033[0m", text))
}

func (r *Renderer) addLineToBuffer(text string) {
	r.screenBuffer = append(r.screenBuffer, text)
}

func (r *Renderer) addEmptyLineToBuffer() {
	r.screenBuffer = append(r.screenBuffer, "")
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