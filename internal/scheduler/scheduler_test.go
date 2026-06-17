package scheduler_test

import (
	"runtime"
	"testing"

	"github.com/philipf/back-cal/internal/scheduler"
)

func TestRegister_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-windows stub test")
	}
	err := scheduler.Register("/usr/local/bin/back-cal")
	if err == nil {
		t.Fatal("expected unsupported OS error, got nil")
	}
}

func TestRegister_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	// Integration test: actually registers tasks in Task Scheduler.
	// Verify manually: open Task Scheduler and look for back-cal-logon and back-cal-unlock.
	err := scheduler.Register(`C:\Windows\System32\cmd.exe`)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
}
