package net

import (
	"fmt"
	"os/exec"
	//"regexp"
	"strconv"
	"strings"
	"time"
)

type NetworkInfo struct {
	Interface      string
	IPAddress      string
	MACAddress     string
	Status         string
	RXBytes        uint64
	TXBytes        uint64
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

	// Remove duplicates - keep only unique interfaces
	return removeDuplicateInterfaces(networks), nil
}

func removeDuplicateInterfaces(networks []NetworkInfo) []NetworkInfo {
	seen := make(map[string]bool)
	var result []NetworkInfo
	
	for _, net := range networks {
		if !seen[net.Interface] {
			seen[net.Interface] = true
			result = append(result, net)
		}
	}
	
	return result
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
				// Extract only the IP address without subnet mask
				ipParts := strings.Split(fields[3], "/")
				iface.IPAddress = ipParts[0]
			}

			// Получаем MAC адрес и статус
			if mac, status, err := getMACAndStatus(iface.Interface); err == nil {
				iface.MACAddress = mac
				iface.Status = status
			}

			// Получаем статистику и скорость
			if rx, tx, err := getNetworkStatsRaw(iface.Interface); err == nil {
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

func getNetworkStatsRaw(iface string) (uint64, uint64, error) {
	// Получаем сырые данные из /sys/class/net/
	rxPath := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", iface)
	txPath := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", iface)

	rxData, err := exec.Command("cat", rxPath).Output()
	if err != nil {
		return 0, 0, err
	}

	txData, err := exec.Command("cat", txPath).Output()
	if err != nil {
		return 0, 0, err
	}

	rxBytes, _ := strconv.ParseUint(strings.TrimSpace(string(rxData)), 10, 64)
	txBytes, _ := strconv.ParseUint(strings.TrimSpace(string(txData)), 10, 64)

	return rxBytes, txBytes, nil
}

func calculateNetworkSpeed(iface string, currentRX, currentTX uint64) (string, string, float64) {
	now := time.Now()
	
	// Проверяем есть ли предыдущие статистики
	if last, exists := lastNetworkStats[iface]; exists {
		timeDiff := now.Sub(last.Time).Seconds()
		if timeDiff > 0 {
			rxSpeed := float64(currentRX-last.RXBytes) / timeDiff
			txSpeed := float64(currentTX-last.TXBytes) / timeDiff
			
			// Конвертируем в человекочитаемый формат
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

func formatSpeed(bytesPerSec float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
	value := bytesPerSec
	unitIndex := 0

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%.0f%s", value, units[unitIndex])
	}
	return fmt.Sprintf("%.1f%s", value, units[unitIndex])
}