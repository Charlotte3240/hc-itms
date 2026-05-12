package services

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"time"
)

func SaveIcon(data []byte, iconDir string, id uint) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if err := os.MkdirAll(iconDir, 0755); err != nil {
		return "", fmt.Errorf("create icon dir: %w", err)
	}

	// Try to decode and re-encode as standard PNG
	img, _, err := image.Decode(bytes.NewReader(data))
	if err == nil {
		filename := fmt.Sprintf("%d_%d.png", id, time.Now().Unix())
		path := filepath.Join(iconDir, filename)

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err == nil && buf.Len() > 0 {
			if err := os.WriteFile(path, buf.Bytes(), 0644); err == nil {
				return path, nil
			}
		}
	}

	// Try CgBI → standard PNG conversion (Apple's non-standard PNG format)
	if converted, err := convertCgbiToStandardPNG(data); err == nil {
		filename := fmt.Sprintf("%d_%d.png", id, time.Now().Unix())
		path := filepath.Join(iconDir, filename)
		if err := os.WriteFile(path, converted, 0644); err == nil {
			return path, nil
		}
	}

	// Fall back to saving raw bytes
	ext := ".png"
	if isJPEG(data) {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d_%d%s", id, time.Now().Unix(), ext)
	path := filepath.Join(iconDir, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write icon: %w", err)
	}

	return path, nil
}

func isJPEG(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
}

func CreatePlaceholderIcon() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 120, 120))
	gray := color.RGBA{200, 200, 200, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{gray}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	return buf.Bytes()
}
