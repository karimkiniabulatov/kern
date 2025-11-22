package video

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Константы для определения ОС
const (
	OSWindows = "windows"
	OSLinux   = "linux"
	OSDarwin  = "darwin"
)

type VideoInfo struct {
	VideoDevices    []VideoDevice
	ActiveStreams   []VideoStream
	GPUEncoders     []VideoEncoder
	EncodingStatus  string
	Framerate       float64
	GPUUtilization  float64
	Bitrate         string
	Resolution      string
	Codec           string
}

type VideoDevice struct {
	Name    string
	ID      string
	Type    string // camera/encoder/decoder
	Formats []string
	Status  string
	Driver  string
}

type VideoStream struct {
	Process    string
	PID        int
	Type       string // capture/encode/decode/playback
	Device     string
	Resolution string
	Framerate  float64
	Bitrate    string
	Codec      string
}

type VideoEncoder struct {
	Name        string
	Type        string // hardware/software
	Codecs      []string
	Active      bool
	Utilization float64
}

func Summary() (*VideoInfo, error) {
	info := &VideoInfo{
		VideoDevices:    []VideoDevice{},
		ActiveStreams:   []VideoStream{},
		GPUEncoders:     []VideoEncoder{},
		EncodingStatus:  "idle",
		Framerate:       0.0,
		GPUUtilization:  0.0,
		Bitrate:         "N/A",
		Resolution:      "N/A",
		Codec:           "N/A",
	}

	// Detect video devices
	info.detectVideoDevices()

	// Detect active video streams
	info.detectActiveStreams()

	// Detect GPU video encoders
	info.detectGPUEncoders()

	// Get video metrics
	info.getVideoMetrics()

	// Гарантируем, что все числовые поля инициализированы
	if info.Framerate < 0 {
		info.Framerate = 0.0
	}
	if info.GPUUtilization < 0 {
		info.GPUUtilization = 0.0
	}

	// Гарантируем, что массивы не nil
	if info.VideoDevices == nil {
		info.VideoDevices = []VideoDevice{}
	}
	if info.ActiveStreams == nil {
		info.ActiveStreams = []VideoStream{}
	}
	if info.GPUEncoders == nil {
		info.GPUEncoders = []VideoEncoder{}
	}

	// Гарантируем, что строковые поля не пустые
	if info.EncodingStatus == "" {
		info.EncodingStatus = "idle"
	}
	if info.Bitrate == "" {
		info.Bitrate = "N/A"
	}
	if info.Resolution == "" {
		info.Resolution = "N/A"
	}
	if info.Codec == "" {
		info.Codec = "N/A"
	}

	return info, nil
}

func (v *VideoInfo) detectVideoDevices() {
	switch runtime.GOOS {
	case OSLinux:
		v.detectVideoDevicesLinux()
	case OSWindows:
		v.detectVideoDevicesWindows()
	case OSDarwin:
		v.detectVideoDevicesMacOS()
	default:
		v.addDefaultDevice()
	}

	// Гарантируем, что есть хотя бы одно устройство
	if len(v.VideoDevices) == 0 {
		v.addDefaultDevice()
	}
}

func (v *VideoInfo) detectVideoDevicesLinux() {
	// Try V4L2 for Linux cameras
	if output, err := exec.Command("ls", "/dev/video*").Output(); err == nil {
		devices := strings.Fields(string(output))
		for _, device := range devices {
			v.VideoDevices = append(v.VideoDevices, VideoDevice{
				Name:   fmt.Sprintf("Camera %s", device),
				ID:     device,
				Type:   "camera",
				Status: "available",
				Driver: "V4L2",
			})
		}
	}

	// Try to get more details with v4l2-ctl
	for i := range v.VideoDevices {
		if output, err := exec.Command("v4l2-ctl", "--device", v.VideoDevices[i].ID, "--list-formats").Output(); err == nil {
			v.parseVideoFormats(i, string(output))
		}
	}

	// Try to detect displays
	if output, err := exec.Command("xrandr", "--listmonitors").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Monitors:") {
				continue
			}
			if strings.Contains(line, "+") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					v.VideoDevices = append(v.VideoDevices, VideoDevice{
						Name:   parts[3],
						ID:     parts[3],
						Type:   "display",
						Status: "active",
						Driver: "X11",
					})
				}
			}
		}
	}

	// Fallback for systems without cameras
	if len(v.VideoDevices) == 0 {
		v.addDefaultDevice()
	}
}

func (v *VideoInfo) detectVideoDevicesWindows() {
	// Detect cameras using WMIC
	if output, err := exec.Command("wmic", "path", "win32_pnpentity", "get", "name").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "camera") || strings.Contains(lineLower, "webcam") ||
				strings.Contains(lineLower, "video") && strings.Contains(lineLower, "device") {
				name := strings.TrimSpace(line)
				if name != "" && name != "Name" {
					v.VideoDevices = append(v.VideoDevices, VideoDevice{
						Name:   name,
						ID:     "unknown",
						Type:   "camera",
						Status: "available",
						Driver: "DirectShow/WMF",
					})
				}
			}
		}
	}

	// Detect displays using PowerShell
	if output, err := exec.Command("powershell", "-Command", "Get-WmiObject -Namespace root\\wmi -Class WmiMonitorBasicDisplayParams | Select-Object InstanceName").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "DISPLAY") && !strings.Contains(line, "InstanceName") {
				displayName := strings.TrimSpace(line)
				if displayName != "" {
					v.VideoDevices = append(v.VideoDevices, VideoDevice{
						Name:   displayName,
						ID:     displayName,
						Type:   "display",
						Status: "active",
						Driver: "Windows Display",
					})
				}
			}
		}
	}

	// Fallback for systems without detected devices
	if len(v.VideoDevices) == 0 {
		v.addDefaultDevice()
	}
}

func (v *VideoInfo) detectVideoDevicesMacOS() {
	// Detect cameras using system_profiler
	if output, err := exec.Command("system_profiler", "SPCameraDataType").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		var currentCamera string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasSuffix(line, ":") && !strings.Contains(line, "Cameras") {
				currentCamera = strings.TrimSuffix(line, ":")
			} else if currentCamera != "" && strings.Contains(line, "Unique ID") {
				v.VideoDevices = append(v.VideoDevices, VideoDevice{
					Name:   currentCamera,
					ID:     "unknown",
					Type:   "camera",
					Status: "available",
					Driver: "AVFoundation",
				})
				currentCamera = ""
			}
		}
	}

	// Detect displays using system_profiler
	if output, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Resolution:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					displayName := fmt.Sprintf("Display %s", strings.TrimSpace(parts[1]))
					v.VideoDevices = append(v.VideoDevices, VideoDevice{
						Name:   displayName,
						ID:     displayName,
						Type:   "display",
						Status: "active",
						Driver: "Quartz/Display",
					})
				}
			}
		}
	}

	// Fallback for systems without detected devices
	if len(v.VideoDevices) == 0 {
		v.addDefaultDevice()
	}
}

func (v *VideoInfo) parseVideoFormats(deviceIndex int, output string) {
	lines := strings.Split(output, "\n")
	var formats []string

	for _, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			start := strings.Index(line, "]") + 2
			if start < len(line) {
				format := strings.TrimSpace(line[start:])
				formats = append(formats, format)
			}
		}
	}

	v.VideoDevices[deviceIndex].Formats = formats
}

func (v *VideoInfo) detectActiveStreams() {
	switch runtime.GOOS {
	case OSLinux:
		v.detectActiveStreamsLinux()
	case OSWindows:
		v.detectActiveStreamsWindows()
	case OSDarwin:
		v.detectActiveStreamsMacOS()
	}

	// Гарантируем, что массив не nil
	if v.ActiveStreams == nil {
		v.ActiveStreams = []VideoStream{}
	}
}

func (v *VideoInfo) detectActiveStreamsLinux() {
	// Find processes using video devices
	if output, err := exec.Command("lsof", "/dev/video*").Output(); err == nil {
		v.parseVideoProcesses(string(output))
	}

	// Check for common video applications
	videoProcesses := []string{
		"ffmpeg", "vlc", "mpv", "obs", "gstreamer",
		"chrome", "firefox", "zoom", "teams", "gst", "chromium", "skype",
	}

	for _, proc := range videoProcesses {
		if output, err := exec.Command("pgrep", "-l", proc).Output(); err == nil {
			v.parseVideoApps(string(output))
		}
	}
}

func (v *VideoInfo) detectActiveStreamsWindows() {
	// Check for common video applications using tasklist
	videoProcesses := []string{
		"ffmpeg.exe", "vlc.exe", "obs64.exe", "obs.exe", "chrome.exe",
		"firefox.exe", "msedge.exe", "zoom.exe", "teams.exe", "skype.exe",
	}

	for _, proc := range videoProcesses {
		if output, err := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", proc), "/FO", "CSV", "/NH").Output(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, proc) {
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(line, proc) {
						fields := strings.Split(line, ",")
						if len(fields) >= 2 {
							pidStr := strings.Trim(fields[1], "\" ")
							pid, err := strconv.Atoi(pidStr)
							if err == nil {
								streamType := v.determineStreamTypeWindows(proc)
								stream := VideoStream{
									Process: proc,
									PID:     pid,
									Type:    streamType,
								}
								v.ActiveStreams = append(v.ActiveStreams, stream)
							}
						}
					}
				}
			}
		}
	}
}

func (v *VideoInfo) detectActiveStreamsMacOS() {
	// Check for common video applications using ps
	videoProcesses := []string{
		"ffmpeg", "vlc", "obs", "QuickTime Player", "FaceTime",
		"Google Chrome", "Firefox", "zoom.us", "Skype",
	}

	for _, proc := range videoProcesses {
		if output, err := exec.Command("pgrep", "-l", "-f", proc).Output(); err == nil {
			v.parseVideoApps(string(output))
		}
	}

	// Additional check using ps for broader capture
	if output, err := exec.Command("ps", "aux").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			for _, proc := range videoProcesses {
				if strings.Contains(line, proc) && !strings.Contains(line, "pgrep") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						pid, err := strconv.Atoi(fields[1])
						if err == nil {
							streamType := v.determineStreamTypeMacOS(proc)
							stream := VideoStream{
								Process: proc,
								PID:     pid,
								Type:    streamType,
							}
							// Avoid duplicates
							found := false
							for _, existing := range v.ActiveStreams {
								if existing.PID == pid {
									found = true
									break
								}
							}
							if !found {
								v.ActiveStreams = append(v.ActiveStreams, stream)
							}
						}
					}
				}
			}
		}
	}
}

func (v *VideoInfo) determineStreamTypeWindows(processName string) string {
	switch processName {
	case "ffmpeg.exe", "obs64.exe", "obs.exe":
		return "encode"
	case "vlc.exe":
		return "playback"
	case "chrome.exe", "firefox.exe", "msedge.exe":
		return "decode"
	case "zoom.exe", "teams.exe", "skype.exe":
		return "capture"
	default:
		return "playback"
	}
}

func (v *VideoInfo) determineStreamTypeMacOS(processName string) string {
	switch processName {
	case "ffmpeg", "obs":
		return "encode"
	case "vlc", "QuickTime Player":
		return "playback"
	case "Google Chrome", "Firefox":
		return "decode"
	case "FaceTime", "zoom.us", "Skype":
		return "capture"
	default:
		return "playback"
	}
}

func (v *VideoInfo) parseVideoProcesses(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/dev/video") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid, _ := strconv.Atoi(fields[1])
				device := ""
				if len(fields) > 8 {
					device = fields[8]
				}
				stream := VideoStream{
					Process: fields[0],
					PID:     pid,
					Type:    "capture",
					Device:  device,
				}
				v.ActiveStreams = append(v.ActiveStreams, stream)
			}
		}
	}
}

func (v *VideoInfo) parseVideoApps(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			pid, _ := strconv.Atoi(fields[0])

			// Determine stream type based on application
			streamType := "playback"
			processName := fields[1]
			if processName == "ffmpeg" || processName == "obs" {
				streamType = "encode"
			} else if strings.Contains(processName, "chrome") || strings.Contains(processName, "firefox") {
				streamType = "decode"
			} else if strings.Contains(processName, "zoom") || strings.Contains(processName, "teams") || strings.Contains(processName, "skype") {
				streamType = "capture"
			}

			stream := VideoStream{
				Process: processName,
				PID:     pid,
				Type:    streamType,
			}
			v.ActiveStreams = append(v.ActiveStreams, stream)
		}
	}
}

func (v *VideoInfo) detectGPUEncoders() {
	switch runtime.GOOS {
	case OSLinux:
		v.detectGPUEncodersLinux()
	case OSWindows:
		v.detectGPUEncodersWindows()
	case OSDarwin:
		v.detectGPUEncodersMacOS()
	}

	// Add software encoders as fallback for all platforms
	v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
		Name:   "CPU x264",
		Type:   "software",
		Codecs: []string{"H.264"},
		Active: false,
	})
	v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
		Name:   "CPU x265",
		Type:   "software",
		Codecs: []string{"H.265"},
		Active: false,
	})

	// Гарантируем, что массив не nil
	if v.GPUEncoders == nil {
		v.GPUEncoders = []VideoEncoder{}
	}
}

func (v *VideoInfo) detectGPUEncodersLinux() {
	// Check NVIDIA NVENC
	if output, err := exec.Command("nvidia-smi", "--query-gpu=name,utilization.enc", "--format=csv,noheader,nounits").Output(); err == nil {
		v.parseNVIDIAEncoders(string(output))
	} else {
		// Fallback: try basic nvidia-smi query
		if output, err := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader").Output(); err == nil {
			if strings.TrimSpace(string(output)) != "" {
				v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
					Name:   "NVIDIA NVENC",
					Type:   "hardware",
					Codecs: []string{"H.264", "H.265", "AV1"},
					Active: false,
				})
			}
		}
	}

	// Check AMD VCE / AMF
	if output, err := exec.Command("rocm-smi", "--showuse").Output(); err == nil {
		v.parseAMDEncoders(string(output))
	} else if output, err := exec.Command("rocminfo").Output(); err == nil {
		if strings.Contains(string(output), "gfx") {
			v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
				Name:   "AMD AMF/VCE",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265"},
				Active: false,
			})
		}
	}

	// Check Intel Quick Sync
	v.detectIntelEncodersLinux()
}

func (v *VideoInfo) detectGPUEncodersWindows() {
	// Check NVIDIA NVENC
	if output, err := exec.Command("nvidia-smi", "--query-gpu=name,utilization.enc", "--format=csv,noheader,nounits").Output(); err == nil {
		v.parseNVIDIAEncoders(string(output))
	} else {
		// Fallback: check if NVIDIA GPU exists
		if output, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output(); err == nil {
			if strings.Contains(strings.ToLower(string(output)), "nvidia") {
				v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
					Name:   "NVIDIA NVENC",
					Type:   "hardware",
					Codecs: []string{"H.264", "H.265", "AV1"},
					Active: false,
				})
			}
		}
	}

	// Check Intel Quick Sync on Windows
	if output, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output(); err == nil {
		outputStr := strings.ToLower(string(output))
		if strings.Contains(outputStr, "intel") {
			v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
				Name:   "Intel Quick Sync",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265", "VP9"},
				Active: false,
			})
		}
	}

	// Check AMD on Windows
	if output, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output(); err == nil {
		outputStr := strings.ToLower(string(output))
		if strings.Contains(outputStr, "amd") || strings.Contains(outputStr, "radeon") {
			v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
				Name:   "AMD AMF/VCE",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265"},
				Active: false,
			})
		}
	}
}

func (v *VideoInfo) detectGPUEncodersMacOS() {
	// Check GPU type using system_profiler
	if output, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
		outputStr := string(output)
		
		// Apple Silicon
		if strings.Contains(outputStr, "Apple M1") || strings.Contains(outputStr, "Apple M2") ||
			strings.Contains(outputStr, "Apple M3") || strings.Contains(outputStr, "Apple M4") {
			v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
				Name:   "Apple Silicon Media Engine",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265", "ProRes"},
				Active: false,
			})
		}
		
		// Intel
		if strings.Contains(outputStr, "Intel") {
			v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
				Name:   "Intel Quick Sync",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265"},
				Active: false,
			})
		}
		
		// AMD (eGPU or older Mac Pro)
		if strings.Contains(outputStr, "AMD") || strings.Contains(outputStr, "Radeon") {
			v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
				Name:   "AMD AMF/VCE",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265"},
				Active: false,
			})
		}
	}
}

func (v *VideoInfo) detectIntelEncodersLinux() {
	// Check for Intel GPU
	if output, err := exec.Command("lspci").Output(); err == nil {
		if strings.Contains(string(output), "Intel Corporation") && strings.Contains(string(output), "VGA") {
			v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
				Name:   "Intel Quick Sync",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265", "VP9"},
				Active: false,
			})
		}
	}
}

func (v *VideoInfo) parseNVIDIAEncoders(output string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) >= 2 {
			util, _ := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
			encoder := VideoEncoder{
				Name:         fmt.Sprintf("NVIDIA NVENC (%s)", strings.TrimSpace(fields[0])),
				Type:         "hardware",
				Codecs:       []string{"H.264", "H.265", "AV1"},
				Active:       util > 0,
				Utilization:  util,
			}
			v.GPUEncoders = append(v.GPUEncoders, encoder)
		} else if len(fields) == 1 {
			// Handle case when only GPU name is available
			encoder := VideoEncoder{
				Name:   fmt.Sprintf("NVIDIA NVENC (%s)", strings.TrimSpace(fields[0])),
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265", "AV1"},
				Active: false,
			}
			v.GPUEncoders = append(v.GPUEncoders, encoder)
		}
	}
}

func (v *VideoInfo) parseAMDEncoders(output string) {
	// Simplified AMD encoder detection
	if strings.Contains(output, "Video Codec") || strings.Contains(output, "encode") {
		encoder := VideoEncoder{
			Name:   "AMD VCE/AMF",
			Type:   "hardware",
			Codecs: []string{"H.264", "H.265"},
			Active: strings.Contains(output, "active"),
		}
		v.GPUEncoders = append(v.GPUEncoders, encoder)
	}
}

func (v *VideoInfo) getVideoMetrics() {
	// Set default metrics based on detected streams
	if len(v.ActiveStreams) > 0 {
		v.Resolution = "1920x1080"
		v.Framerate = 30.0
		v.Bitrate = "4.5 Mbps"
		v.Codec = "H.264"

		// Determine encoding status
		encoding := false
		decoding := false
		capture := false

		for _, stream := range v.ActiveStreams {
			switch stream.Type {
			case "encode":
				encoding = true
			case "decode":
				decoding = true
			case "capture":
				capture = true
			}
		}

		if encoding {
			v.EncodingStatus = "encoding"
		} else if decoding {
			v.EncodingStatus = "decoding"
		} else if capture {
			v.EncodingStatus = "capturing"
		} else {
			v.EncodingStatus = "active"
		}

		// Calculate GPU utilization
		totalUtil := 0.0
		activeEncoders := 0
		for _, encoder := range v.GPUEncoders {
			if encoder.Active {
				totalUtil += encoder.Utilization
				activeEncoders++
			}
		}

		// If no encoder utilization data, try to get general GPU utilization
		if activeEncoders == 0 {
			switch runtime.GOOS {
			case OSLinux, OSWindows:
				if output, err := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits").Output(); err == nil {
					if util, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64); err == nil {
						v.GPUUtilization = util
					}
				}
			case OSDarwin:
				// macOS doesn't have nvidia-smi by default, use alternative methods if needed
				v.GPUUtilization = 0
			}
		} else {
			v.GPUUtilization = totalUtil / float64(activeEncoders)
		}
	} else {
		v.EncodingStatus = "idle"
		v.Resolution = "N/A"
		v.Framerate = 0
		v.Bitrate = "N/A"
		v.Codec = "N/A"
		v.GPUUtilization = 0
	}

	// Гарантируем корректные значения для гистограмм
	if v.Framerate < 0 {
		v.Framerate = 0.0
	}
	if v.GPUUtilization < 0 {
		v.GPUUtilization = 0.0
	}
	if v.GPUUtilization > 100 {
		v.GPUUtilization = 100.0
	}
}

func (v *VideoInfo) addDefaultDevice() {
	v.VideoDevices = append(v.VideoDevices, VideoDevice{
		Name:   "Default Video Device",
		Type:   "virtual",
		Status: "available",
	})
}