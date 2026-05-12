package services

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

type IPAMetadata struct {
	Name         string
	BundleID     string
	Version      string
	BuildNumber  string
	MinOSVersion string
	IsEnterprise bool
	IconData     []byte
}

func ExtractIPAMetadata(filePath string) (*IPAMetadata, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("open ipa: %w", err)
	}
	defer reader.Close()

	var infoPlistFile *zip.File
	var mobileprovisionFile *zip.File
	appDir := ""

	for _, f := range reader.File {
		name := f.Name
		if strings.HasPrefix(name, "Payload/") && strings.HasSuffix(name, ".app/") {
			appDir = name
		}
		if strings.HasPrefix(name, "Payload/") && strings.HasSuffix(name, ".app/Info.plist") {
			fCopy := f
			infoPlistFile = fCopy
		}
		if strings.HasPrefix(name, "Payload/") && strings.Contains(name, ".app/Embedded.mobileprovision") {
			fCopy := f
			mobileprovisionFile = fCopy
		}
	}

	if infoPlistFile == nil {
		return nil, fmt.Errorf("Info.plist not found in IPA")
	}

	// Parse Info.plist
	rc, err := infoPlistFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open Info.plist: %w", err)
	}
	defer rc.Close()

	plistData, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read Info.plist: %w", err)
	}

	var info struct {
		BundleIdentifier   string                 `plist:"CFBundleIdentifier"`
		ShortVersionString string                 `plist:"CFBundleShortVersionString"`
		Version            string                 `plist:"CFBundleVersion"`
		DisplayName        string                 `plist:"CFBundleDisplayName"`
		Name               string                 `plist:"CFBundleName"`
		MinimumOSVersion   string                 `plist:"MinimumOSVersion"`
		Icons              map[string]interface{} `plist:"CFBundleIcons"`
		IconsIpad          map[string]interface{} `plist:"CFBundleIcons~ipad"`
	}

	_, err = plist.Unmarshal(plistData, &info)
	if err != nil {
		return nil, fmt.Errorf("parse Info.plist: %w", err)
	}

	appName := info.DisplayName
	if appName == "" {
		appName = info.Name
	}

	// Determine enterprise signing
	isEnterprise := false
	if mobileprovisionFile != nil {
		mpRC, err := mobileprovisionFile.Open()
		if err == nil {
			mpData, _ := io.ReadAll(mpRC)
			mpRC.Close()
			mpStr := string(mpData)
			if strings.Contains(mpStr, "ProvisionsAllDevices") {
				start := strings.Index(mpStr, "<?xml")
				end := strings.LastIndex(mpStr, "</plist>")
				if start >= 0 && end >= 0 {
					plistStr := mpStr[start : end+8]
					if strings.Contains(plistStr, "<key>ProvisionsAllDevices</key>") &&
						strings.Contains(plistStr, "<true/>") {
						isEnterprise = true
					}
				}
			}
		}
	}

	// Extract icon - find the best matching icon
	iconData := extractIconFromIPA(&reader.Reader, appDir, info.Icons, info.IconsIpad)

	return &IPAMetadata{
		Name:         appName,
		BundleID:     info.BundleIdentifier,
		Version:      info.ShortVersionString,
		BuildNumber:  info.Version,
		MinOSVersion: info.MinimumOSVersion,
		IsEnterprise: isEnterprise,
		IconData:     iconData,
	}, nil
}

func extractIconFromIPA(reader *zip.Reader, appDir string, icons, iconsIpad map[string]interface{}) []byte {
	// Try to find icon name from CFBundleIcons
	iconNames := getIconNames(icons, iconsIpad)

	// Try exact icon file names first
	for _, name := range iconNames {
		for _, f := range reader.File {
			if strings.HasPrefix(f.Name, appDir) && strings.HasSuffix(f.Name, "/"+name) {
				data := readFileFromZip(f)
				if len(data) > 0 {
					return data
				}
			}
		}
	}

	// Fallback: find any AppIcon*.png, preferring larger ones
	var bestData []byte
	bestSize := 0

	for _, f := range reader.File {
		if !strings.HasPrefix(f.Name, appDir) {
			continue
		}
		base := filepath.Base(f.Name)
		if strings.HasPrefix(base, "AppIcon") && strings.HasSuffix(base, ".png") {
			data := readFileFromZip(f)
			if len(data) > bestSize {
				bestData = data
				bestSize = len(data)
			}
		}
	}

	if bestData != nil {
		return bestData
	}

	// Last resort: any PNG in the app directory
	for _, f := range reader.File {
		if !strings.HasPrefix(f.Name, appDir) || !strings.HasSuffix(f.Name, ".png") {
			continue
		}
		data := readFileFromZip(f)
		if len(data) > bestSize {
			bestData = data
			bestSize = len(data)
		}
	}

	return bestData
}

func getIconNames(icons, iconsIpad map[string]interface{}) []string {
	var names []string

	for _, iconMap := range []map[string]interface{}{icons, iconsIpad} {
		if iconMap == nil {
			continue
		}
		if primary, ok := iconMap["CFBundlePrimaryIcon"].(map[string]interface{}); ok {
			if files, ok := primary["CFBundleIconFiles"].([]interface{}); ok {
				for _, f := range files {
					if name, ok := f.(string); ok {
						if !strings.HasSuffix(name, ".png") {
							names = append(names, name+"@2x.png")
							names = append(names, name+"@3x.png")
							names = append(names, name+".png")
						} else {
							names = append(names, name)
						}
					}
				}
			}
		}
	}

	return names
}

func readFileFromZip(f *zip.File) []byte {
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
