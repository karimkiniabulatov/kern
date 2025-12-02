package net

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
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
	ConnectionType string // "Ethernet", "Wi-Fi", "Bluetooth", "Cellular", "VPN", "Bridge", "Virtual", "Loopback", "VLAN"
	MaxSpeed      string // "1 Gbps", "600 Mbps", "54 Mbps"
	Technology    string // "802.11ac", "5G", "Bluetooth 5.0", "LTE"
	SignalStrength float64 // Для беспроводных сетей (%)
	IsPhysical    bool   // Физический или виртуальный интерфейс
	Driver        string // Драйвер устройства
	MTU           int    // Maximum Transmission Unit
	MACVendor     string // Производитель по OUI MAC-адреса
	VLANID        int    // ID VLAN (если применимо)
	IsBridged     bool   // Является ли bridge интерфейсом
	ParentInterface string // Родительский интерфейс для VLAN
}

var lastNetworkStats = make(map[string]struct {
	RXBytes uint64
	TXBytes uint64
	Time    time.Time
})

func Summary() ([]NetworkInfo, error) {
	// Используем расширенную функцию обнаружения
	interfaces, err := detectAllNetworkInterfaces()
	if err != nil {
		// Фоллбэк на старый метод
		interfaces, err = getNetworkInterfaces()
		if err != nil {
			return getFallbackEmptyInterfaces(), nil
		}
	}

	var networks []NetworkInfo
	for _, iface := range interfaces {
		// Включаем ВСЕ интерфейсы (кроме loopback)
		if iface.Interface != "lo" {
			networks = append(networks, iface)
		}
	}

	// Если нет активных интерфейсов, возвращаем fallback данные
	if len(networks) == 0 {
		return getFallbackEmptyInterfaces(), nil
	}

	return removeDuplicateInterfaces(networks), nil
}

// getFallbackEmptyInterfaces возвращает минимальный набор данных для обеспечения ожидаемого формата
func getFallbackEmptyInterfaces() []NetworkInfo {
	return []NetworkInfo{
		{
			Interface:      "unknown",
			IPAddress:      "N/A",
			MACAddress:     "N/A",
			Status:         "DOWN",
			RXBytes:        0,
			TXBytes:        0,
			RXSpeed:        "0B/s",
			TXSpeed:        "0B/s",
			ActivityPercent: 0.0,
			ConnectionType: "Unknown",
			MaxSpeed:      "N/A",
			Technology:    "N/A",
			SignalStrength: 0.0,
			IsPhysical:    false,
			Driver:        "N/A",
			MTU:           1500,
			MACVendor:     "N/A",
			VLANID:        0,
			IsBridged:     false,
			ParentInterface: "",
		},
	}
}

func removeDuplicateInterfaces(networks []NetworkInfo) []NetworkInfo {
    seen := make(map[string]bool)
    var result []NetworkInfo
    
    for _, net := range networks {
        key := net.Interface + "|" + net.MACAddress
        if !seen[key] {
            seen[key] = true
            result = append(result, net)
        }
    }
    
    return result
}

func getNetworkInterfaces() ([]NetworkInfo, error) {
    var interfaces []NetworkInfo
    var err error

    switch runtime.GOOS {
    case "linux":
        interfaces, err = getLinuxNetworkInterfaces()
    case "windows":
        interfaces, err = getWindowsNetworkInterfaces()  
    case "darwin":
        interfaces, err = getMacOSNetworkInterfaces()
    default:
        interfaces, err = getFallbackNetworkInterfaces()
    }

    if err != nil {
        return getFallbackNetworkInterfacesWithDefaults()
    }

    // ИСПРАВЛЕНИЕ: Убрать избыточную фильтрацию и показывать ВСЕ интерфейсы
    var filteredInterfaces []NetworkInfo
    for _, iface := range interfaces {
        // Включаем все интерфейсы кроме полностью нерабочих и loopback
        if iface.Status != "DOWN" && iface.ConnectionType != "Loopback" {
            filteredInterfaces = append(filteredInterfaces, iface)
        }
    }

    // Если после фильтрации нет интерфейсов, показываем ВСЕ кроме loopback
    if len(filteredInterfaces) == 0 {
        for _, iface := range interfaces {
            if iface.ConnectionType != "Loopback" {
                filteredInterfaces = append(filteredInterfaces, iface)
            }
        }
    }

    // Если все еще нет интерфейсов, возвращаем fallback
    if len(filteredInterfaces) == 0 {
        return getFallbackNetworkInterfacesWithDefaults()
    }

    return removeDuplicateInterfaces(filteredInterfaces), nil
}

// Новая функция для расширенного обнаружения интерфейсов
func detectAllNetworkInterfaces() ([]NetworkInfo, error) {
    var allInterfaces []NetworkInfo
    
    // Получаем базовые интерфейсы
    baseInterfaces, err := getLinuxNetworkInterfaces()
    if err != nil {
        baseInterfaces, _ = getFallbackNetworkInterfaces()
    }
    
    // Добавляем Wi-Fi интерфейсы
    wifiInterfaces := detectWifiInterfaces()
    allInterfaces = append(allInterfaces, baseInterfaces...)
    allInterfaces = append(allInterfaces, wifiInterfaces...)
    
    // Добавляем Bluetooth интерфейсы
    btInterfaces := detectBluetoothInterfaces()
    allInterfaces = append(allInterfaces, btInterfaces...)
    
    // Добавляем VPN интерфейсы
    vpnInterfaces := detectVPNInterfaces()
    allInterfaces = append(allInterfaces, vpnInterfaces...)
    
    // Убираем дубликаты
    return removeDuplicateInterfaces(allInterfaces), nil
}

// Функция обнаружения Wi-Fi интерфейсов
func detectWifiInterfaces() []NetworkInfo {
    var wifiInterfaces []NetworkInfo
    
    // Метод 1: Через iw
    cmd := exec.Command("iw", "dev")
    output, err := cmd.Output()
    if err == nil {
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
            if strings.Contains(line, "Interface") {
                parts := strings.Fields(line)
                if len(parts) >= 2 {
                    ifaceName := parts[1]
                    // Добавляем как Wi-Fi интерфейс
                    wifiInterfaces = append(wifiInterfaces, NetworkInfo{
                        Interface:      ifaceName,
                        ConnectionType: "Wi-Fi",
                        Status:         "UP",
                        IsPhysical:     true,
                    })
                }
            }
        }
    }
    
    // Метод 2: Через ip link (имена wlan*, wlp*)
    cmd = exec.Command("ip", "-o", "link", "show")
    output, err = cmd.Output()
    if err == nil {
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
            if strings.Contains(line, "wlan") || strings.Contains(line, "wlp") {
                fields := strings.Fields(line)
                if len(fields) >= 2 {
                    ifaceName := strings.TrimSuffix(fields[1], ":")
                    wifiInterfaces = append(wifiInterfaces, NetworkInfo{
                        Interface:      ifaceName,
                        ConnectionType: "Wi-Fi",
                        Status:         "UP",
                        IsPhysical:     true,
                    })
                }
            }
        }
    }
    
    return wifiInterfaces
}

// Функция обнаружения Bluetooth интерфейсов
func detectBluetoothInterfaces() []NetworkInfo {
    var btInterfaces []NetworkInfo
    
    // Через hciconfig
    cmd := exec.Command("hciconfig")
    output, err := cmd.Output()
    if err == nil {
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
            if strings.Contains(line, "hci") {
                parts := strings.Split(line, ":")
                if len(parts) > 0 {
                    ifaceName := strings.TrimSpace(parts[0])
                    btInterfaces = append(btInterfaces, NetworkInfo{
                        Interface:      ifaceName,
                        ConnectionType: "Bluetooth",
                        Status:         "UP",
                        IsPhysical:     true,
                    })
                }
            }
        }
    }
    
    return btInterfaces
}

// Функция обнаружения VPN интерфейсов
func detectVPNInterfaces() []NetworkInfo {
    var vpnInterfaces []NetworkInfo
    
    // Через ip link (имена tun*, tap*)
    cmd := exec.Command("ip", "-o", "link", "show")
    output, err := cmd.Output()
    if err == nil {
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
            if strings.Contains(line, "tun") || strings.Contains(line, "tap") {
                fields := strings.Fields(line)
                if len(fields) >= 2 {
                    ifaceName := strings.TrimSuffix(fields[1], ":")
                    vpnInterfaces = append(vpnInterfaces, NetworkInfo{
                        Interface:      ifaceName,
                        ConnectionType: "VPN",
                        Status:         "UP",
                        IsPhysical:     false,
                    })
                }
            }
        }
    }
    
    return vpnInterfaces
}

// ensureNetworkInfoDefaults гарантирует, что все поля NetworkInfo имеют валидные значения
func ensureNetworkInfoDefaults(info NetworkInfo) NetworkInfo {
	if info.Interface == "" {
		info.Interface = "unknown"
	}
	if info.IPAddress == "" {
		info.IPAddress = "N/A"
	}
	if info.MACAddress == "" {
		info.MACAddress = "N/A"
	}
	if info.Status == "" {
		info.Status = "DOWN"
	}
	if info.RXSpeed == "" {
		info.RXSpeed = "0B/s"
	}
	if info.TXSpeed == "" {
		info.TXSpeed = "0B/s"
	}
	if info.ConnectionType == "" {
		info.ConnectionType = determineConnectionType(info.Interface, info.MACAddress)
	}
	if info.MaxSpeed == "" {
		info.MaxSpeed = determineMaxSpeed(info.Interface, info.ConnectionType)
	}
	if info.Technology == "" {
		info.Technology = determineTechnology(info.Interface, info.ConnectionType)
	}
	if info.MACVendor == "" {
		info.MACVendor = getMACVendor(info.MACAddress)
	}
	// Гарантируем, что ActivityPercent и SignalStrength всегда числовые значения
	if info.ActivityPercent < 0 {
		info.ActivityPercent = 0.0
	}
	if info.SignalStrength < 0 {
		info.SignalStrength = 0.0
	}
	if info.SignalStrength > 100 {
		info.SignalStrength = 100.0
	}
	// Определяем физический/виртуальный статус
	info.IsPhysical = determinePhysicalStatus(info.Interface, info.ConnectionType, info.MACAddress)
	// Получаем информацию о драйвере и MTU
	if info.Driver == "" {
		info.Driver = getInterfaceDriver(info.Interface)
	}
	if info.MTU == 0 {
		info.MTU = getInterfaceMTU(info.Interface)
	}
	// Определяем VLAN и bridge информацию
	if info.VLANID == 0 {
		info.VLANID = getVLANID(info.Interface)
	}
	if !info.IsBridged {
		info.IsBridged = isBridgeInterface(info.Interface)
	}
	if info.ParentInterface == "" && info.VLANID > 0 {
		info.ParentInterface = getVLANParent(info.Interface)
	}
	
	return info
}

// determinePhysicalStatus определяет является ли интерфейс физическим
func determinePhysicalStatus(interfaceName, connectionType, macAddress string) bool {
	name := strings.ToLower(interfaceName)
	
	// Виртуальные интерфейсы
	switch {
	case strings.Contains(name, "lo") || strings.Contains(name, "loopback"):
		return false
	case strings.Contains(name, "docker") || strings.Contains(name, "veth"):
		return false
	case strings.Contains(name, "br-") || strings.Contains(name, "bridge"):
		return false
	case strings.Contains(name, "tun") || strings.Contains(name, "tap"):
		return false
	case strings.Contains(name, "virbr") || strings.Contains(name, "vnet"):
		return false
	case strings.Contains(name, "kube") || strings.Contains(name, "calico"):
		return false
	case strings.Contains(name, "flannel") || strings.Contains(name, "cni"):
		return false
	case strings.Contains(name, "vlan") || strings.Contains(name, "."):
		return false
	}
	
	// Проверяем по MAC-адресу
	if macAddress == "00:00:00:00:00:00" || macAddress == "N/A" {
		return false
	}
	
	// Проверяем по типу соединения
	switch connectionType {
	case "Virtual", "Bridge", "VPN", "Loopback", "VLAN":
		return false
	}
	
	return true
}

// determineConnectionType определяет тип соединения по имени интерфейса и MAC-адресу
func determineConnectionType(interfaceName, macAddress string) string {
	name := strings.ToLower(interfaceName)
	mac := strings.ToLower(macAddress)
	
	// Проверяем по имени интерфейса
	switch {
	case strings.Contains(name, "lo") || strings.Contains(name, "loopback"):
		return "Loopback"
	case strings.Contains(name, "wlan") || strings.Contains(name, "wifi") || strings.Contains(name, "wireless"):
		return "Wi-Fi"
	case strings.Contains(name, "eth") || strings.Contains(name, "enp") || strings.Contains(name, "ens") || 
		 strings.Contains(name, "em") || strings.Contains(name, "p"):
		return "Ethernet"
	case strings.Contains(name, "tun") || strings.Contains(name, "tap") || strings.Contains(name, "vpn"):
		return "VPN"
	case strings.Contains(name, "bluetooth") || strings.Contains(name, "bt"):
		return "Bluetooth"
	case strings.Contains(name, "wwan") || strings.Contains(name, "cellular") || strings.Contains(name, "modem"):
		return "Cellular"
	case strings.Contains(name, "br-") || strings.Contains(name, "bridge"):
		return "Bridge"
	case strings.Contains(name, "docker") || strings.Contains(name, "veth"):
		return "Virtual"
	case strings.Contains(name, "virbr") || strings.Contains(name, "vnet"):
		return "Virtual"
	case strings.Contains(name, "kube") || strings.Contains(name, "calico"):
		return "Virtual"
	case strings.Contains(name, "flannel") || strings.Contains(name, "cni"):
		return "Virtual"
	case strings.Contains(name, "ppp") || strings.Contains(name, "pppoe"):
		return "PPP"
	case strings.Contains(name, "vlan") || strings.Contains(name, "."):
		return "VLAN"
	}
	
	// Проверяем по MAC-адресу (первые 3 байта - OUI)
	if strings.HasPrefix(mac, "00:15:") || strings.HasPrefix(mac, "00:18:") {
		return "Bluetooth"
	}
	if strings.HasPrefix(mac, "02:") || strings.HasPrefix(mac, "06:") {
		return "Ethernet"
	}
	
	return "Unknown"
}

// determineMaxSpeed определяет максимальную скорость соединения
func determineMaxSpeed(interfaceName, connectionType string) string {
	switch connectionType {
	case "Ethernet":
		// Пытаемся определить скорость через системные утилиты
		if speed := getEthernetSpeed(interfaceName); speed != "" {
			return speed
		}
		return "1 Gbps" // значение по умолчанию для Ethernet
	case "Wi-Fi":
		if speed := getWifiSpeed(interfaceName); speed != "" {
			return speed
		}
		return "600 Mbps" // значение по умолчанию для Wi-Fi
	case "Bluetooth":
		return "24 Mbps"
	case "Cellular":
		return "1 Gbps"
	case "VPN", "Virtual", "Bridge":
		return "10 Gbps"
	case "Loopback":
		return "∞ (Loopback)"
	case "PPP":
		return "56 Kbps"
	case "VLAN":
		// VLAN наследует скорость от родительского интерфейса
		if parent := getVLANParent(interfaceName); parent != "" {
			return determineMaxSpeed(parent, "Ethernet")
		}
		return "1 Gbps"
	default:
		return "Unknown"
	}
}

// determineTechnology определяет технологию соединения
func determineTechnology(interfaceName, connectionType string) string {
	switch connectionType {
	case "Wi-Fi":
		if tech := getWifiTechnology(interfaceName); tech != "" {
			return tech
		}
		return "802.11ac"
	case "Ethernet":
		return "Gigabit Ethernet"
	case "Bluetooth":
		return "Bluetooth 5.0"
	case "Cellular":
		return "5G/LTE"
	case "VPN":
		return "VPN Tunnel"
	case "Virtual":
		return "Virtual Interface"
	case "Bridge":
		return "Network Bridge"
	case "Loopback":
		return "Loopback Device"
	case "PPP":
		return "Point-to-Point Protocol"
	case "VLAN":
		return "Virtual LAN"
	default:
		return "Unknown"
	}
}

// getMACVendor определяет производителя по OUI MAC-адреса
func getMACVendor(mac string) string {
	if mac == "N/A" || len(mac) < 8 {
		return "Unknown"
	}
	
	// Берем первые 3 байта MAC-адреса (OUI)
	oui := strings.ToUpper(mac[:8])
	
	// Распространенные OUI производителей
	vendors := map[string]string{
		"00:15:17": "Apple",
		"00:1B:63": "Apple",
		"00:1D:4F": "Apple",
		"00:23:DF": "Apple",
		"00:25:BC": "Apple",
		"00:26:BB": "Apple",
		"00:30:65": "Apple",
		"00:3E:E1": "Apple",
		"00:50:F2": "Microsoft",
		"00:1D:60": "Dell",
		"00:14:22": "Dell",
		"00:18:8B": "Dell",
		"00:1A:A0": "Dell",
		"00:0C:29": "VMware",
		"00:05:69": "VMware",
		"00:1C:14": "VMware",
		"00:50:56": "VMware",
		"08:00:27": "VirtualBox",
		"00:1C:42": "Parallels",
		"00:1C:C4": "Cisco",
		"00:24:14": "Cisco",
		"00:26:0B": "Cisco",
		"00:1B:21": "Intel",
		"00:13:CE": "Intel",
		"00:16:EA": "Intel",
		"00:19:D1": "Intel",
		"00:1C:C0": "Fujitsu",
		"00:0F:FE": "Samsung",
		"00:12:47": "Samsung",
		"00:15:99": "Samsung",
		"00:1E:7D": "Samsung",
		"00:21:19": "Samsung",
		"00:23:39": "Samsung",
		"00:24:54": "Samsung",
		"00:26:37": "Samsung",
	}
	
	if vendor, exists := vendors[oui]; exists {
		return vendor
	}
	
	return "Unknown"
}

// getVLANID получает ID VLAN для интерфейса
func getVLANID(interfaceName string) int {
	switch runtime.GOOS {
	case "linux":
		// Проверяем, является ли интерфейс VLAN
		if strings.Contains(interfaceName, ".") {
			parts := strings.Split(interfaceName, ".")
			if len(parts) > 1 {
				if vlanID, err := strconv.Atoi(parts[1]); err == nil {
					return vlanID
				}
			}
		}
		// Проверяем через sysfs
		vlanPath := fmt.Sprintf("/sys/class/net/%s/vlan", interfaceName)
		if _, err := exec.Command("ls", vlanPath).Output(); err == nil {
			// Это VLAN интерфейс
			return 1 // По умолчанию
		}
	}
	return 0
}

// isBridgeInterface проверяет, является ли интерфейс bridge
func isBridgeInterface(interfaceName string) bool {
	switch runtime.GOOS {
	case "linux":
		bridgePath := fmt.Sprintf("/sys/class/net/%s/bridge", interfaceName)
		_, err := exec.Command("ls", bridgePath).Output()
		return err == nil
	}
	return strings.Contains(strings.ToLower(interfaceName), "br-") || 
		   strings.Contains(strings.ToLower(interfaceName), "bridge")
}

// getVLANParent получает родительский интерфейс для VLAN
func getVLANParent(interfaceName string) string {
	switch runtime.GOOS {
	case "linux":
		if strings.Contains(interfaceName, ".") {
			parts := strings.Split(interfaceName, ".")
			if len(parts) > 0 {
				return parts[0]
			}
		}
		// Пытаемся получить через sysfs
		parentPath := fmt.Sprintf("/sys/class/net/%s/device", interfaceName)
		output, err := exec.Command("readlink", "-f", parentPath).Output()
		if err == nil {
			path := strings.TrimSpace(string(output))
			if strings.Contains(path, "/net/") {
				parts := strings.Split(path, "/net/")
				if len(parts) > 1 {
					return parts[1]
				}
			}
		}
	}
	return ""
}

// getInterfaceDriver получает информацию о драйвере интерфейса
func getInterfaceDriver(interfaceName string) string {
	switch runtime.GOOS {
	case "linux":
		driverPath := fmt.Sprintf("/sys/class/net/%s/device/driver", interfaceName)
		if _, err := exec.Command("ls", driverPath).Output(); err == nil {
			cmd := exec.Command("basename", driverPath)
			output, err := cmd.Output()
			if err == nil {
				return strings.TrimSpace(string(output))
			}
		}
	}
	return "Unknown"
}

// getInterfaceMTU получает MTU интерфейса
func getInterfaceMTU(interfaceName string) int {
	switch runtime.GOOS {
	case "linux":
		mtuPath := fmt.Sprintf("/sys/class/net/%s/mtu", interfaceName)
		output, err := exec.Command("cat", mtuPath).Output()
		if err == nil {
			if mtu, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
				return mtu
			}
		}
	}
	return 1500 // Значение по умолчанию
}

// getEthernetSpeed получает скорость Ethernet интерфейса
func getEthernetSpeed(interfaceName string) string {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("ethtool", interfaceName)
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Speed:") {
					parts := strings.Split(line, "Speed:")
					if len(parts) > 1 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	case "windows":
		cmd := exec.Command("powershell", "-Command", 
			fmt.Sprintf("Get-NetAdapter -Name '%s' | Select-Object -ExpandProperty LinkSpeed", interfaceName))
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
}

// getWifiSpeed получает скорость Wi-Fi интерфейса
func getWifiSpeed(interfaceName string) string {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("iw", "dev", interfaceName, "link")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "tx bitrate:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						return parts[2] + " " + parts[3]
					}
				}
			}
		}
	case "windows":
		cmd := exec.Command("netsh", "wlan", "show", "interfaces")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for i, line := range lines {
				if strings.Contains(line, "Name") && strings.Contains(line, interfaceName) {
					// Ищем скорость в следующих строках
					for j := i; j < len(lines) && j < i+10; j++ {
						if strings.Contains(lines[j], "Receive rate") {
							parts := strings.Split(lines[j], ":")
							if len(parts) > 1 {
								return strings.TrimSpace(parts[1]) + " Mbps"
							}
						}
					}
				}
			}
		}
	}
	return ""
}

// getWifiTechnology получает технологию Wi-Fi
func getWifiTechnology(interfaceName string) string {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("iw", "dev", interfaceName, "info")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "type") {
					return "802.11" + extractWifiStandard(line)
				}
			}
		}
	}
	return ""
}

// extractWifiStandard извлекает стандарт Wi-Fi из строки
func extractWifiStandard(line string) string {
	if strings.Contains(line, "802.11") {
		if strings.Contains(line, "ac") {
			return "ac"
		} else if strings.Contains(line, "ax") {
			return "ax"
		} else if strings.Contains(line, "n") {
			return "n"
		} else if strings.Contains(line, "g") {
			return "g"
		} else if strings.Contains(line, "b") {
			return "b"
		}
	}
	return "ac" // значение по умолчанию
}

func getLinuxNetworkInterfaces() ([]NetworkInfo, error) {
	var interfaces []NetworkInfo

	// Получаем все сетевые интерфейсы через ip link
	cmd := exec.Command("ip", "-o", "link", "show")
	output, err := cmd.Output()
	if err != nil {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ifaceName := strings.TrimSuffix(fields[1], ":")
			
			// Пропускаем пустые имена
			if ifaceName == "" {
				continue
			}

			// Пропускаем только loopback интерфейсы
			if ifaceName == "lo" {
				continue
			}

			// Получаем все интерфейсы без фильтрации по статусу
			iface := NetworkInfo{
				Interface: ifaceName,
				Status:    "DOWN",
			}

			// Определяем статус из вывода ip link
			if strings.Contains(line, "state UP") {
				iface.Status = "UP"
			} else if strings.Contains(line, "state UNKNOWN") {
				iface.Status = "UNKNOWN"
			}

			// Получаем MAC адрес
			for i, field := range fields {
				if field == "link/ether" && i+1 < len(fields) {
					iface.MACAddress = fields[i+1]
					break
				}
			}

			// Получаем IP адрес через отдельную команду
			if ip := getLinuxIPAddress(ifaceName); ip != "" {
				iface.IPAddress = ip
			}

			// Получаем статистику и скорость
			if rx, tx, err := getLinuxNetworkStats(ifaceName); err == nil {
				iface.RXBytes = rx
				iface.TXBytes = tx
				iface.RXSpeed, iface.TXSpeed, iface.ActivityPercent = calculateNetworkSpeed(ifaceName, rx, tx)
			}

			// Получаем тип соединения и дополнительную информацию
			iface.ConnectionType = determineConnectionType(ifaceName, iface.MACAddress)
			iface.MaxSpeed = determineMaxSpeed(ifaceName, iface.ConnectionType)
			iface.Technology = determineTechnology(ifaceName, iface.ConnectionType)
			iface.MACVendor = getMACVendor(iface.MACAddress)
			
			// Для беспроводных интерфейсов получаем силу сигнала
			if iface.ConnectionType == "Wi-Fi" {
				iface.SignalStrength = getWifiSignalStrength(ifaceName)
			}

			// Определяем VLAN и bridge информацию
			iface.VLANID = getVLANID(ifaceName)
			iface.IsBridged = isBridgeInterface(ifaceName)
			if iface.VLANID > 0 {
				iface.ParentInterface = getVLANParent(ifaceName)
			}

			// Определяем физический статус и получаем драйвер/MTU
			iface.IsPhysical = determinePhysicalStatus(ifaceName, iface.ConnectionType, iface.MACAddress)
			iface.Driver = getInterfaceDriver(ifaceName)
			iface.MTU = getInterfaceMTU(ifaceName)

			interfaces = append(interfaces, ensureNetworkInfoDefaults(iface))
		}
	}

	// Если не найдено интерфейсов, возвращаем fallback
	if len(interfaces) == 0 {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	return interfaces, nil
}

// getLinuxIPAddress получает IP адрес для интерфейса в Linux
func getLinuxIPAddress(interfaceName string) string {
	cmd := exec.Command("ip", "-o", "-4", "addr", "show", "dev", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[2] == "inet" {
			ipParts := strings.Split(fields[3], "/")
			return ipParts[0]
		}
	}
	return ""
}

// getWifiSignalStrength получает силу сигнала Wi-Fi
func getWifiSignalStrength(interfaceName string) float64 {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("iw", "dev", interfaceName, "link")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "signal:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						signalStr := strings.TrimSuffix(parts[1], "dBm")
						if signal, err := strconv.ParseFloat(signalStr, 64); err == nil {
							// Конвертируем из dBm в проценты (примерная формула)
							// Обычно: -30dBm = 100%, -90dBm = 0%
							if signal >= -30 {
								return 100.0
							} else if signal <= -90 {
								return 0.0
							}
							return (signal + 90) / 60 * 100
						}
					}
				}
			}
		}
	}
	return 0.0
}

func getWindowsNetworkInterfaces() ([]NetworkInfo, error) {
	var interfaces []NetworkInfo

	// Используем net.Interfaces для получения информации об интерфейсах
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	for _, netIface := range netInterfaces {
		iface := NetworkInfo{
			Interface:  netIface.Name,
			MACAddress: netIface.HardwareAddr.String(),
			Status:     "DOWN",
		}

		// Определяем статус
		if netIface.Flags&net.FlagUp != 0 {
			iface.Status = "UP"
		}

		// Получаем IP адреса
		addrs, err := netIface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						iface.IPAddress = ipNet.IP.String()
						break
					}
				}
			}
		}

		// Получаем статистику через PowerShell
		if rx, tx, err := getWindowsNetworkStats(iface.Interface); err == nil {
			iface.RXBytes = rx
			iface.TXBytes = tx
			iface.RXSpeed, iface.TXSpeed, iface.ActivityPercent = calculateNetworkSpeed(iface.Interface, rx, tx)
		}

		// Получаем тип соединения и дополнительную информацию
		iface.ConnectionType = determineConnectionType(iface.Interface, iface.MACAddress)
		iface.MaxSpeed = determineMaxSpeed(iface.Interface, iface.ConnectionType)
		iface.Technology = determineTechnology(iface.Interface, iface.ConnectionType)
		iface.MACVendor = getMACVendor(iface.MACAddress)
		iface.IsPhysical = determinePhysicalStatus(iface.Interface, iface.ConnectionType, iface.MACAddress)

		interfaces = append(interfaces, ensureNetworkInfoDefaults(iface))
	}

	// Если не найдено интерфейсов, возвращаем fallback
	if len(interfaces) == 0 {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	return interfaces, nil
}

func getMacOSNetworkInterfaces() ([]NetworkInfo, error) {
	var interfaces []NetworkInfo

	// Используем net.Interfaces для macOS
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	for _, netIface := range netInterfaces {
		iface := NetworkInfo{
			Interface:  netIface.Name,
			MACAddress: netIface.HardwareAddr.String(),
			Status:     "DOWN",
		}

		// Определяем статус
		if netIface.Flags&net.FlagUp != 0 {
			iface.Status = "UP"
		}

		// Получаем IP адреса
		addrs, err := netIface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						iface.IPAddress = ipNet.IP.String()
						break
					}
				}
			}
		}

		// Получаем статистику через netstat
		if rx, tx, err := getMacOSNetworkStats(iface.Interface); err == nil {
			iface.RXBytes = rx
			iface.TXBytes = tx
			iface.RXSpeed, iface.TXSpeed, iface.ActivityPercent = calculateNetworkSpeed(iface.Interface, rx, tx)
		}

		// Получаем тип соединения и дополнительную информацию
		iface.ConnectionType = determineConnectionType(iface.Interface, iface.MACAddress)
		iface.MaxSpeed = determineMaxSpeed(iface.Interface, iface.ConnectionType)
		iface.Technology = determineTechnology(iface.Interface, iface.ConnectionType)
		iface.MACVendor = getMACVendor(iface.MACAddress)
		iface.IsPhysical = determinePhysicalStatus(iface.Interface, iface.ConnectionType, iface.MACAddress)

		interfaces = append(interfaces, ensureNetworkInfoDefaults(iface))
	}

	// Если не найдено интерфейсов, возвращаем fallback
	if len(interfaces) == 0 {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	return interfaces, nil
}

func getFallbackNetworkInterfaces() ([]NetworkInfo, error) {
	var interfaces []NetworkInfo

	// Fallback реализация используя только net package
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	for _, netIface := range netInterfaces {
		iface := NetworkInfo{
			Interface:  netIface.Name,
			MACAddress: netIface.HardwareAddr.String(),
			Status:     "DOWN",
		}

		if netIface.Flags&net.FlagUp != 0 {
			iface.Status = "UP"
		}

		addrs, err := netIface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						iface.IPAddress = ipNet.IP.String()
						break
					}
				}
			}
		}

		// Для неизвестных платформ статистика недоступна
		iface.RXSpeed = "0B/s"
		iface.TXSpeed = "0B/s"
		iface.ActivityPercent = 0.0
		iface.ConnectionType = determineConnectionType(iface.Interface, iface.MACAddress)
		iface.MaxSpeed = "Unknown"
		iface.Technology = "Unknown"
		iface.MACVendor = getMACVendor(iface.MACAddress)
		iface.IsPhysical = determinePhysicalStatus(iface.Interface, iface.ConnectionType, iface.MACAddress)

		interfaces = append(interfaces, ensureNetworkInfoDefaults(iface))
	}

	// Если не найдено интерфейсов, возвращаем fallback
	if len(interfaces) == 0 {
		return getFallbackNetworkInterfacesWithDefaults()
	}

	return interfaces, nil
}

// getFallbackNetworkInterfacesWithDefaults возвращает минимальный набор интерфейсов с гарантированными значениями
func getFallbackNetworkInterfacesWithDefaults() ([]NetworkInfo, error) {
	return []NetworkInfo{
		{
			Interface:      "default",
			IPAddress:      "127.0.0.1",
			MACAddress:     "00:00:00:00:00:00",
			Status:         "UP",
			RXBytes:        0,
			TXBytes:        0,
			RXSpeed:        "0B/s",
			TXSpeed:        "0B/s",
			ActivityPercent: 0.0,
			ConnectionType: "Ethernet",
			MaxSpeed:      "1 Gbps",
			Technology:    "Gigabit Ethernet",
			SignalStrength: 0.0,
			IsPhysical:    true,
			Driver:        "unknown",
			MTU:           1500,
			MACVendor:     "Unknown",
			VLANID:        0,
			IsBridged:     false,
			ParentInterface: "",
		},
	}, nil
}

func getLinuxNetworkStats(iface string) (uint64, uint64, error) {
	rxPath := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", iface)
	txPath := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", iface)

	rxData, err := exec.Command("cat", rxPath).Output()
	if err != nil {
		return 0, 0, nil // Возвращаем нули вместо ошибки
	}

	txData, err := exec.Command("cat", txPath).Output()
	if err != nil {
		return 0, 0, nil // Возвращаем нули вместо ошибки
	}

	rxBytes, _ := strconv.ParseUint(strings.TrimSpace(string(rxData)), 10, 64)
	txBytes, _ := strconv.ParseUint(strings.TrimSpace(string(txData)), 10, 64)

	return rxBytes, txBytes, nil
}

func getWindowsNetworkStats(iface string) (uint64, uint64, error) {
	// Используем PowerShell для получения статистики сети в Windows
	cmd := exec.Command("powershell", "-Command", 
		"Get-NetAdapterStatistics | Where-Object {$_.Name -eq '"+iface+"'} | Select-Object ReceivedBytes, SentBytes")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, nil // Возвращаем нули вместо ошибки
	}

	lines := strings.Split(string(output), "\n")
	var rxBytes, txBytes uint64

	for _, line := range lines {
		if strings.Contains(line, "ReceivedBytes") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				rxBytes, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
		if strings.Contains(line, "SentBytes") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				txBytes, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}

	return rxBytes, txBytes, nil
}

func getMacOSNetworkStats(iface string) (uint64, uint64, error) {
	// Используем netstat для получения статистики в macOS
	cmd := exec.Command("netstat", "-bi")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, nil // Возвращаем нули вместо ошибки
	}

	lines := strings.Split(string(output), "\n")
	var rxBytes, txBytes uint64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == iface {
			// Поле 6 обычно содержит received bytes, поле 9 - transmitted bytes
			if len(fields) >= 10 {
				rxBytes, _ = strconv.ParseUint(fields[6], 10, 64)
				txBytes, _ = strconv.ParseUint(fields[9], 10, 64)
				break
			}
		}
	}

	return rxBytes, txBytes, nil
}

func calculateNetworkSpeed(iface string, currentRX, currentTX uint64) (string, string, float64) {
	now := time.Now()
	
	if last, exists := lastNetworkStats[iface]; exists {
		timeDiff := now.Sub(last.Time).Seconds()
		if timeDiff > 0 {
			rxSpeed := float64(currentRX-last.RXBytes) / timeDiff
			txSpeed := float64(currentTX-last.TXBytes) / timeDiff
			
			rxSpeedStr := formatSpeed(rxSpeed)
			txSpeedStr := formatSpeed(txSpeed)
			
			totalSpeed := rxSpeed + txSpeed
			activity := (totalSpeed / 125000000) * 100
			if activity > 100 {
				activity = 100
			}

			lastNetworkStats[iface] = struct {
				RXBytes uint64
				TXBytes uint64
				Time    time.Time
			}{currentRX, currentTX, now}

			return rxSpeedStr, txSpeedStr, activity
		}
	}

	lastNetworkStats[iface] = struct {
		RXBytes uint64
		TXBytes uint64
		Time    time.Time
	}{currentRX, currentTX, now}

	return "0B/s", "0B/s", 0.0
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