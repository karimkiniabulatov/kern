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
	ShowGPU     bool   `json:"show_gpu"`
	ShowAI      bool   `json:"show_ai"`
	ShowMining  bool   `json:"show_mining"`
	
	// NEW: User preferences for default view
	LastUsedModules *LastUsedModules `json:"last_used_modules,omitempty"`
	
	// NEW: Гистограмма настроек
	ShowHistograms    bool `json:"show_histograms"`
	HistogramSegments int  `json:"histogram_segments"` // Количество сегментов в гистограмме
}

type LastUsedModules struct {
	ShowDisk   bool `json:"show_disk"`
	ShowCPU    bool `json:"show_cpu"`
	ShowMem    bool `json:"show_mem"`
	ShowNet    bool `json:"show_net"`
	ShowGPU    bool `json:"show_gpu"`
	ShowAI     bool `json:"show_ai"`
	ShowMining bool `json:"show_mining"`
}

func (c *Config) T(key string) string {
	return i18n.GetTranslation(c.Language, key)
}

// UpdateLastUsedModules updates the last used module preferences
func (c *Config) UpdateLastUsedModules(showDisk, showCPU, showMem, showNet, showGPU, showAI, showMining bool) {
	if c.LastUsedModules == nil {
		c.LastUsedModules = &LastUsedModules{}
	}
	
	c.LastUsedModules.ShowDisk = showDisk
	c.LastUsedModules.ShowCPU = showCPU
	c.LastUsedModules.ShowMem = showMem
	c.LastUsedModules.ShowNet = showNet
	c.LastUsedModules.ShowGPU = showGPU
	c.LastUsedModules.ShowAI = showAI
	c.LastUsedModules.ShowMining = showMining
	
	// Save the updated config
	c.Save()
}

// UseLastUsedModules applies the last used module preferences to the current config
func (c *Config) UseLastUsedModules() {
	if c.LastUsedModules != nil {
		c.ShowDisk = c.LastUsedModules.ShowDisk
		c.ShowCPU = c.LastUsedModules.ShowCPU
		c.ShowMem = c.LastUsedModules.ShowMem
		c.ShowNet = c.LastUsedModules.ShowNet
		c.ShowGPU = c.LastUsedModules.ShowGPU
		c.ShowAI = c.LastUsedModules.ShowAI
		c.ShowMining = c.LastUsedModules.ShowMining
	}
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
		ShowGPU:     false,
		ShowAI:      false,
		ShowMining:  false,
		ShowHistograms: true,
		HistogramSegments: 20, // 20 сегментов = 5% на элемент
	}
}

// GetDefaultConfig returns a default configuration (public version for main package)
func GetDefaultConfig(language string) *Config {
	return getDefaultConfig(language)
}