package cpu

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type CPUInfo struct {
	SocketID     int      // Номер сокета/процессора
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
	NUMANode     int     // Нод NUMA
	Physical     bool    // Физический процессор
}

type CPUStats struct {
	Total uint64
	Idle  uint64
}

type SystemCPUInfo struct {
	TotalSockets int        // Общее количество сокетов
	TotalCores   int        // Общее количество ядер
	TotalThreads int        // Общее количество потоков
	CPUs         []*CPUInfo // Информация по каждому процессору
}

var (
	lastCPUStats   []CPUStats
	lastCoreStats  [][]CPUStats
	statsMutex     sync.RWMutex
)

func Summary() (*SystemCPUInfo, error) {
	systemInfo := &SystemCPUInfo{
		CPUs: make([]*CPUInfo, 0),
	}
	
	// Получаем информацию о всех процессорах
	cpus, err := getAllCPUInfo()
	if err != nil {
		// Фолбэк: создаем один процессор с базовой информацией
		basicCPU := &CPUInfo{
			SocketID:     0,
			Model:        "Unknown",
			Vendor:       "Unknown",
			Architecture: runtime.GOARCH,
			Cores:        runtime.NumCPU(),
			Threads:      runtime.NumCPU(),
			Usage:        0.0,
			Frequency:    "Unknown",
			Load1:        0.0,
			Load5:        0.0,
			Load15:       0.0,
			CoreUsage:    make([]float64, 0),
			NUMANode:     0,
			Physical:     true,
		}
		systemInfo.CPUs = append(systemInfo.CPUs, basicCPU)
		systemInfo.TotalSockets = 1
		systemInfo.TotalCores = basicCPU.Cores
		systemInfo.TotalThreads = basicCPU.Threads
		return systemInfo, nil
	}

	systemInfo.CPUs = cpus
	
	// Считаем общую статистику
	for _, cpu := range cpus {
		systemInfo.TotalSockets++
		systemInfo.TotalCores += cpu.Cores
		systemInfo.TotalThreads += cpu.Threads
	}

	return systemInfo, nil
}

// Новая функция для получения информации о всех процессорах
func getAllCPUInfo() ([]*CPUInfo, error) {
    switch runtime.GOOS {
    case "linux":
        return getLinuxAllCPUInfo()
    case "windows":
        return getWindowsAllCPUInfo()
    case "darwin":
        return getDarwinAllCPUInfo()
    default:
        return getDefaultCPUInfo()
    }
}

func getLinuxAllCPUInfo() ([]*CPUInfo, error) {
	var cpus []*CPUInfo

	// Получаем информацию о количестве сокетов через lscpu
	cmd := exec.Command("lscpu")
	output, err := cmd.Output()
	if err != nil {
		return getDefaultCPUInfo()
	}

	lines := strings.Split(string(output), "\n")
	var sockets, coresPerSocket, threadsPerCore int
	var model, vendor, arch string

	for _, line := range lines {
		if strings.HasPrefix(line, "Socket(s):") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				sockets, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		} else if strings.HasPrefix(line, "Core(s) per socket:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				coresPerSocket, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		} else if strings.HasPrefix(line, "Thread(s) per core:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				threadsPerCore, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		} else if strings.HasPrefix(line, "Model name:") {
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
		}
	}

	// Если не удалось определить количество сокетов, используем значение по умолчанию
	if sockets == 0 {
		sockets = 1
	}
	if coresPerSocket == 0 {
		coresPerSocket = runtime.NumCPU()
	}
	if threadsPerCore == 0 {
		threadsPerCore = 1
	}

	// Получаем среднюю загрузку системы
	load1, load5, load15, err := getLoadAverage()
	if err != nil {
		// Логируем ошибку, но продолжаем
		log.Printf("Failed to get load average: %v", err)
		load1, load5, load15 = 0.0, 0.0, 0.0
	}

	// Получаем информацию о NUMA нодах
	numaNodes := getLinuxNUMANodes()

	// Создаем информацию для каждого сокета
	for i := 0; i < sockets; i++ {
		cpu := &CPUInfo{
			SocketID:     i,
			Model:        model,
			Vendor:       vendor,
			Architecture: arch,
			Cores:        coresPerSocket,
			Threads:      coresPerSocket * threadsPerCore,
			Usage:        0.0,
			Load1:        load1,
			Load5:        load5,
			Load15:       load15,
			NUMANode:     i % len(numaNodes), // Распределяем по нодам циклически
			Physical:     true,
		}

		// Получаем частоту для процессора
		if freq, err := getCPUFrequency(); err == nil {
			cpu.Frequency = freq
		} else {
			cpu.Frequency = "Unknown"
		}

		// Получаем загрузку для процессора
		if usage, err := getCPUUsage(); err == nil {
			cpu.Usage = usage
		}

		// Получаем загрузку по ядрам для этого процессора
		if coreUsage, err := getPerCoreUsage(); err == nil && len(coreUsage) > 0 {
			// Распределяем ядра между процессорами
			coresPerCPU := len(coreUsage) / sockets
			startCore := i * coresPerCPU
			endCore := startCore + coresPerCPU
			if i == sockets-1 { // Последний процессор получает все оставшиеся ядра
				endCore = len(coreUsage)
			}
			if startCore < len(coreUsage) {
				cpu.CoreUsage = coreUsage[startCore:endCore]
			}
		}

		cpus = append(cpus, cpu)
	}

	return cpus, nil
}

func getLinuxNUMANodes() []int {
	// Проверяем наличие NUMA нод
	nodes := []int{0} // По умолчанию одна нода
	
	// Пытаемся прочитать информацию о NUMA из sysfs
	files, err := os.ReadDir("/sys/devices/system/node")
	if err != nil {
		return nodes
	}

	nodes = []int{}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "node") {
			if nodeID, err := strconv.Atoi(strings.TrimPrefix(file.Name(), "node")); err == nil {
				nodes = append(nodes, nodeID)
			}
		}
	}

	if len(nodes) == 0 {
		return []int{0}
	}

	return nodes
}

func getWindowsAllCPUInfo() ([]*CPUInfo, error) {
	var cpus []*CPUInfo

	// Использование WMI для получения информации о всех процессорах
	cmd := exec.Command("wmic", "cpu", "get", "Name,NumberOfCores,NumberOfLogicalProcessors,MaxClockSpeed", "/format:csv")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i == 0 || strings.TrimSpace(line) == "" {
				continue // Пропускаем заголовок и пустые строки
			}
			
			fields := strings.Split(line, ",")
			if len(fields) >= 5 {
				cpu := &CPUInfo{
					SocketID:     i - 1, // начинаем с 0
					Physical:     true,
					NUMANode:     0, // По умолчанию для Windows
					Architecture: runtime.GOARCH,
				}
				
				// Имя процессора
				if len(fields) >= 2 {
					cpu.Model = strings.TrimSpace(fields[1])
					cpu.Model = strings.ReplaceAll(cpu.Model, "CPU", "")
					cpu.Model = strings.Split(cpu.Model, "@")[0]
					cpu.Model = strings.TrimSpace(cpu.Model)
				}
				
				// Количество ядер
				if len(fields) >= 3 {
					cpu.Cores, _ = strconv.Atoi(strings.TrimSpace(fields[2]))
				}
				
				// Количество логических процессоров
				if len(fields) >= 4 {
					cpu.Threads, _ = strconv.Atoi(strings.TrimSpace(fields[3]))
				}
				
				// Максимальная частота
				if len(fields) >= 5 {
					if freqMHz, err := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64); err == nil {
						cpu.Frequency = fmt.Sprintf("%.0f MHz", freqMHz)
					}
				}
				
				// Получаем производителя отдельно
				cmdVendor := exec.Command("wmic", "cpu", "get", "Manufacturer")
				if outputVendor, err := cmdVendor.Output(); err == nil {
					vendorLines := strings.Split(string(outputVendor), "\n")
					if len(vendorLines) >= 2+i {
						cpu.Vendor = strings.TrimSpace(vendorLines[1+i])
					}
				}
				
				cpus = append(cpus, cpu)
			}
		}
	}

	// Если через WMI не получилось, используем старый метод как фолбэк
	if len(cpus) == 0 {
		cpu := &CPUInfo{
			SocketID:     0,
			NUMANode:     0,
			Physical:     true,
			Architecture: runtime.GOARCH,
		}

		// Получаем модель процессора
		cmdModel := exec.Command("wmic", "cpu", "get", "Name")
		outputModel, err := cmdModel.Output()
		if err == nil {
			lines := strings.Split(string(outputModel), "\n")
			if len(lines) >= 2 {
				cpu.Model = strings.TrimSpace(lines[1])
				cpu.Model = strings.ReplaceAll(cpu.Model, "CPU", "")
				cpu.Model = strings.Split(cpu.Model, "@")[0]
				cpu.Model = strings.TrimSpace(cpu.Model)
			}
		}

		// Получаем производителя
		cmdVendor := exec.Command("wmic", "cpu", "get", "Manufacturer")
		outputVendor, err := cmdVendor.Output()
		if err == nil {
			lines := strings.Split(string(outputVendor), "\n")
			if len(lines) >= 2 {
				cpu.Vendor = strings.TrimSpace(lines[1])
			}
		}

		// Получаем количество ядер и потоков
		cmdCores := exec.Command("wmic", "cpu", "get", "NumberOfCores")
		outputCores, err := cmdCores.Output()
		if err == nil {
			lines := strings.Split(string(outputCores), "\n")
			if len(lines) >= 2 {
				cpu.Cores, _ = strconv.Atoi(strings.TrimSpace(lines[1]))
			}
		} else {
			cpu.Cores = runtime.NumCPU()
		}

		cmdThreads := exec.Command("wmic", "cpu", "get", "NumberOfLogicalProcessors")
		outputThreads, err := cmdThreads.Output()
		if err == nil {
			lines := strings.Split(string(outputThreads), "\n")
			if len(lines) >= 2 {
				cpu.Threads, _ = strconv.Atoi(strings.TrimSpace(lines[1]))
			}
		} else {
			cpu.Threads = runtime.NumCPU()
		}

		// Получаем частоту
		if freq, err := getCPUFrequency(); err == nil {
			cpu.Frequency = freq
		} else {
			cpu.Frequency = "Unknown"
		}

		cpus = append(cpus, cpu)
	}
	
	return cpus, nil
}

func getDarwinAllCPUInfo() ([]*CPUInfo, error) {
	var cpus []*CPUInfo

	// Использование sysctl для получения детальной информации
	cmd := exec.Command("sysctl", "-n", "hw.physicalcpu_max", "hw.logicalcpu_max", "hw.cpufrequency_max")
	output, err := cmd.Output()
	if err == nil {
		fields := strings.Fields(string(output))
		if len(fields) >= 3 {
			physicalCPU, _ := strconv.Atoi(fields[0])
			logicalCPU, _ := strconv.Atoi(fields[1])
			cpufrequency, _ := strconv.ParseFloat(fields[2], 64)
			
			// На macOS обычно один физический процессор
			cpu := &CPUInfo{
				SocketID:     0,
				NUMANode:     0,
				Physical:     true,
				Cores:        physicalCPU,
				Threads:      logicalCPU,
			}
			
			// Преобразуем частоту из Гц в МГц
			if cpufrequency > 0 {
				freqMHz := cpufrequency / 1000000
				cpu.Frequency = fmt.Sprintf("%.0f MHz", freqMHz)
			}

			cpus = append(cpus, cpu)
		}
	}

	// Если через sysctl не получилось, используем старый метод как фолбэк
	if len(cpus) == 0 {
		cpu := &CPUInfo{
			SocketID: 0,
			NUMANode: 0,
			Physical: true,
		}

		// Получаем модель процессора
		cmdModel := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		outputModel, err := cmdModel.Output()
		if err == nil {
			cpu.Model = strings.TrimSpace(string(outputModel))
			cpu.Model = strings.ReplaceAll(cpu.Model, "CPU", "")
			cpu.Model = strings.Split(cpu.Model, "@")[0]
			cpu.Model = strings.TrimSpace(cpu.Model)
		}

		// Получаем производителя
		cmdVendor := exec.Command("sysctl", "-n", "machdep.cpu.vendor")
		outputVendor, err := cmdVendor.Output()
		if err == nil {
			cpu.Vendor = strings.TrimSpace(string(outputVendor))
		}

		// Получаем архитектуру
		cmdArch := exec.Command("uname", "-m")
		outputArch, err := cmdArch.Output()
		if err == nil {
			cpu.Architecture = strings.TrimSpace(string(outputArch))
		} else {
			cpu.Architecture = runtime.GOARCH
		}

		// Получаем количество ядер
		cmdCores := exec.Command("sysctl", "-n", "hw.physicalcpu")
		outputCores, err := cmdCores.Output()
		if err == nil {
			cpu.Cores, _ = strconv.Atoi(strings.TrimSpace(string(outputCores)))
		} else {
			cpu.Cores = runtime.NumCPU()
		}

		// Получаем количество потоков
		cmdThreads := exec.Command("sysctl", "-n", "hw.logicalcpu")
		outputThreads, err := cmdThreads.Output()
		if err == nil {
			cpu.Threads, _ = strconv.Atoi(strings.TrimSpace(string(outputThreads)))
		} else {
			cpu.Threads = runtime.NumCPU()
		}

		// Получаем частоту
		if freq, err := getCPUFrequency(); err == nil {
			cpu.Frequency = freq
		} else {
			cpu.Frequency = "Unknown"
		}

		cpus = append(cpus, cpu)
	}
	
	return cpus, nil
}

func getDefaultCPUInfo() ([]*CPUInfo, error) {
	cpu := &CPUInfo{
		SocketID:     0,
		Model:        "Unknown",
		Vendor:       "Unknown",
		Architecture: runtime.GOARCH,
		Cores:        runtime.NumCPU(),
		Threads:      runtime.NumCPU(),
		Usage:        0.0,
		Frequency:    "Unknown",
		Load1:        0.0,
		Load5:        0.0,
		Load15:       0.0,
		CoreUsage:    make([]float64, 0),
		NUMANode:     0,
		Physical:     true,
	}
	return []*CPUInfo{cpu}, nil
}

// Остальные функции остаются без изменений...
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
	return nil, fmt.Errorf("per-core usage not implemented for Windows")
}

func getDarwinPerCoreUsage() ([]float64, error) {
	return nil, fmt.Errorf("per-core usage not implemented for macOS")
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

func getLinuxLoadAverage() (float64, float64, float64, error) {
    data, err := os.ReadFile("/proc/loadavg")
    if err != nil {
        return 0, 0, 0, err
    }

    fields := strings.Fields(string(data))
    if len(fields) >= 3 {
        load1, err1 := strconv.ParseFloat(fields[0], 64)
        load5, err2 := strconv.ParseFloat(fields[1], 64)
        load15, err3 := strconv.ParseFloat(fields[2], 64)
        
        if err1 == nil && err2 == nil && err3 == nil {
            return load1, load5, load15, nil
        }
    }

    return 0, 0, 0, fmt.Errorf("invalid loadavg format")
}

// Добавить функцию для Windows
func getWindowsLoadAverage() (float64, float64, float64, error) {
    // Используем PowerShell для получения средней загрузки на Windows
    cmd := exec.Command("powershell", "-Command", 
        "Get-WmiObject Win32_Processor | Measure-Object -Property LoadPercentage -Average | Select-Object -ExpandProperty Average")
    
    output, err := cmd.Output()
    if err != nil {
        return 0, 0, 0, err
    }
    
    avgStr := strings.TrimSpace(string(output))
    if avgStr == "" {
        return 0, 0, 0, fmt.Errorf("no output from PowerShell")
    }
    
    avg, err := strconv.ParseFloat(avgStr, 64)
    if err != nil {
        return 0, 0, 0, err
    }
    
    // Для Windows возвращаем одно значение для всех трех интервалов
    return avg, avg, avg, nil
}

// Добавить функцию для macOS
func getDarwinLoadAverage() (float64, float64, float64, error) {
    cmd := exec.Command("sysctl", "-n", "vm.loadavg")
    output, err := cmd.Output()
    if err != nil {
        return 0, 0, 0, err
    }
    
    outputStr := strings.TrimSpace(string(output))
    // Формат: { 1.23 0.56 0.34 }
    outputStr = strings.Trim(outputStr, "{} ")
    fields := strings.Fields(outputStr)
    
    if len(fields) >= 3 {
        load1, err1 := strconv.ParseFloat(fields[0], 64)
        load5, err2 := strconv.ParseFloat(fields[1], 64)
        load15, err3 := strconv.ParseFloat(fields[2], 64)
        
        if err1 == nil && err2 == nil && err3 == nil {
            return load1, load5, load15, nil
        }
    }
    
    return 0, 0, 0, fmt.Errorf("invalid loadavg format")
}

// Обновить функцию getLoadAverage
func getLoadAverage() (float64, float64, float64, error) {
    switch runtime.GOOS {
    case "linux":
        return getLinuxLoadAverage()
    case "windows":
        return getWindowsLoadAverage()
    case "darwin":
        return getDarwinLoadAverage()
    default:
        return 0, 0, 0, fmt.Errorf("load average not supported on %s", runtime.GOOS)
    }
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