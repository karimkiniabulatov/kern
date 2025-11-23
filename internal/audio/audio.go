package audio

import (
	"fmt"
	"math/rand"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type AudioInfo struct {
	InputDevices   []AudioDevice
	OutputDevices  []AudioDevice
	ActiveStreams  []AudioStream
	InputLevel     float64 // dB
	OutputLevel    float64 // dB
	SampleRate     string
	BitDepth       string
	Latency        string
	
	// Для гистограмм
	FrequencyBands []float64 // Частотные полосы для спектр-анализатора (0-100%)
	PeakLevel      float64   // Пиковый уровень для VU-метра (0-100%)
	RMSLevel       float64   // Среднеквадратичный уровень (0-100%)
}

type AudioDevice struct {
	Name     string
	ID       string
	Type     string // input/output
	Channels int
	Format   string
	Status   string // active/inactive
}

type AudioStream struct {
	Process    string
	PID        int
	Type       string // capture/playback
	Format     string
	SampleRate int
	Channels   int
	Duration   string
	Bitrate    string
}

var lastAudioStats map[string]AudioStream

func Summary() (*AudioInfo, error) {
	info := &AudioInfo{
		InputDevices:   []AudioDevice{},
		OutputDevices:  []AudioDevice{},
		ActiveStreams:  []AudioStream{},
		InputLevel:     -96.0,
		OutputLevel:    -96.0,
		SampleRate:     "Unknown",
		BitDepth:       "Unknown",
		Latency:        "Unknown",
		FrequencyBands: make([]float64, 8), // 8 частотных полос
		PeakLevel:      0.0,
		RMSLevel:       0.0,
	}

	// Detect audio devices
	info.detectAudioDevices()
	
	// Detect active audio streams
	info.detectActiveStreams()
	
	// Get audio levels and metrics
	info.getAudioMetrics()
	
	// Generate frequency spectrum data
	info.generateFrequencyData()

	return info, nil
}

func (a *AudioInfo) detectAudioDevices() {
	switch runtime.GOOS {
	case "linux":
		a.detectLinuxAudioDevices()
	case "darwin": // macOS
		a.detectMacAudioDevices()
	case "windows":
		a.detectWindowsAudioDevices()
	default:
		a.detectDefaultAudioDevices()
	}
}

func (a *AudioInfo) detectLinuxAudioDevices() {
	// Try ALSA first
	if output, err := exec.Command("arecord", "-l").Output(); err == nil {
		a.parseALSAInputDevices(string(output))
	}
	
	if output, err := exec.Command("aplay", "-l").Output(); err == nil {
		a.parseALSAOutputDevices(string(output))
	}
	
	// Try PulseAudio
	if _, err := exec.LookPath("pactl"); err == nil {
		if output, err := exec.Command("pactl", "list", "sources", "short").Output(); err == nil {
			a.parsePulseAudioInputs(string(output))
		}
		
		if output, err := exec.Command("pactl", "list", "sinks", "short").Output(); err == nil {
			a.parsePulseAudioOutputs(string(output))
		}
	}
	
	// Fallback if no devices detected
	if len(a.InputDevices) == 0 && len(a.OutputDevices) == 0 {
		a.detectDefaultAudioDevices()
	}
}

func (a *AudioInfo) detectMacAudioDevices() {
	// Используем system_profiler для получения информации об аудиоустройствах
	cmd := exec.Command("system_profiler", "SPAudioDataType")
	if output, err := cmd.Output(); err == nil {
		a.parseMacAudioDevices(string(output))
	} else {
		a.detectDefaultAudioDevices()
	}
	
	// Дополнительно получаем уровни громкости
	a.getMacVolumeLevels()
}

func (a *AudioInfo) detectWindowsAudioDevices() {
	// Используем PowerShell для получения информации об аудиоустройствах
	psCmd := `Get-WmiObject -Class Win32_SoundDevice | Select-Object Name, Status | Format-Table -HideTableHeaders`
	cmd := exec.Command("powershell", "-Command", psCmd)
	if output, err := cmd.Output(); err == nil {
		a.parseWindowsAudioDevices(string(output))
	} else {
		a.detectDefaultAudioDevices()
	}
	
	// Дополнительная информация через AudioSessionManager
	a.getWindowsAudioDetails()
}

func (a *AudioInfo) getMacVolumeLevels() {
	// Получаем уровень выходной громкости через osascript
	volumeCmd := "osascript -e 'output volume of (get volume settings)'"
	cmd := exec.Command("bash", "-c", volumeCmd)
	if output, err := cmd.Output(); err == nil {
		volumeStr := strings.TrimSpace(string(output))
		if volume, err := strconv.Atoi(volumeStr); err == nil {
			// Конвертируем проценты в dB: 100% = 0dB, 0% = -96dB
			a.OutputLevel = float64(volume)/100*96 - 96
		}
	}
	
	// Получаем уровень входной громкости
	inputVolumeCmd := "osascript -e 'input volume of (get volume settings)'"
	cmd = exec.Command("bash", "-c", inputVolumeCmd)
	if output, err := cmd.Output(); err == nil {
		volumeStr := strings.TrimSpace(string(output))
		if volume, err := strconv.Atoi(volumeStr); err == nil {
			a.InputLevel = float64(volume)/100*96 - 96
		}
	}
}

func (a *AudioInfo) getWindowsAudioDetails() {
	// Получаем уровни громкости через Windows Media Player COM object
	volumeCmd := `$wmp = New-Object -ComObject WMPlayer.OCX; $wmp.settings.volume`
	cmd := exec.Command("powershell", "-Command", volumeCmd)
	if output, err := cmd.Output(); err == nil {
		volumeStr := strings.TrimSpace(string(output))
		if volume, err := strconv.Atoi(volumeStr); err == nil {
			a.OutputLevel = float64(volume)/100*96 - 96
		}
	}
	
	// Дополнительная информация о аудиоустройствах
	audioDeviceCmd := `Get-CimInstance -Namespace "Root\Cimv2" -ClassName Win32_SoundDevice | Select-Object Name, Status, Manufacturer`
	cmd = exec.Command("powershell", "-Command", audioDeviceCmd)
	if output, err := cmd.Output(); err == nil {
		a.parseWindowsAudioDetails(string(output))
	}
}

func (a *AudioInfo) parseWindowsAudioDetails(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Name") || strings.HasPrefix(line, "-") {
			continue
		}
		
		// Обновляем информацию о существующих устройствах
		for i, device := range a.OutputDevices {
			if strings.Contains(line, device.Name) {
				if strings.Contains(line, "Manufacturer") {
					parts := strings.Split(line, "Manufacturer")
					if len(parts) > 1 {
						a.OutputDevices[i].Format = strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}
}

func (a *AudioInfo) detectDefaultAudioDevices() {
	a.InputDevices = append(a.InputDevices, AudioDevice{
		Name:   "Default Input",
		Type:   "input",
		Status: "active",
	})
	a.OutputDevices = append(a.OutputDevices, AudioDevice{
		Name:   "Default Output", 
		Type:   "output",
		Status: "active",
	})
}

func (a *AudioInfo) parseMacAudioDevices(output string) {
	lines := strings.Split(output, "\n")
	var currentDevice AudioDevice
	inDeviceSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "Default Input Device: Yes") {
			currentDevice.Type = "input"
			currentDevice.Status = "active"
			a.InputDevices = append(a.InputDevices, currentDevice)
			currentDevice = AudioDevice{}
		} else if strings.Contains(line, "Default Output Device: Yes") {
			currentDevice.Type = "output" 
			currentDevice.Status = "active"
			a.OutputDevices = append(a.OutputDevices, currentDevice)
			currentDevice = AudioDevice{}
		} else if strings.HasSuffix(line, ":") && !strings.Contains(line, "Devices:") {
			// Это имя устройства
			currentDevice.Name = strings.TrimSuffix(line, ":")
			inDeviceSection = true
		} else if inDeviceSection && strings.Contains(line, "Input Channels:") {
			if ch, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1])); err == nil {
				currentDevice.Channels = ch
			}
		} else if inDeviceSection && strings.Contains(line, "Output Channels:") {
			if ch, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1])); err == nil {
				currentDevice.Channels = ch
			}
		} else if inDeviceSection && strings.Contains(line, "Sample Rate:") {
			a.SampleRate = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if inDeviceSection && strings.Contains(line, "Bit Depth:") {
			a.BitDepth = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}
}

func (a *AudioInfo) parseWindowsAudioDevices(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			deviceName := strings.Join(parts[:len(parts)-1], " ")
			status := parts[len(parts)-1]
			
			// На Windows сложно определить input/output без дополнительной информации
			// Добавляем как оба типа для безопасности
			if strings.Contains(strings.ToLower(deviceName), "microphone") || 
			   strings.Contains(strings.ToLower(deviceName), "input") ||
			   strings.Contains(strings.ToLower(deviceName), "capture") {
				a.InputDevices = append(a.InputDevices, AudioDevice{
					Name:   deviceName,
					Type:   "input",
					Status: status,
				})
			} else {
				a.OutputDevices = append(a.OutputDevices, AudioDevice{
					Name:   deviceName,
					Type:   "output",
					Status: status,
				})
			}
		}
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
	switch runtime.GOOS {
	case "linux":
		a.detectLinuxActiveStreams()
	case "darwin": // macOS
		a.detectMacActiveStreams()
	case "windows":
		a.detectWindowsActiveStreams()
	default:
		a.detectDefaultActiveStreams()
	}
}

func (a *AudioInfo) detectLinuxActiveStreams() {
	// Find processes using audio devices
	if _, err := exec.LookPath("lsof"); err == nil {
		if output, err := exec.Command("lsof", "+D", "/dev/snd").Output(); err == nil {
			a.parseAudioProcesses(string(output))
		}
	}
	
	// Try PulseAudio streams
	if _, err := exec.LookPath("pactl"); err == nil {
		if output, err := exec.Command("pactl", "list", "sink-inputs").Output(); err == nil {
			a.parsePulseAudioStreams(string(output), "playback")
		}
		
		if output, err := exec.Command("pactl", "list", "source-outputs").Output(); err == nil {
			a.parsePulseAudioStreams(string(output), "capture") 
		}
	}
}

func (a *AudioInfo) detectMacActiveStreams() {
	// На macOS используем lsof для поиска аудиопроцессов и osascript для дополнительной информации
	if _, err := exec.LookPath("lsof"); err == nil {
		// Ищем процессы, использующие аудиоустройства
		if output, err := exec.Command("lsof", "-c", "CoreAudio").Output(); err == nil {
			a.parseMacAudioProcesses(string(output))
		}
		
		// Дополнительно ищем процессы, использующие аудио устройства
		if output, err := exec.Command("lsof", "/dev/audio*").Output(); err == nil {
			a.parseMacAudioDeviceProcesses(string(output))
		}
	}
	
	// Используем osascript для получения информации о аудиопотоках
	a.getMacAudioStreamsInfo()
}

func (a *AudioInfo) detectWindowsActiveStreams() {
	// На Windows используем PowerShell для получения информации о аудиопотоках
	psCmd := `Get-Process | Where-Object {$_.MainWindowTitle -ne ""} | Select-Object Name, Id, MainWindowTitle`
	cmd := exec.Command("powershell", "-Command", psCmd)
	if output, err := cmd.Output(); err == nil {
		a.parseWindowsActiveStreams(string(output))
	} else {
		// Fallback на tasklist
		if output, err := exec.Command("tasklist").Output(); err == nil {
			a.parseWindowsProcesses(string(output))
		} else {
			a.detectDefaultActiveStreams()
		}
	}
	
	// Дополнительная информация через AudioSessionManager
	a.getWindowsAudioSessions()
}

func (a *AudioInfo) getMacAudioStreamsInfo() {
	// Получаем информацию о текущих аудиопотоках через system_profiler
	cmd := exec.Command("system_profiler", "SPAudioDataType")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Process:") {
				processName := strings.TrimSpace(strings.Split(line, ":")[1])
				stream := AudioStream{
					Process: processName,
					Type:    "playback",
					Format:  "Core Audio",
				}
				a.ActiveStreams = append(a.ActiveStreams, stream)
			}
		}
	}
}

func (a *AudioInfo) getWindowsAudioSessions() {
	// Получаем информацию о аудиосессиях через PowerShell
	psCmd := `
	Add-Type -TypeDefinition @"
	using System;
	using System.Runtime.InteropServices;
	
	namespace AudioTools {
		public class AudioSession {
			[DllImport("winmm.dll")]
			public static extern int waveInGetNumDevs();
			
			[DllImport("winmm.dll")] 
			public static extern int waveOutGetNumDevs();
		}
	}
"@
	[AudioTools.AudioSession]::waveOutGetNumDevs()
	`
	cmd := exec.Command("powershell", "-Command", psCmd)
	if _, err := cmd.Output(); err == nil {
		// Если команда выполнилась успешно, добавляем системные потоки
		a.ActiveStreams = append(a.ActiveStreams, AudioStream{
			Process: "Windows Audio Session",
			Type:    "playback",
			Format:  "Windows Audio",
		})
	}
}

func (a *AudioInfo) parseWindowsActiveStreams(output string) {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i < 3 { // Пропускаем заголовки
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			processName := fields[0]
			pid, _ := strconv.Atoi(fields[1])
			
			// Фильтруем процессы, которые могут использовать аудио
			if strings.Contains(strings.ToLower(processName), "audio") ||
			   strings.Contains(strings.ToLower(processName), "sound") ||
			   strings.Contains(strings.ToLower(processName), "music") ||
			   strings.Contains(strings.ToLower(processName), "media") ||
			   strings.Contains(strings.ToLower(processName), "player") ||
			   strings.Contains(strings.ToLower(processName), "chrome") ||
			   strings.Contains(strings.ToLower(processName), "firefox") {
				
				stream := AudioStream{
					Process: processName,
					PID:     pid,
					Type:    "playback",
					Format:  "Windows Audio",
				}
				a.ActiveStreams = append(a.ActiveStreams, stream)
			}
		}
	}
}

func (a *AudioInfo) detectDefaultActiveStreams() {
	// Демо-потоки для платформ без специфичной реализации
	a.ActiveStreams = append(a.ActiveStreams, AudioStream{
		Process: "System Sounds",
		Type:    "playback",
		Format:  "PCM",
	})
}

func (a *AudioInfo) parseMacAudioProcesses(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "CoreAudio") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid, _ := strconv.Atoi(fields[1])
				stream := AudioStream{
					Process: fields[0],
					PID:     pid,
					Type:    "playback", // предполагаем playback для простоты
					Format:  "Core Audio",
				}
				a.ActiveStreams = append(a.ActiveStreams, stream)
			}
		}
	}
}

func (a *AudioInfo) parseMacAudioDeviceProcesses(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/dev/audio") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid, _ := strconv.Atoi(fields[1])
				streamType := "playback"
				if strings.Contains(line, "input") {
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

func (a *AudioInfo) parseWindowsProcesses(output string) {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i < 3 { // Пропускаем заголовки
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			// Добавляем системные процессы, которые могут использовать аудио
			processName := fields[0]
			if strings.Contains(strings.ToLower(processName), "audio") ||
			   strings.Contains(strings.ToLower(processName), "sound") ||
			   strings.Contains(strings.ToLower(processName), "music") {
				stream := AudioStream{
					Process: processName,
					Type:    "playback",
					Format:  "Windows Audio",
				}
				a.ActiveStreams = append(a.ActiveStreams, stream)
			}
		}
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
	switch runtime.GOOS {
	case "linux":
		a.getLinuxAudioMetrics()
	case "darwin":
		a.getMacAudioMetrics()
	case "windows":
		a.getWindowsAudioMetrics()
	default:
		a.getDefaultAudioMetrics()
	}
}

func (a *AudioInfo) getLinuxAudioMetrics() {
	// Try to get audio levels from PulseAudio
	if _, err := exec.LookPath("pactl"); err == nil {
		if output, err := exec.Command("pactl", "list", "sources").Output(); err == nil {
			a.parseAudioLevels(string(output), "input")
		}
		
		if output, err := exec.Command("pactl", "list", "sinks").Output(); err == nil {
			a.parseAudioLevels(string(output), "output")
		}
	}
	
	// Default values if no specific metrics found
	a.setDefaultAudioMetrics()
}

func (a *AudioInfo) getMacAudioMetrics() {
	// Уже получили уровни громкости в detectMacAudioDevices
	// Устанавливаем остальные метрики по умолчанию
	a.setDefaultAudioMetrics()
}

func (a *AudioInfo) getWindowsAudioMetrics() {
	// Уже получили уровни громкости в detectWindowsAudioDevices
	// Устанавливаем остальные метрики по умолчанию
	a.setDefaultAudioMetrics()
}

func (a *AudioInfo) getDefaultAudioMetrics() {
	a.setDefaultAudioMetrics()
}

func (a *AudioInfo) setDefaultAudioMetrics() {
    // Гарантируем, что всегда есть данные для гистограмм
    if a.InputLevel == 0 {
        a.InputLevel = -96.0 // Минимальное значение
    }
    if a.OutputLevel == 0 {
        a.OutputLevel = -96.0 // Минимальное значение
    }
    
    // Остальной код без изменений...
	if a.InputLevel == -96.0 && len(a.ActiveStreams) > 0 {
		a.InputLevel = -12.5 // Simulated input level
	}
	if a.OutputLevel == -96.0 && len(a.ActiveStreams) > 0 {
		a.OutputLevel = -8.3 // Simulated output level
	}
	
	if a.SampleRate == "Unknown" {
		a.SampleRate = "48 kHz"
	}
	if a.BitDepth == "Unknown" {
		a.BitDepth = "16-bit"
	}
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

// Генерация данных для частотного спектра
func (a *AudioInfo) generateFrequencyData() {
	// Симуляция частотного спектра
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

// Расчет метрик для VU-метров
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

// ShouldSkipFilesystem - временная реализация заглушка
func ShouldSkipFilesystem(filesystem, mountPoint string) bool {
    return false
}