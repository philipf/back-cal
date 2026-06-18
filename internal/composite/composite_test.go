package composite_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/philipf/back-cal/internal/composite"
)

func solidImage(w, h int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func loadPNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output PNG: %v", err)
	}
	return img
}

func assertColor(t *testing.T, img image.Image, x, y int, want color.RGBA) {
	t.Helper()
	got := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	if got != want {
		t.Errorf("pixel (%d,%d): got %v, want %v", x, y, got, want)
	}
}

func TestComposite_OutputDimensions(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "wallpaper.png")

	src := solidImage(50, 30, color.RGBA{R: 255, A: 255})
	cfg := composite.Config{
		MonitorWidth:  200,
		MonitorHeight: 100,
		BgColor:       "#000000",
		PaddingTop:    10,
		PaddingRight:  10,
		OutputPath:    out,
	}

	if err := composite.Composite(src, cfg); err != nil {
		t.Fatalf("Composite: %v", err)
	}

	result := loadPNG(t, out)
	b := result.Bounds()
	if b.Dx() != 200 || b.Dy() != 100 {
		t.Errorf("output dimensions: got %dx%d, want 200x100", b.Dx(), b.Dy())
	}
}

func TestComposite_CalendarTopRight(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "wallpaper.png")

	red := color.RGBA{R: 255, A: 255}
	black := color.RGBA{A: 255}

	// 20x10 red calendar on a 100x80 canvas, 5px padding each side
	src := solidImage(20, 10, red)
	cfg := composite.Config{
		MonitorWidth:  100,
		MonitorHeight: 80,
		BgColor:       "#000000",
		PaddingTop:    5,
		PaddingRight:  5,
		OutputPath:    out,
	}

	if err := composite.Composite(src, cfg); err != nil {
		t.Fatalf("Composite: %v", err)
	}

	result := loadPNG(t, out)

	// Calendar should start at x=75 (100 - 20 - 5), y=5
	assertColor(t, result, 75, 5, red)   // top-left corner of calendar
	assertColor(t, result, 94, 5, red)   // top-right corner of calendar (x=100-5-1=94)
	assertColor(t, result, 75, 14, red)  // bottom-left corner of calendar (y=5+10-1=14)
	assertColor(t, result, 94, 14, red)  // bottom-right corner of calendar

	// Background pixels outside calendar area
	assertColor(t, result, 0, 0, black)   // top-left of canvas
	assertColor(t, result, 0, 79, black)  // bottom-left of canvas
	assertColor(t, result, 74, 5, black)  // just left of calendar
	assertColor(t, result, 75, 4, black)  // just above calendar
	assertColor(t, result, 75, 15, black) // just below calendar
	assertColor(t, result, 95, 5, black)  // right padding area
}

func TestComposite_ScaleResizesCalendar(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "wallpaper.png")

	red := color.RGBA{R: 255, A: 255}

	// 20x10 red calendar scaled 2x -> 40x20, top-right on a 200x100 canvas, no padding
	src := solidImage(20, 10, red)
	cfg := composite.Config{
		MonitorWidth:  200,
		MonitorHeight: 100,
		BgColor:       "#000000",
		Scale:         2,
		OutputPath:    out,
	}

	if err := composite.Composite(src, cfg); err != nil {
		t.Fatalf("Composite: %v", err)
	}

	result := loadPNG(t, out)

	// Scaled calendar occupies x in [160,199], y in [0,19].
	// Centre pixels are solidly red regardless of edge resampling.
	assertColor(t, result, 180, 10, red)
	// A pixel beyond the scaled height stays background.
	assertColor(t, result, 180, 25, color.RGBA{A: 255})
}

func TestComposite_ScaleZeroIsNativeSize(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "wallpaper.png")

	red := color.RGBA{R: 255, A: 255}

	// Scale 0 (unset default) must behave as native size: 20x10 at top-right.
	src := solidImage(20, 10, red)
	cfg := composite.Config{
		MonitorWidth:  100,
		MonitorHeight: 80,
		BgColor:       "#000000",
		Scale:         0,
		PaddingTop:    5,
		PaddingRight:  5,
		OutputPath:    out,
	}

	if err := composite.Composite(src, cfg); err != nil {
		t.Fatalf("Composite: %v", err)
	}

	result := loadPNG(t, out)
	assertColor(t, result, 75, 5, red)  // top-left of calendar at native position
	assertColor(t, result, 94, 14, red) // bottom-right of calendar at native position
}

func TestComposite_BackgroundColor(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "wallpaper.png")

	src := solidImage(10, 10, color.RGBA{G: 255, A: 255})
	cfg := composite.Config{
		MonitorWidth:  100,
		MonitorHeight: 100,
		BgColor:       "#1a2b3c",
		PaddingTop:    50,
		PaddingRight:  50,
		OutputPath:    out,
	}

	if err := composite.Composite(src, cfg); err != nil {
		t.Fatalf("Composite: %v", err)
	}

	result := loadPNG(t, out)
	// Top-left corner is pure background — far from the calendar
	assertColor(t, result, 0, 0, color.RGBA{R: 0x1a, G: 0x2b, B: 0x3c, A: 255})
}

func TestComposite_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "subdir", "wallpaper.png")

	src := solidImage(10, 10, color.RGBA{A: 255})
	cfg := composite.Config{
		MonitorWidth: 50, MonitorHeight: 50,
		BgColor: "#000000", OutputPath: out,
	}

	if err := composite.Composite(src, cfg); err != nil {
		t.Fatalf("Composite: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestParseHexColor_Invalid(t *testing.T) {
	src := solidImage(1, 1, color.RGBA{A: 255})
	cfg := composite.Config{
		MonitorWidth: 10, MonitorHeight: 10,
		BgColor:    "zzzzzz",
		OutputPath: filepath.Join(t.TempDir(), "out.png"),
	}
	if err := composite.Composite(src, cfg); err == nil {
		t.Fatal("expected error for invalid hex color, got nil")
	}
}
