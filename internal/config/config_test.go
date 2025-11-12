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