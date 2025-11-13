package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := getDefaultConfig("en")
	
	if cfg.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", cfg.Language)
	}
	
	if cfg.RefreshRate != 2 {
		t.Errorf("Expected refresh rate 2, got %d", cfg.RefreshRate)
	}
}

func TestConfigSaveLoad(t *testing.T) {
	cfg := &Config{
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

	loadedCfg, err := Load("")
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}

	if loadedCfg.Language != cfg.Language {
		t.Errorf("Language mismatch: expected %s, got %s", cfg.Language, loadedCfg.Language)
	}
}

func TestModulePreferences(t *testing.T) {
	cfg := getDefaultConfig("en")
	
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