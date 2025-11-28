package mem

import (
	"bytes"
	"fmt"
	"os/exec"
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
	Free             string
	Available        string
	SwapTotal        string
	SwapUsed         string
	SwapFree         string
	UsagePercent     float64
	SwapUsagePercent float64
	Modules          []MemoryModule
}

type MemoryModule struct {
	Slot         string
	Size         string
	SizeBytes    uint64  // НОВОЕ: размер в байтах для вычислений
	Type         string
	Speed        string
	Manufacturer string
	PartNumber   string
	Timings      string
	UsagePercent float64 // НОВОЕ: процент использования модуля
}

var (
	lastMemUpdate time.Time
	memCache      *MemoryInfo
	memMutex      sync.RWMutex
	cacheDuration = 500 * time.Millisecond
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
			Modules:          []MemoryModule{},
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

	usedMemory := virtMem.Total - virtMem.Available

	modules, err := getMemoryModules()
	if err != nil {
		modules = []MemoryModule{}
	}

	// НОВОЕ: Вычисляем проценты использования для каждого модуля памяти
	totalMemoryBytes := virtMem.Total
	for i := range modules {
		if totalMemoryBytes > 0 && modules[i].SizeBytes > 0 {
			modules[i].UsagePercent = float64(modules[i].SizeBytes) / float64(totalMemoryBytes) * 100
		} else {
			modules[i].UsagePercent = 0.0
		}
	}

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
		Modules:          modules,
	}

	if info.UsagePercent < 0 {
		info.UsagePercent = 0.0
	}
	if info.SwapUsagePercent < 0 {
		info.SwapUsagePercent = 0.0
	}

	return info, nil
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
	cmd := exec.Command("dmidecode", "--type", "memory")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []MemoryModule{}, err
	}

	output := out.String()
	return parseDMIDecodeOutput(output), nil
}

func parseDMIDecodeOutput(output string) []MemoryModule {
	var modules []MemoryModule
	blocks := strings.Split(output, "Memory Device")
	
	for _, block := range blocks[1:] {
		module := MemoryModule{}
		
		lines := strings.Split(block, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			
			switch {
			case strings.HasPrefix(line, "Size:"):
				sizeStr := strings.TrimSpace(strings.TrimPrefix(line, "Size:"))
				module.Size = sizeStr
				module.SizeBytes = parseMemorySize(sizeStr) // НОВОЕ: парсим размер в байтах
			case strings.HasPrefix(line, "Type:"):
				module.Type = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
			case strings.HasPrefix(line, "Speed:"):
				module.Speed = strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
			case strings.HasPrefix(line, "Manufacturer:"):
				module.Manufacturer = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
			case strings.HasPrefix(line, "Part Number:"):
				module.PartNumber = strings.TrimSpace(strings.TrimPrefix(line, "Part Number:"))
			case strings.HasPrefix(line, "Locator:"):
				module.Slot = strings.TrimSpace(strings.TrimPrefix(line, "Locator:"))
			}
		}
		
		if module.Size != "" && module.Size != "No Module Installed" {
			modules = append(modules, module)
		}
	}
	
	return modules
}

func parseWMICMemory() ([]MemoryModule, error) {
	cmd := exec.Command("wmic", "memorychip", "get", 
		"BankLabel,Capacity,MemoryType,Speed,Manufacturer,PartNumber", "/format:csv")
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
		if len(fields) < 6 {
			continue
		}
		
		module := MemoryModule{
			Slot:         strings.TrimSpace(fields[1]),
			Manufacturer: strings.TrimSpace(fields[4]),
			PartNumber:   strings.TrimSpace(fields[5]),
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
			currentModule.SizeBytes = parseMemorySize(sizeStr) // НОВОЕ: парсим размер в байтах
		case strings.HasPrefix(line, "Type:"):
			currentModule.Type = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
		case strings.HasPrefix(line, "Speed:"):
			currentModule.Speed = strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
		case strings.HasPrefix(line, "Manufacturer:"):
			currentModule.Manufacturer = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
		case strings.HasPrefix(line, "Part Number:"):
			currentModule.PartNumber = strings.TrimSpace(strings.TrimPrefix(line, "Part Number:"))
		}
	}
	
	if currentModule.Slot != "" {
		modules = append(modules, currentModule)
	}
	
	return modules
}

// НОВАЯ ФУНКЦИЯ: Парсинг размера памяти из строки в байты
func parseMemorySize(sizeStr string) uint64 {
	if sizeStr == "" || sizeStr == "No Module Installed" {
		return 0
	}

	// Удаляем пробелы и приводим к нижнему регистру
	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))
	
	// Парсим числовое значение и единицу измерения
	var value float64
	var unit string
	
	// Разбираем строку типа "16 GB", "8 gb", "8192 mb" и т.д.
	_, err := fmt.Sscanf(sizeStr, "%f %s", &value, &unit)
	if err != nil {
		// Пробуем парсить без пробела: "16gb", "8gb"
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

	// Конвертируем в байты в зависимости от единицы измерения
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
		// Если единица не распознана, предполагаем что это байты
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