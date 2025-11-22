package tests

import (
	"testing"
	config "github.com/karimkiniabulatov/kern/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.GetDefaultConfig("en")
	
	if cfg.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", cfg.Language)
	}
	
	if cfg.RefreshRate != 2 {
		t.Errorf("Expected refresh rate 2, got %d", cfg.RefreshRate)
	}
}

func TestConfigSaveLoad(t *testing.T) {
	cfg := &config.Config{
		Language:    "ru",
		RefreshRate: 5,
		Colors:      true,
		ShowDisk:    true,
		ShowCPU:     true,
		ShowMem:     true,
		ShowNet:     true,
	}

	err := cfg.Save()
	if err != nil {
		t.Errorf("Failed to save config: %v", err)
	}

	loadedCfg, err := config.Load("")
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}

	if loadedCfg.Language != cfg.Language {
		t.Errorf("Language mismatch: expected %s, got %s", cfg.Language, loadedCfg.Language)
	}
}

func TestModulePreferences(t *testing.T) {
	cfg := config.GetDefaultConfig("en")
	
	// Test updating last used modules
	cfg.UpdateLastUsedModules(true, false, true, false, true, false, true)
	
	if cfg.LastUsedModules == nil {
		t.Error("LastUsedModules should not be nil")
	}
	
	if !cfg.LastUsedModules.ShowDisk {
		t.Error("ShowDisk should be true")
	}
	
	if cfg.LastUsedModules.ShowCPU {
		t.Error("ShowCPU should be false")
	}
}

// Добавить тест для новых полей конфигурации
func TestConfigNewFields(t *testing.T) {
    cfg := config.GetDefaultConfig("en")
    
    // Проверить новые поля гистограмм
    if !cfg.ShowHistograms {
        t.Error("ShowHistograms should be true by default")
    }
    if cfg.HistogramSegments != 20 {
        t.Errorf("Expected HistogramSegments 20, got %d", cfg.HistogramSegments)
    }
    
    // Проверить новые модули
    if cfg.ShowGPU {
        t.Error("ShowGPU should be false by default")
    }
    if cfg.ShowAI {
        t.Error("ShowAI should be false by default")
    }
    if cfg.ShowMining {
        t.Error("ShowMining should be false by default")
    }
    if cfg.ShowAudio {
        t.Error("ShowAudio should be false by default")
    }
    if cfg.ShowVideo {
        t.Error("ShowVideo should be false by default")
    }
}

// Добавить тест для LastUsedModules
func TestLastUsedModules(t *testing.T) {
    cfg := config.GetDefaultConfig("en")
    
    // Обновить последние использованные модули
    cfg.UpdateLastUsedModules(true, false, true, false, true, false, true, false, true)
    
    if cfg.LastUsedModules == nil {
        t.Error("LastUsedModules should not be nil after update")
    }
    
    // Проверить сохраненные значения
    if !cfg.LastUsedModules.ShowDisk {
        t.Error("ShowDisk should be true in LastUsedModules")
    }
    if cfg.LastUsedModules.ShowCPU {
        t.Error("ShowCPU should be false in LastUsedModules")
    }
    if !cfg.LastUsedModules.ShowMem {
        t.Error("ShowMem should be true in LastUsedModules")
    }
}