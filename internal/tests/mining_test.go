package tests

import (
	"testing"
	mining "github.com/karimkiniabulatov/kern/internal/mining"
)

func TestMiningSummary(t *testing.T) {
	info, err := mining.Summary()
	if err != nil {
		t.Fatalf("Failed to get mining info: %v", err)
	}

	// It's normal to have no mining processes running
	if info.Algorithm == "" {
		t.Log("No mining activity detected (normal in test environment)")
	}
}