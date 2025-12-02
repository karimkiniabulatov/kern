//go:build linux

package gpu

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

func detectGPUsViaPCIHardware() ([]*GPUInfo, error) {
    gpus := detectPCILinux()
    return gpus, nil
}

func detectPCILinux() []*GPUInfo {
    var gpus []*GPUInfo
    
    // Сканируем PCI устройства в /sys/bus/pci/devices
    devicesPath := "/sys/bus/pci/devices"
    entries, err := os.ReadDir(devicesPath)
    if err != nil {
        return gpus
    }
    
    for _, entry := range entries {
        devicePath := filepath.Join(devicesPath, entry.Name())
        
        // Проверяем класс устройства (03 - display controller)
        classPath := filepath.Join(devicePath, "class")
        classData, err := os.ReadFile(classPath)
        if err != nil {
            continue
        }
        
        // Класс 0x0300 - VGA compatible controller
        classStr := strings.TrimSpace(string(classData))
        if len(classStr) < 4 {
            continue
        }
        
        class, err := strconv.ParseUint(classStr[2:4], 16, 8)
        if err != nil {
            continue
        }
        
        // 03 = display controller class
        if class == 0x03 {
            gpu := parsePCIDeviceLinux(devicePath, entry.Name())
            if gpu != nil {
                gpus = append(gpus, gpu)
            }
        }
    }
    
    return gpus
}

func parsePCIDeviceLinux(devicePath, deviceName string) *GPUInfo {
    // Читаем vendor и device ID
    vendorPath := filepath.Join(devicePath, "vendor")
    deviceIDPath := filepath.Join(devicePath, "device")
    
    vendorData, err := os.ReadFile(vendorPath)
    if err != nil {
        return nil
    }
    
    deviceData, err := os.ReadFile(deviceIDPath)
    if err != nil {
        return nil
    }
    
    vendorStr := strings.TrimSpace(string(vendorData))
    deviceStr := strings.TrimSpace(string(deviceData))
    
    if len(vendorStr) < 3 || len(deviceStr) < 3 {
        return nil
    }
    
    vendorID := vendorStr[2:] // убираем 0x
    deviceID := deviceStr[2:] // убираем 0x
    
    // Определяем производителя
    vendorName := getVendorName(vendorID)
    deviceName := getDeviceName(vendorID, deviceID)
    
    gpu := &GPUInfo{
        Model:           fmt.Sprintf("%s %s", vendorName, deviceName),
        DriverVersion:   "Hardware Detected (PCI ID: " + vendorID + ":" + deviceID + ")",
        GPUTemp:         0.0,
        MemoryTotal:     "0 MB",
        MemoryUsed:      "0 MB",
        MemoryFree:      "0 MB",
        Utilization:     0.0,
        PowerDraw:       "0 W",
        PowerLimit:      "0 W",
        FanSpeed:        0.0,
        ClockCore:       "0 MHz",
        ClockMemory:     "0 MHz",
        PerformanceState: "Active",
    }
    
    return gpu
}