//go:build linux

package disk

import (
	"net/url"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
)

// detectAllStorageDevicesLinux обнаруживает все устройства хранения в Linux
func detectAllStorageDevicesLinux() ([]DiskInfo, error) {
    var devices []DiskInfo
    
    // Сканируем устройства в /sys/block
    sysBlockPath := "/sys/block"
    entries, err := os.ReadDir(sysBlockPath)
    if err != nil {
        return devices, err
    }
    
    for _, entry := range entries {
        deviceName := entry.Name()
        
        // Пропускаем виртуальные устройства
        if strings.HasPrefix(deviceName, "loop") || 
           strings.HasPrefix(deviceName, "ram") ||
           strings.HasPrefix(deviceName, "zram") {
            continue
        }
        
        devicePath := filepath.Join(sysBlockPath, deviceName)
        
        // Проверяем, является ли это устройством хранения
        if isStorageDevice(devicePath) {
            device := parseStorageDeviceLinux(devicePath, deviceName)
            if device != nil {
                devices = append(devices, *device)
            }
        }
    }
    
    return devices, nil
}

func isStorageDevice(devicePath string) bool {
    // Проверяем наличие файла размеров
    sizePath := filepath.Join(devicePath, "size")
    if _, err := os.Stat(sizePath); err != nil {
        return false
    }
    
    // Проверяем, что это не read-only устройство (типа CD-ROM)
    roPath := filepath.Join(devicePath, "ro")
    if data, err := os.ReadFile(roPath); err == nil {
        if strings.TrimSpace(string(data)) == "1" {
            return false // CD-ROM
        }
    }
    
    return true
}

func parseStorageDeviceLinux(devicePath, deviceName string) *DiskInfo {
    device := &DiskInfo{
        Filesystem: "/dev/" + deviceName,
        Physical:   true,
        SMARTStatus: "UNKNOWN",
    }
    
    // Получаем модель и серийный номер из udev
    udevCmd := exec.Command("udevadm", "info", "--query=property", "--name=/dev/" + deviceName)
    if output, err := udevCmd.Output(); err == nil {
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
            if strings.HasPrefix(line, "ID_MODEL=") {
                device.Model = strings.TrimPrefix(line, "ID_MODEL=")
            } else if strings.HasPrefix(line, "ID_SERIAL_SHORT=") {
                device.Serial = strings.TrimPrefix(line, "ID_SERIAL_SHORT=")
            } else if strings.HasPrefix(line, "ID_MODEL_ENC=") {
                // Если модель не найдена, используем закодированную версию
                if device.Model == "" {
                    encoded := strings.TrimPrefix(line, "ID_MODEL_ENC=")
                    if decoded, err := url.QueryUnescape(encoded); err == nil {
                        device.Model = decoded
                    }
                }
            }
        }
    }
    
    // Если модель все еще не найдена, пробуем другие методы
    if device.Model == "" {
        // Пробуем через hdparm
        hdparmCmd := exec.Command("hdparm", "-I", "/dev/"+deviceName)
        if output, err := hdparmCmd.Output(); err == nil {
            lines := strings.Split(string(output), "\n")
            for _, line := range lines {
                if strings.Contains(line, "Model Number:") {
                    device.Model = strings.TrimSpace(strings.TrimPrefix(line, "Model Number:"))
                } else if strings.Contains(line, "Serial Number:") {
                    device.Serial = strings.TrimSpace(strings.TrimPrefix(line, "Serial Number:"))
                }
            }
        }
    }
    
    // Получаем размер в секторах
    sizePath := filepath.Join(devicePath, "size")
    if data, err := os.ReadFile(sizePath); err == nil {
        sizeStr := strings.TrimSpace(string(data))
        // Размер в секторах по 512 байт
        if sectors, err := strconv.ParseUint(sizeStr, 10, 64); err == nil {
            sizeBytes := sectors * 512
            device.Size = formatBytes(sizeBytes)
        }
    }
    
    // Получаем модель из sys (резервный метод)
    if device.Model == "" {
        modelPath := filepath.Join(devicePath, "device", "model")
        if data, err := os.ReadFile(modelPath); err == nil {
            device.Model = strings.TrimSpace(string(data))
        }
    }
    
    // Получаем серийный номер из sys (резервный метод)
    if device.Serial == "" {
        serialPath := filepath.Join(devicePath, "device", "serial")
        if data, err := os.ReadFile(serialPath); err == nil {
            device.Serial = strings.TrimSpace(string(data))
        }
    }
    
    // Получаем vendor (резервный метод)
    if device.Model != "" {
        vendorPath := filepath.Join(devicePath, "device", "vendor")
        if data, err := os.ReadFile(vendorPath); err == nil {
            vendor := strings.TrimSpace(string(data))
            if vendor != "" && !strings.Contains(device.Model, vendor) {
                device.Model = vendor + " " + device.Model
            }
        }
    }
    
    // Определяем тип (SSD/HDD)
    rotationalPath := filepath.Join(devicePath, "queue", "rotational")
    if data, err := os.ReadFile(rotationalPath); err == nil {
        if strings.TrimSpace(string(data)) == "0" {
            device.DiskType = "SSD"
        } else {
            device.DiskType = "HDD"
        }
    }
    
    // Проверяем съемность
    removablePath := filepath.Join(devicePath, "removable")
    if data, err := os.ReadFile(removablePath); err == nil {
        if strings.TrimSpace(string(data)) == "1" {
            device.DiskType = "Removable"
            // Для съемных устройств пытаемся определить USB
            if isUSBDevice(devicePath) {
                device.DiskType = "USB"
            }
        }
    }
    
    // Проверяем NVMe
    if strings.HasPrefix(deviceName, "nvme") {
        device.DiskType = "NVMe"
    }
    
    // Получаем SMART статус
    smartStatus := getSMARTStatus(deviceName)
    if smartStatus != "Unavailable" {
        device.SMARTStatus = smartStatus
    }
    
    return device
}

func isUSBDevice(devicePath string) bool {
    // Проверяем, является ли устройство USB
    subsystemPath := filepath.Join(devicePath, "device", "subsystem")
    if data, err := exec.Command("readlink", "-f", subsystemPath).Output(); err == nil {
        path := strings.TrimSpace(string(data))
        if strings.Contains(path, "usb") {
            return true
        }
    }
    
    // Альтернативный метод: проверка через uevents
    ueventPath := filepath.Join(devicePath, "device", "uevent")
    if data, err := os.ReadFile(ueventPath); err == nil {
        content := string(data)
        if strings.Contains(content, "DRIVER=usb-storage") || 
           strings.Contains(content, "MODALIAS=usb:") {
            return true
        }
    }
    
    return false
}