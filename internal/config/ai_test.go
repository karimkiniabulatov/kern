package ai

import "testing"

func TestAISummary(t *testing.T) {
	info, err := Summary()
	if err != nil {
		t.Fatalf("Failed to get AI info: %v", err)
	}

	// It's normal to have no AI processes running
	if info.ProcessCount == 0 {
		t.Log("No AI processes detected (normal in test environment)")
	}
}