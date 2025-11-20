package video

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type VideoInfo struct {
	VideoDevices   []VideoDevice
	ActiveStreams  []VideoStream
	GPUEncoders    []VideoEncoder
	Resolution     string
	FrameRate      float64
	Bitrate        string
	Codec          string
	EncodingStatus string // idle/encoding/decoding
}

type VideoDevice struct {
	Name     string
	ID       string
	Type     string // camera/encoder/decoder
	Formats  []string
	Status   string
	Driver   string
}

type VideoStream struct {
	Process    string
	PID        int
	Type       string // capture/encode/decode/playback
	Device     string
	Resolution string
	FrameRate  float64
	Bitrate    string
	Codec      string
}

type VideoEncoder struct {
	Name      string
	Type      string // hardware/software
	Codecs    []string
	Active    bool
	Utilization float64
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
		"chrome", "firefox", "zoom", "teams",
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
				stream := VideoStream{
					Process: fields[0],
					PID:     pid,
					Type:    "capture",
					Device:  fields[8],
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
			if fields[1] == "ffmpeg" || fields[1] == "obs" {
				streamType = "encode"
			} else if strings.Contains(fields[1], "chrome") || strings.Contains(fields[1], "firefox") {
				streamType = "decode"
			}
			
			stream := VideoStream{
				Process: fields[1],
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
	}
	
	// Check AMD VCE
	if output, err := exec.Command("rocm-smi", "--showuse").Output(); err == nil {
		v.parseAMDEncoders(string(output))
	}
	
	// Check Intel Quick Sync
	v.detectIntelEncoders()
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
		}
	}
}

func (v *VideoInfo) parseAMDEncoders(output string) {
	// Simplified AMD encoder detection
	if strings.Contains(output, "Video Codec") {
		encoder := VideoEncoder{
			Name:    "AMD VCE",
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
			}
			v.GPUEncoders = append(v.GPUEncoders, encoder)
		}
	}
}

func (v *VideoInfo) getVideoMetrics() {
	// Set default metrics based on detected streams
	if len(v.ActiveStreams) > 0 {
		v.Resolution = "1920x1080"
		v.FrameRate = 30.0
		v.Bitrate = "4.5 Mbps"
		v.Codec = "H.264"
		
		// Determine encoding status
		encoding := false
		for _, stream := range v.ActiveStreams {
			if stream.Type == "encode" {
				encoding = true
				break
			}
		}
		
		if encoding {
			v.EncodingStatus = "encoding"
		} else {
			v.EncodingStatus = "active"
		}
	} else {
		v.EncodingStatus = "idle"
		v.Resolution = "N/A"
		v.FrameRate = 0
		v.Bitrate = "N/A"
		v.Codec = "N/A"
	}
}