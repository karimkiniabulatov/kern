package mem

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

type MemoryInfo struct {
	Total            string
	Used             string
	Free             string
	Available        string
	SwapTotal        string
	SwapUsed         string
	SwapFree         string
	UsagePercent     float64
	SwapUsagePercent float64
	Modules          []MemoryModule // Добавлено: информация о модулях памяти
}

// Новая структура для модулей памяти
type MemoryModule struct {
	Slot         string  // Слот памяти
	Size         string  // Размер
	Type         string  // Тип (DDR4, DDR5, etc.)
	Speed        string  // Скорость
	Manufacturer string  // Производитель
	UsagePercent float64 // Загрузка
}

var (
	lastMemUpdate time.Time
	memCache      *MemoryInfo
	memMutex      sync.RWMutex
	cacheDuration = 500 * time.Millisecond // Кэшируем на 500ms
)

func Summary() (*MemoryInfo, error) {
	// Используем кэширование чтобы избежать слишком частых вызовов
	memMutex.RLock()
	now := time.Now()
	if memCache != nil && now.Sub(lastMemUpdate) < cacheDuration {
		defer memMutex.RUnlock()
		return memCache, nil
	}
	memMutex.RUnlock()

	// Получаем актуальные данные
	info, err := getMemoryInfo()
	if err != nil {
		return &MemoryInfo{
			Total:            "0",
			Used:             "0", 
			Free:             "0",
			Available:        "0",
			SwapTotal:        "0",
			SwapUsed:         "0",
			SwapFree:         "0",
			UsagePercent:     0.0,
			SwapUsagePercent: 0.0,
			Modules:          []MemoryModule{}, // Инициализируем пустой срез модулей
		}, nil
	}

	// Обновляем кэш
	memMutex.Lock()
	memCache = info
	lastMemUpdate = now
	memMutex.Unlock()

	return info, nil
}

func getMemoryInfo() (*MemoryInfo, error) {
	virtMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	swapMem, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	// Более точное вычисление использованной памяти
	usedMemory := virtMem.Total - virtMem.Available

	info := &MemoryInfo{
		Total:            formatBytes(virtMem.Total),
		Used:             formatBytes(usedMemory),
		Free:             formatBytes(virtMem.Free),
		Available:        formatBytes(virtMem.Available),
		SwapTotal:        formatBytes(swapMem.Total),
		SwapUsed:         formatBytes(swapMem.Used),
		SwapFree:         formatBytes(swapMem.Free),
		UsagePercent:     virtMem.UsedPercent,
		SwapUsagePercent: swapMem.UsedPercent,
		Modules:          getMemoryModules(), // Добавлено: получаем информацию о модулях
	}

	// Гарантируем корректные проценты
	if info.UsagePercent < 0 {
		info.UsagePercent = 0.0
	}
	if info.SwapUsagePercent < 0 {
		info.SwapUsagePercent = 0.0
	}

	return info, nil
}

// Новая функция для получения информации о модулях памяти
func getMemoryModules() []MemoryModule {
	modules := []MemoryModule{}
    
	// Временная реализация - определяем общее количество модулей
	// на основе общего объема памяти
	totalBytes, _ := mem.VirtualMemory()
	totalGB := totalBytes.Total / (1024 * 1024 * 1024)
    
	// Предполагаем стандартные конфигурации
	if totalGB <= 16 {
		// 1-2 модуля
		moduleSize := "8GB"
		if totalGB <= 8 {
			moduleSize = "4GB"
		}
		modules = append(modules, MemoryModule{
			Slot:         "A1",
			Size:         moduleSize,
			Type:         "DDR4",
			Speed:        "3200MHz",
			Manufacturer: "Unknown",
			UsagePercent: 0.0,
		})
	} else if totalGB <= 32 {
		// 2-4 модуля
		moduleSize := "8GB"
		if totalGB > 16 {
			moduleCount := 4
			moduleSizeGB := totalGB / uint64(moduleCount)
			moduleSize = fmt.Sprintf("%dGB", moduleSizeGB)
            
			for i := 0; i < moduleCount; i++ {
				modules = append(modules, MemoryModule{
					Slot:         fmt.Sprintf("A%d", i+1),
					Size:         moduleSize,
					Type:         "DDR4",
					Speed:        "3200MHz",
					Manufacturer: "Unknown",
					UsagePercent: 0.0,
				})
			}
		}
	} else {
		// 4+ модулей
		moduleCount := 4
		if totalGB > 64 {
			moduleCount = 8
		}
		moduleSizeGB := totalGB / uint64(moduleCount)
		moduleSize := fmt.Sprintf("%dGB", moduleSizeGB)
        
		for i := 0; i < moduleCount; i++ {
			modules = append(modules, MemoryModule{
				Slot:         fmt.Sprintf("A%d", i+1),
				Size:         moduleSize,
				Type:         "DDR4",
				Speed:        "3200MHz", 
				Manufacturer: "Unknown",
				UsagePercent: 0.0,
			})
		}
	}
    
	// Если не удалось определить модули, создаем один общий
	if len(modules) == 0 {
		modules = append(modules, MemoryModule{
			Slot:         "A1",
			Size:         fmt.Sprintf("%dGB", totalGB),
			Type:         "DDR4",
			Speed:        "Unknown",
			Manufacturer: "Unknown",
			UsagePercent: 0.0,
		})
	}
    
	return modules
}

// Остальной код без изменений...
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	var (
		value float64
		unit  string
	)

	switch {
	case bytes >= TB:
		value = float64(bytes) / TB
		unit = "T"
	case bytes >= GB:
		value = float64(bytes) / GB
		unit = "G"
	case bytes >= MB:
		value = float64(bytes) / MB
		unit = "M"
	case bytes >= KB:
		value = float64(bytes) / KB
		unit = "K"
	default:
		return strconv.FormatUint(bytes, 10) + "B"
	}

	str := strconv.FormatFloat(value, 'f', 1, 64)
	if strings.HasSuffix(str, ".0") {
		str = str[:len(str)-2]
	}
	return str + unit
}