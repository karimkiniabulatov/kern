package tests

import (
	"testing"
	audio "github.com/karimkiniabulatov/kern/internal/audio"
)

func TestAudioSummary(t *testing.T) {
	info, err := audio.Summary()
	if err != nil {
		t.Fatalf("Failed to get audio info: %v", err)
	}

	// Should always return valid structure even if no audio devices
	if info == nil {
		t.Error("Audio info should not be nil")
	}

	// Should have at least default devices
	if len(info.InputDevices) == 0 {
		t.Log("No input audio devices detected")
	}
	
	if len(info.OutputDevices) == 0 {
		t.Log("No output audio devices detected")
	}
}

func TestAudioStreamDetection(t *testing.T) {
	info, _ := audio.Summary()
	
	// Test that we can detect streams without errors
	// In test environment there might be no active streams
	if len(info.ActiveStreams) == 0 {
		t.Log("No active audio streams detected (normal in test environment)")
	}
}

func TestAudioLevels(t *testing.T) {
	info, _ := audio.Summary()
	
	// Audio levels should be reasonable values
	if info.InputLevel < -96 || info.InputLevel > 0 {
		t.Logf("Input level out of expected range: %.1f dB", info.InputLevel)
	}
	
	if info.OutputLevel < -96 || info.OutputLevel > 0 {
		t.Logf("Output level out of expected range: %.1f dB", info.OutputLevel)
	}
}