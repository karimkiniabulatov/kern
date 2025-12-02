//go:build windows

package gpu

import (
    "fmt"
    "os/exec"
    "strings"
)

func detectGPUsViaPCIHardware() ([]*GPUInfo, error) {
    return detectPCIWindows()
}

func detectPCIWindows() []*GPUInfo {
    var gpus []*GPUInfo
    
    // Используем PowerShell для получения информации о GPU
    cmd := exec.Command("powershell", "-Command", 
        "Get-WmiObject Win32_VideoController | Select-Object Name,AdapterCompatibility,AdapterRAM,DriverVersion,VideoProcessor")
    
    output, err := cmd.Output()
    if err != nil {
        // Если не удалось, возвращаем заглушку
        return []*GPUInfo{{
            Model:           "Display Adapter (Windows)",
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
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.Contains(line, "Name") {
            continue
        }
        
        fields := strings.Fields(line)
        if len(fields) == 0 {
            continue
        }
        
        // Собираем имя устройства
        model := strings.Join(fields, " ")
        if model != "" {
            gpu := &GPUInfo{
                Model:           model,
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
        }
    }
    
    if len(gpus) == 0 {
        gpus = append(gpus, &GPUInfo{
            Model:           "Display Adapter (Windows)",
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