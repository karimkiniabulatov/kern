package tests

import (
	"testing"
	cpu "github.com/karimkiniabulatov/kern/internal/cpu"
)

func TestCPUInfo(t *testing.T) {
    info, err := cpu.Summary()
    if err != nil {
        t.Fatalf("Failed to get CPU info: %v", err)
    }

    // Проверяем наличие CPU
    if len(info.CPUs) == 0 {
        t.Log("No CPU data available (might be normal in test environment)")
        return
    }

    // Проверяем первый CPU
    firstCPU := info.CPUs[0]
    if firstCPU.Model == "" {
        t.Error("CPU model should not be empty")
    }

    if firstCPU.Cores <= 0 {
        t.Errorf("CPU cores should be positive, got %d", firstCPU.Cores)
    }

    if firstCPU.Usage < 0 || firstCPU.Usage > 100 {
        t.Errorf("CPU usage should be between 0 and 100, got %.2f", firstCPU.Usage)
    }
}
/*
func TestCPUUsageCalculation(t *testing.T) {
	// Test that usage calculation doesn't panic
	_, err := cpu.GetCPUUsage()
	if err != nil {
		t.Logf("getCPUUsage returned error (might be normal on some systems): %v", err)
	}
}*/