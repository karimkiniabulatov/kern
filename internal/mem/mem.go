package mem

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type MemoryInfo struct {
	Total           string
	Used            string
	Free            string
	Available       string
	SwapTotal       string
	SwapUsed        string
	SwapFree        string
	UsagePercent    float64
	SwapUsagePercent float64
}

func Summary() (*MemoryInfo, error) {
	// Используем free для получения информации о памяти
	cmd := exec.Command("free", "-h")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseFreeOutput(string(output))
}

func parseFreeOutput(output string) (*MemoryInfo, error) {
	lines := strings.Split(output, "\n")
	info := &MemoryInfo{}

	// Получаем raw данные для расчета процентов
	cmdRaw := exec.Command("free")
	outputRaw, err := cmdRaw.Output()
	if err == nil {
		info.UsagePercent, info.SwapUsagePercent = calculateUsagePercent(string(outputRaw))
	}

	// Парсим основную память
	if len(lines) >= 2 {
		memFields := parseFreeLine(lines[1])
		if len(memFields) >= 7 {
			info.Total = memFields[1]
			info.Used = memFields[2]
			info.Free = memFields[3]
			info.Available = memFields[6]
		}
	}

	// Парсим swap
	if len(lines) >= 3 {
		swapFields := parseFreeLine(lines[2])
		if len(swapFields) >= 5 {
			info.SwapTotal = swapFields[1]
			info.SwapUsed = swapFields[2]
			info.SwapFree = swapFields[3]
		}
	}

	return info, nil
}

func parseFreeLine(line string) []string {
	re := regexp.MustCompile(`\s+`)
	fields := re.Split(strings.TrimSpace(line), -1)
	return fields
}

func calculateUsagePercent(output string) (float64, float64) {
	lines := strings.Split(output, "\n")
	var memPercent, swapPercent float64

	// Memory usage
	if len(lines) >= 2 {
		memFields := parseFreeLine(lines[1])
		if len(memFields) >= 7 {
			total, _ := strconv.ParseFloat(memFields[1], 64)
			used, _ := strconv.ParseFloat(memFields[2], 64)
			if total > 0 {
				memPercent = (used / total) * 100
			}
		}
	}

	// Swap usage
	if len(lines) >= 3 {
		swapFields := parseFreeLine(lines[2])
		if len(swapFields) >= 5 {
			total, _ := strconv.ParseFloat(swapFields[1], 64)
			used, _ := strconv.ParseFloat(swapFields[2], 64)
			if total > 0 {
				swapPercent = (used / total) * 100
			}
		}
	}

	return memPercent, swapPercent
}