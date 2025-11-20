package audio

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"math/rand"
)

type AudioInfo struct {
	InputDevices    []AudioDevice
	OutputDevices   []AudioDevice
	ActiveStreams   []AudioStream
	InputLevel      float64  // dB
	OutputLevel     float64  // dB
	SampleRate      string
	BitDepth        string
	Latency         string
	
	// NEW: Для гистограмм
	FrequencyBands  []float64 // Частотные полосы для спектр-анализатора (0-100%)
	PeakLevel       float64   // Пиковый уровень для VU-метра (0-100%)
	RMSLevel        float64   // Среднеквадратичный уровень (0-100%)
}

type AudioDevice struct {
	Name      string
	ID        string
	Type      string // input/output
	Channels  int
	Format    string
	Status    string // active/inactive
}

type AudioStream struct {
	Process     string
	PID         int
	Type        string // capture/playback
	Format      string
	SampleRate  int
	Channels    int
	Duration    string
	Bitrate     string
}

var lastAudioStats map[string]AudioStream

func Summary() (*AudioInfo, error) {
	info := &AudioInfo{
		InputDevices:  []AudioDevice{},
		OutputDevices: []AudioDevice{},
		ActiveStreams: []AudioStream{},
		FrequencyBands: make([]float64, 8), // 8 частотных полос
	}

	// Detect audio devices
	info.detectAudioDevices()
	
	// Detect active audio streams
	info.detectActiveStreams()
	
	// Get audio levels and metrics
	info.getAudioMetrics()
	
	// Generate frequency spectrum data (симуляция для демонстрации)
	info.generateFrequencyData()

	return info, nil
}

func (a *AudioInfo) detectAudioDevices() {
	// Try ALSA first (Linux)
	if output, err := exec.Command("arecord", "-l").Output(); err == nil {
		a.parseALSAInputDevices(string(output))
	}
	
	if output, err := exec.Command("aplay", "-l").Output(); err == nil {
		a.parseALSAOutputDevices(string(output))
	}
	
	// Try PulseAudio
	if output, err := exec.Command("pactl", "list", "sources", "short").Output(); err == nil {
		a.parsePulseAudioInputs(string(output))
	}
	
	if output, err := exec.Command("pactl", "list", "sinks", "short").Output(); err == nil {
		a.parsePulseAudioOutputs(string(output))
	}
	
	// Fallback for systems without specific audio tools
	if len(a.InputDevices) == 0 && len(a.OutputDevices) == 0 {
		a.InputDevices = append(a.InputDevices, AudioDevice{
			Name:   "Default Input",
			Type:   "input",
			Status: "unknown",
		})
		a.OutputDevices = append(a.OutputDevices, AudioDevice{
			Name:   "Default Output", 
			Type:   "output",
			Status: "unknown",
		})
	}
}

func (a *AudioInfo) parseALSAInputDevices(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "card") && strings.Contains(line, "device") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				device := AudioDevice{
					Name:   strings.Join(fields[5:], " "),
					ID:     fmt.Sprintf("%s:%s", fields[1], fields[3]),
					Type:   "input",
					Status: "active",
				}
				a.InputDevices = append(a.InputDevices, device)
			}
		}
	}
}

func (a *AudioInfo) parseALSAOutputDevices(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "card") && strings.Contains(line, "device") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				device := AudioDevice{
					Name:   strings.Join(fields[5:], " "),
					ID:     fmt.Sprintf("%s:%s", fields[1], fields[3]),
					Type:   "output", 
					Status: "active",
				}
				a.OutputDevices = append(a.OutputDevices, device)
			}
		}
	}
}

func (a *AudioInfo) parsePulseAudioInputs(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			device := AudioDevice{
				Name:   fields[1],
				ID:     fields[0],
				Type:   "input",
				Status: "active",
			}
			a.InputDevices = append(a.InputDevices, device)
		}
	}
}

func (a *AudioInfo) parsePulseAudioOutputs(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			device := AudioDevice{
				Name:   fields[1],
				ID:     fields[0],
				Type:   "output",
				Status: "active",
			}
			a.OutputDevices = append(a.OutputDevices, device)
		}
	}
}

func (a *AudioInfo) detectActiveStreams() {
	// Find processes using audio devices
	if output, err := exec.Command("lsof", "+D", "/dev/snd").Output(); err == nil {
		a.parseAudioProcesses(string(output))
	}
	
	// Try PulseAudio streams
	if output, err := exec.Command("pactl", "list", "sink-inputs").Output(); err == nil {
		a.parsePulseAudioStreams(string(output), "playback")
	}
	
	if output, err := exec.Command("pactl", "list", "source-outputs").Output(); err == nil {
		a.parsePulseAudioStreams(string(output), "capture") 
	}
}

func (a *AudioInfo) parseAudioProcesses(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "pcm") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid, _ := strconv.Atoi(fields[1])
				streamType := "playback"
				if strings.Contains(line, "capture") {
					streamType = "capture"
				}
				
				stream := AudioStream{
					Process: fields[0],
					PID:     pid,
					Type:    streamType,
					Format:  "PCM",
				}
				a.ActiveStreams = append(a.ActiveStreams, stream)
			}
		}
	}
}

func (a *AudioInfo) parsePulseAudioStreams(output, streamType string) {
	lines := strings.Split(output, "\n")
	currentStream := AudioStream{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "Sink Input #") || strings.HasPrefix(line, "Source Output #") {
			if currentStream.Process != "" {
				a.ActiveStreams = append(a.ActiveStreams, currentStream)
			}
			currentStream = AudioStream{Type: streamType}
		} else if strings.Contains(line, "application.name =") {
			currentStream.Process = strings.Trim(strings.Split(line, "=")[1], " \"")
		} else if strings.Contains(line, "audio.channels =") {
			if ch, err := strconv.Atoi(strings.Trim(strings.Split(line, "=")[1], " \"")); err == nil {
				currentStream.Channels = ch
			}
		} else if strings.Contains(line, "audio.rate =") {
			if rate, err := strconv.Atoi(strings.Trim(strings.Split(line, "=")[1], " \"")); err == nil {
				currentStream.SampleRate = rate
			}
		}
	}
	
	if currentStream.Process != "" {
		a.ActiveStreams = append(a.ActiveStreams, currentStream)
	}
}

func (a *AudioInfo) getAudioMetrics() {
	// Try to get audio levels from PulseAudio
	if output, err := exec.Command("pactl", "list", "sources").Output(); err == nil {
		a.parseAudioLevels(string(output), "input")
	}
	
	if output, err := exec.Command("pactl", "list", "sinks").Output(); err == nil {
		a.parseAudioLevels(string(output), "output")
	}
	
	// Default values for demonstration
	if a.InputLevel == 0 && len(a.ActiveStreams) > 0 {
		a.InputLevel = -12.5 // Simulated input level
	}
	if a.OutputLevel == 0 && len(a.ActiveStreams) > 0 {
		a.OutputLevel = -8.3 // Simulated output level
	}
	
	a.SampleRate = "48 kHz"
	a.BitDepth = "16-bit"
	a.Latency = "23.4 ms"
	
	// Calculate peak and RMS levels for VU meters
	a.calculateVUMetrics()
}

func (a *AudioInfo) parseAudioLevels(output, direction string) {
	lines := strings.Split(output, "\n")
	var currentLevel float64
	
	for _, line := range lines {
		if strings.Contains(line, "Volume:") {
			// Parse volume level - simplified
			parts := strings.Split(line, "/")
			if len(parts) >= 2 {
				levelStr := strings.TrimSpace(parts[1])
				if percent, err := strconv.ParseFloat(strings.TrimSuffix(levelStr, "%"), 64); err == nil {
					currentLevel = percent/100*96 - 96 // Convert to dB scale
				}
			}
		}
	}
	
	if direction == "input" {
		a.InputLevel = currentLevel
	} else {
		a.OutputLevel = currentLevel
	}
}

// NEW: Генерация данных для частотного спектра
func (a *AudioInfo) generateFrequencyData() {
	// Симуляция частотного спектра (в реальности нужно получать от аудиодрайвера)
	for i := range a.FrequencyBands {
		// Базовый уровень шума
		level := rand.Float64() * 20
		
		// Добавляем пики на разных частотах для реалистичности
		switch i {
		case 1, 2: // Низкие частоты
			level += rand.Float64() * 40
		case 3, 4: // Средние частоты  
			level += rand.Float64() * 60
		case 5, 6: // Высокие частоты
			level += rand.Float64() * 30
		}
		
		if level > 100 {
			level = 100
		}
		a.FrequencyBands[i] = level
	}
}

// NEW: Расчет метрик для VU-метров
func (a *AudioInfo) calculateVUMetrics() {
	// Конвертируем dB уровни в проценты для VU-метров
	// -96dB = 0%, 0dB = 100%
	a.PeakLevel = (a.InputLevel + 96) / 96 * 100
	if a.PeakLevel < 0 {
		a.PeakLevel = 0
	} else if a.PeakLevel > 100 {
		a.PeakLevel = 100
	}
	
	// RMS обычно на 6-12dB ниже пика
	a.RMSLevel = a.PeakLevel * 0.7
	if a.RMSLevel < 0 {
		a.RMSLevel = 0
	}
}