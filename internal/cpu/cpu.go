package cpu

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
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

var lastCPUStats []CPUStats

type CPUStats struct {
	Total uint64
	Idle  uint64
}

func Summary() (*CPUInfo, error) {
	info := &CPUInfo{}

	// Получаем базовую информацию
	if model, vendor, arch, cores, threads, err := getCPUInfo(); err == nil {
		info.Model = model
		info.Vendor = vendor
		info.Architecture = arch
		info.Cores = cores
		info.Threads = threads
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
	}

	// Получаем загрузку по ядрам
	info.CoreUsage = getPerCoreUsage()

	return info, nil
}

func getCPUInfo() (string, string, string, int, int, error) {
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
				// Упрощаем модель
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

func getCPUUsage() (float64, error) {
	// Используем /proc/stat для расчета загрузки CPU
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
		if i == 4 { // idle time
			idle = val
		}
	}

	// Сохраняем текущие статистики
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

	// Первый запуск - сохраняем статистики
	lastCPUStats = []CPUStats{currentStats}
	return 0, nil
}

func getLoadAverage() (float64, float64, float64, error) {
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

func getPerCoreUsage() []float64 {
	// Упрощенная реализация - возвращаем одинаковую загрузку для всех ядер
	// В реальной реализации нужно парсить /proc/stat для каждого ядра
	usage, _ := getCPUUsage()
	coreCount := runtime.NumCPU()
	usagePerCore := make([]float64, coreCount)
	for i := range usagePerCore {
		usagePerCore[i] = usage
	}
	return usagePerCore
}