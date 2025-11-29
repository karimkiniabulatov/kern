package mem

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

type MemoryInfo struct {
	Total            string
	Used             string
	UsedBytes        uint64  // НОВОЕ: размер в байтах для точных вычислений
	Free             string
	Available        string
	AvailableBytes   uint64  // НОВОЕ: размер в байтах
	SwapTotal        string
	SwapUsed         string
	SwapFree         string
	UsagePercent     float64
	SwapUsagePercent float64
	Modules          []MemoryModule
	
	// НОВЫЕ ПОЛЯ для детальной информации
	UsedByProcesses  string  // Память, используемая процессами (исключая кэш/буферы)
	Cached           string  // Кэшированная память
	Buffers          string  // Буферы
	Active           string  // Активная память
	Inactive         string  // Неактивная память
	Shared           string  // Разделяемая память
}

type MemoryModule struct {
	Slot         string
	Size         string
	SizeBytes    uint64
	Type         string
	Speed        string
	Manufacturer string
	PartNumber   string
	SerialNumber string // ДОБАВЛЕНО: серийный номер
	Timings      string
	UsagePercent float64
}

var (
    lastMemUpdate   time.Time
    memCache        *MemoryInfo
    memMutex        sync.RWMutex
    cacheDuration   = 1 * time.Second // Увеличиваем кэш до 1 секунды для стабильности
)

func Summary() (*MemoryInfo, error) {
    memMutex.RLock()
    now := time.Now()
    if memCache != nil && now.Sub(lastMemUpdate) < cacheDuration {
        defer memMutex.RUnlock()
        return memCache, nil
    }
    memMutex.RUnlock()

    info, err := getMemoryInfo()
    if err != nil {
        // Возвращаем кэшированные данные при ошибке, если они есть
        memMutex.RLock()
        if memCache != nil {
            defer memMutex.RUnlock()
            return memCache, nil
        }
        memMutex.RUnlock()
        
        // Или создаем минимальный набор данных
        return &MemoryInfo{
            Total:            "0",
            Used:             "0",
            UsedBytes:        0,
            Free:             "0",
            Available:        "0",
            AvailableBytes:   0,
            SwapTotal:        "0",
            SwapUsed:         "0",
            SwapFree:         "0",
            UsagePercent:     0.0,
            SwapUsagePercent: 0.0,
            Modules:          []MemoryModule{},
            UsedByProcesses:  "0",
            Cached:           "0",
            Buffers:          "0",
            Active:           "0",
            Inactive:         "0",
            Shared:           "0",
        }, nil
    }

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

    // ИСПРАВЛЕННЫЙ РАСЧЕТ: используем реально доступную память для расчета использования
    usedMemory := virtMem.Total - virtMem.Available
    usagePercent := float64(usedMemory) / float64(virtMem.Total) * 100
    
    // Гарантируем корректные границы
    if usagePercent < 0 {
        usagePercent = 0.0
    }
    if usagePercent > 100 {
        usagePercent = 100.0
    }

    // Память, используемая только процессами (исключая кэш)
    usedByProcesses := virtMem.Used

    modules, err := getMemoryModules()
    if err != nil {
        modules = []MemoryModule{}
    }

    // Получаем детальную информацию о памяти
    cached, buffers, active, inactive, shared := getDetailedMemoryInfo()

    // Обновляем процент использования для каждого модуля памяти
    for i := range modules {
        if modules[i].SizeBytes > 0 {
            modules[i].UsagePercent = calculateModuleUsage(
                modules[i].SizeBytes, 
                virtMem.Total, 
                usedMemory, // используем corrected usedMemory
            )
        } else {
            modules[i].UsagePercent = 0.0
        }
    }

    info := &MemoryInfo{
        Total:            formatBytes(virtMem.Total),
        Used:             formatBytes(usedMemory),
        UsedBytes:        usedMemory,
        Free:             formatBytes(virtMem.Free),
        Available:        formatBytes(virtMem.Available),
        AvailableBytes:   virtMem.Available,
        SwapTotal:        formatBytes(swapMem.Total),
        SwapUsed:         formatBytes(swapMem.Used),
        SwapFree:         formatBytes(swapMem.Free),
        UsagePercent:     usagePercent, // Используем корректное значение из gopsutil
        SwapUsagePercent: swapMem.UsedPercent,
        Modules:          modules,
        UsedByProcesses:  formatBytes(usedByProcesses),
        Cached:           formatBytes(cached),
        Buffers:          formatBytes(buffers),
        Active:           formatBytes(active),
        Inactive:         formatBytes(inactive),
        Shared:           formatBytes(shared),
    }

    // Дополнительная проверка для Swap
    if info.SwapUsagePercent < 0 {
        info.SwapUsagePercent = 0.0
    }

    return info, nil
}

// НОВАЯ ФУНКЦИЯ: Получение детальной информации о памяти
func getDetailedMemoryInfo() (uint64, uint64, uint64, uint64, uint64) {
	var cached, buffers, active, inactive, shared uint64

	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/meminfo")
		if err != nil {
			// Возвращаем нулевые значения при ошибке
			return 0, 0, 0, 0, 0
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
			value *= 1024 // Конвертируем из KB в байты

			switch fields[0] {
			case "Cached:":
				cached = value
			case "Buffers:":
				buffers = value
			case "Active:":
				active = value
			case "Inactive:":
				inactive = value
			case "Shmem:":
				shared = value
			}
		}
	case "windows":
		// Для Windows используем другую логику
		if output, err := exec.Command("wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize", "/value").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			var freeMem, totalMem uint64
			
			for _, line := range lines {
				if strings.HasPrefix(line, "FreePhysicalMemory=") {
					if val, err := strconv.ParseUint(strings.TrimPrefix(line, "FreePhysicalMemory="), 10, 64); err == nil {
						freeMem = val * 1024 // KB to bytes
					}
				} else if strings.HasPrefix(line, "TotalVisibleMemorySize=") {
					if val, err := strconv.ParseUint(strings.TrimPrefix(line, "TotalVisibleMemorySize="), 10, 64); err == nil {
						totalMem = val * 1024 // KB to bytes
					}
				}
			}
			
			// Для Windows приблизительные значения
			if totalMem > 0 && freeMem > 0 {
				used := totalMem - freeMem
				cached = used * 20 / 100 // Предполагаем 20% кэша
				buffers = used * 5 / 100 // Предполагаем 5% буферов
				active = used * 70 / 100 // Предполагаем 70% активной памяти
			}
		}
	case "darwin":
		// Для macOS используем vm_stat
		if output, err := exec.Command("vm_stat").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			pageSize := uint64(4096) // Стандартный размер страницы в macOS
			
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) < 2 {
					continue
				}
				
				value, err := strconv.ParseUint(strings.TrimSuffix(fields[1], "."), 10, 64)
				if err != nil {
					continue
				}
				value *= pageSize
				
				switch fields[0] {
				case "Pages", "Pages:":
					// Общая информация
				case "FileCache:":
					cached = value
				case "Purgeable":
					inactive = value
				}
			}
		}
	}

	return cached, buffers, active, inactive, shared
}

func getMemoryModules() ([]MemoryModule, error) {
	switch runtime.GOOS {
	case "linux":
		return parseDMIDecode()
	case "windows":
		return parseWMICMemory()
	case "darwin":
		return parseMacMemory()
	default:
		return []MemoryModule{}, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func parseDMIDecode() ([]MemoryModule, error) {
	// Проверяем доступность dmidecode
	if _, err := exec.LookPath("dmidecode"); err != nil {
		return []MemoryModule{}, fmt.Errorf("dmidecode not available")
	}

	cmd := exec.Command("sudo", "dmidecode", "--type", "memory")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// Пробуем без sudo
		cmd = exec.Command("dmidecode", "--type", "memory")
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			return []MemoryModule{}, err
		}
	}

	output := out.String()
	return parseDMIDecodeOutput(output), nil
}

func parseDMIDecodeOutput(output string) []MemoryModule {
    var modules []MemoryModule
    blocks := strings.Split(output, "Memory Device")
    
    // Пропускаем первый блок (заголовок)
    for i := 1; i < len(blocks); i++ {
        module := MemoryModule{}
        lines := strings.Split(blocks[i], "\n")
        
        for _, line := range lines {
            line = strings.TrimSpace(line)
            
            switch {
            case strings.HasPrefix(line, "Size:"):
                sizeStr := strings.TrimSpace(strings.TrimPrefix(line, "Size:"))
                module.Size = sizeStr
                module.SizeBytes = parseMemorySize(sizeStr)
            case strings.HasPrefix(line, "Type:"):
                module.Type = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
            case strings.HasPrefix(line, "Speed:"):
                module.Speed = strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
            case strings.HasPrefix(line, "Manufacturer:"):
                module.Manufacturer = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
            case strings.HasPrefix(line, "Part Number:"):
                module.PartNumber = strings.TrimSpace(strings.TrimPrefix(line, "Part Number:"))
            case strings.HasPrefix(line, "Serial Number:"):
                serial := strings.TrimSpace(strings.TrimPrefix(line, "Serial Number:"))
                if serial != "" && serial != "Unknown" && serial != "Not Specified" {
                    module.SerialNumber = serial
                }
            case strings.HasPrefix(line, "Locator:"):
                module.Slot = strings.TrimSpace(strings.TrimPrefix(line, "Locator:"))
            case strings.HasPrefix(line, "Bank Locator:"):
                // Используем Bank Locator как дополнительный идентификатор
                if module.Slot == "" {
                    module.Slot = strings.TrimSpace(strings.TrimPrefix(line, "Bank Locator:"))
                }
            }
        }
        
        // ВАЖНОЕ ИСПРАВЛЕНИЕ: принимаем ВСЕ модули, даже с неизвестным размером
        // но фильтруем полностью пустые слоты
        if module.Slot != "" {
            // Если размер не определен, помечаем как "Unknown"
            if module.Size == "" || module.Size == "No Module Installed" {
                module.Size = "Unknown"
                module.SizeBytes = 0
            }
            
            // Проверяем уникальность по слоту + серийному номеру
            if isTrulyUniqueModule(modules, module.Slot, module.SerialNumber) {
                modules = append(modules, module)
            }
        }
    }
    
    return modules
}

// НОВАЯ ФУНКЦИЯ для строгой проверки уникальности
func isTrulyUniqueModule(modules []MemoryModule, slot string, serial string) bool {
    for _, m := range modules {
        // Считаем уникальным, если отличается слот ИЛИ серийный номер
        if m.Slot == slot && m.SerialNumber == serial {
            return false
        }
        // Если слот тот же, но серийный номер разный - это разные модули
        if m.Slot == slot && m.SerialNumber != serial {
            return true // это другой модуль в том же слоте? маловероятно, но возможно
        }
    }
    return true
}

// УЛУЧШЕННАЯ ФУНКЦИЯ расчета использования модулей
func calculateModuleUsage(moduleSize uint64, totalSystemMemory uint64, usedSystemMemory uint64) float64 {
    if totalSystemMemory == 0 || moduleSize == 0 {
        return 0.0
    }
    
    // Более точный расчет: предполагаем равномерное распределение данных по модулям
    // Это упрощение, но лучше чем постоянный 0.8%
    moduleRatio := float64(moduleSize) / float64(totalSystemMemory)
    usagePercent := (float64(usedSystemMemory) / float64(totalSystemMemory)) * 100
    
    // Возвращаем общий процент использования, скорректированный на долю модуля
    return usagePercent * moduleRatio
}

// Улучшенная функция для альтернативного получения информации о памяти
func parseAlternativeMemoryInfo() []MemoryModule {
    var modules []MemoryModule
    
    // Пробуем разные методы получения информации о памяти
    switch runtime.GOOS {
    case "linux":
        // Метод 1: Попробуем через lshw
        if output, err := exec.Command("lshw", "-class", "memory", "-quiet").Output(); err == nil {
            modules = parseLshwOutput(string(output))
        }
        
        // Метод 2: Если lshw не сработал, пробуем через /proc/meminfo и dmidecode без sudo
        if len(modules) == 0 {
            if output, err := exec.Command("dmidecode", "-t", "memory").Output(); err == nil {
                modules = parseDMIDecodeOutput(string(output))
            }
        }
        
        // Метод 3: Парсим /sys/bus/drivers для получения информации о слотах
        if len(modules) == 0 {
            modules = parseSysfsMemoryInfo()
        }
        
    case "windows":
        // Для Windows используем WMI
        cmd := exec.Command("wmic", "memorychip", "get", "BankLabel,Capacity,Speed,MemoryType,Manufacturer,PartNumber,SerialNumber", "/format:csv")
        if output, err := cmd.Output(); err == nil {
            modules = parseWMICMemoryOutput(string(output))
        }
    }
    
    return modules
}

// НОВАЯ ФУНКЦИЯ для парсинга информации из sysfs (Linux)
func parseSysfsMemoryInfo() []MemoryModule {
    var modules []MemoryModule
    
    // Ищем информацию о памяти в /sys/devices
    sysfsPath := "/sys/devices/system/edac/mc"
    if _, err := os.Stat(sysfsPath); err == nil {
        // Есть информация о контроллере памяти
        mcDirs, err := filepath.Glob(filepath.Join(sysfsPath, "mc*"))
        if err == nil {
            for _, mcDir := range mcDirs {
                dimmDirs, _ := filepath.Glob(filepath.Join(mcDir, "dimm*"))
                for i, dimmDir := range dimmDirs {
                    module := MemoryModule{
                        Slot:  fmt.Sprintf("DIMM%d", i),
                        Size:  "Unknown",
                        Type:  "DDR",
                    }
                    
                    // Пытаемся получить размер из разных мест
                    if sizeFile, err := os.ReadFile(filepath.Join(dimmDir, "size")); err == nil {
                        if sizeMB, err := strconv.Atoi(strings.TrimSpace(string(sizeFile))); err == nil {
                            module.SizeBytes = uint64(sizeMB) * 1024 * 1024
                            module.Size = formatBytes(module.SizeBytes)
                        }
                    }
                    
                    modules = append(modules, module)
                }
            }
        }
    }
    
    return modules
}

// НОВАЯ ФУНКЦИЯ для парсинга вывода lshw
func parseLshwOutput(output string) []MemoryModule {
    var modules []MemoryModule
    lines := strings.Split(output, "\n")
    moduleCount := 0
    
    for _, line := range lines {
        if strings.Contains(line, "memory") && (strings.Contains(line, "GiB") || strings.Contains(line, "MiB")) {
            fields := strings.Fields(line)
            if len(fields) >= 3 {
                module := MemoryModule{
                    Slot:  fmt.Sprintf("DIMM%d", moduleCount),
                    Size:  fields[2],
                    Type:  "DDR4", // Предполагаем
                    Speed: "Unknown",
                }
                module.SizeBytes = parseMemorySize(module.Size)
                modules = append(modules, module)
                moduleCount++
            }
        }
    }
    
    return modules
}

// НОВАЯ ФУНКЦИЯ для парсинга вывода WMIC
func parseWMICMemoryOutput(output string) []MemoryModule {
    var modules []MemoryModule
    lines := strings.Split(output, "\n")
    
    for i, line := range lines {
        if i == 0 || strings.TrimSpace(line) == "" {
            continue
        }
        fields := strings.Split(line, ",")
        if len(fields) >= 7 {
            module := MemoryModule{
                Slot:         strings.TrimSpace(fields[1]),
                Manufacturer: strings.TrimSpace(fields[5]),
                PartNumber:   strings.TrimSpace(fields[6]),
                SerialNumber: strings.TrimSpace(fields[7]),
            }
            
            // Parse capacity
            if capacity, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64); err == nil {
                module.SizeBytes = capacity
                module.Size = formatBytes(capacity)
            }
            
            // Parse memory type
            if memType, err := strconv.Atoi(strings.TrimSpace(fields[3])); err == nil {
                module.Type = memoryTypeToString(memType)
            }
            
            // Parse speed
            if speed, err := strconv.Atoi(strings.TrimSpace(fields[4])); err == nil {
                module.Speed = fmt.Sprintf("%d MHz", speed)
            }
            
            modules = append(modules, module)
        }
    }
    
    return modules
}

func parseWMICMemory() ([]MemoryModule, error) {
	cmd := exec.Command("wmic", "memorychip", "get", 
		"BankLabel,Capacity,MemoryType,Speed,Manufacturer,PartNumber,SerialNumber", "/format:csv")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []MemoryModule{}, err
	}

	output := out.String()
	return parseWMICOutput(output), nil
}

func parseWMICOutput(output string) []MemoryModule {
	var modules []MemoryModule
	lines := strings.Split(output, "\n")
	
	for _, line := range lines[1:] {
		fields := strings.Split(line, ",")
		if len(fields) < 7 {
			continue
		}
		
		module := MemoryModule{
			Slot:         strings.TrimSpace(fields[1]),
			Manufacturer: strings.TrimSpace(fields[4]),
			PartNumber:   strings.TrimSpace(fields[5]),
			SerialNumber: strings.TrimSpace(fields[6]), // Добавляем серийный номер
		}
		
		// Parse capacity
		if capacity, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64); err == nil {
			module.SizeBytes = capacity
			module.Size = formatBytes(capacity)
		}
		
		// Parse memory type
		if memType, err := strconv.Atoi(strings.TrimSpace(fields[3])); err == nil {
			module.Type = memoryTypeToString(memType)
		}
		
		// Parse speed
		if speed, err := strconv.Atoi(strings.TrimSpace(fields[4])); err == nil {
			module.Speed = fmt.Sprintf("%d MHz", speed)
		}
		
		modules = append(modules, module)
	}
	
	return modules
}

func parseMacMemory() ([]MemoryModule, error) {
	cmd := exec.Command("system_profiler", "SPMemoryDataType")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []MemoryModule{}, err
	}

	output := out.String()
	return parseMacMemoryOutput(output), nil
}

func parseMacMemoryOutput(output string) []MemoryModule {
	var modules []MemoryModule
	
	lines := strings.Split(output, "\n")
	var currentModule MemoryModule
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		switch {
		case strings.HasPrefix(line, "BANK"):
			if currentModule.Slot != "" {
				modules = append(modules, currentModule)
			}
			currentModule = MemoryModule{Slot: strings.TrimSpace(line)}
			
		case strings.HasPrefix(line, "Size:"):
			sizeStr := strings.TrimSpace(strings.TrimPrefix(line, "Size:"))
			currentModule.Size = sizeStr
			currentModule.SizeBytes = parseMemorySize(sizeStr)
		case strings.HasPrefix(line, "Type:"):
			currentModule.Type = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
		case strings.HasPrefix(line, "Speed:"):
			currentModule.Speed = strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
		case strings.HasPrefix(line, "Manufacturer:"):
			currentModule.Manufacturer = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
		case strings.HasPrefix(line, "Part Number:"):
			currentModule.PartNumber = strings.TrimSpace(strings.TrimPrefix(line, "Part Number:"))
		case strings.HasPrefix(line, "Serial Number:"):
			currentModule.SerialNumber = strings.TrimSpace(strings.TrimPrefix(line, "Serial Number:"))
		}
	}
	
	if currentModule.Slot != "" {
		modules = append(modules, currentModule)
	}
	
	return modules
}

// Функция parseMemorySize без изменений...
func parseMemorySize(sizeStr string) uint64 {
	if sizeStr == "" || sizeStr == "No Module Installed" {
		return 0
	}

	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))
	
	var value float64
	var unit string
	
	_, err := fmt.Sscanf(sizeStr, "%f %s", &value, &unit)
	if err != nil {
		for i, char := range sizeStr {
			if char < '0' || char > '9' {
				if char == '.' {
					continue
				}
				valueStr := sizeStr[:i]
				unit = sizeStr[i:]
				value, _ = strconv.ParseFloat(valueStr, 64)
				break
			}
		}
	}

	switch unit {
	case "tb", "t":
		return uint64(value * 1024 * 1024 * 1024 * 1024)
	case "gb", "g":
		return uint64(value * 1024 * 1024 * 1024)
	case "mb", "m":
		return uint64(value * 1024 * 1024)
	case "kb", "k":
		return uint64(value * 1024)
	default:
		return uint64(value)
	}
}

func memoryTypeToString(memType int) string {
	switch memType {
	case 20: return "DDR"
	case 21: return "DDR2"
	case 24: return "DDR3"
	case 26: return "DDR4"
	case 34: return "DDR5"
	default: return fmt.Sprintf("Unknown (%d)", memType)
	}
}

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