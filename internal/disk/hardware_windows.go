//go:build windows

package disk

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "strconv"
    "strings"
)

// WindowsPhysicalDisk представляет физический диск в Windows
type WindowsPhysicalDisk struct {
    DeviceId      int    `json:"DeviceId"`
    MediaType     string `json:"MediaType"`
    BusType       string `json:"BusType"`
    Size          uint64 `json:"Size"`
    FriendlyName  string `json:"FriendlyName"`
    SerialNumber  string `json:"SerialNumber"`
	Manufacturer  string `json:"Manufacturer"` 
    OperationalStatus string `json:"OperationalStatus"`
    HealthStatus  string `json:"HealthStatus"`
}

// detectAllStorageDevicesWindows обнаруживает все устройства хранения в Windows
func detectAllStorageDevicesWindows() ([]DiskInfo, error) {
    var devices []DiskInfo

    // Метод 1: PowerShell для получения детальной информации о физических дисках
    devices1, err1 := getPhysicalDisksViaPowerShell()
    if err1 == nil && len(devices1) > 0 {
        devices = append(devices, devices1...)
    }

    // Метод 2: WMI для получения базовой информации
    devices2, err2 := getStorageDevicesViaWMI()
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
                MountedOn:  "Unknown",
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

// getPhysicalDisksViaPowerShell получает информацию о физических дисках через PowerShell
func getPhysicalDisksViaPowerShell() ([]DiskInfo, error) {
    var disks []DiskInfo

    // Команда PowerShell для получения информации о физических дисках
    cmd := exec.Command("powershell", "-Command", 
        "Get-PhysicalDisk | Select-Object DeviceId, MediaType, BusType, Size, FriendlyName, SerialNumber, OperationalStatus, HealthStatus, Manufacturer | ConvertTo-Json")

    output, err := cmd.Output()
    if err != nil {
        return disks, err
    }

    // Декодируем JSON
    var physicalDisks []WindowsPhysicalDisk
    if err := json.Unmarshal(output, &physicalDisks); err != nil {
        // Пробуем как одиночный объект
        var singleDisk WindowsPhysicalDisk
        if err := json.Unmarshal(output, &singleDisk); err == nil {
            physicalDisks = []WindowsPhysicalDisk{singleDisk}
        }
    }

    // Конвертируем в DiskInfo
    for _, pd := range physicalDisks {
        disk := DiskInfo{
            Filesystem:  fmt.Sprintf("PhysicalDisk%d", pd.DeviceId),
            Size:        formatBytes(pd.Size),
            Used:        "Unknown",
            Available:   "Unknown",
            UsePercent:  0.0,
            MountedOn:   fmt.Sprintf("Disk %d", pd.DeviceId),
            Physical:    true,
            DiskType:    convertMediaType(pd.MediaType),
            Model:       pd.FriendlyName,
            Serial:      pd.SerialNumber,
            Vendor:      pd.Manufacturer, // Добавляем вендора
            SMARTStatus: convertHealthStatus(pd.HealthStatus),
        }

        // Получаем дополнительную информацию через Get-Disk
        if details, err := getDiskDetails(pd.DeviceId); err == nil {
            if details.PartitionStyle != "" {
                disk.DiskType = details.PartitionStyle
            }
        }

        disks = append(disks, disk)
    }

    return disks, nil
}

// WindowsDiskDetails содержит детальную информацию о диске
type WindowsDiskDetails struct {
    PartitionStyle string `json:"PartitionStyle"`
    NumberOfPartitions int `json:"NumberOfPartitions"`
    IsBoot bool `json:"IsBoot"`
    IsSystem bool `json:"IsSystem"`
}

// getDiskDetails получает детальную информацию о диске
func getDiskDetails(deviceId int) (WindowsDiskDetails, error) {
    var details WindowsDiskDetails

    cmd := exec.Command("powershell", "-Command", 
        fmt.Sprintf("Get-Disk -Number %d | Select-Object PartitionStyle, NumberOfPartitions, IsBoot, IsSystem | ConvertTo-Json", deviceId))

    output, err := cmd.Output()
    if err != nil {
        return details, err
    }

    if err := json.Unmarshal(output, &details); err != nil {
        return details, err
    }

    return details, nil
}

// getStorageDevicesViaWMI получает информацию о устройствах хранения через WMI
func getStorageDevicesViaWMI() ([]DiskInfo, error) {
    var disks []DiskInfo

    cmd := exec.Command("wmic", "diskdrive", "get", "Caption,Size,Model,SerialNumber,MediaType,InterfaceType", "/format:csv")
    output, err := cmd.Output()
    if err != nil {
        return disks, err
    }

    lines := strings.Split(string(output), "\n")
    for i, line := range lines {
        if i == 0 || strings.TrimSpace(line) == "" {
            continue
        }

        fields := strings.Split(line, ",")
        if len(fields) >= 7 {
            disk := DiskInfo{
                Filesystem: strings.TrimSpace(fields[1]),
                Model:      strings.TrimSpace(fields[3]),
                Serial:     strings.TrimSpace(fields[4]),
                Physical:   true,
                DiskType:   strings.TrimSpace(fields[5]),
                SMARTStatus: "UNKNOWN",
            }

            // Размер
            if sizeStr := strings.TrimSpace(fields[2]); sizeStr != "" {
                if size, err := strconv.ParseUint(sizeStr, 10, 64); err == nil {
                    disk.Size = formatBytes(size)
                }
            }

            // Тип интерфейса
            if interfaceType := strings.TrimSpace(fields[6]); interfaceType != "" {
                if disk.DiskType == "Unknown" {
                    disk.DiskType = interfaceType
                }
            }

            disks = append(disks, disk)
        }
    }

    return disks, nil
}

// convertMediaType конвертирует тип носителя Windows в читаемый формат
func convertMediaType(mediaType string) string {
    switch strings.ToLower(mediaType) {
    case "ssd", "solid state drive":
        return "SSD"
    case "hdd", "hard disk drive":
        return "HDD"
    case "sas", "scsi":
        return "Enterprise"
    case "virtual", "vhd":
        return "Virtual"
    case "removable":
        return "Removable"
    default:
        return "Unknown"
    }
}

// convertHealthStatus конвертирует статус здоровья Windows
func convertHealthStatus(healthStatus string) string {
    switch strings.ToLower(healthStatus) {
    case "healthy", "ok":
        return "PASSED"
    case "warning":
        return "WARNING"
    case "unhealthy", "critical":
        return "FAILED"
    default:
        return "UNKNOWN"
    }
}

// formatBytes форматирует байты в читаемый формат
func formatBytes(bytes uint64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
        TB = GB * 1024
    )

    switch {
    case bytes >= TB:
        return fmt.Sprintf("%.1f TB", float64(bytes)/float64(TB))
    case bytes >= GB:
        return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
    case bytes >= MB:
        return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
    case bytes >= KB:
        return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
    default:
        return fmt.Sprintf("%d B", bytes)
    }
}
