package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	translations = make(map[string]map[string]string)
	mu           sync.RWMutex
	supportedLanguages = []string{
		"en", "ru", "es", "fr", "de", "it", "pt", "nl", "pl", "sv", "da", "no", "fi",
		"cs", "hu", "ro", "bg", "hr", "sk", "sl", "et", "lv", "lt", "uk", "sr", "bs",
		"mk", "sq", "el", "ja", "ko", "zh", "ar", "hi", "id", "vi", "th", "tr", "he",
		"fa", "ur", "bn", "ta", "te", "ml", "kn", "gu", "mr", "pa", "ne",
	}
)

func init() {
	loadTranslations()
}

func loadTranslations() {
	mu.Lock()
	defer mu.Unlock()

	// Поиск файлов переводов в разных местах
	possiblePaths := []string{
		"i18n",
		"/usr/local/share/kern/i18n", 
		"/usr/share/kern/i18n",
		"./i18n",
		"~/.config/kern/i18n",
	}

	// Добавляем путь относительно исполняемого файла
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		possiblePaths = append(possiblePaths, filepath.Join(exeDir, "i18n"))
	}

	for _, basePath := range possiblePaths {
		// Expand ~ в пути
		if basePath[:2] == "~/" {
			if home, err := os.UserHomeDir(); err == nil {
				basePath = filepath.Join(home, basePath[2:])
			}
		}

		files, err := filepath.Glob(filepath.Join(basePath, "active.*.json"))
		if err != nil {
			continue
		}

		for _, file := range files {
			lang := filepath.Base(file)[7:9] // Извлекаем код языка из имени файла
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			var translationData struct {
				Translations map[string]map[string]string `json:"translations"`
			}

			if err := json.Unmarshal(data, &translationData); err != nil {
				continue
			}

			// Flatten структуру переводов
			flatTranslations := make(map[string]string)
			for section, sectionMap := range translationData.Translations {
				for key, value := range sectionMap {
					flatKey := section + "." + key
					flatTranslations[flatKey] = value
				}
			}

			translations[lang] = flatTranslations
		}
	}

	// Добавляем английский как fallback если он не загружен
	if _, exists := translations["en"]; !exists {
		translations["en"] = getDefaultEnglishTranslations()
	}
}

func GetTranslation(lang, key string) string {
	mu.RLock()
	defer mu.RUnlock()

	// Проверяем поддержку языка
	if !IsLanguageSupported(lang) {
		lang = "en"
	}

	// Пробуем запрошенный язык
	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[key]; exists {
			return translation
		}
	}

	// Fallback to English
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[key]; exists {
			return translation
		}
	}

	return key
}

func IsLanguageSupported(lang string) bool {
	for _, supported := range supportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

func GetSupportedLanguages() []string {
	return supportedLanguages
}

// DownloadLanguage загружает языковой пакет с GitHub
func DownloadLanguage(lang string) error {
	if !IsLanguageSupported(lang) {
		return fmt.Errorf("language %s is not supported", lang)
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/karimkiniabulatov/kern/main/i18n/active.%s.json", lang)
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download language pack: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("language pack not found for %s", lang)
	}

	var translationData struct {
		Translations map[string]map[string]string `json:"translations"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&translationData); err != nil {
		return fmt.Errorf("failed to parse language pack: %v", err)
	}

	// Сохраняем локально
	configPath, err := getI18nConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(configPath, fmt.Sprintf("active.%s.json", lang))
	data, err := json.MarshalIndent(translationData, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// Перезагружаем переводы
	loadTranslations()
	
	return nil
}

func getI18nConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "kern", "i18n"), nil
}

func getDefaultEnglishTranslations() map[string]string {
    return map[string]string{
        "common.title":           "kern - System Monitor",
        "common.refresh":         "Refresh",
        "common.seconds":         "seconds",
        "common.error":           "Error",
        "common.unknown":         "Unknown",
        "common.total":           "Total",
        "common.used":            "Used",
        "common.free":            "Free",
        "common.available":       "Available",
        "common.status":          "Status",
        "common.up":              "UP",
        "common.down":            "DOWN",

        "cpu.title":              "CPU Information",
        "cpu.model":              "Model",
        "cpu.cores":              "Cores",
        "cpu.threads":            "Threads",
        "cpu.usage":              "Usage",
        "cpu.frequency":          "Frequency",
        "cpu.load":               "Load",
        "cpu.load_1min":          "1 min",
        "cpu.load_5min":          "5 min",
        "cpu.load_15min":         "15 min",
        "cpu.load_average":       "Load Average",
        "cpu.core":               "Core",
        "cpu.overall_usage":      "Overall Usage",
        "cpu.core_usage":         "Core Usage",

        "memory.title":           "Memory Information",
        "memory.ram":             "RAM",
        "memory.swap":            "Swap",
        "memory.physical":        "Physical Memory",
        "memory.virtual":         "Virtual Memory",

        "disk.title":             "Disk Information",
        "disk.filesystem":        "Filesystem",
        "disk.size":              "Size",
        "disk.mounted":           "Mounted On",
        "disk.usage":             "Usage",
        "disk.storage":           "Storage",

        "network.title":          "Network Information",
        "network.interface":      "Interface",
        "network.received":       "Received",
        "network.sent":           "Sent",
        "network.bytes":          "bytes",
        "network.packets":        "packets",
        "network.errors":         "Errors",
        "network.interfaces":     "Network Interfaces",

        // NEW: GPU translations
        "gpu.title":              "GPU Information",
        "gpu.model":              "Model",
        "gpu.driver":             "Driver",
        "gpu.temperature":        "Temperature",
        "gpu.memory":             "Memory",
        "gpu.utilization":        "Utilization",
        "gpu.power":              "Power",
        "gpu.clock_core":         "Core Clock",
        "gpu.clock_memory":       "Memory Clock",
        "gpu.fan_speed":          "Fan Speed",

        // NEW: AI translations
        "ai.title":               "AI Training",
        "ai.framework":           "Framework",
        "ai.processes":           "Processes",
        "ai.vram":                "VRAM",
        "ai.model":               "Model",
        "ai.batch_size":          "Batch Size",
        "ai.throughput":          "Throughput",
        "ai.epoch":               "Epoch",
        "ai.loss":                "Loss",
        "ai.accuracy":            "Accuracy",
        "ai.training_time":       "Training Time",
        "ai.no_training":         "No AI training detected",

        // NEW: Mining translations
        "mining.title":           "Mining Information",
        "mining.algorithm":       "Algorithm",
        "mining.hashrate":        "Hashrate",
        "mining.shares":          "Shares",
        "mining.temperature":     "Temperature",
        "mining.power":           "Power",
        "mining.efficiency":      "Efficiency",
        "mining.uptime":          "Uptime",
        "mining.revenue_24h":     "24h Revenue",
        "mining.pool":            "Pool",
        "mining.not_detected":    "No mining activity detected",

        "ui.histogram":           "Histogram",
        "ui.vertical":            "Vertical",
        "ui.horizontal":          "Horizontal",
        "ui.loading":             "Loading...",
        "ui.no_data":             "No data",
        "ui.percent":             "percent",
        "ui.press_quit":          "Press 'q' to quit",
        "ui.refresh_every":       "Auto-refresh every",

        "commands.help":          "Help",
        "commands.version":       "Version",
        "commands.all":           "All information",
        "commands.disk_only":     "Disk only",
        "commands.cpu_only":      "CPU only",
        "commands.memory_only":   "Memory only",
        "commands.network_only":  "Network only",
        "commands.gpu_only":      "GPU only",
        "commands.ai_only":       "AI training only",
        "commands.mining_only":   "Mining only",
        "commands.refresh_rate":  "Refresh rate",
        "commands.language":      "Language",
        "commands.remote":        "Remote mode",
        "commands.port":          "Port",

        "errors.command_not_found":    "Command not found: %s",
        "errors.invalid_port":         "Invalid port: %d",
        "errors.invalid_refresh":      "Invalid refresh interval: %d",
        "errors.config_load_failed":   "Failed to load configuration",
        "errors.module_failed":        "Module %s error: %v",
        "errors.nvidia_smi_not_found": "nvidia-smi not found. GPU monitoring requires NVIDIA drivers",
        "errors.rocm_smi_not_found":   "rocm-smi not found. AMD GPU monitoring requires ROCm",

        "status.monitoring_started":   "Monitoring started",
        "status.updating_every":       "Updating every %d seconds",
        "status.press_ctrl_c":         "Press Ctrl+C to exit",
        "status.remote_mode":          "Remote mode on port %d",
        "status.gpu_detected":         "GPU detected: %s",
        "status.ai_training_detected": "AI training processes detected: %d",
        "status.mining_detected":      "Mining activity detected: %s",
    }
}