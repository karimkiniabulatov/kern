package tests

import (
	"testing"
	net "github.com/karimkiniabulatov/kern/internal/net"
)

func TestNetworkSummary(t *testing.T) {
	networks, err := net.Summary()
	if err != nil {
		t.Fatalf("Failed to get network info: %v", err)
	}

	// It's normal to have no networks in test environment
	if len(networks) == 0 {
		t.Log("No network interfaces found (might be normal in test environment)")
		return
	}

	for _, net := range networks {
		if net.Interface == "" {
			t.Error("Network interface should not be empty")
		}
	}
}

func TestFormatSpeed(t *testing.T) {
	testCases := []struct {
		input    float64
		expected string
	}{
		{100, "100B/s"},
		{1024, "1.0KB/s"},
		{1048576, "1.0MB/s"},
		{1073741824, "1.0GB/s"},
	}

	for _, tc := range testCases {
		result := net.FormatSpeed(tc.input)
		if result != tc.expected {
			t.Errorf("formatSpeed(%.0f) = %s, want %s", tc.input, result, tc.expected)
		}
	}
}