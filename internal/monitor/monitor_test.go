package monitor_test

import (
	"runtime"
	"testing"

	"github.com/philipf/back-cal/internal/monitor"
)

func TestPrimaryMonitor_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-windows stub test, skipping on windows")
	}
	w, h, err := monitor.PrimaryMonitor()
	if err == nil {
		t.Fatal("expected unsupported OS error, got nil")
	}
	if w != 0 || h != 0 {
		t.Errorf("expected 0x0, got %dx%d", w, h)
	}
}

func TestPrimaryMonitor_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	w, h, err := monitor.PrimaryMonitor()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive dimensions, got %dx%d", w, h)
	}
	t.Logf("primary monitor: %dx%d", w, h)
}
