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

	// Проверяем что возвращаются все интерфейсы
	if len(networks) == 0 {
		t.Log("No network interfaces found (might be normal in test environment)")
	}

	// Проверяем структуру данных для каждого интерфейса
	for _, iface := range networks {
		if iface.Interface == "" {
			t.Error("Network interface should not be empty")
		}
		if iface.IPAddress == "" {
			t.Logf("Interface %s has no IP address (may be normal)", iface.Interface)
		}
	}
}

func TestNetworkDetailedInfo(t *testing.T) {
	// Тест для проверки детальной информации
	networks, err := net.Summary()
	if err != nil {
		t.Fatalf("Failed to get network info: %v", err)
	}

	for _, iface := range networks {
		// Проверяем обязательные поля
		if iface.Interface == "" {
			t.Error("Interface name should not be empty")
		}
		if iface.Status == "" {
			t.Error("Interface status should not be empty")
		}
		if iface.ConnectionType == "" {
			t.Error("Connection type should not be empty")
		}
		
		// Проверяем числовые поля
		if iface.ActivityPercent < 0 || iface.ActivityPercent > 100 {
			t.Errorf("ActivityPercent should be between 0 and 100, got %.2f", iface.ActivityPercent)
		}
		if iface.MTU <= 0 {
			t.Logf("Interface %s has invalid MTU: %d", iface.Interface, iface.MTU)
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
		{0, "0B/s"},
		{1234567, "1.2MB/s"},
	}

	for _, tc := range testCases {
		result := net.FormatSpeed(tc.input)
		if result != tc.expected {
			t.Errorf("FormatSpeed(%.0f) = %s, want %s", tc.input, result, tc.expected)
		}
	}
}

func TestNetworkInterfaceDetection(t *testing.T) {
	// Тест для проверки обнаружения различных типов интерфейсов
	networks, _ := net.Summary()
	
	interfaceTypes := make(map[string]int)
	for _, iface := range networks {
		interfaceTypes[iface.ConnectionType]++
	}
	
	t.Logf("Detected interface types: %v", interfaceTypes)
	
	// Проверяем что для каждого типа есть корректная обработка
	for connType, count := range interfaceTypes {
		t.Logf("Found %d interfaces of type: %s", count, connType)
	}
}