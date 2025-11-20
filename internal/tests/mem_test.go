package tests

import (
	"testing"
	mem "github.com/karimkiniabulatov/kern/internal/mem"
)

func TestMemorySummary(t *testing.T) {
	info, err := mem.Summary()
	if err != nil {
		t.Fatalf("Failed to get memory info: %v", err)
	}

	if info.Total == "" {
		t.Error("Total memory should not be empty")
	}

	if info.UsagePercent < 0 || info.UsagePercent > 100 {
		t.Errorf("Memory usage percent should be between 0 and 100, got %.2f", info.UsagePercent)
	}
}