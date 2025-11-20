package tests

import (
	"testing"
	gpu "github.com/karimkiniabulatov/kern/internal/gpu"
)

func TestGPUSummary(t *testing.T) {
	info, err := gpu.Summary()
	if err != nil {
		t.Fatalf("Failed to get GPU info: %v", err)
	}

	if info.Model == "" {
		t.Log("No GPU detected (might be normal on systems without GPU)")
	}
}