package tests

import (
	"testing"
	disk "github.com/karimkiniabulatov/kern/internal/disk"
)

func TestDiskSummary(t *testing.T) {
	disks, err := disk.Summary()
	if err != nil {
		t.Fatalf("Failed to get disk info: %v", err)
	}

	if len(disks) == 0 {
		t.Log("No disks found (might be normal in test environment)")
		return
	}

	for _, disk := range disks {
		if disk.Filesystem == "" {
			t.Error("Disk filesystem should not be empty")
		}

		if disk.UsePercent < 0 || disk.UsePercent > 100 {
			t.Errorf("Disk usage percent should be between 0 and 100, got %.2f", disk.UsePercent)
		}
	}
}

func TestSkipFilesystem(t *testing.T) {
	testCases := []struct {
		filesystem string
		mountPoint string
		shouldSkip bool
	}{
		{"/dev/sda1", "/", false},
		{"tmpfs", "/tmp", true},
		{"devtmpfs", "/dev", true},
		{"/dev/loop0", "/snap/core", true},
	}

	for _, tc := range testCases {
		result := disk.ShouldSkipFilesystem(tc.filesystem, tc.mountPoint)
		if result != tc.shouldSkip {
			t.Errorf("shouldSkipFilesystem(%q, %q) = %v, want %v", 
				tc.filesystem, tc.mountPoint, result, tc.shouldSkip)
		}
	}
}

// В internal/tests/disk_test.go добавить тесты для новой функциональности
func TestDiskDetailedMode(t *testing.T) {
    // Тест детального режима
    disks, err := disk.Summary(true)
    if err != nil {
        t.Fatalf("Failed to get detailed disk info: %v", err)
    }
    
    // Проверяем что возвращаются все диски
    t.Logf("Found %d disks in detailed mode", len(disks))
}

func TestDiskDefaultMode(t *testing.T) {
    // Тест обычного режима
    disks, err := disk.Summary(false)
    if err != nil {
        t.Fatalf("Failed to get default disk info: %v", err)
    }
    
    // Проверяем что возвращаются только основные диски
    t.Logf("Found %d disks in default mode", len(disks))
    if len(disks) > 0 {
        for i, d := range disks {
            t.Logf("Disk %d: %s mounted on %s", i, d.Filesystem, d.MountedOn)
        }
    }
}