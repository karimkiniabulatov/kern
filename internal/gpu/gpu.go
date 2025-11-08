package gpu

import (
    "fmt"
    "os/exec"
    "strconv"
    "strings"
)

type GPUInfo struct {
    Model         string
    DriverVersion string
    GPUTemp       float64
    MemoryTotal   string
    MemoryUsed    string
    MemoryFree    string
    Utilization  float64
    PowerDraw     string
    PowerLimit    string
    FanSpeed     float64
    ClockCore    string
    ClockMemory  string
    PerformanceState string
}

func Summary() (*GPUInfo, error) {
    info := &GPUInfo{}

    // Try to get NVIDIA GPU info using nvidia-smi
    if output, err := exec.Command("nvidia-smi", "--query-gpu=name,driver_version,temperature.gpu,memory.total,memory.used,memory.free,utilization.gpu,power.draw,power.limit,fan.speed,clocks.current.graphics,clocks.current.memory,performance.state", "--format=csv,noheader,nounits").Output(); err == nil {
        return parseNvidiaSMIOutput(string(output))
    }

    // Try AMD GPU (using rocm-smi if available)
    if output, err := exec.Command("rocm-smi", "--showproductname", "--showtemp", "--showuse", "--showmemuse").Output(); err == nil {
        return parseAMDSMIOutput(string(output))
    }

    // Fallback: check for GPU devices
    info.Model = "Generic GPU"
    info.DriverVersion = "Unknown"
    
    return info, nil
}

func parseNvidiaSMIOutput(output string) (*GPUInfo, error) {
    lines := strings.Split(strings.TrimSpace(output), "\n")
    if len(lines) == 0 {
        return nil, fmt.Errorf("no GPU data available")
    }

    // Take first GPU
    fields := strings.Split(lines[0], ", ")
    if len(fields) < 13 {
        return nil, fmt.Errorf("insufficient GPU data")
    }

    info := &GPUInfo{
        Model:          strings.TrimSpace(fields[0]),
        DriverVersion:  strings.TrimSpace(fields[1]),
        PerformanceState: strings.TrimSpace(fields[12]),
    }

    // Parse numeric values
    if temp, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64); err == nil {
        info.GPUTemp = temp
    }
    
    info.MemoryTotal = strings.TrimSpace(fields[3]) + " MB"
    info.MemoryUsed = strings.TrimSpace(fields[4]) + " MB" 
    info.MemoryFree = strings.TrimSpace(fields[5]) + " MB"

    if util, err := strconv.ParseFloat(strings.TrimSpace(fields[6]), 64); err == nil {
        info.Utilization = util
    }

    info.PowerDraw = strings.TrimSpace(fields[7]) + " W"
    info.PowerLimit = strings.TrimSpace(fields[8]) + " W"

    if fan, err := strconv.ParseFloat(strings.TrimSpace(fields[9]), 64); err == nil {
        info.FanSpeed = fan
    }

    info.ClockCore = strings.TrimSpace(fields[10]) + " MHz"
    info.ClockMemory = strings.TrimSpace(fields[11]) + " MHz"

    return info, nil
}

func parseAMDSMIOutput(output string) (*GPUInfo, error) {
    // Simplified AMD GPU parsing
    info := &GPUInfo{
        Model: "AMD GPU",
    }

    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, "Product Name") {
            parts := strings.Split(line, ":")
            if len(parts) > 1 {
                info.Model = strings.TrimSpace(parts[1])
            }
        } else if strings.Contains(line, "Temperature") {
            parts := strings.Split(line, ":")
            if len(parts) > 1 {
                if temp, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
                    info.GPUTemp = temp
                }
            }
        } else if strings.Contains(line, "GPU use") {
            parts := strings.Split(line, ":")
            if len(parts) > 1 {
                if util, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(parts[1], "%")), 64); err == nil {
                    info.Utilization = util
                }
            }
        }
    }

    return info, nil
}