package config

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/karimkiniabulatov/kern/internal/i18n"
)

type Config struct {
    Language    string `json:"language"`
    RefreshRate int    `json:"refresh_rate"`
    Colors      bool   `json:"colors"`
    Theme       string `json:"theme"`
    DetailedCPU bool   `json:"detailed_cpu"`
    ShowDisk    bool   `json:"show_disk"`
    ShowCPU     bool   `json:"show_cpu"`
    ShowMem     bool   `json:"show_mem"`
    ShowNet     bool   `json:"show_net"`
    ShowGPU     bool   `json:"show_gpu"`     // NEW
    ShowAI      bool   `json:"show_ai"`      // NEW
    ShowMining  bool   `json:"show_mining"`  // NEW
}

func (c *Config) T(key string) string {
    return i18n.GetTranslation(c.Language, key)
}

func Load(language string) (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return getDefaultConfig(language), nil
	}

	configFile := filepath.Join(configPath, "kern.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return getDefaultConfig(language), nil
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return getDefaultConfig(language), nil
	}

	// Override language if provided via command line
	if language != "" {
		config.Language = language
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configPath, "kern.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "kern"), nil
}

func getDefaultConfig(language string) *Config {
    if language == "" {
        language = "en" // Default to English
    }
    
    return &Config{
        Language:    language,
        RefreshRate: 2,
        Colors:      true,
        Theme:       "default",
        DetailedCPU: false,
        ShowDisk:    true,  
        ShowCPU:     true,  
        ShowMem:     true,   
        ShowNet:     true,
        ShowGPU:     false,  // NEW: GPU disabled by default (requires nvidia-smi)
        ShowAI:      false,  // NEW: AI disabled by default  
        ShowMining:  false,  // NEW: Mining disabled by default
    }
}