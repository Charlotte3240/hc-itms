package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/shogo82148/androidbinary/apk"
)

type APKMetadata struct {
	Name        string
	PackageName string
	Version     string
	VersionCode string
	IconData    []byte
}

func ExtractAPKMetadata(filePath string) (*APKMetadata, error) {
	apkFile, err := apk.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer apkFile.Close()

	pkgName := apkFile.PackageName()
	manifest := apkFile.Manifest()

	versionName := manifest.VersionName.MustString()
	versionCode := manifest.VersionCode.MustInt32()

	appName, err := apkFile.Label(nil)
	if err != nil || appName == "" {
		appName = pkgName
	}

	// Method 1: Try androidbinary library
	var iconData []byte
	iconImg, err := apkFile.Icon(nil)
	if err == nil && iconImg != nil {
		var buf bytes.Buffer
		encodeErr := png.Encode(&buf, iconImg)
		if encodeErr == nil && buf.Len() > 100 {
			iconData = buf.Bytes()
		}
	}

	// Method 2: Try extracting from ZIP directly
	if len(iconData) == 0 {
		iconData = extractAPKIconFromZip(filePath)
	}

	// Method 3: Create a colored placeholder with first letter
	if len(iconData) == 0 {
		iconData = createAppIcon(appName)
	}

	return &APKMetadata{
		Name:        appName,
		PackageName: pkgName,
		Version:     versionName,
		VersionCode: fmt.Sprintf("%d", versionCode),
		IconData:    iconData,
	}, nil
}

func extractAPKIconFromZip(filePath string) []byte {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil
	}
	defer reader.Close()

	type iconCandidate struct {
		data []byte
		size int
		prio int
	}

	priorities := map[string]int{
		"mipmap-xxxhdpi": 1,
		"mipmap-xxhdpi":  2,
		"mipmap-xhdpi":   3,
		"mipmap-hdpi":    4,
		"mipmap-mdpi":    5,
		"drawable-xxxhdpi": 6,
		"drawable-xxhdpi":  7,
		"drawable-xhdpi":   8,
		"drawable":         9,
		"mipmap":           10,
	}

	iconNames := []string{"ic_launcher.png", "ic_launcher_round.png"}
	var candidates []iconCandidate

	for _, f := range reader.File {
		base := filepath.Base(f.Name)
		isIcon := false
		for _, name := range iconNames {
			if base == name {
				isIcon = true
				break
			}
		}
		if !isIcon {
			continue
		}

		prio := 99
		for dir, p := range priorities {
			if strings.Contains(f.Name, dir) {
				prio = p
				break
			}
		}

		data := readZipFile(f)
		if len(data) > 0 {
			candidates = append(candidates, iconCandidate{data: data, size: len(data), prio: prio})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by priority (lower is better), then by size (larger is better)
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.prio < best.prio || (c.prio == best.prio && c.size > best.size) {
			best = c
		}
	}

	return best.data
}

func createAppIcon(name string) []byte {
	size := 120
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background color based on name hash
	hash := 0
	for _, c := range name {
		hash += int(c)
	}
	colors := []color.RGBA{
		{66, 133, 244, 255},   // Blue
		{219, 68, 55, 255},    // Red
		{244, 180, 0, 255},    // Yellow
		{15, 157, 88, 255},    // Green
		{171, 71, 188, 255},   // Purple
		{255, 112, 67, 255},   // Orange
		{0, 172, 193, 255},    // Cyan
	}
	bgColor := colors[hash%len(colors)]
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw first character
	if len(name) > 0 {
		drawCenteredText(img, strings.ToUpper(string([]rune(name)[0])), size/2, size/2+10)
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func drawCenteredText(img *image.RGBA, text string, x, y int) {
	// Simple bitmap font for uppercase letters
	// This creates a basic placeholder - for production you'd use a real font library
	letter := map[string][]string{
		"A": {"01110", "10001", "11111", "10001", "10001"},
		"B": {"11110", "10001", "11110", "10001", "11110"},
		"C": {"01111", "10000", "10000", "10000", "01111"},
		"D": {"11110", "10001", "10001", "10001", "11110"},
		"E": {"11111", "10000", "11110", "10000", "11111"},
		"F": {"11111", "10000", "11110", "10000", "10000"},
		"G": {"01111", "10000", "10011", "10001", "01111"},
		"H": {"10001", "10001", "11111", "10001", "10001"},
		"I": {"11111", "00100", "00100", "00100", "11111"},
		"J": {"00111", "00010", "00010", "10010", "01100"},
		"K": {"10001", "10010", "11100", "10010", "10001"},
		"L": {"10000", "10000", "10000", "10000", "11111"},
		"M": {"10001", "11011", "10101", "10001", "10001"},
		"N": {"10001", "11001", "10101", "10011", "10001"},
		"O": {"01110", "10001", "10001", "10001", "01110"},
		"P": {"11110", "10001", "11110", "10000", "10000"},
		"Q": {"01110", "10001", "10101", "10010", "01101"},
		"R": {"11110", "10001", "11110", "10010", "10001"},
		"S": {"01111", "10000", "01110", "00001", "11110"},
		"T": {"11111", "00100", "00100", "00100", "00100"},
		"U": {"10001", "10001", "10001", "10001", "01110"},
		"V": {"10001", "10001", "10001", "01010", "00100"},
		"W": {"10001", "10001", "10101", "11011", "10001"},
		"X": {"10001", "01010", "00100", "01010", "10001"},
		"Y": {"10001", "01010", "00100", "00100", "00100"},
		"Z": {"11111", "00010", "00100", "01000", "11111"},
		"0": {"01110", "10011", "10101", "11001", "01110"},
		"1": {"00100", "01100", "00100", "00100", "01110"},
		"2": {"01110", "10001", "00110", "01000", "11111"},
		"3": {"11111", "00010", "00110", "10001", "01110"},
		"4": {"10010", "10010", "11111", "00010", "00010"},
		"5": {"11111", "10000", "11110", "00001", "11110"},
		"6": {"01110", "10000", "11110", "10001", "01110"},
		"7": {"11111", "00001", "00010", "00100", "01000"},
		"8": {"01110", "10001", "01110", "10001", "01110"},
		"9": {"01110", "10001", "01111", "00001", "01110"},
		".": {"00000", "00000", "00000", "00000", "00100"},
		"-": {"00000", "00000", "11111", "00000", "00000"},
	}

	pattern, ok := letter[text]
	if !ok {
		pattern = letter["?"]
		if pattern == nil {
			return
		}
	}

	white := color.RGBA{255, 255, 255, 255}
	scale := 8
	startX := x - (len(pattern[0])*scale)/2
	startY := y - (len(pattern)*scale)/2

	for row, line := range pattern {
		for col, ch := range line {
			if ch == '1' {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						px := startX + col*scale + dx
						py := startY + row*scale + dy
						if px >= 0 && px < 120 && py >= 0 && py < 120 {
							img.Set(px, py, white)
						}
					}
				}
			}
		}
	}
}

func readZipFile(f *zip.File) []byte {
	rc, err := f.Open()
	if err != nil {
		return nil
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil
	}
	return data
}
