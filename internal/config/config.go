package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Language    string `json:"language"`
	RefreshRate int    `json:"refresh_rate"`
	Colors      bool   `json:"colors"`
	Theme       string `json:"theme"`
	DetailedCPU bool   `json:"detailed_cpu"` // Добавляем флаг для детального CPU
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
	}
}