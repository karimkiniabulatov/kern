//go:build linux || windows || darwin

package gpu

import (
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
)

// PCIDevice представляет PCI устройство
type PCIDevice struct {
    VendorID uint16
    DeviceID uint16
    Class    uint8
    Subclass uint8
    Vendor   string
    Device   string
    Bus      string
    Slot     string
    Function string
}

// detectGPUsViaPCIHardware - аппаратное обнаружение через PCI
func detectGPUsViaPCIHardware() ([]*GPUInfo, error) {
    var gpus []*GPUInfo
    
    switch runtime.GOOS {
    case "linux":
        gpus = detectPCILinux()
    case "windows":
        gpus = detectPCIWindows()
    case "darwin":
        gpus = detectPCIMacOS()
    }
    
    return gpus, nil
}

// detectPCILinux - обнаружение PCI на Linux через /sys/bus/pci
func detectPCILinux() []*GPUInfo {
    var gpus []*GPUInfo
    
    // Сканируем PCI устройства в /sys/bus/pci/devices
    devicesPath := "/sys/bus/pci/devices"
    entries, err := os.ReadDir(devicesPath)
    if err != nil {
        return gpus
    }
    
    for _, entry := range entries {
        devicePath := devicesPath + "/" + entry.Name()
        
        // Проверяем класс устройства (03 - display controller)
        classPath := devicePath + "/class"
        classData, err := os.ReadFile(classPath)
        if err != nil {
            continue
        }
        
        // Класс 0x0300 - VGA compatible controller
        // Класс 0x0380 - Display controller
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

// parsePCIDeviceLinux - парсит информацию о PCI устройстве на Linux
func parsePCIDeviceLinux(devicePath, deviceName string) *GPUInfo {
    // Читаем vendor и device ID
    vendorPath := devicePath + "/vendor"
    deviceIDPath := devicePath + "/device"
    
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
        Model:           fmt.Sprintf("%s %s (PCI ID: %s:%s)", vendorName, deviceName, vendorID, deviceID),
        DriverVersion:   "Hardware Detected",
        GPUTemp:         0.0,
        MemoryTotal:     "0 MB",  // Будет определено через драйвер
        MemoryUsed:      "0 MB",
        MemoryFree:      "0 MB",
        Utilization:     0.0,     // 0% для отображения гистограммы
        PowerDraw:       "0 W",
        PowerLimit:      "0 W",
        FanSpeed:        0.0,
        ClockCore:       "0 MHz",
        ClockMemory:     "0 MHz",
        PerformanceState: "Active",
    }
    
    return gpu
}

// getVendorName - возвращает имя производителя по ID
func getVendorName(vendorID string) string {
    switch vendorID {
    case "10de":
        return "NVIDIA"
    case "1002":
        return "AMD"
    case "8086":
        return "Intel"
    case "102b":
        return "Matrox"
    case "1a03":
        return "ASPEED"
    default:
        return "Unknown Vendor"
    }
}

// getDeviceName - возвращает примерное имя устройства
func getDeviceName(vendorID, deviceID string) string {
    // База данных известных устройств (упрощенная)
    devices := map[string]map[string]string{
        "10de": { // NVIDIA
            "1b80": "GeForce GTX 1080",
            "1c81": "GeForce GTX 1060",
            "1f08": "GeForce RTX 2060",
            "2204": "GeForce RTX 3080",
        },
        "1002": { // AMD
            "67df": "Radeon RX 580",
            "7340": "Radeon RX 5700 XT",
            "73bf": "Radeon RX 6900 XT",
        },
        "8086": { // Intel
            "5912": "HD Graphics 630",
            "9bc5": "UHD Graphics 630",
            "4c8a": "Iris Xe Graphics",
        },
    }
    
    if vendorDevices, ok := devices[vendorID]; ok {
        if name, ok := vendorDevices[deviceID]; ok {
            return name
        }
    }
    
    return fmt.Sprintf("Device %s", deviceID)
}

// detectPCIWindows - обнаружение PCI на Windows через реестр/WMI
func detectPCIWindows() []*GPUInfo {
    var gpus []*GPUInfo
    
    // Здесь будет реализация через реестр Windows
    // HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Enum\PCI
    // или через WMI: Win32_PnPEntity WHERE PNPClass = 'Display'
    
    // Временная заглушка - будет возвращать общую информацию
    gpu := &GPUInfo{
        Model:           "Display Adapter (Windows PCI)",
        DriverVersion:   "Hardware Detected",
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
    
    gpus = append(gpus, gpu)
    return gpus
}

// detectPCIMacOS - обнаружение PCI на macOS через IORegistry
func detectPCIMacOS() []*GPUInfo {
    var gpus []*GPUInfo
    
    // Использовать IORegistryEntryCreateCFProperties
    // или системные вызовы для получения информации о PCI
    
    gpu := &GPUInfo{
        Model:           "Graphics Controller (macOS)",
        DriverVersion:   "Hardware Detected",
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
    
    gpus = append(gpus, gpu)
    return gpus
}