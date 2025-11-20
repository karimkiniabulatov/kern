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

	if info.Model == "" {
		t.Error("CPU model should not be empty")
	}

	if info.Cores <= 0 {
		t.Errorf("CPU cores should be positive, got %d", info.Cores)
	}

	if info.Usage < 0 || info.Usage > 100 {
		t.Errorf("CPU usage should be between 0 and 100, got %.2f", info.Usage)
	}
}

func TestCPUUsageCalculation(t *testing.T) {
	// Test that usage calculation doesn't panic
	_, err := cpu.GetCPUUsage()
	if err != nil {
		t.Logf("getCPUUsage returned error (might be normal on some systems): %v", err)
	}
}