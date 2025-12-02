//go:build linux

package disk

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// detectAllStorageDevices обнаруживает все устройства хранения в Linux
func detectAllStorageDevices() ([]DiskInfo, error) {
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
            device := parseStorageDevice(devicePath, deviceName)
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

func parseStorageDevice(devicePath, deviceName string) *DiskInfo {
    device := &DiskInfo{
        Filesystem: "/dev/" + deviceName,
        Physical:   true,
    }
    
    // Получаем размер
    sizePath := filepath.Join(devicePath, "size")
    if data, err := os.ReadFile(sizePath); err == nil {
        sizeStr := strings.TrimSpace(string(data))
        // Размер в секторах по 512 байт
        // Конвертируем в байты и форматируем
    }
    
    // Получаем модель
    modelPath := filepath.Join(devicePath, "device", "model")
    if data, err := os.ReadFile(modelPath); err == nil {
        device.Model = strings.TrimSpace(string(data))
    }
    
    // Получаем серийный номер
    serialPath := filepath.Join(devicePath, "device", "serial")
    if data, err := os.ReadFile(serialPath); err == nil {
        device.Serial = strings.TrimSpace(string(data))
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
        }
    }
    
    return device
}