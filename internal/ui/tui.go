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

	// GPU monitoring
	if t.config.ShowGPU {
		if gpuData, exists := data["gpu"]; exists {
			row = t.renderGPU(row, width, gpuData)
		}
	}

	// AI training monitoring  
	if t.config.ShowAI {
		if aiData, exists := data["ai"]; exists {
			row = t.renderAI(row, width, aiData)
		}
	}

	// Mining monitoring
	if t.config.ShowMining {
		if miningData, exists := data["mining"]; exists {
			row = t.renderMining(row, width, miningData)
		}
	}

	// Audio monitoring
	if t.config.ShowAudio {
		if audioData, exists := data["audio"]; exists {
			row = t.renderAudio(row, width, audioData)
		}
	}

	// Video monitoring
	if t.config.ShowVideo {
		if videoData, exists := data["video"]; exists {
			row = t.renderVideo(row, width, videoData)
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

// Mining rendering function
func (t *TUI) renderMining(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("mining.title"), width)

	switch miningData := data.(type) {
	case *mining.MiningInfo:
		if miningData.Algorithm != "" {
			row = t.printLine(row, 0, fmt.Sprintf("%s: %s (%s)", 
				t.config.T("mining.algorithm"), miningData.Algorithm, miningData.Currency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			
			if miningData.Hashrate != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.hashrate"), miningData.Hashrate), tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}

			if miningData.SharesValid > 0 {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %d / %d (%.1f%%)", 
					t.config.T("mining.shares"), miningData.SharesValid, 
					miningData.SharesValid+miningData.SharesInvalid,
					float64(miningData.SharesValid)/float64(miningData.SharesValid+miningData.SharesInvalid)*100), 
					tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}

			if miningData.PowerConsumption != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.power"), miningData.PowerConsumption), tcell.StyleDefault.Foreground(tcell.ColorYellow), width)
			}

			if miningData.Temperature > 0 {
				tempGraph := t.createCompactGraph(miningData.Temperature, 10)
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s %.1fC", 
					t.config.T("mining.temperature"), tempGraph, miningData.Temperature), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
			}

			if miningData.Efficiency != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.efficiency"), miningData.Efficiency), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}

			if miningData.Uptime != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.uptime"), miningData.Uptime), tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}

			if miningData.Revenue24h != "" {
				row = t.printLine(row, 0, fmt.Sprintf("%s: %s", 
					t.config.T("mining.revenue_24h"), miningData.Revenue24h), tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
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

// Audio rendering function
func (t *TUI) renderAudio(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, "Audio Streams", width)

	switch audioData := data.(type) {
	case *audio.AudioInfo:
		// Input devices
		if len(audioData.InputDevices) > 0 {
			row = t.printLine(row, 0, "Input Devices:", tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, device := range audioData.InputDevices {
				if i >= 2 { // Limit to 2 devices
					break
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s [%s]", device.Name, device.Status), 
					tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		}

		// Output devices
		if len(audioData.OutputDevices) > 0 {
			row = t.printLine(row, 0, "Output Devices:", tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true), width)
			for i, device := range audioData.OutputDevices {
				if i >= 2 { // Limit to 2 devices
					break
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s [%s]", device.Name, device.Status), 
					tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
			}
		}

		// Active streams
		if len(audioData.ActiveStreams) > 0 {
			row = t.printLine(row, 0, "Active Streams:", tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true), width)
			for i, stream := range audioData.ActiveStreams {
				if i >= 3 { // Limit to 3 streams
					break
				}
				streamType := "Playback"
				if stream.Type == "capture" {
					streamType = "Capture"
				}
				row = t.printLine(row, 2, fmt.Sprintf("%s: %s (%s)", streamType, stream.Process, stream.Format), 
					tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
			}
		} else {
			row = t.printLine(row, 0, "No active audio streams", tcell.StyleDefault.Foreground(tcell.ColorGray), width)
		}

		// Audio levels
		if audioData.InputLevel != 0 || audioData.OutputLevel != 0 {
			row = t.printLine(row, 0, "Audio Levels:", tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true), width)
			
			if audioData.InputLevel != 0 {
				inputGraph := t.createCompactGraph((audioData.InputLevel + 96) / 96 * 100, 10)
				row = t.printLine(row, 2, fmt.Sprintf("Input: %s %.1f dB", inputGraph, audioData.InputLevel), 
					tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}
			
			if audioData.OutputLevel != 0 {
				outputGraph := t.createCompactGraph((audioData.OutputLevel + 96) / 96 * 100, 10)
				row = t.printLine(row, 2, fmt.Sprintf("Output: %s %.1f dB", outputGraph, audioData.OutputLevel), 
					tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
			}
		}

		// Audio metrics
		row = t.printLine(row, 0, fmt.Sprintf("Sample Rate: %s | Bit Depth: %s | Latency: %s", 
			audioData.SampleRate, audioData.BitDepth, audioData.Latency), 
			tcell.StyleDefault.Foreground(tcell.ColorAqua), width)

	case map[string]string:
		// Handle error case
		if errorMsg, exists := audioData["error"]; exists {
			row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorMsg), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}
	default:
		row = t.printLine(row, 0, "No audio data available", tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

// Video rendering function - УПРОЩЕННАЯ ВЕРСИЯ
func (t *TUI) renderVideo(startRow int, width int, data interface{}) int {
	row := t.renderHeader(startRow, "Video Streams", width)

	switch videoData := data.(type) {
	case *video.VideoInfo:
		// Базовая информация о видео
		row = t.printLine(row, 0, "Video Monitoring Active", tcell.StyleDefault.Foreground(tcell.ColorGreen), width)
		
		// Простая информация о доступных устройствах
		if len(videoData.VideoDevices) > 0 {
			row = t.printLine(row, 0, fmt.Sprintf("Devices: %d available", len(videoData.VideoDevices)), 
				tcell.StyleDefault.Foreground(tcell.ColorAqua), width)
		}
		
		// Активные потоки
		if len(videoData.ActiveStreams) > 0 {
			row = t.printLine(row, 0, fmt.Sprintf("Active Streams: %d", len(videoData.ActiveStreams)), 
				tcell.StyleDefault.Foreground(tcell.ColorLightCoral), width)
		} else {
			row = t.printLine(row, 0, "No active video streams", tcell.StyleDefault.Foreground(tcell.ColorGray), width)
		}

		// GPU энкодеры
		if len(videoData.GPUEncoders) > 0 {
			row = t.printLine(row, 0, fmt.Sprintf("GPU Encoders: %d available", len(videoData.GPUEncoders)), 
				tcell.StyleDefault.Foreground(tcell.ColorYellow), width)
		}

	case map[string]string:
		// Handle error case
		if errorMsg, exists := videoData["error"]; exists {
			row = t.printLine(row, 0, fmt.Sprintf("Error: %s", errorMsg), tcell.StyleDefault.Foreground(tcell.ColorRed), width)
		}
	default:
		row = t.printLine(row, 0, "No video data available", tcell.StyleDefault.Foreground(tcell.ColorGray), width)
	}
	return row + 1
}

func (t *TUI) renderHeader(startRow int, title string, width int) int {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true).Background(tcell.ColorDarkBlue)
	t.printCentered(startRow, fmt.Sprintf(" %s ", strings.ToUpper(title)), style, width)
	return startRow + 1
}

func (t *TUI) renderFooter(startRow int, width int, height int) {
	if startRow >= height-1 {
		return
	}

	footerText := "Press 'q' or Ctrl+C to exit | kern v1.2.1"
	style := tcell.StyleDefault.Foreground(tcell.ColorGray)
	t.printCentered(height-1, footerText, style, width)
}

func (t *TUI) printLine(row int, indent int, text string, style tcell.Style, width int) int {
	screenWidth, screenHeight := t.screen.Size()
	if row >= screenHeight-1 {
		return row
	}

	// Handle text that might be longer than screen width
	if len(text) > width-indent {
		text = text[:width-indent-3] + "..."
	}

	for i := 0; i < indent; i++ {
		t.screen.SetContent(i, row, ' ', nil, style)
	}

	for i, ch := range text {
		if indent+i >= width {
			break
		}
		t.screen.SetContent(indent+i, row, ch, nil, style)
	}

	return row + 1
}

func (t *TUI) printCentered(row int, text string, style tcell.Style, width int) {
	screenWidth, screenHeight := t.screen.Size()
	if row < 0 || row >= screenHeight {
		return
	}

	start := (width - len(text)) / 2
	if start < 0 {
		start = 0
	}

	for i, ch := range text {
		if start+i >= width {
			break
		}
		t.screen.SetContent(start+i, row, ch, nil, style)
	}
}

func (t *TUI) createCompactGraph(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100) * float64(width))
	graph := ""

	for i := 0; i < width; i++ {
		if i < filled {
			graph += "█"
		} else {
			graph += "░"
		}
	}

	return graph
}

func (t *TUI) getDeviceType(filesystem string, mountPoint string) string {
	if strings.Contains(filesystem, "nvme") || strings.Contains(filesystem, "ssd") {
		return "SSD"
	} else if strings.Contains(filesystem, "sd") {
		return "HDD"
	} else if mountPoint == "/" {
		return "System"
	}
	return "Storage"
}