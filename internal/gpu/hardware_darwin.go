//go:build darwin

package gpu

import (
    "fmt"
    "os/exec"
    "strings"
)

func detectGPUsViaPCIHardware() ([]*GPUInfo, error) {
    gpus := detectPCIMacOS()
    return gpus, nil
}

func detectPCIMacOS() []*GPUInfo {
    var gpus []*GPUInfo
    
    // Используем system_profiler для получения информации о GPU
    cmd := exec.Command("system_profiler", "SPDisplaysDataType")
    
    output, err := cmd.Output()
    if err != nil {
        // Если не удалось, возвращаем заглушку
        return []*GPUInfo{{
            Model:           "Graphics Controller (macOS)",
            DriverVersion:   "Unknown",
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
        }}
    }
    
    lines := strings.Split(string(output), "\n")
    var currentGPU *GPUInfo
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        if strings.Contains(line, "Chipset Model:") {
            if currentGPU != nil {
                gpus = append(gpus, currentGPU)
            }
            parts := strings.Split(line, ":")
            if len(parts) > 1 {
                currentGPU = &GPUInfo{
                    Model:           strings.TrimSpace(parts[1]),
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
            }
        } else if currentGPU != nil && strings.Contains(line, "VRAM") {
            // Пытаемся извлечь VRAM информацию
            parts := strings.Split(line, ":")
            if len(parts) > 1 {
                currentGPU.MemoryTotal = strings.TrimSpace(parts[1])
            }
        }
    }
    
    if currentGPU != nil {
        gpus = append(gpus, currentGPU)
    }
    
    if len(gpus) == 0 {
        gpus = append(gpus, &GPUInfo{
            Model:           "Graphics Controller (macOS)",
            DriverVersion:   "Unknown",
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
        })
    }
    
    return gpus
}