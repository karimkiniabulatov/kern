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

	if len(info) == 0 {
		t.Log("No GPU detected (might be normal on systems without GPU)")
		return
	}

	// Проверяем отсутствие дублирования
	seenModels := make(map[string]bool)
	for _, gpuInfo := range info {
		if gpuInfo.Model == "" {
			t.Error("GPU model should not be empty")
		}
		
		// Проверка на дублирование
		if seenModels[gpuInfo.Model] {
			t.Errorf("Duplicate GPU model detected: %s", gpuInfo.Model)
		}
		seenModels[gpuInfo.Model] = true
		
		// Проверяем числовые поля
		if gpuInfo.Utilization < 0 || gpuInfo.Utilization > 100 {
			t.Errorf("GPU utilization should be between 0 and 100, got %.2f", gpuInfo.Utilization)
		}
		if gpuInfo.GPUTemp < 0 || gpuInfo.GPUTemp > 120 {
			t.Errorf("GPU temperature should be between 0 and 120, got %.2f", gpuInfo.GPUTemp)
		}
		if gpuInfo.FanSpeed < 0 || gpuInfo.FanSpeed > 100 {
			t.Errorf("GPU fan speed should be between 0 and 100, got %.2f", gpuInfo.FanSpeed)
		}
	}
}

func TestGPUDetectionMethods(t *testing.T) {
	// Тест для проверки различных методов обнаружения
	// Проверяем что detectAllGPUsEnhanced не вызывает дублирующие функции
	info, err := gpu.Summary()
	if err != nil {
		t.Logf("GPU detection error (may be normal): %v", err)
		return
	}
	
	// Проверяем что все GPU имеют уникальные идентификаторы
	gpuIDs := make(map[string]bool)
	for _, gpuInfo := range info {
		// Создаем уникальный ID на основе модели и серийного номера
		gpuID := gpuInfo.Model
		if gpuID != "" {
			if gpuIDs[gpuID] {
				t.Errorf("Duplicate GPU ID found: %s", gpuID)
			}
			gpuIDs[gpuID] = true
		}
	}
	
	t.Logf("Found %d unique GPUs", len(gpuIDs))
}

func TestGPUInfoCompleteness(t *testing.T) {
	// Проверяем что все поля GPUInfo инициализированы
	info, err := gpu.Summary()
	if err != nil {
		t.Logf("GPU detection error (may be normal): %v", err)
		return
	}
	
	for _, gpuInfo := range info {
		// Проверяем обязательные поля
		if gpuInfo.Model == "" {
			t.Error("GPU model should not be empty")
		}
		if gpuInfo.DriverVersion == "" {
			t.Log("GPU driver version is empty (may be normal for generic detection)")
		}
		if gpuInfo.PerformanceState == "" {
			t.Log("GPU performance state is empty (may be normal)")
		}
		
		// Проверяем строки памяти
		if gpuInfo.MemoryTotal == "" {
			t.Log("GPU memory total is empty (may be normal)")
		}
		if gpuInfo.MemoryUsed == "" {
			t.Log("GPU memory used is empty (may be normal)")
		}
		if gpuInfo.MemoryFree == "" {
			t.Log("GPU memory free is empty (may be normal)")
		}
	}
}