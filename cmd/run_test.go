package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRanToday_NoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "last-run.txt")
	if ranToday(path) {
		t.Error("expected false for missing file")
	}
}

func TestRanToday_TodayDate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "last-run.txt")
	os.WriteFile(path, []byte(today()), 0o644)
	if !ranToday(path) {
		t.Error("expected true for today's date")
	}
}

func TestRanToday_YesterdayDate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "last-run.txt")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	os.WriteFile(path, []byte(yesterday), 0o644)
	if ranToday(path) {
		t.Error("expected false for yesterday's date")
	}
}

func TestWriteToday_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "last-run.txt")
	if err := writeToday(path); err != nil {
		t.Fatalf("writeToday: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != today() {
		t.Errorf("got %q, want %q", data, today())
	}
}

func TestResolvedPath_Default(t *testing.T) {
	p := resolvedPath("paths.nonexistent_key", "myfile.txt")
	if filepath.Base(p) != "myfile.txt" {
		t.Errorf("unexpected path: %s", p)
	}
}
