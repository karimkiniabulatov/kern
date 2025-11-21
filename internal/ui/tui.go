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
	"github.com/karimkiniabulatov/kern/internal/audio"
	"github.com/karimkiniabulatov/kern/internal/video"
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
		row = t.renderLogo(row, t.width)
	}

	// Render only enabled modules
	if t.config.ShowDisk {
		if diskData, exists := data["disk"]; exists {
			row = t.renderDisk(row, t.width, diskData)
		}
	}

	if t.config.ShowMem {
		if memData, exists := data["mem"]; exists {
			row = t.renderMemory(row, t.width, memData)
		}
	}

	if t.config.ShowNet {
		if netData, exists := data["net"]; exists {
			row = t.renderNetwork(row, t.width, netData)
		}
	}

	if t.config.ShowCPU {
		if cpuData, exists := data["cpu"]; exists {
			row = t.renderCPU(row, t.width, cpuData)
		}
	}

	// GPU monitoring
	if t.config.ShowGPU {
		if gpuData, exists := data["gpu"]; exists {
			row = t.renderGPU(row, t.width, gpuData)
		}
	}

	// AI training monitoring  
	if t.config.ShowAI {
		if aiData, exists := data["ai"]; exists {
			row = t.renderAI(row, t.width, aiData)
		}
	}

	// Mining monitoring
	if t.config.ShowMining {
		if miningData, exists := data["mining"]; exists {
			row = t.renderMining(row, t.width, miningData)
		}
	}

	// Audio monitoring
	if t.config.ShowAudio {
		if audioData, exists := data["audio"]; exists {
			row = t.renderAudio(row, t.width, audioData)
		}
	}

	// Video monitoring
	if t.config.ShowVideo {
		if videoData, exists := data["video"]; exists {
			row = t.renderVideo(row, t.width, videoData)
		}
	}

	t.renderFooter(row, t.width, t.height)
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
		" kern v1.2.1 - System Monitoring Tool",
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

		graph := t.createCompactGraph(cpuInfo.Usage, 10)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f%%",
			t.config.T("cpu.usage"), graph, cpuInfo.Usage), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)

		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("cpu.frequency"), cpuInfo.Frequency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %.2f, %.2f, %.2f",
			t.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

		if t.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
			row = t.printLine(row, 0, fmt.Sprintf("%s:", t.config.T("cpu.core_usage")), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			for i, usage := range cpuInfo.CoreUsage {
				coreGraph := t.createCompactGraph(usage, 10)
				// Форматируем номер ядра с ведущим нулем для ядер 0-9
				coreNumber := fmt.Sprintf("%02d", i+1)
				row = t.printLine(row, 2, fmt.Sprintf("%s %s: %s %.1f%%",
					t.config.T("cpu.core"), coreNumber, coreGraph, usage), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}
		}
	}
	return row + 1
}

func (t *TUI) renderMemory(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("memory.title"), width)

	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		ramGraph := t.createCompactGraph(memInfo.UsagePercent, 10)
		row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s %s %.1f%%",
			t.config.T("memory.ram"), memInfo.Used, memInfo.Total, ramGraph, memInfo.UsagePercent), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)

		row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("common.available"), memInfo.Available), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" {
			swapGraph := t.createCompactGraph(memInfo.SwapUsagePercent, 10)
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

				diskGraph := t.createCompactGraph(d.UsePercent, 10)
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

				activityGraph := t.createCompactGraph(netInfo.ActivityPercent, 10)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f%%",
					"Activity", activityGraph, netInfo.ActivityPercent), tcell.StyleDefault.Foreground(tcell.ColorFuchsia), width)

				row = t.printLine(row, 0, fmt.Sprintf("%s: %s / %s↑",
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

// GPU rendering function
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
			tempGraph := t.createCompactGraph(gpuData.GPUTemp, 10)
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1fC", 
				t.config.T("gpu.temperature"), tempGraph, gpuData.GPUTemp), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}

		if gpuData.Utilization > 0 {
			utilGraph := t.createCompactGraph(gpuData.Utilization, 10)
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
		
	case map[string]interface{}:
		// Handle error case
		if errorMsg, exists := gpuData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}
		}
	default:
		row = t.printLine(row, 0, "No GPU data available", tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

// AI training rendering function
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
					t.config.T("ai.accuracy"), aiData.Accuracy*100), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}

			if aiData.TrainingTime != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("ai.training_time"), aiData.TrainingTime), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		} else {
			row = t.printLine(row, 0, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
		}
	case map[string]interface{}:
		if errorMsg, exists := aiData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}
		}
	default:
		row = t.printLine(row, 0, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

// Mining rendering function
func (t *TUI) renderMining(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("mining.title"), width)

	switch miningData := data.(type) {
	case *mining.MiningInfo:
		if miningData.Algorithm != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("mining.algorithm"), miningData.Algorithm), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			
			if miningData.Hashrate != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("mining.hashrate"), miningData.Hashrate), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}

			if miningData.SharesValid > 0 {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %d valid, %d invalid", 
					t.config.T("mining.shares"), miningData.SharesValid, miningData.SharesInvalid), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}

			if miningData.Temperature > 0 {
				tempGraph := t.createCompactGraph(miningData.Temperature, 10)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1fC", 
					t.config.T("mining.temperature"), tempGraph, miningData.Temperature), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}

			if miningData.PowerConsumption != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.power"), miningData.PowerConsumption), tcell.StyleDefault.Foreground(tcell.ColorYellow), width)
			}

			if miningData.Efficiency != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.efficiency"), miningData.Efficiency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}

			if miningData.Uptime != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.uptime"), miningData.Uptime), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}

			if miningData.Pool != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.pool"), miningData.Pool), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}

			if miningData.Revenue24h != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.revenue_24h"), miningData.Revenue24h), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}
		} else {
			row = t.printLine(row, 0, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
		}
	case map[string]interface{}:
		if errorMsg, exists := miningData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}
		}
	default:
		row = t.printLine(row, 0, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

// Audio rendering function
func (t *TUI) renderAudio(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("audio.title"), width)

	switch audioData := data.(type) {
	case *audio.AudioInfo:
		// Input devices
		if len(audioData.InputDevices) > 0 {
			row = t.printLine(row, 0, t.config.T("audio.input_devices"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, device := range audioData.InputDevices {
				if i >= 2 { // Limit to 2 devices for space
					break
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s (%s)", device.Name, device.Status), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		}

		// Output devices
		if len(audioData.OutputDevices) > 0 {
			row = t.printLine(row, 0, t.config.T("audio.output_devices"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, device := range audioData.OutputDevices {
				if i >= 2 { // Limit to 2 devices for space
					break
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s (%s)", device.Name, device.Status), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		}

		// Audio levels
		if audioData.InputLevel != 0 || audioData.OutputLevel != 0 {
			row = t.printLine(row, 0, t.config.T("audio.levels"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			
			if audioData.InputLevel != 0 {
				inputGraph := t.createCompactGraph((audioData.InputLevel+96)/96*100, 10)
				row = t.printLine(row, 2, fmt.Sprintf("%s: %s %.1f dB", 
					t.config.T("audio.input"), inputGraph, audioData.InputLevel), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}
			
			if audioData.OutputLevel != 0 {
				outputGraph := t.createCompactGraph((audioData.OutputLevel+96)/96*100, 10)
				row = t.printLine(row, 2, fmt.Sprintf("%s: %s %.1f dB", 
					t.config.T("audio.output"), outputGraph, audioData.OutputLevel), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}
		}

		// Active streams
		if len(audioData.ActiveStreams) > 0 {
			row = t.printLine(row, 0, t.config.T("audio.active_streams"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, stream := range audioData.ActiveStreams {
				if i >= 3 { // Limit to 3 streams for space
					break
				}
				streamType := t.config.T("audio.playback")
				if stream.Type == "capture" {
					streamType = t.config.T("audio.capture")
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s: %s", streamType, stream.Process), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}
		}

		// Audio format info
		if audioData.SampleRate != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s, %s, %s", 
				t.config.T("audio.format"), audioData.SampleRate, audioData.BitDepth, audioData.Latency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		}

		// VU Meter simulation
		if audioData.PeakLevel > 0 {
			row = t.printLine(row, 0, t.config.T("audio.vu_meter"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			vuGraph := t.createCompactGraph(audioData.PeakLevel, 20)
			row = t.printLine(row, 2, fmt.Sprintf("%s: %s", t.config.T("audio.peak"), vuGraph), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}

	case map[string]interface{}:
		if errorMsg, exists := audioData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("common.error"), errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}
		}
	default:
		row = t.printLine(row, 0, t.config.T("audio.no_audio"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

// Video rendering function
func (t *TUI) renderVideo(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("video.title"), width)

	switch videoData := data.(type) {
	case *video.VideoInfo:
		// Video devices
		if len(videoData.VideoDevices) > 0 {
			row = t.printLine(row, 0, t.config.T("video.devices"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, device := range videoData.VideoDevices {
				if i >= 2 { // Limit to 2 devices for space
					break
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s (%s)", device.Name, device.Status), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		}

		// Active streams
		if len(videoData.ActiveStreams) > 0 {
			row = t.printLine(row, 0, t.config.T("video.active_streams"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, stream := range videoData.ActiveStreams {
				if i >= 3 { // Limit to 3 streams for space
					break
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s: %s", stream.Type, stream.Process), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}
		}

		// GPU encoders
		if len(videoData.GPUEncoders) > 0 {
			row = t.printLine(row, 0, t.config.T("video.encoders"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, encoder := range videoData.GPUEncoders {
				if i >= 2 { // Limit to 2 encoders for space
					break
				}
				status := t.config.T("video.inactive")
				if encoder.Active {
					status = t.config.T("video.active")
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s (%s) - %s", encoder.Name, encoder.Type, status), tcell.StyleDefault.Foreground(tcell.ColorYellow), width)
			}
		}

		// Encoding status
		if videoData.EncodingStatus != "" {
			statusColor := tcell.ColorGray
			if videoData.EncodingStatus == "active" || videoData.EncodingStatus == "encoding" {
				statusColor = tcell.ColorGreen
			}
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
				t.config.T("video.encoding_status"), videoData.EncodingStatus), tcell.StyleDefault.Foreground(statusColor), width)
		}

		// Video metrics
		if videoData.Bitrate != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
				t.config.T("video.bitrate"), videoData.Bitrate), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		}

		if videoData.Resolution != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
				t.config.T("video.resolution"), videoData.Resolution), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		}

		// GPU utilization for video
		if videoData.GPUUtilization > 0 {
			gpuGraph := t.createCompactGraph(videoData.GPUUtilization, 10)
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1f%%", 
				t.config.T("video.gpu_usage"), gpuGraph, videoData.GPUUtilization), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
		}

	case map[string]interface{}:
		if errorMsg, exists := videoData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", t.config.T("common.error"), errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}
		}
	default:
		row = t.printLine(row, 0, t.config.T("video.no_video"), tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

func (t *TUI) renderHeader(startRow int, title string, width int) int {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true).Background(tcell.ColorDarkBlue)
	
	// Create header line with padding
	header := fmt.Sprintf(" %s ", title)
	t.printCentered(startRow, header, style, width)
	
	return startRow + 1
}

func (t *TUI) renderFooter(startRow int, width int, height int) {
	if startRow >= height-1 {
		return
	}

	footerText := fmt.Sprintf("Press 'q' or Ctrl+C to exit | Refresh: %ds | kern v1.2.1", t.config.RefreshRate)
	style := tcell.StyleDefault.Foreground(tcell.ColorGray)
	
	// Ensure footer is at the bottom
	footerRow := height - 1
	t.printCentered(footerRow, footerText, style, width)
}

func (t *TUI) printCentered(row int, text string, style tcell.Style, width int) {
	if row < 0 {
		return
	}

	screenWidth, screenHeight := t.screen.Size()
	if row >= screenHeight {
		return
	}

	x := (width - len(text)) / 2
	if x < 0 {
		x = 0
	}

	for i, ch := range text {
		if x+i >= screenWidth {
			break
		}
		t.screen.SetContent(x+i, row, ch, nil, style)
	}
}

func (t *TUI) printLine(row int, indent int, text string, style tcell.Style, width int) int {
	if row < 0 {
		return row
	}

	screenWidth, screenHeight := t.screen.Size()
	if row >= screenHeight {
		return row
	}

	// Handle text that might be longer than screen width
	if len(text) > width-indent {
		text = text[:width-indent-3] + "..."
	}

	for i, ch := range text {
		if indent+i >= screenWidth {
			break
		}
		t.screen.SetContent(indent+i, row, ch, nil, style)
	}

	return row + 1
}

func (t *TUI) createCompactGraph(percent float64, length int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100) * float64(length))
	if filled > length {
		filled = length
	}

	graph := strings.Repeat("█", filled)
	empty := strings.Repeat("░", length-filled)
	
	return "[" + graph + empty + "]"
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