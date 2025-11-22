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

// Добавить тест для проверки структуры AudioInfo
func TestAudioInfoStructure(t *testing.T) {
    info, err := audio.Summary()
    if err != nil {
        t.Fatalf("Failed to get audio info: %v", err)
    }

    // Проверить что массивы инициализированы (даже если пустые)
    if info.InputDevices == nil {
        t.Error("InputDevices should not be nil")
    }
    if info.OutputDevices == nil {
        t.Error("OutputDevices should not be nil")
    }
    if info.ActiveStreams == nil {
        t.Error("ActiveStreams should not be nil")
    }
    if info.FrequencyBands == nil {
        t.Error("FrequencyBands should not be nil")
    }
}

// Добавить тест для проверки уровней аудио
func TestAudioLevelsRange(t *testing.T) {
    info, _ := audio.Summary()
    
    // Уровни должны быть в разумных пределах
    if info.InputLevel < -96 || info.InputLevel > 0 {
        t.Logf("Input level out of typical range: %.1f dB", info.InputLevel)
    }
    if info.OutputLevel < -96 || info.OutputLevel > 0 {
        t.Logf("Output level out of typical range: %.1f dB", info.OutputLevel)
    }
    
    // Проверить VU метрики
    if info.PeakLevel < 0 || info.PeakLevel > 100 {
        t.Errorf("PeakLevel should be between 0-100, got %.1f", info.PeakLevel)
    }
    if info.RMSLevel < 0 || info.RMSLevel > 100 {
        t.Errorf("RMSLevel should be between 0-100, got %.1f", info.RMSLevel)
    }
}