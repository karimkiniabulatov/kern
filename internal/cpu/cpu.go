package cpu

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type CPUInfo struct {
	Model        string
	Vendor       string
	Architecture string
	Cores        int
	Threads      int
	Usage        float64
	Frequency    string
	Load1        float64
	Load5        float64
	Load15       float64
	CoreUsage    []float64
}

type CPUStats struct {
	Total uint64
	Idle  uint64
}

var (
	lastCPUStats   []CPUStats
	lastCoreStats  [][]CPUStats
	statsMutex     sync.RWMutex
)

func Summary() (*CPUInfo, error) {
	info := &CPUInfo{}
	
	// Всегда инициализировать поля, необходимые для гистограмм
	info.Usage = 0.0
	info.Load1 = 0.0
	info.Load5 = 0.0
	info.Load15 = 0.0
	info.CoreUsage = make([]float64, 0)

	// Получаем базовую информацию
	if model, vendor, arch, cores, threads, err := getCPUInfo(); err == nil {
		info.Model = model
		info.Vendor = vendor
		info.Architecture = arch
		info.Cores = cores
		info.Threads = threads
	} else {
		// Гарантируем значения по умолчанию, если реальные данные недоступны
		info.Model = "Unknown"
		info.Vendor = "Unknown"
		info.Architecture = runtime.GOARCH
		info.Cores = runtime.NumCPU()
		info.Threads = runtime.NumCPU()
	}

	// Получаем загрузку
	if usage, err := getCPUUsage(); err == nil {
		info.Usage = usage
	}

	// Получаем среднюю загрузку
	if load1, load5, load15, err := getLoadAverage(); err == nil {
		info.Load1 = load1
		info.Load5 = load5
		info.Load15 = load15
	}

	// Получаем частоту
	if freq, err := getCPUFrequency(); err == nil {
		info.Frequency = freq
	} else {
		info.Frequency = "Unknown"
	}

	// Получаем загрузку по ядрам
	if coreUsage, err := getPerCoreUsage(); err == nil && len(coreUsage) > 0 {
		info.CoreUsage = coreUsage
	} else {
		// Фолбэк: создаем массив с общим значением
		info.CoreUsage = make([]float64, info.Cores)
		for i := range info.CoreUsage {
			info.CoreUsage[i] = info.Usage
		}
	}

	return info, nil
}

// НОВАЯ ФУНКЦИЯ: правильное получение загрузки по ядрам
func getPerCoreUsage() ([]float64, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxPerCoreUsage()
	case "windows":
		return getWindowsPerCoreUsage()
	case "darwin":
		return getDarwinPerCoreUsage()
	default:
		return nil, fmt.Errorf("per-core usage not supported on %s", runtime.GOOS)
	}
}

func getLinuxPerCoreUsage() ([]float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var currentStats []CPUStats
	var coreUsage []float64

	// Собираем статистику для всех CPU (включая общий и ядра)
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				continue
			}

			var total, idle uint64
			for i := 1; i < len(fields); i++ {
				val, _ := strconv.ParseUint(fields[i], 10, 64)
				total += val
				if i == 4 { // Поле idle
					idle = val
				}
			}

			currentStats = append(currentStats, CPUStats{Total: total, Idle: idle})
		}
	}

	// Пропускаем общую статистику (cpu), начинаем с ядер (cpu0, cpu1, ...)
	if len(currentStats) <= 1 {
		return nil, fmt.Errorf("no core statistics found")
	}

	statsMutex.RLock()
	hasPrevious := len(lastCoreStats) > 0 && len(lastCoreStats[0]) == len(currentStats)
	statsMutex.RUnlock()

	if hasPrevious {
		statsMutex.RLock()
		last := lastCoreStats[0]
		statsMutex.RUnlock()

		// Обрабатываем каждое ядро, начиная с cpu1 (индекс 1)
		for i := 1; i < len(currentStats); i++ {
			if i < len(last) {
				totalDiff := currentStats[i].Total - last[i].Total
				idleDiff := currentStats[i].Idle - last[i].Idle

				if totalDiff > 0 {
					usage := 100.0 * float64(totalDiff-idleDiff) / float64(totalDiff)
					coreUsage = append(coreUsage, usage)
				} else {
					coreUsage = append(coreUsage, 0.0)
				}
			} else {
				coreUsage = append(coreUsage, 0.0)
			}
		}
	} else {
		// Первый запуск - заполняем нулями
		for i := 1; i < len(currentStats); i++ {
			coreUsage = append(coreUsage, 0.0)
		}
	}

	// Сохраняем текущую статистику
	statsMutex.Lock()
	if len(lastCoreStats) == 0 {
		lastCoreStats = make([][]CPUStats, 1)
	}
	lastCoreStats[0] = currentStats
	statsMutex.Unlock()

	return coreUsage, nil
}

// Заглушки для других платформ
func getWindowsPerCoreUsage() ([]float64, error) {
	// На Windows сложно получить загрузку по ядрам без WMI
	// Возвращаем nil чтобы использовать фолбэк
	return nil, fmt.Errorf("per-core usage not implemented for Windows")
}

func getDarwinPerCoreUsage() ([]float64, error) {
	// На macOS можно использовать sysctl или top
	return nil, fmt.Errorf("per-core usage not implemented for macOS")
}

func getCPUInfo() (string, string, string, int, int, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxCPUInfo()
	case "windows":
		return getWindowsCPUInfo()
	case "darwin":
		return getDarwinCPUInfo()
	default:
		return "Unknown", "Unknown", runtime.GOARCH, runtime.NumCPU(), runtime.NumCPU(), nil
	}
}

func getLinuxCPUInfo() (string, string, string, int, int, error) {
	cmd := exec.Command("lscpu")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", "Unknown", "Unknown", runtime.NumCPU(), runtime.NumCPU(), nil
	}

	lines := strings.Split(string(output), "\n")
	var model, vendor, arch string
	var cores, threads int

	for _, line := range lines {
		if strings.HasPrefix(line, "Model name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				model = strings.TrimSpace(parts[1])
				model = strings.ReplaceAll(model, "CPU", "")
				model = strings.Split(model, "@")[0]
				model = strings.TrimSpace(model)
			}
		} else if strings.HasPrefix(line, "Vendor ID:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				vendor = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Architecture:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				arch = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Core(s) per socket:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cores, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		} else if strings.HasPrefix(line, "CPU(s):") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				threads, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		}
	}

	if cores == 0 {
		cores = runtime.NumCPU()
	}
	if threads == 0 {
		threads = runtime.NumCPU()
	}

	return model, vendor, arch, cores, threads, nil
}

func getWindowsCPUInfo() (string, string, string, int, int, error) {
	// Получаем модель процессора
	cmdModel := exec.Command("wmic", "cpu", "get", "Name")
	outputModel, err := cmdModel.Output()
	model := "Unknown"
	if err == nil {
		lines := strings.Split(string(outputModel), "\n")
		if len(lines) >= 2 {
			model = strings.TrimSpace(lines[1])
			model = strings.ReplaceAll(model, "CPU", "")
			model = strings.Split(model, "@")[0]
			model = strings.TrimSpace(model)
		}
	}

	// Получаем производителя
	cmdVendor := exec.Command("wmic", "cpu", "get", "Manufacturer")
	outputVendor, err := cmdVendor.Output()
	vendor := "Unknown"
	if err == nil {
		lines := strings.Split(string(outputVendor), "\n")
		if len(lines) >= 2 {
			vendor = strings.TrimSpace(lines[1])
		}
	}

	// Получаем количество ядер и потоков
	cmdCores := exec.Command("wmic", "cpu", "get", "NumberOfCores")
	outputCores, err := cmdCores.Output()
	cores := runtime.NumCPU()
	if err == nil {
		lines := strings.Split(string(outputCores), "\n")
		if len(lines) >= 2 {
			cores, _ = strconv.Atoi(strings.TrimSpace(lines[1]))
		}
	}

	cmdThreads := exec.Command("wmic", "cpu", "get", "NumberOfLogicalProcessors")
	outputThreads, err := cmdThreads.Output()
	threads := runtime.NumCPU()
	if err == nil {
		lines := strings.Split(string(outputThreads), "\n")
		if len(lines) >= 2 {
			threads, _ = strconv.Atoi(strings.TrimSpace(lines[1]))
		}
	}

	return model, vendor, runtime.GOARCH, cores, threads, nil
}

func getDarwinCPUInfo() (string, string, string, int, int, error) {
	// Получаем модель процессора
	cmdModel := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	outputModel, err := cmdModel.Output()
	model := "Unknown"
	if err == nil {
		model = strings.TrimSpace(string(outputModel))
		model = strings.ReplaceAll(model, "CPU", "")
		model = strings.Split(model, "@")[0]
		model = strings.TrimSpace(model)
	}

	// Получаем производителя
	cmdVendor := exec.Command("sysctl", "-n", "machdep.cpu.vendor")
	outputVendor, err := cmdVendor.Output()
	vendor := "Unknown"
	if err == nil {
		vendor = strings.TrimSpace(string(outputVendor))
	}

	// Получаем архитектуру
	cmdArch := exec.Command("uname", "-m")
	outputArch, err := cmdArch.Output()
	arch := runtime.GOARCH
	if err == nil {
		arch = strings.TrimSpace(string(outputArch))
	}

	// Получаем количество ядер
	cmdCores := exec.Command("sysctl", "-n", "hw.physicalcpu")
	outputCores, err := cmdCores.Output()
	cores := runtime.NumCPU()
	if err == nil {
		cores, _ = strconv.Atoi(strings.TrimSpace(string(outputCores)))
	}

	// Получаем количество потоков
	cmdThreads := exec.Command("sysctl", "-n", "hw.logicalcpu")
	outputThreads, err := cmdThreads.Output()
	threads := runtime.NumCPU()
	if err == nil {
		threads, _ = strconv.Atoi(strings.TrimSpace(string(outputThreads)))
	}

	return model, vendor, arch, cores, threads, nil
}

func getCPUUsage() (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxCPUUsage()
	case "windows", "darwin":
		// TODO: Реализовать для Windows и macOS
		return 0, nil
	default:
		return 0, nil
	}
}

func getLinuxCPUUsage() (float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0, nil
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 8 {
		return 0, nil
	}

	var total, idle uint64
	for i := 1; i < len(fields); i++ {
		val, _ := strconv.ParseUint(fields[i], 10, 64)
		total += val
		if i == 4 {
			idle = val
		}
	}

	currentStats := CPUStats{Total: total, Idle: idle}

	if len(lastCPUStats) > 0 {
		last := lastCPUStats[0]
		totalDiff := total - last.Total
		idleDiff := idle - last.Idle

		if totalDiff > 0 {
			usage := 100.0 * float64(totalDiff-idleDiff) / float64(totalDiff)
			lastCPUStats[0] = currentStats
			return usage, nil
		}
	}

	lastCPUStats = []CPUStats{currentStats}
	return 0, nil
}

func getLoadAverage() (float64, float64, float64, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxLoadAverage()
	case "windows", "darwin":
		// TODO: Реализовать для Windows и macOS
		return 0, 0, 0, nil
	default:
		return 0, 0, 0, nil
	}
}

func getLinuxLoadAverage() (float64, float64, float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		load1, _ := strconv.ParseFloat(fields[0], 64)
		load5, _ := strconv.ParseFloat(fields[1], 64)
		load15, _ := strconv.ParseFloat(fields[2], 64)
		return load1, load5, load15, nil
	}

	return 0, 0, 0, nil
}

func getCPUFrequency() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxCPUFrequency()
	case "windows":
		return getWindowsCPUFrequency()
	case "darwin":
		return getDarwinCPUFrequency()
	default:
		return "Unknown", nil
	}
}

func getLinuxCPUFrequency() (string, error) {
	cmd := exec.Command("lscpu")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CPU max MHz:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				freq := strings.TrimSpace(parts[1])
				if freqMHz, err := strconv.ParseFloat(freq, 64); err == nil {
					return fmt.Sprintf("%.0f MHz", freqMHz), nil
				}
			}
		}
	}

	return "Unknown", nil
}

func getWindowsCPUFrequency() (string, error) {
	cmd := exec.Command("wmic", "cpu", "get", "MaxClockSpeed")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		freq := strings.TrimSpace(lines[1])
		if freqMHz, err := strconv.ParseFloat(freq, 64); err == nil {
			return fmt.Sprintf("%.0f MHz", freqMHz), nil
		}
	}

	return "Unknown", nil
}

func getDarwinCPUFrequency() (string, error) {
	cmd := exec.Command("sysctl", "-n", "hw.cpufrequency")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", err
	}

	freqStr := strings.TrimSpace(string(output))
	if freqHz, err := strconv.ParseFloat(freqStr, 64); err == nil {
		freqMHz := freqHz / 1000000
		return fmt.Sprintf("%.0f MHz", freqMHz), nil
	}

	return "Unknown", nil
}