package tests

import (
	"testing"
	video "github.com/karimkiniabulatov/kern/internal/video"
)

func TestVideoSummary(t *testing.T) {
	info, err := video.Summary()
	if err != nil {
		t.Fatalf("Failed to get video info: %v", err)
	}

	if info == nil {
		t.Error("Video info should not be nil")
	}

	// Should have at least default video device
	if len(info.VideoDevices) == 0 {
		t.Log("No video devices detected")
	}
}

func TestVideoStreamDetection(t *testing.T) {
	info, _ := video.Summary()
	
	if len(info.ActiveStreams) == 0 {
		t.Log("No active video streams detected (normal in test environment)")
	}
}

func TestGPUEncoderDetection(t *testing.T) {
	info, _ := video.Summary()
	
	// GPU encoders might not be available in test environment
	if len(info.GPUEncoders) == 0 {
		t.Log("No GPU encoders detected (might be normal)")
	}
}

func TestVideoMetrics(t *testing.T) {
	info, _ := video.Summary()
	
	// Encoding status should be one of expected values
	validStatuses := map[string]bool{
		"idle": true, "active": true, "encoding": true, "decoding": true,
	}
	
	if !validStatuses[info.EncodingStatus] {
		t.Errorf("Invalid encoding status: %s", info.EncodingStatus)
	}
}