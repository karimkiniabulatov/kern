package ui

import (
	"fmt"
	"strconv"
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
		row = t.renderLogo(row)
	}

	// Render only enabled modules
	if t.config.ShowDisk {
		if diskData, exists := data["disk"]; exists {
			row = t.renderDisk(row, diskData)
		}
	}

	if t.config.ShowMem {
		if memData, exists := data["mem"]; exists {
			row = t.renderMemory(row, memData)
		}
	}

	if t.config.ShowNet {
		if netData, exists := data["net"]; exists {
			row = t.renderNetwork(row, netData)
		}
	}

	if t.config.ShowCPU {
		if cpuData, exists := data["cpu"]; exists {
			row = t.renderCPU(row, cpuData)
		}
	}

	// GPU monitoring
	if t.config.ShowGPU {
		if gpuData, exists := data["gpu"]; exists {
			row = t.renderGPU(row, gpuData)
		}
	}

	// AI training monitoring  
	if t.config.ShowAI {
		if aiData, exists := data["ai"]; exists {
			row = t.renderAI(row, aiData)
		}
	}

	// Mining monitoring
	if t.config.ShowMining {
		if miningData, exists := data["mining"]; exists {
			row = t.renderMining(row, miningData)
		}
	}

	// Audio monitoring
	if t.config.ShowAudio {
		if audioData, exists := data["audio"]; exists {
			row = t.renderAudio(row, audioData)
		}
	}

	// Video monitoring
	if t.config.ShowVideo {
		if videoData, exists := data["video"]; exists {
			row = t.renderVideo(row, videoData)
		}
	}

	t.renderFooter(row)
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

func (t *TUI) renderLogo(startRow int) int {
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
		t.printCentered(startRow+i, line, cyan)
	}

	return startRow + len(logo) + 1
}

func (t *TUI) renderCPU(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("cpu.title"))

	if cpuInfo, ok := data.(*cpu.CPUInfo); ok {
		row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("cpu.model"), cpuInfo.Model), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		row = t.printSimple(row, fmt.Sprintf("%s: %d %s, %d %s",
			t.config.T("cpu.cores"), cpuInfo.Cores, t.config.T("cpu.cores"),
			cpuInfo.Threads, tconfig.T("cpu.threads")), tcell.StyleDefault.Foreground(tcell.ColorAqua))

		// Usage with graph on new line
		usageGraph := t.createSolidGraph(cpuInfo.Usage)
		row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", t.config.T("cpu.usage"), cpuInfo.Usage, usageGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))

		row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("cpu.frequency"), cpuInfo.Frequency), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		row = t.printSimple(row, fmt.Sprintf("%s: %.2f, %.2f, %.2f",
			t.config.T("cpu.load_average"), cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15), tcell.StyleDefault.Foreground(tcell.ColorAqua))

		if t.config.DetailedCPU && len(cpuInfo.CoreUsage) > 0 {
			row = t.printSimple(row, fmt.Sprintf("%s:", t.config.T("cpu.core_usage")), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			
			// Группируем ядра для компактного отображения (по 3 в строку)
			coresPerLine := 3
			for i := 0; i < len(cpuInfo.CoreUsage); i += coresPerLine {
				line := "  "
				for j := 0; j < coresPerLine && i+j < len(cpuInfo.CoreUsage); j++ {
					coreIdx := i + j
					usage := cpuInfo.CoreUsage[coreIdx]
					
					// Новый формат: "01: 000.1%" с фиксированной шириной
					coreGraph := t.createSolidGraph(usage)
					line += fmt.Sprintf(" %02d: %05.1f%%%s", coreIdx+1, usage, coreGraph)
					
					if j < coresPerLine-1 && i+j+1 < len(cpuInfo.CoreUsage) {
						line += " |"
					}
				}
				row = t.printSimple(row, line, tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}
		}
	}
	return row + 1
}

func (t *TUI) renderMemory(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("memory.title"))

	if memInfo, ok := data.(*mem.MemoryInfo); ok {
		// RAM usage with detailed information
		ramGraph := t.createSolidGraph(memInfo.UsagePercent)
		row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%.1f%%) %s",
			t.config.T("memory.ram"), memInfo.Used, memInfo.Total, 
			memInfo.UsagePercent, ramGraph), tcell.StyleDefault.Foreground(tcell.ColorGreen))

		row = t.printSimple(row, fmt.Sprintf("%s: %s | %s: %s", 
			t.config.T("common.available"), memInfo.Available,
			t.config.T("common.free"), memInfo.Free), tcell.StyleDefault.Foreground(tcell.ColorAqua))

		if memInfo.SwapTotal != "0B" && memInfo.SwapTotal != "" && memInfo.SwapTotal != "0" {
			// Swap usage with graph
			swapGraph := t.createSolidGraph(memInfo.SwapUsagePercent)
			row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%.1f%%) %s",
				t.config.T("memory.swap"), memInfo.SwapUsed, memInfo.SwapTotal, 
				memInfo.SwapUsagePercent, swapGraph), tcell.StyleDefault.Foreground(tcell.ColorFuchsia))
		}
	}
	return row + 1
}

func (t *TUI) renderDisk(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("disk.title"))

	if disks, ok := data.([]disk.DiskInfo); ok {
		count := 0
		for _, d := range disks {
			if strings.HasPrefix(d.Filesystem, "/dev/") && count < 3 {
				devType := t.getDeviceType(d.Filesystem, d.MountedOn)
				mountPoint := d.MountedOn
				if mountPoint == "/" {
					mountPoint = "ROOT"
				}

				row = t.printSimple(row, fmt.Sprintf("%s: %s (%s)", t.config.T("disk.filesystem"), d.Filesystem, devType), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("disk.mounted"), mountPoint), tcell.StyleDefault.Foreground(tcell.ColorAqua))

				// Disk usage with graph
				diskGraph := t.createSolidGraph(d.UsePercent)
				row = t.printSimple(row, fmt.Sprintf("%s: %s / %s %.1f%% %s",
					t.config.T("disk.usage"), d.Used, d.Size, d.UsePercent, diskGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))

				count++
				if count < 3 {
					row++
				}
			}
		}
	}
	return row + 1
}

func (t *TUI) renderNetwork(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("network.title"))

	if networks, ok := data.([]net.NetworkInfo); ok {
		count := 0
		for _, netInfo := range networks {
			if netInfo.Status == "UP" && netInfo.Interface != "lo" && count < 2 {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("network.interface"), netInfo.Interface), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				row = t.printSimple(row, fmt.Sprintf("%s: %s", "IP Address", netInfo.IPAddress), tcell.StyleDefault.Foreground(tcell.ColorAqua))
				row = t.printSimple(row, fmt.Sprintf("%s: %s", "MAC Address", netInfo.MACAddress), tcell.StyleDefault.Foreground(tcell.ColorAqua))

				// Activity with graph
				activityGraph := t.createSolidGraph(netInfo.ActivityPercent)
				row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", "Activity", netInfo.ActivityPercent, activityGraph), tcell.StyleDefault.Foreground(tcell.ColorFuchsia))

				row = t.printSimple(row, fmt.Sprintf("%s: %s / %s↑", "Speed", netInfo.RXSpeed, netInfo.TXSpeed), tcell.StyleDefault.Foreground(tcell.ColorAqua))

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
func (t *TUI) renderGPU(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("gpu.title"))

	// Handle both *gpu.GPUInfo and error cases
	switch gpuData := data.(type) {
	case *gpu.GPUInfo:
		row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("gpu.model"), gpuData.Model), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		
		if gpuData.DriverVersion != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("gpu.driver"), gpuData.DriverVersion), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		}
		
		if gpuData.GPUTemp > 0 {
			// Temperature with graph
			tempGraph := t.createSolidGraph(gpuData.GPUTemp)
			row = t.printSimple(row, fmt.Sprintf("%s: %.1fC %s", 
				t.config.T("gpu.temperature"), gpuData.GPUTemp, tempGraph), tcell.StyleDefault.Foreground(tcell.ColorRed))
		}

		if gpuData.Utilization > 0 {
			// Utilization with graph
			utilGraph := t.createSolidGraph(gpuData.Utilization)
			row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", 
				t.config.T("gpu.utilization"), gpuData.Utilization, utilGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
		}

		if gpuData.MemoryUsed != "" && gpuData.MemoryTotal != "" {
			// Добавляем гистограмму для использования памяти
			usedMB := extractMemoryMB(gpuData.MemoryUsed)
			totalMB := extractMemoryMB(gpuData.MemoryTotal)
			if usedMB > 0 && totalMB > 0 {
				memoryPercent := float64(usedMB) / float64(totalMB) * 100
				memoryGraph := t.createSolidGraph(memoryPercent)
				row = t.printSimple(row, fmt.Sprintf("%s: %s / %s (%.1f%%) %s", 
					t.config.T("gpu.memory"), gpuData.MemoryUsed, gpuData.MemoryTotal, 
					memoryPercent, memoryGraph), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			} else {
				row = t.printSimple(row, fmt.Sprintf("%s: %s / %s", 
					t.config.T("gpu.memory"), gpuData.MemoryUsed, gpuData.MemoryTotal), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
		}
		
		if gpuData.PowerDraw != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s", 
				t.config.T("gpu.power"), gpuData.PowerDraw), tcell.StyleDefault.Foreground(tcell.ColorYellow))
		}
			
		if gpuData.ClockCore != "" && gpuData.ClockMemory != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s | %s: %s", 
				t.config.T("gpu.clock_core"), gpuData.ClockCore, 
				t.config.T("gpu.clock_memory"), gpuData.ClockMemory), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		}
		
	case map[string]interface{}:
		// Handle error case
		if errorMsg, exists := gpuData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.printSimple(row, "No GPU data available", tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

// AI training rendering function
func (t *TUI) renderAI(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("ai.title"))

	switch aiData := data.(type) {
	case *ai.AIInfo:
		if aiData.ProcessCount > 0 {
			if aiData.Framework != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("ai.framework"), aiData.Framework), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
			row = t.printSimple(row, fmt.Sprintf("%s: %d", t.config.T("ai.processes"), aiData.ProcessCount), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			
			if aiData.VRAMUsage != "" && aiData.VRAMTotal != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s / %s", 
					t.config.T("ai.vram"), aiData.VRAMUsage, aiData.VRAMTotal), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if aiData.ModelName != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("ai.model"), aiData.ModelName), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if aiData.BatchSize > 0 {
				row = t.printSimple(row, fmt.Sprintf("%s: %d | %s: %.1f samples/sec", 
					t.config.T("ai.batch_size"), aiData.BatchSize,
					t.config.T("ai.throughput"), aiData.Throughput), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}

			if aiData.Epoch > 0 {
				row = t.printSimple(row, fmt.Sprintf("%s: %d | %s: %.3f | %s: %.1f%%", 
					t.config.T("ai.epoch"), aiData.Epoch,
					t.config.T("ai.loss"), aiData.Loss,
					t.config.T("ai.accuracy"), aiData.Accuracy*100), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}

			if aiData.TrainingTime != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("ai.training_time"), aiData.TrainingTime), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
		} else {
			row = t.printSimple(row, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	case map[string]interface{}:
		if errorMsg, exists := aiData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.printSimple(row, t.config.T("ai.no_training"), tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

// Mining rendering function
func (t *TUI) renderMining(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("mining.title"))

	switch miningData := data.(type) {
	case *mining.MiningInfo:
		if miningData.Algorithm != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("mining.algorithm"), miningData.Algorithm), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			
			if miningData.Hashrate != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("mining.hashrate"), miningData.Hashrate), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}

			if miningData.SharesValid > 0 {
				row = t.printSimple(row, fmt.Sprintf("%s: %d valid, %d invalid", 
					t.config.T("mining.shares"), miningData.SharesValid, miningData.SharesInvalid), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			}

			if miningData.Temperature > 0 {
				// Temperature with graph
				tempGraph := t.createSolidGraph(miningData.Temperature)
				row = t.printSimple(row, fmt.Sprintf("%s: %.1fC %s", 
					t.config.T("mining.temperature"), miningData.Temperature, tempGraph), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}

			if miningData.PowerConsumption != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.power"), miningData.PowerConsumption), tcell.StyleDefault.Foreground(tcell.ColorYellow))
			}

			if miningData.Efficiency != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.efficiency"), miningData.Efficiency), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if miningData.Uptime != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.uptime"), miningData.Uptime), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if miningData.Pool != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.pool"), miningData.Pool), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}

			if miningData.Revenue24h != "" {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", 
					t.config.T("mining.revenue_24h"), miningData.Revenue24h), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			}
		} else {
			row = t.printSimple(row, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	case map[string]interface{}:
		if errorMsg, exists := miningData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("Error: %s", errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.printSimple(row, t.config.T("mining.not_detected"), tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

// Audio rendering function
func (t *TUI) renderAudio(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("audio.title"))

	switch audioData := data.(type) {
	case *audio.AudioInfo:
		// Input devices
		if len(audioData.InputDevices) > 0 {
			row = t.printSimple(row, t.config.T("audio.input_devices"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			for i, device := range audioData.InputDevices {
				if i >= 2 { // Limit to 2 devices for space
					break
				}
				row = t.printSimple(row, fmt.Sprintf("  %s (%s)", device.Name, device.Status), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
		}

		// Output devices
		if len(audioData.OutputDevices) > 0 {
			row = t.printSimple(row, t.config.T("audio.output_devices"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			for i, device := range audioData.OutputDevices {
				if i >= 2 { // Limit to 2 devices for space
					break
				}
				row = t.printSimple(row, fmt.Sprintf("  %s (%s)", device.Name, device.Status), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
		}

		// Audio levels
		if audioData.InputLevel != 0 || audioData.OutputLevel != 0 {
			row = t.printSimple(row, t.config.T("audio.levels"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			
			if audioData.InputLevel != 0 {
				// Input level with graph
				inputLevelPercent := (audioData.InputLevel + 96) / 96 * 100
				inputGraph := t.createSolidGraph(inputLevelPercent)
				row = t.printSimple(row, fmt.Sprintf("  %s: %.1f dB %s", 
					t.config.T("audio.input"), audioData.InputLevel, inputGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}
			
			if audioData.OutputLevel != 0 {
				// Output level with graph
				outputLevelPercent := (audioData.OutputLevel + 96) / 96 * 100
				outputGraph := t.createSolidGraph(outputLevelPercent)
				row = t.printSimple(row, fmt.Sprintf("  %s: %.1f dB %s", 
					t.config.T("audio.output"), audioData.OutputLevel, outputGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
			}
		}

		// Active streams
		if len(audioData.ActiveStreams) > 0 {
			row = t.printSimple(row, t.config.T("audio.active_streams"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			for i, stream := range audioData.ActiveStreams {
				if i >= 3 { // Limit to 3 streams for space
					break
				}
				streamType := t.config.T("audio.playback")
				if stream.Type == "capture" {
					streamType = t.config.T("audio.capture")
				}
				row = t.printSimple(row, fmt.Sprintf("  %s: %s", streamType, stream.Process), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			}
		}

		// Audio format info
		if audioData.SampleRate != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s, %s, %s", 
				t.config.T("audio.format"), audioData.SampleRate, audioData.BitDepth, audioData.Latency), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		}

		// VU Meter simulation
		if audioData.PeakLevel > 0 {
			row = t.printSimple(row, t.config.T("audio.vu_meter"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			// Peak level with graph
			vuGraph := t.createSolidGraph(audioData.PeakLevel)
			row = t.printSimple(row, fmt.Sprintf("  %s: %.1f%% %s", 
				t.config.T("audio.peak"), audioData.PeakLevel, vuGraph), tcell.StyleDefault.Foreground(tcell.ColorRed))
		}

	case map[string]interface{}:
		if errorMsg, exists := audioData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("common.error"), errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.printSimple(row, t.config.T("audio.no_audio"), tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

// Video rendering function
func (t *TUI) renderVideo(startRow int, data interface{}) int {
	row := t.renderHeader(startRow, t.config.T("video.title"))

	switch videoData := data.(type) {
	case *video.VideoInfo:
		// Video devices
		if len(videoData.VideoDevices) > 0 {
			row = t.printSimple(row, t.config.T("video.devices"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			for i, device := range videoData.VideoDevices {
				if i >= 2 { // Limit to 2 devices for space
					break
				}
				row = t.printSimple(row, fmt.Sprintf("  %s (%s)", device.Name, device.Status), tcell.StyleDefault.Foreground(tcell.ColorAqua))
			}
		}

		// Active streams
		if len(videoData.ActiveStreams) > 0 {
			row = t.printSimple(row, t.config.T("video.active_streams"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			for i, stream := range videoData.ActiveStreams {
				if i >= 3 { // Limit to 3 streams for space
					break
				}
				row = t.printSimple(row, fmt.Sprintf("  %s: %s", stream.Type, stream.Process), tcell.StyleDefault.Foreground(tcell.ColorGreen))
			}
		}

		// GPU encoders
		if len(videoData.GPUEncoders) > 0 {
			row = t.printSimple(row, t.config.T("video.encoders"), tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true))
			for i, encoder := range videoData.GPUEncoders {
				if i >= 2 { // Limit to 2 encoders for space
					break
				}
				status := t.config.T("video.inactive")
				if encoder.Active {
					status = t.config.T("video.active")
				}
				row = t.printSimple(row, fmt.Sprintf("  %s (%s) - %s", encoder.Name, encoder.Type, status), tcell.StyleDefault.Foreground(tcell.ColorYellow))
			}
		}

		// Encoding status
		if videoData.EncodingStatus != "" {
			statusColor := tcell.ColorGray
			if videoData.EncodingStatus == "active" || videoData.EncodingStatus == "encoding" {
				statusColor = tcell.ColorGreen
			}
			row = t.printSimple(row, fmt.Sprintf("%s: %s", 
				t.config.T("video.encoding_status"), videoData.EncodingStatus), tcell.StyleDefault.Foreground(statusColor))
		}

		// Video metrics
		if videoData.Bitrate != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s", 
				t.config.T("video.bitrate"), videoData.Bitrate), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		}

		if videoData.Resolution != "" {
			row = t.printSimple(row, fmt.Sprintf("%s: %s", 
				t.config.T("video.resolution"), videoData.Resolution), tcell.StyleDefault.Foreground(tcell.ColorAqua))
		}

		// GPU utilization for video
		if videoData.GPUUtilization > 0 {
			// GPU usage with graph
			gpuGraph := t.createSolidGraph(videoData.GPUUtilization)
			row = t.printSimple(row, fmt.Sprintf("%s: %.1f%% %s", 
				t.config.T("video.gpu_usage"), videoData.GPUUtilization, gpuGraph), tcell.StyleDefault.Foreground(tcell.ColorLightCoral))
		}

		// ГИСТОГРАММА ЧАСТОТЫ КАДРОВ (если есть активные потоки)
		if videoData.Framerate > 0 {
			// Нормализуем частоту кадров для гистограммы (предполагаем макс 60 fps)
			fpsPercent := (videoData.Framerate / 60.0) * 100
			if fpsPercent > 100 {
				fpsPercent = 100
			}
			fpsGraph := t.createSolidGraph(fpsPercent)
			row = t.printSimple(row, fmt.Sprintf("%s: %.1f fps %s", 
				t.config.T("video.frame_rate"), videoData.Framerate, fpsGraph), 
				tcell.StyleDefault.Foreground(tcell.ColorGreen))
		}

	case map[string]interface{}:
		if errorMsg, exists := videoData["error"]; exists {
			if errorStr, ok := errorMsg.(string); ok {
				row = t.printSimple(row, fmt.Sprintf("%s: %s", t.config.T("common.error"), errorStr), tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		}
	default:
		row = t.printSimple(row, t.config.T("video.no_video"), tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	return row + 1
}

func (t *TUI) renderHeader(startRow int, title string) int {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true).Background(tcell.ColorDarkBlue)
	
	// Create header line with padding
	header := fmt.Sprintf(" %s ", title)
	t.printSimple(startRow, header, style)
	
	return startRow + 1
}

func (t *TUI) renderFooter(startRow int) {
	if startRow >= t.height-1 {
		return
	}

	footerText := fmt.Sprintf("Press 'q' or Ctrl+C to exit | Refresh: %ds | kern v1.2.1", t.config.RefreshRate)
	style := tcell.StyleDefault.Foreground(tcell.ColorGray)
	
	// Обеспечиваем что футер внизу и выровнен по левому краю
	footerRow := t.height - 1
	
	// Обрезаем текст если он слишком длинный
	if len(footerText) > t.width {
		footerText = footerText[:t.width-3] + "..."
	}
	
	// Выводим с левого края
	t.printSimple(footerRow, footerText, style)
}

// Упрощенная функция вывода - всегда с начала строки
func (t *TUI) printSimple(row int, text string, style tcell.Style) int {
	if row < 0 || row >= t.height {
		return row
	}

	// Обрабатываем текст с учетом ширины символов
	x := 0
	for _, ch := range text {
		if x >= t.width {
			break
		}
		t.screen.SetContent(x, row, ch, nil, style)
		x++
	}
	return row + 1
}

// Функция для центрированного вывода
func (t *TUI) printCentered(row int, text string, style tcell.Style) {
	if row < 0 {
		return
	}

	if row >= t.height {
		return
	}

	x := (t.width - len(text)) / 2
	if x < 0 {
		x = 0
	}

	for i, ch := range text {
		if x+i >= t.width {
			break
		}
		t.screen.SetContent(x+i, row, ch, nil, style)
	}
}

// Сплошная гистограмма с фиксированным количеством сегментов (20 сегментов = 5% на элемент)
func (t *TUI) createSolidGraph(percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	segments := 20
	filled := int((percent / 100) * float64(segments))
	if filled > segments {
		filled = segments
	}

	// Используем более совместимые символы для гистограммы
	// █ и ░ могут иметь проблемы с отображением, используем блоки
	graph := strings.Repeat("█", filled)  // Полный блок
	empty := strings.Repeat("░", segments-filled) // Светлый блок
	
	// Альтернатива если есть проблемы с юникодом:
	// graph := strings.Repeat("■", filled)
	// empty := strings.Repeat("□", segments-filled)
	
	// Или ASCII версия:
	// graph := strings.Repeat("#", filled)
	// empty := strings.Repeat(".", segments-filled)

	return graph + empty
}

// Старая функция для совместимости
func (t *TUI) createCompactGraph(percent float64) string {
	return t.createSolidGraph(percent)
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

// Вспомогательная функция для извлечения числового значения памяти из строки
func extractMemoryMB(memoryStr string) int {
	// Пример: "8192 MB" -> 8192
	parts := strings.Fields(memoryStr)
	if len(parts) >= 2 {
		if value, err := strconv.Atoi(parts[0]); err == nil {
			return value
		}
	}
	return 0
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