package gpu

import "testing"

func TestGPUSummary(t *testing.T) {
	info, err := Summary()
	if err != nil {
		t.Fatalf("Failed to get GPU info: %v", err)
	}

	if info.Model == "" {
		t.Log("No GPU detected (might be normal on systems without GPU)")
	}
}