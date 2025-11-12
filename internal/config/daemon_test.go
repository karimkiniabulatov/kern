package service

import "testing"

func TestDaemonConfig(t *testing.T) {
	dm := NewDaemonManager()
	cfg := dm.GetConfig()

	if cfg.Port <= 0 || cfg.Port > 65535 {
		t.Errorf("Invalid port number: %d", cfg.Port)
	}

	if !cfg.Enabled {
		t.Error("Daemon should be enabled by default")
	}
}

func TestDaemonManagerCreation(t *testing.T) {
	dm := NewDaemonManager()
	
	if dm == nil {
		t.Error("DaemonManager should not be nil")
	}

	status := dm.Status()
	if status["port"] == nil {
		t.Error("Status should contain port information")
	}
}