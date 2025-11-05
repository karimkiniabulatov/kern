package cpu

import (
	"os" //добавлено
	"os/exec"
	"strconv"
	"strings"
)

type CPUInfo struct {
	Model      string
	Cores      int
	Threads    int
	Usage      float64
	Frequency  string
	Load1      float64
	Load5      float64
	Load15     float64
}

func Summary() (*CPUInfo, error) {
	info := &CPUInfo{}

	// Get CPU model and cores from /proc/cpuinfo
	if model, cores, err := getCPUInfo(); err == nil {
		info.Model = model
		info.Cores = cores
	}

	// Get load average
	if load1, load5, load15, err := getLoadAverage(); err == nil {
		info.Load1 = load1
		info.Load5 = load5
		info.Load15 = load15
	}

	// Get CPU frequency
	if freq, err := getCPUFrequency(); err == nil {
		info.Frequency = freq
	}

	return info, nil
}

func getCPUInfo() (string, int, error) {
	cmd := exec.Command("lscpu")
	output, err := cmd.Output()
	if err != nil {
		return "", 0, err
	}

	lines := strings.Split(string(output), "\n")
	var model string
	var cores int

	for _, line := range lines {
		if strings.HasPrefix(line, "Model name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				model = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "CPU(s):") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cores, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		}
	}

	return model, cores, nil
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
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CPU max MHz:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]) + " MHz", nil
			}
		}
	}

	return "Unknown", nil
}