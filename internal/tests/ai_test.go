package tests

import (
	"testing"
	ai "github.com/karimkiniabulatov/kern/internal/ai"
)

func TestAISummary(t *testing.T) {
	info, err := ai.Summary()
	if err != nil {
		t.Fatalf("Failed to get AI info: %v", err)
	}

	// It's normal to have no AI processes running
	if info.ProcessCount == 0 {
		t.Log("No AI processes detected (normal in test environment)")
	}
}