package composite

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/image/bmp"
)

// Config holds compositing parameters.
type Config struct {
	MonitorWidth  int
	MonitorHeight int
	BgColor       string // hex color, e.g. "#000000"
	PaddingTop    int
	PaddingRight  int
	OutputPath    string
}

// DecodeBMP decodes raw BMP bytes into an image.Image.
func DecodeBMP(data []byte) (image.Image, error) {
	img, err := bmp.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decoding BMP: %w", err)
	}
	return img, nil
}

// Composite places src top-right on a full-screen canvas and writes it as PNG to cfg.OutputPath.
func Composite(src image.Image, cfg Config) error {
	bg, err := parseHexColor(cfg.BgColor)
	if err != nil {
		return fmt.Errorf("parsing background color: %w", err)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, cfg.MonitorWidth, cfg.MonitorHeight))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	srcBounds := src.Bounds()
	x := cfg.MonitorWidth - srcBounds.Dx() - cfg.PaddingRight
	y := cfg.PaddingTop
	draw.Draw(canvas, image.Rect(x, y, x+srcBounds.Dx(), y+srcBounds.Dy()), src, srcBounds.Min, draw.Over)

	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	f, err := os.Create(cfg.OutputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, canvas); err != nil {
		return fmt.Errorf("encoding PNG: %w", err)
	}
	return nil
}

func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q: expected #RRGGBB", "#"+s)
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("invalid red in %q: %w", "#"+s, err)
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("invalid green in %q: %w", "#"+s, err)
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("invalid blue in %q: %w", "#"+s, err)
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}
