package mem

import (
	"strconv"
	"strings"

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
}

func Summary() (*MemoryInfo, error) {
	return getMemoryInfo()
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

	info := &MemoryInfo{
		Total:            formatBytes(virtMem.Total),
		Used:             formatBytes(virtMem.Used),
		Free:             formatBytes(virtMem.Free),
		Available:        formatBytes(virtMem.Available),
		SwapTotal:        formatBytes(swapMem.Total),
		SwapUsed:         formatBytes(swapMem.Used),
		SwapFree:         formatBytes(swapMem.Free),
		UsagePercent:     virtMem.UsedPercent,
		SwapUsagePercent: swapMem.UsedPercent,
	}

	return info, nil
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