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

// Добавить тест для проверки структуры AIInfo
func TestAIInfoStructure(t *testing.T) {
    info, err := ai.Summary()
    if err != nil {
        t.Fatalf("Failed to get AI info: %v", err)
    }

    // Проверить что все поля инициализированы
    if info.Framework == "" {
        t.Error("Framework should not be empty")
    }
    if info.VRAMUsage == "" {
        t.Error("VRAMUsage should not be empty")
    }
    if info.VRAMTotal == "" {
        t.Error("VRAMTotal should not be empty")
    }
}

// Добавить тест для проверки числовых полей
func TestAINumericFields(t *testing.T) {
    info, _ := ai.Summary()
    
    if info.ProcessCount < 0 {
        t.Errorf("ProcessCount should be non-negative, got %d", info.ProcessCount)
    }
    if info.BatchSize < 0 {
        t.Errorf("BatchSize should be non-negative, got %d", info.BatchSize)
    }
    if info.Throughput < 0 {
        t.Errorf("Throughput should be non-negative, got %f", info.Throughput)
    }
}