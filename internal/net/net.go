package net

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type NetworkInfo struct {
	Interface      string
	IPAddress      string
	MACAddress     string
	Status         string
	RXBytes        string
	TXBytes        string
	RXSpeed        string
	TXSpeed        string
	ActivityPercent float64
}

var lastNetworkStats = make(map[string]struct {
	RXBytes uint64
	TXBytes uint64
	Time    time.Time
})

func Summary() ([]NetworkInfo, error) {
	interfaces, err := getNetworkInterfaces()
	if err != nil {
		return nil, err
	}

	var networks []NetworkInfo
	for _, iface := range interfaces {
		if iface.Status == "UP" && iface.Interface != "lo" {
			networks = append(networks, iface)
		}
	}

	return networks, nil
}

func getNetworkInterfaces() ([]NetworkInfo, error) {
	var interfaces []NetworkInfo

	// Получаем базовую информацию об интерфейсах
	cmd := exec.Command("ip", "-o", "addr", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			iface := NetworkInfo{
				Interface: fields[1],
				Status:    "DOWN",
			}

			// Получаем IP адрес
			if len(fields) >= 6 && fields[2] == "inet" {
				iface.IPAddress = fields[3]
			}

			// Получаем MAC адрес и статус
			if mac, status, err := getMACAndStatus(iface.Interface); err == nil {
				iface.MACAddress = mac
				iface.Status = status
			}

			// Получаем статистику и скорость
			if rx, tx, err := getNetworkStats(iface.Interface); err == nil {
				iface.RXBytes = rx
				iface.TXBytes = tx
				iface.RXSpeed, iface.TXSpeed, iface.ActivityPercent = calculateNetworkSpeed(iface.Interface, rx, tx)
			}

			interfaces = append(interfaces, iface)
		}
	}

	return interfaces, nil
}

func getMACAndStatus(iface string) (string, string, error) {
	cmd := exec.Command("ip", "link", "show", iface)
	output, err := cmd.Output()
	if err != nil {
		return "", "DOWN", err
	}

	lines := strings.Split(string(output), "\n")
	var mac, status string

	for _, line := range lines {
		if strings.Contains(line, "link/ether") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				mac = fields[1]
			}
		}
		if strings.Contains(line, "state") {
			if strings.Contains(line, "state UP") {
				status = "UP"
			} else {
				status = "DOWN"
			}
		}
	}

	return mac, status, nil
}

func getNetworkStats(iface string) (string, string, error) {
	// Получаем статистику в человекочитаемом формате
	cmd := exec.Command("ip", "-s", "-h", "link", "show", iface)
	output, err := cmd.Output()
	if err != nil {
		return "0", "0", err
	}

	var rxBytes, txBytes string
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if strings.Contains(line, "RX:") && i+1 < len(lines) {
			rxLine := strings.TrimSpace(lines[i+1])
			rxBytes = extractBytes(rxLine)
		}
		if strings.Contains(line, "TX:") && i+1 < len(lines) {
			txLine := strings.TrimSpace(lines[i+1])
			txBytes = extractBytes(txLine)
		}
	}

	return rxBytes, txBytes, nil
}

func extractBytes(line string) string {
	re := regexp.MustCompile(`(\d+\.?\d*)([KMGT]?)(i?)B`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		return matches[1] + matches[2] + matches[3] + "B"
	}
	return "0B"
}

func calculateNetworkSpeed(iface, rxBytes, txBytes string) (string, string, float64) {
	now := time.Now()
	
	// Конвертируем в байты для расчета
	currentRX := convertToBytes(rxBytes)
	currentTX := convertToBytes(txBytes)

	// Проверяем есть ли предыдущие статистики
	if last, exists := lastNetworkStats[iface]; exists {
		timeDiff := now.Sub(last.Time).Seconds()
		if timeDiff > 0 {
			rxSpeed := float64(currentRX-last.RXBytes) / timeDiff
			txSpeed := float64(currentTX-last.TXBytes) / timeDiff
			
			// Конвертируем обратно в человекочитаемый формат
			rxSpeedStr := formatSpeed(rxSpeed)
			txSpeedStr := formatSpeed(txSpeed)
			
			// Рассчитываем общую активность (в процентах от 1 Гбит/с)
			totalSpeed := rxSpeed + txSpeed
			activity := (totalSpeed / 125000000) * 100 // 1 Гбит/с = 125000000 байт/с
			if activity > 100 {
				activity = 100
			}

			// Обновляем последние статистики
			lastNetworkStats[iface] = struct {
				RXBytes uint64
				TXBytes uint64
				Time    time.Time
			}{currentRX, currentTX, now}

			return rxSpeedStr, txSpeedStr, activity
		}
	}

	// Первое измерение
	lastNetworkStats[iface] = struct {
		RXBytes uint64
		TXBytes uint64
		Time    time.Time
	}{currentRX, currentTX, now}

	return "0B/s", "0B/s", 0
}

func convertToBytes(size string) uint64 {
	re := regexp.MustCompile(`(\d+\.?\d*)([KMGT]?)(i?)B`)
	matches := re.FindStringSubmatch(size)
	if len(matches) < 3 {
		return 0
	}

	value, _ := strconv.ParseFloat(matches[1], 64)
	unit := matches[2]

	switch unit {
	case "K":
		return uint64(value * 1024)
	case "M":
		return uint64(value * 1024 * 1024)
	case "G":
		return uint64(value * 1024 * 1024 * 1024)
	case "T":
		return uint64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return uint64(value)
	}
}

func formatSpeed(bytesPerSec float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
	value := bytesPerSec
	unitIndex := 0

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.1f%s", value, units[unitIndex])
}

// Вспомогательная функция для форматирования
func fmt.Sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}