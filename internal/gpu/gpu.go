package gpu

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type GPUInfo struct {
	Model           string
	DriverVersion   string
	GPUTemp         float64
	MemoryTotal     string
	MemoryUsed      string
	MemoryFree      string
	Utilization     float64
	PowerDraw       string
	PowerLimit      string
	FanSpeed        float64
	ClockCore       string
	ClockMemory     string
	PerformanceState string
}

// Summary возвращает информацию о всех GPU в системе
func Summary() ([]*GPUInfo, error) {
	return detectAllGPUs()
}

// Универсальное обнаружение GPU
func detectAllGPUs() ([]*GPUInfo, error) {
	var gpus []*GPUInfo
    
	// NVIDIA
	if nvidiaGPUs, err := detectNVIDIA(); err == nil {
		gpus = append(gpus, nvidiaGPUs...)
	}
    
	// AMD 
	if amdGPUs, err := detectAMD(); err == nil {
		gpus = append(gpus, amdGPUs...)
	}
    
	// Intel
	if intelGPUs, err := detectIntel(); err == nil {
		gpus = append(gpus, intelGPUs...)
	}
    
	if len(gpus) > 0 {
		return gpus, nil
	}
    
	// Platform-specific fallback detection
	if genericGPUs := detectGenericGPUs(); len(genericGPUs) > 0 {
		return genericGPUs, nil
	}
    
	// Return empty slice if no GPUs found
	return gpus, fmt.Errorf("no GPU devices detected")
}

// detectNVIDIA обнаруживает NVIDIA GPU через nvidia-smi
func detectNVIDIA() ([]*GPUInfo, error) {
	output, err := exec.Command("nvidia-smi", "--query-gpu=name,driver_version,temperature.gpu,memory.total,memory.used,memory.free,utilization.gpu,power.draw,power.limit,fan.speed,clocks.current.graphics,clocks.current.memory,performance.state", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, err
	}
	return parseNvidiaSMIOutput(string(output))
}

// detectAMD обнаруживает AMD GPU через rocm-smi
func detectAMD() ([]*GPUInfo, error) {
	output, err := exec.Command("rocm-smi", "--showproductname", "--showtemp", "--showuse", "--showmemuse", "--showdriverversion").Output()
	if err != nil {
		return nil, err
	}
	return parseAMDSMIOutput(string(output))
}

// detectIntel обнаруживает Intel GPU через intel_gpu_top или sysfs
func detectIntel() ([]*GPUInfo, error) {
	var gpus []*GPUInfo
	
	// Попробуем использовать intel_gpu_top если доступен
	if output, err := exec.Command("intel_gpu_top", "-J").Output(); err == nil {
		if intelGPUs, err := parseIntelGPUOutput(string(output)); err == nil && len(intelGPUs) > 0 {
			return intelGPUs, nil
		}
	}
	
	// Fallback: проверка через sysfs (Linux)
	if runtime.GOOS == "linux" {
		if intelGPUs, err := detectIntelViaSysfs(); err == nil && len(intelGPUs) > 0 {
			return intelGPUs, nil
		}
	}
	
	// Fallback: универсальное обнаружение Intel
	if genericIntel := detectGenericIntel(); len(genericIntel) > 0 {
		return genericIntel, nil
	}
	
	return gpus, fmt.Errorf("Intel GPU not detected")
}

// parseIntelGPUOutput парсит вывод intel_gpu_top
func parseIntelGPUOutput(output string) ([]*GPUInfo, error) {
	var gpus []*GPUInfo
	
	// Базовая реализация парсинга вывода intel_gpu_top
	// В реальной реализации здесь должен быть парсинг JSON вывода
	info := &GPUInfo{
		Model:           "Intel GPU",
		DriverVersion:   "Unknown",
		GPUTemp:         0.0,
		MemoryTotal:     "0 MB",
		MemoryUsed:      "0 MB",
		MemoryFree:      "0 MB",
		Utilization:     0.0,
		PowerDraw:       "0 W",
		PowerLimit:      "0 W",
		FanSpeed:        0.0,
		ClockCore:       "0 MHz",
		ClockMemory:     "0 MHz",
		PerformanceState: "Unknown",
	}
	
	gpus = append(gpus, info)
	return gpus, nil
}

// detectIntelViaSysfs обнаруживает Intel GPU через sysfs на Linux
func detectIntelViaSysfs() ([]*GPUInfo, error) {
	var gpus []*GPUInfo
	
	// Проверяем наличие Intel GPU в sysfs
	if _, err := exec.Command("test", "-d", "/sys/class/drm/card0/device").Output(); err != nil {
		return nil, err
	}
	
	// Читаем информацию о GPU
	if vendor, err := exec.Command("cat", "/sys/class/drm/card0/device/vendor").Output(); err == nil {
		if strings.Contains(string(vendor), "8086") { // Intel vendor ID
			info := &GPUInfo{
				Model:           "Intel Integrated Graphics",
				DriverVersion:   "Unknown",
				PerformanceState: "Active",
			}
			
			// Пытаемся получить дополнительную информацию
			if model, err := exec.Command("cat", "/sys/class/drm/card0/device/name").Output(); err == nil {
				info.Model = strings.TrimSpace(string(model))
			}
			
			gpus = append(gpus, info)
		}
	}
	
	if len(gpus) > 0 {
		return gpus, nil
	}
	
	return nil, fmt.Errorf("Intel GPU not found in sysfs")
}

// detectGenericIntel универсальное обнаружение Intel GPU
func detectGenericIntel() []*GPUInfo {
	var gpus []*GPUInfo
	
	switch runtime.GOOS {
	case "darwin": // macOS
		if output, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "Chipset Model:") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						model := strings.TrimSpace(parts[1])
						if strings.Contains(model, "Intel") {
							gpus = append(gpus, &GPUInfo{
								Model: "Intel " + model,
								DriverVersion: "Unknown",
							})
						}
					}
				}
			}
		}
	case "windows":
		// Улучшенное обнаружение Intel GPU на Windows через WMI
		if intelGPUs, err := getWindowsIntelGPUInfo(); err == nil && len(intelGPUs) > 0 {
			return intelGPUs
		}
		
		// Fallback: базовое обнаружение через wmic
		if output, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name", "/value").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Name=") {
					gpuName := strings.TrimPrefix(line, "Name=")
					gpuName = strings.TrimSpace(gpuName)
					if strings.Contains(strings.ToLower(gpuName), "intel") {
						gpus = append(gpus, &GPUInfo{
							Model: gpuName,
							DriverVersion: "Unknown",
						})
					}
				}
			}
		}
	case "linux":
		if output, err := exec.Command("lspci").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "VGA compatible controller") && strings.Contains(line, "Intel") {
					gpus = append(gpus, &GPUInfo{
						Model: strings.TrimSpace(line),
						DriverVersion: "Unknown",
					})
				}
			}
		}
	}
	
	return gpus
}

// getWindowsIntelGPUInfo улучшенное обнаружение Intel GPU на Windows через WMI
func getWindowsIntelGPUInfo() ([]*GPUInfo, error) {
	var gpus []*GPUInfo

	// Использование WMI для получения детальной информации о GPU Intel
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "Name,DriverVersion,AdapterRAM,VideoProcessor,AdapterCompatibility", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	gpuMap := make(map[string]*GPUInfo)
	
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		
		fields := strings.Split(line, ",")
		if len(fields) >= 6 {
			deviceID := strings.TrimSpace(fields[1])
			name := strings.TrimSpace(fields[2])
			driverVersion := strings.TrimSpace(fields[3])
			adapterRAM := strings.TrimSpace(fields[4])
			videoProcessor := strings.TrimSpace(fields[5])
			adapterCompatibility := strings.TrimSpace(fields[6])

			// Фильтруем только Intel GPU
			if !strings.Contains(strings.ToLower(name), "intel") && 
			   !strings.Contains(strings.ToLower(adapterCompatibility), "intel") &&
			   !strings.Contains(strings.ToLower(videoProcessor), "intel") {
				continue
			}

			gpu := &GPUInfo{
				Model:         name,
				DriverVersion: driverVersion,
				PerformanceState: "Active",
			}

			// Обработка памяти
			if adapterRAM != "" {
				if ramBytes, err := strconv.ParseUint(adapterRAM, 10, 64); err == nil {
					ramMB := ramBytes / 1024 / 1024
					gpu.MemoryTotal = fmt.Sprintf("%d MB", ramMB)
				}
			}

			// Используем VideoProcessor как дополнительный идентификатор
			if videoProcessor != "" && videoProcessor != "Unknown" {
				if gpu.Model == "" || gpu.Model == "Intel GPU" {
					gpu.Model = videoProcessor
				}
			}

			gpuMap[deviceID] = gpu
		}
	}

	// Преобразуем map в slice
	for _, gpu := range gpuMap {
		gpus = append(gpus, gpu)
	}

	if len(gpus) > 0 {
		return gpus, nil
	}

	return nil, fmt.Errorf("no Intel GPUs found via WMI")
}

// parseNvidiaSMIOutput парсит вывод nvidia-smi для нескольких GPU
func parseNvidiaSMIOutput(output string) ([]*GPUInfo, error) {
	var gpus []*GPUInfo
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no GPU data available")
	}

	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) < 13 {
			continue // Пропускаем некорректные строки
		}

		info := &GPUInfo{
			Model:            strings.TrimSpace(fields[0]),
			DriverVersion:    strings.TrimSpace(fields[1]),
			PerformanceState: strings.TrimSpace(fields[12]),
		}

		// Parse numeric values with fallbacks
		if temp, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64); err == nil {
			info.GPUTemp = temp
		} else {
			info.GPUTemp = 0.0
		}
		
		info.MemoryTotal = strings.TrimSpace(fields[3]) + " MB"
		info.MemoryUsed = strings.TrimSpace(fields[4]) + " MB" 
		info.MemoryFree = strings.TrimSpace(fields[5]) + " MB"

		if util, err := strconv.ParseFloat(strings.TrimSpace(fields[6]), 64); err == nil {
			info.Utilization = util
		} else {
			info.Utilization = 0.0
		}

		info.PowerDraw = strings.TrimSpace(fields[7]) + " W"
		info.PowerLimit = strings.TrimSpace(fields[8]) + " W"

		if fan, err := strconv.ParseFloat(strings.TrimSpace(fields[9]), 64); err == nil {
			info.FanSpeed = fan
		} else {
			info.FanSpeed = 0.0
		}

		info.ClockCore = strings.TrimSpace(fields[10]) + " MHz"
		info.ClockMemory = strings.TrimSpace(fields[11]) + " MHz"

		gpus = append(gpus, info)
	}

	return gpus, nil
}

// parseAMDSMIOutput парсит вывод rocm-smi для нескольких GPU
func parseAMDSMIOutput(output string) ([]*GPUInfo, error) {
	var gpus []*GPUInfo
	
	lines := strings.Split(output, "\n")
	var currentGPU *GPUInfo
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Новый GPU начинается с "card"
		if strings.HasPrefix(line, "card") {
			if currentGPU != nil {
				gpus = append(gpus, currentGPU)
			}
			currentGPU = &GPUInfo{
				Model:           "AMD GPU",
				DriverVersion:   "Unknown",
				GPUTemp:         0.0,
				MemoryTotal:     "0 MB",
				MemoryUsed:      "0 MB",
				MemoryFree:      "0 MB",
				Utilization:     0.0,
				PowerDraw:       "0 W",
				PowerLimit:      "0 W",
				FanSpeed:        0.0,
				ClockCore:       "0 MHz",
				ClockMemory:     "0 MHz",
				PerformanceState: "Unknown",
			}
			continue
		}

		if currentGPU == nil {
			continue
		}

		switch {
		case strings.Contains(line, "Product Name"):
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				currentGPU.Model = strings.TrimSpace(parts[1])
			}
		case strings.Contains(line, "Driver version"):
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				currentGPU.DriverVersion = strings.TrimSpace(parts[1])
			}
		case strings.Contains(line, "Temperature"):
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				tempStr := strings.TrimSpace(parts[1])
				tempStr = strings.TrimSuffix(tempStr, "c")
				tempStr = strings.TrimSuffix(tempStr, "C")
				if temp, err := strconv.ParseFloat(strings.TrimSpace(tempStr), 64); err == nil {
					currentGPU.GPUTemp = temp
				}
			}
		case strings.Contains(line, "GPU use"):
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				utilStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "%"))
				if util, err := strconv.ParseFloat(utilStr, 64); err == nil {
					currentGPU.Utilization = util
				}
			}
		case strings.Contains(line, "Memory use"):
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				memInfo := strings.TrimSpace(parts[1])
				// Parse memory information like "1234 MB / 5678 MB"
				if memParts := strings.Split(memInfo, "/"); len(memParts) == 2 {
					currentGPU.MemoryUsed = strings.TrimSpace(memParts[0])
					currentGPU.MemoryTotal = strings.TrimSpace(memParts[1])
				}
			}
		}
	}

	// Добавляем последний GPU
	if currentGPU != nil {
		gpus = append(gpus, currentGPU)
	}

	return gpus, nil
}

// detectGenericGPUs возвращает информацию о GPU для универсального обнаружения
func detectGenericGPUs() []*GPUInfo {
	var gpus []*GPUInfo
	info := &GPUInfo{
		Model:         "",
		DriverVersion: "Unknown",
	}

	switch runtime.GOOS {
	case "darwin": // macOS
		info.Model = detectMacGPU()
	case "windows":
		info.Model = detectWindowsGPU()
	case "linux":
		info.Model = detectLinuxGPU()
	default:
		info.Model = "Generic GPU (Unknown platform)"
	}

	if info.Model == "" {
		info.Model = "Generic GPU (No specific GPU detected)"
	}

	// Для универсального обнаружения предполагаем одну GPU
	gpus = append(gpus, info)
	return gpus
}

func detectMacGPU() string {
	// Use system_profiler to get GPU info on macOS
	if output, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Chipset Model:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					return "Apple " + strings.TrimSpace(parts[1])
				}
			}
			// Look for Intel integrated graphics
			if strings.Contains(line, "Intel") && i > 0 && strings.Contains(lines[i-1], "Chipset Model:") {
				return strings.TrimSpace(line)
			}
		}
	}
	return "Integrated Graphics (Apple)"
}

func detectWindowsGPU() string {
	// Use wmic to get basic GPU info on Windows
	if output, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name", "/value").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Name=") {
				gpuName := strings.TrimPrefix(line, "Name=")
				gpuName = strings.TrimSpace(gpuName)
				if gpuName != "" {
					return gpuName
				}
			}
		}
	}
	return "Integrated Graphics (Windows)"
}

func detectLinuxGPU() string {
	// Check for common GPU detection methods on Linux
	commands := []struct {
		cmd  string
		args []string
	}{
		{"lspci", []string{}},
		{"lshw", []string{"-C", "display"}},
	}

	for _, c := range commands {
		if output, err := exec.Command(c.cmd, c.args...).Output(); err == nil {
			outputStr := string(output)
			
			// Check for Intel integrated graphics
			if strings.Contains(outputStr, "Intel") && 
			   (strings.Contains(outputStr, "Graphics") || strings.Contains(outputStr, "HD Graphics") || strings.Contains(outputStr, "UHD Graphics")) {
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(line, "Intel") {
						return strings.TrimSpace(line)
					}
				}
			}
			
			// Check for AMD integrated graphics
			if strings.Contains(outputStr, "AMD") && strings.Contains(outputStr, "Graphics") {
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(line, "AMD") && strings.Contains(line, "Graphics") {
						return strings.TrimSpace(line)
					}
				}
			}
		}
	}

	// Check /proc/cpuinfo for GPU info (for ARM devices like Raspberry Pi)
	if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
		if data, err := exec.Command("cat", "/proc/cpuinfo").Output(); err == nil {
			if strings.Contains(string(data), "VideoCore") {
				return "Broadcom VideoCore (Raspberry Pi)"
			}
		}
	}

	return "Integrated Graphics (Linux)"
}