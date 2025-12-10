//go:build darwin

package disk

import (
    "fmt"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
)

// Добавляем поле SizeBytes в DiskInfo для macOS
type DiskInfoExtended struct {
    DiskInfo
    SizeBytes uint64
}

// detectAllStorageDevicesDarwin обнаруживает все устройства хранения в macOS
func detectAllStorageDevicesDarwin(detailed bool) ([]DiskInfo, error) {
    var devices []DiskInfo

    // Метод 1: system_profiler для получения детальной информации о хранилище
    devices1, err1 := getStorageViaSystemProfiler()
    if err1 == nil && len(devices1) > 0 {
        devices = append(devices, devices1...)
    }

    // Метод 2: diskutil для получения информации о дисках
    devices2, err2 := getStorageViaDiskUtil()
    if err2 == nil && len(devices2) > 0 {
        devices = append(devices, devices2...)
    }

    // Если ничего не найдено, возвращаем минимальный набор данных
    if len(devices) == 0 {
        devices = []DiskInfo{
            {
                Filesystem: "Unknown",
                Size:       "0 GB",
                Used:       "0 GB",
                Available:  "0 GB",
                UsePercent: 0.0,
                MountedOn:  "/",
                Physical:   true,
                DiskType:   "Unknown",
                Model:      "Unknown",
                Serial:     "Unknown",
                SMARTStatus: "UNKNOWN",
            },
        }
    }

    return devices, nil
}

// getStorageViaSystemProfiler получает информацию о хранилище через system_profiler
func getStorageViaSystemProfiler() ([]DiskInfo, error) {
    var disks []DiskInfo

    cmd := exec.Command("system_profiler", "SPStorageDataType")
    output, err := cmd.Output()
    if err != nil {
        return disks, err
    }

    lines := strings.Split(string(output), "\n")
    var currentDisk *DiskInfoExtended
    var inStorageSection bool

    for _, line := range lines {
        line = strings.TrimSpace(line)

        // Начало нового раздела хранилища
        if strings.HasPrefix(line, "Storage:") {
            if currentDisk != nil {
                disks = append(disks, currentDisk.DiskInfo)
            }
            currentDisk = &DiskInfoExtended{
                DiskInfo: DiskInfo{
                    Physical:   true,
                    SMARTStatus: "UNKNOWN",
                },
            }
            inStorageSection = true
            continue
        }

        if !inStorageSection || currentDisk == nil {
            continue
        }

        // Конец раздела
        if line == "" && currentDisk.Filesystem != "" {
            disks = append(disks, currentDisk.DiskInfo)
            currentDisk = nil
            inStorageSection = false
            continue
        }

        // Парсим информацию о диске
        switch {
        case strings.HasPrefix(line, "BSD Name:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                currentDisk.Filesystem = "/dev/" + strings.TrimSpace(parts[1])
            }
        case strings.HasPrefix(line, "Mount Point:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                currentDisk.MountedOn = strings.TrimSpace(parts[1])
            }
        case strings.HasPrefix(line, "Capacity:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                sizeStr := strings.TrimSpace(parts[1])
                if size, err := parseStorageSize(sizeStr); err == nil {
                    currentDisk.Size = formatBytesMac(size)
                    currentDisk.SizeBytes = size
                }
            }
        case strings.HasPrefix(line, "Available:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                availStr := strings.TrimSpace(parts[1])
                if avail, err := parseStorageSize(availStr); err == nil {
                    currentDisk.Available = formatBytesMac(avail)
                    // Рассчитываем использованное пространство
                    if currentDisk.SizeBytes > 0 {
                        used := currentDisk.SizeBytes - avail
                        currentDisk.Used = formatBytesMac(used)
                        currentDisk.UsePercent = float64(used) / float64(currentDisk.SizeBytes) * 100
                    }
                }
            }
        case strings.HasPrefix(line, "Device Name:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                currentDisk.Model = strings.TrimSpace(parts[1])
            }
        case strings.HasPrefix(line, "Media Name:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                if currentDisk.Model == "" {
                    currentDisk.Model = strings.TrimSpace(parts[1])
                }
            }
        case strings.HasPrefix(line, "Medium Type:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                diskType := strings.TrimSpace(parts[1])
                currentDisk.DiskType = convertMacDiskType(diskType)
            }
        case strings.Contains(line, "Serial Number"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                currentDisk.Serial = strings.TrimSpace(parts[1])
            }
        case strings.Contains(line, "SMART"):
            if strings.Contains(line, "Verified") || strings.Contains(line, "Supported") {
                currentDisk.SMARTStatus = "PASSED"
            } else if strings.Contains(line, "Failing") {
                currentDisk.SMARTStatus = "FAILED"
            }
        }
    }

    // Добавляем последний диск
    if currentDisk != nil && currentDisk.Filesystem != "" {
        disks = append(disks, currentDisk.DiskInfo)
    }

    return disks, nil
}

// getStorageViaDiskUtil получает информацию о дисках через diskutil
func getStorageViaDiskUtil() ([]DiskInfo, error) {
    var disks []DiskInfo

    cmd := exec.Command("diskutil", "list")
    output, err := cmd.Output()
    if err != nil {
        return disks, err
    }

    lines := strings.Split(string(output), "\n")
    var currentDevice string
    var inDeviceSection bool

    for _, line := range lines {
        line = strings.TrimSpace(line)

        // Начало нового устройства
        if strings.HasPrefix(line, "/dev/") {
            if currentDevice != "" {
                // Получаем детали для предыдущего устройства
                if disk, err := getDiskDetails(currentDevice); err == nil {
                    disks = append(disks, disk)
                }
            }
            currentDevice = strings.Fields(line)[0]
            inDeviceSection = true
            continue
        }

        if !inDeviceSection || currentDevice == "" {
            continue
        }

        // Конец раздела устройства
        if line == "" {
            inDeviceSection = false
            continue
        }
    }

    // Обрабатываем последнее устройство
    if currentDevice != "" {
        if disk, err := getDiskDetails(currentDevice); err == nil {
            disks = append(disks, disk)
        }
    }

    return disks, nil
}

// getDiskDetails получает детальную информацию о конкретном диске
func getDiskDetails(device string) (DiskInfo, error) {
    disk := DiskInfo{
        Filesystem: device,
        Physical:   true,
        SMARTStatus: "UNKNOWN",
        Vendor:     "Unknown", // Инициализируем поле Vendor
    }

    // Получаем информацию через diskutil info
    cmd := exec.Command("diskutil", "info", device)
    output, err := cmd.Output()
    if err != nil {
        return disk, err
    }

    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        switch {
        case strings.HasPrefix(line, "Device / Media Name:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                disk.Model = strings.TrimSpace(parts[1])
                // Пытаемся определить производителя из названия модели
                disk.Vendor = extractVendorFromModel(disk.Model)
            }
        case strings.HasPrefix(line, "Disk Size:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                sizeStr := strings.TrimSpace(parts[1])
                if size, err := parseStorageSize(sizeStr); err == nil {
                    disk.Size = formatBytesMac(size)
                }
            }
        case strings.HasPrefix(line, "Device Identifier:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                disk.Serial = strings.TrimSpace(parts[1])
            }
        case strings.HasPrefix(line, "Protocol:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                protocol := strings.TrimSpace(parts[1])
                disk.DiskType = convertMacProtocol(protocol)
            }
        case strings.HasPrefix(line, "SMART Status:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                status := strings.TrimSpace(parts[1])
                disk.SMARTStatus = convertMacSMARTStatus(status)
            }
        case strings.HasPrefix(line, "Mount Point:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                disk.MountedOn = strings.TrimSpace(parts[1])
            }
        case strings.HasPrefix(line, "Vendor:"):
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                vendor := strings.TrimSpace(parts[1])
                if vendor != "" {
                    disk.Vendor = vendor
                }
            }
        }
    }

    // Если Vendor не найден в diskutil info, пытаемся получить из system_profiler
    if disk.Vendor == "Unknown" || disk.Vendor == "" {
        if vendor, err := getVendorFromSystemProfiler(device); err == nil && vendor != "" {
            disk.Vendor = vendor
        }
    }

    // Если точка монтирования не найдена, пробуем получить через df
    if disk.MountedOn == "" {
        if mount, err := getMountPoint(device); err == nil {
            disk.MountedOn = mount
        }
    }

    return disk, nil
}

func extractVendorFromModel(model string) string {
    modelUpper := strings.ToUpper(model)
    
    // Список известных производителей
    vendors := map[string][]string{
        "Apple":     {"APPLE", "MAC", "IMAC", "MACBOOK", "MAC PRO", "MAC MINI"},
        "Samsung":   {"SAMSUNG", "EVO", "PRO", "QVO", "SSD 9", "SSD 8"},
        "WD":        {"WESTERN DIGITAL", "WD", "MY PASSPORT", "MY BOOK", "BLACK", "BLUE", "RED", "GOLD"},
        "Seagate":   {"SEAGATE", "EXPANSION", "BACKUP PLUS", "IRONWOLF", "BARRACUDA"},
        "Toshiba":   {"TOSHIBA", "CANVIO", "EXTERNAL"},
        "SanDisk":   {"SANDISK", "EXTREME", "ULTRA"},
        "Crucial":   {"CRUCIAL", "MX", "BX", "P"},
        "Intel":     {"INTEL", "OPTANE", "SSDSC", "SSDPE"},
        "Kingston":  {"KINGSTON", "KC", "A", "SA", "UV"},
        "ADATA":     {"ADATA", "SU", "XPG", "SX"},
        "Transcend": {"TRANSCEND", "TS", "ESD"},
        "HP":        {"HP ", "HEWLETT"},
        "Dell":      {"DELL", "POWEREDGE"},
        "Lenovo":    {"LENOVO", "THINKPAD", "THINKCENTRE"},
        "HGST":      {"HGST", "HITACHI"},
        "PNY":       {"PNY", "CS", "XC"},
        "Corsair":   {"CORSAIR", "FORCE", "MP"},
    }

    for vendor, keywords := range vendors {
        for _, keyword := range keywords {
            if strings.Contains(modelUpper, keyword) {
                return vendor
            }
        }
    }

    // Проверяем распространенные префиксы
    if strings.HasPrefix(modelUpper, "APPLE") {
        return "Apple"
    } else if strings.HasPrefix(modelUpper, "WD") {
        return "Western Digital"
    } else if strings.HasPrefix(modelUpper, "ST") {
        return "Seagate"
    } else if strings.HasPrefix(modelUpper, "HTS") || strings.HasPrefix(modelUpper, "HTE") {
        return "HGST"
    } else if strings.HasPrefix(modelUpper, "TOSHIBA") {
        return "Toshiba"
    }

    return "Unknown"
}

// getVendorFromSystemProfiler получает информацию о производителе через system_profiler
func getVendorFromSystemProfiler(device string) (string, error) {
    // Получаем базовое имя устройства (без /dev/ и номеров разделов)
    baseDevice := strings.TrimPrefix(device, "/dev/")
    baseDevice = regexp.MustCompile(`\d+$`).ReplaceAllString(baseDevice, "")
    
    // Пробуем получить информацию через system_profiler SPSerialATADataType
    cmd := exec.Command("system_profiler", "SPSerialATADataType")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    lines := strings.Split(string(output), "\n")
    var inDeviceSection bool
    var currentVendor string

    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        // Ищем раздел для нашего устройства
        if strings.Contains(line, baseDevice) && strings.Contains(line, "BSD Name") {
            inDeviceSection = true
            continue
        }

        if inDeviceSection {
            if strings.HasPrefix(line, "Vendor:") {
                parts := strings.SplitN(line, ":", 2)
                if len(parts) == 2 {
                    currentVendor = strings.TrimSpace(parts[1])
                }
            }
            
            // Конец раздела
            if line == "" && currentVendor != "" {
                return currentVendor, nil
            }
        }
    }

    return "", fmt.Errorf("vendor not found in system_profiler")
}


// getMountPoint получает точку монтирования для устройства
func getMountPoint(device string) (string, error) {
    cmd := exec.Command("df", device)
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    lines := strings.Split(string(output), "\n")
    if len(lines) >= 2 {
        fields := strings.Fields(lines[1])
        if len(fields) >= 6 {
            return fields[5], nil
        }
    }

    return "", fmt.Errorf("mount point not found")
}

// parseStorageSize парсит размер хранилища из строки macOS
func parseStorageSize(sizeStr string) (uint64, error) {
    // Пример: "500.1 GB (500,107,862,016 bytes)"
    re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([KMGTP]?B)`)
    matches := re.FindStringSubmatch(sizeStr)
    
    if len(matches) == 3 {
        value, err := strconv.ParseFloat(matches[1], 64)
        if err != nil {
            return 0, err
        }
        
        unit := strings.ToUpper(matches[2])
        switch unit {
        case "KB", "K":
            return uint64(value * 1024), nil
        case "MB", "M":
            return uint64(value * 1024 * 1024), nil
        case "GB", "G":
            return uint64(value * 1024 * 1024 * 1024), nil
        case "TB", "T":
            return uint64(value * 1024 * 1024 * 1024 * 1024), nil
        case "PB", "P":
            return uint64(value * 1024 * 1024 * 1024 * 1024 * 1024), nil
        default:
            return uint64(value), nil
        }
    }
    
    // Пробуем найти байты в скобках
    re = regexp.MustCompile(`\(([\d,]+)\s+bytes\)`)
    matches = re.FindStringSubmatch(sizeStr)
    if len(matches) == 2 {
        bytesStr := strings.ReplaceAll(matches[1], ",", "")
        return strconv.ParseUint(bytesStr, 10, 64)
    }
    
    return 0, fmt.Errorf("cannot parse size: %s", sizeStr)
}

// formatBytesMac форматирует байты для macOS
func formatBytesMac(bytes uint64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    
    div, exp := uint64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// convertMacDiskType конвертирует тип диска macOS
func convertMacDiskType(macType string) string {
    switch strings.ToLower(macType) {
    case "solid state", "ssd", "flash storage":
        return "SSD"
    case "rotational", "hard disk", "hdd":
        return "HDD"
    case "external", "removable":
        return "External"
    case "virtual", "disk image":
        return "Virtual"
    case "network", "afp", "smb", "nfs":
        return "Network"
    default:
        return "Unknown"
    }
}

// convertMacProtocol конвертирует протокол macOS в тип диска
func convertMacProtocol(protocol string) string {
    switch strings.ToLower(protocol) {
    case "sata", "sas":
        return "SATA/SAS"
    case "pci", "nvme", "apple proprietary":
        return "NVMe"
    case "usb", "firewire", "thunderbolt":
        return "External"
    default:
        return "Unknown"
    }
}

// convertMacSMARTStatus конвертирует статус SMART macOS
func convertMacSMARTStatus(status string) string {
    switch strings.ToLower(status) {
    case "verified", "passed", "healthy":
        return "PASSED"
    case "failing", "failed", "not supported":
        return "FAILED"
    default:
        return "UNKNOWN"
    }
}