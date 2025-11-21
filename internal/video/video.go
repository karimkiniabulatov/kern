package video

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
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
	Name         string
	Type         string // hardware/software
	Codecs       []string
	Active       bool
	Utilization  float64
	// Убрано поле Status, так как используется Active для отслеживания состояния
}

func Summary() (*VideoInfo, error) {
	info := &VideoInfo{
		VideoDevices:  []VideoDevice{},
		ActiveStreams: []VideoStream{},
		GPUEncoders:   []VideoEncoder{},
	}

	// Detect video devices
	info.detectVideoDevices()
	
	// Detect active video streams
	info.detectActiveStreams()
	
	// Detect GPU video encoders
	info.detectGPUEncoders()
	
	// Get video metrics
	info.getVideoMetrics()

	return info, nil
}

func (v *VideoInfo) detectVideoDevices() {
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
		v.VideoDevices = append(v.VideoDevices, VideoDevice{
			Name:   "Default Video Device",
			Type:   "virtual",
			Status: "available",
		})
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
					Active: false, // Используем Active вместо Status
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
				Active: false, // Используем Active вместо Status
			})
		}
	}
	
	// Check Intel Quick Sync
	v.detectIntelEncoders()
	
	// Add software encoders as fallback
	v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
		Name:   "CPU x264",
		Type:   "software",
		Codecs: []string{"H.264"},
		Active: false, // Используем Active вместо Status
	})
	v.GPUEncoders = append(v.GPUEncoders, VideoEncoder{
		Name:   "CPU x265",
		Type:   "software",
		Codecs: []string{"H.265"},
		Active: false, // Используем Active вместо Status
	})
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
			Name:    "AMD VCE/AMF",
			Type:    "hardware", 
			Codecs:  []string{"H.264", "H.265"},
			Active:  strings.Contains(output, "active"),
		}
		v.GPUEncoders = append(v.GPUEncoders, encoder)
	}
}

func (v *VideoInfo) detectIntelEncoders() {
	// Check for Intel GPU
	if output, err := exec.Command("lspci").Output(); err == nil {
		if strings.Contains(string(output), "Intel Corporation") && strings.Contains(string(output), "VGA") {
			encoder := VideoEncoder{
				Name:   "Intel Quick Sync",
				Type:   "hardware",
				Codecs: []string{"H.264", "H.265", "VP9"},
				Active: false,
			}
			v.GPUEncoders = append(v.GPUEncoders, encoder)
		}
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
			if output, err := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits").Output(); err == nil {
				if util, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64); err == nil {
					v.GPUUtilization = util
				}
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
}