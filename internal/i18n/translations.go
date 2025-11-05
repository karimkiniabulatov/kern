package i18n

var translations = map[string]map[string]string{
	"en": {
		"title":           "kern - System Monitor",
		"cpu_info":        "CPU Information",
		"memory_info":     "Memory Information", 
		"disk_info":       "Disk Information",
		"network_info":    "Network Information",
		"press_quit":      "Press 'q' to quit",
		"refresh_every":   "Auto-refresh every",
		"seconds":         "seconds",
		"overall_usage":   "Overall Usage",
		"core":            "Core",
		"ram":             "RAM",
		"swap":            "Swap",
	},
	"ru": {
		"title":           "kern - Мониторинг системы",
		"cpu_info":        "Информация о процессоре",
		"memory_info":     "Информация о памяти", 
		"disk_info":       "Дисковая информация",
		"network_info":    "Сетевая информация",
		"press_quit":      "Нажмите 'q' для выхода",
		"refresh_every":   "Автообновление каждые",
		"seconds":         "секунд",
		"overall_usage":   "Общая загрузка",
		"core":            "Ядро",
		"ram":             "ОЗУ",
		"swap":            "Своп",
	},
}

func GetTranslation(lang, key string) string {
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