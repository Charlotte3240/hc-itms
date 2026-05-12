package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Charlotte3240/hc-itms/config"
	"github.com/Charlotte3240/hc-itms/database"
	"github.com/Charlotte3240/hc-itms/models"
	"github.com/Charlotte3240/hc-itms/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VersionHandler struct {
	cfg *config.Config
}

func NewVersionHandler(cfg *config.Config) *VersionHandler {
	return &VersionHandler{cfg: cfg}
}

func (h *VersionHandler) Upload(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid app id"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if file.Size > h.cfg.Storage.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	changelog := c.PostForm("changelog")
	ext := strings.ToLower(filepath.Ext(file.Filename))

	if ext != ".ipa" && ext != ".apk" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only .ipa and .apk files are supported"})
		return
	}

	platform := "ios"
	if ext == ".apk" {
		platform = "android"
	}

	// Save uploaded file
	uploadDir := filepath.Join(h.cfg.Storage.UploadDir, platform)
	os.MkdirAll(uploadDir, 0755)

	tempPath := filepath.Join(uploadDir, fmt.Sprintf("upload_%d%s", appID, ext))
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Extract metadata
	var app *models.App
	var version *models.Version

	if platform == "ios" {
		app, version, err = h.processIPA(uint(appID), tempPath, file.Size, changelog)
	} else {
		app, version, err = h.processAPK(uint(appID), tempPath, file.Size, changelog)
	}

	if err != nil {
		os.Remove(tempPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"app": app, "version": version})
}

func (h *VersionHandler) processIPA(appID uint, filePath string, fileSize int64, changelog string) (*models.App, *models.Version, error) {
	meta, err := services.ExtractIPAMetadata(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract IPA metadata: %w", err)
	}

	// Find or create app
	var app models.App
	result := database.DB.Where("bundle_id = ? AND platform = ?", meta.BundleID, "ios").First(&app)
	if result.Error == gorm.ErrRecordNotFound {
		app = models.App{
			Name:          meta.Name,
			Platform:      "ios",
			BundleID:      meta.BundleID,
			LatestVersion: meta.Version,
			IsEnterprise:  meta.IsEnterprise,
		}
		database.DB.Create(&app)
	} else if result.Error != nil {
		return nil, nil, fmt.Errorf("database error: %w", result.Error)
	}

	// Check for duplicate version
	var existingVer models.Version
	if database.DB.Where("app_id = ? AND version = ? AND build_number = ?", app.ID, meta.Version, meta.BuildNumber).First(&existingVer).Error == nil {
		os.Remove(filePath)
		return nil, nil, fmt.Errorf("version %s (%s) already exists", meta.Version, meta.BuildNumber)
	}

	// Rename file to final path
	finalPath := filepath.Join(h.cfg.Storage.UploadDir, "ios",
		fmt.Sprintf("%d_%s_%s.ipa", app.ID, meta.Version, meta.BuildNumber))
	os.Rename(filePath, finalPath)

	// Save icon
	iconPath := ""
	if len(meta.IconData) > 0 {
		iconPath, _ = services.SaveIcon(meta.IconData, h.cfg.Storage.IconDir, app.ID)
	}

	// Generate plist
	plistContent := services.GeneratePlist(
		h.cfg.Server.BaseURL, app.ID, 0, // version ID will be updated after create
		meta.BundleID, meta.Version, meta.Name,
	)
	plistPath := filepath.Join(h.cfg.Storage.UploadDir, "ios",
		fmt.Sprintf("%d_%s_%s.plist", app.ID, meta.Version, meta.BuildNumber))
	os.WriteFile(plistPath, []byte(plistContent), 0644)

	version := models.Version{
		AppID:        app.ID,
		Version:      meta.Version,
		BuildNumber:  meta.BuildNumber,
		FilePath:     finalPath,
		FileSize:     fileSize,
		PlistPath:    plistPath,
		MinOSVersion: meta.MinOSVersion,
		Changelog:    changelog,
	}
	database.DB.Create(&version)

	// Update plist with correct version ID
	plistContent = services.GeneratePlist(
		h.cfg.Server.BaseURL, app.ID, version.ID,
		meta.BundleID, meta.Version, meta.Name,
	)
	os.WriteFile(plistPath, []byte(plistContent), 0644)

	// Update app
	updates := map[string]interface{}{
		"latest_version": meta.Version,
	}
	if iconPath != "" {
		updates["icon_path"] = iconPath
	}
	database.DB.Model(&app).Updates(updates)
	database.DB.First(&app, app.ID)

	return &app, &version, nil
}

func (h *VersionHandler) processAPK(appID uint, filePath string, fileSize int64, changelog string) (*models.App, *models.Version, error) {
	meta, err := services.ExtractAPKMetadata(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract APK metadata: %w", err)
	}

	// Find or create app
	var app models.App
	result := database.DB.Where("bundle_id = ? AND platform = ?", meta.PackageName, "android").First(&app)
	if result.Error == gorm.ErrRecordNotFound {
		app = models.App{
			Name:          meta.Name,
			Platform:      "android",
			BundleID:      meta.PackageName,
			LatestVersion: meta.Version,
		}
		database.DB.Create(&app)
	} else if result.Error != nil {
		return nil, nil, fmt.Errorf("database error: %w", result.Error)
	}

	// Check for duplicate version
	var existingVer models.Version
	if database.DB.Where("app_id = ? AND version = ? AND build_number = ?", app.ID, meta.Version, meta.VersionCode).First(&existingVer).Error == nil {
		os.Remove(filePath)
		return nil, nil, fmt.Errorf("version %s (%s) already exists", meta.Version, meta.VersionCode)
	}

	// Rename file to final path
	finalPath := filepath.Join(h.cfg.Storage.UploadDir, "android",
		fmt.Sprintf("%d_%s_%s.apk", app.ID, meta.Version, meta.VersionCode))
	os.Rename(filePath, finalPath)

	// Save icon
	iconPath := ""
	if len(meta.IconData) > 0 {
		iconPath, _ = services.SaveIcon(meta.IconData, h.cfg.Storage.IconDir, app.ID)
	}

	version := models.Version{
		AppID:       app.ID,
		Version:     meta.Version,
		BuildNumber: meta.VersionCode,
		FilePath:    finalPath,
		FileSize:    fileSize,
		Changelog:   changelog,
	}
	database.DB.Create(&version)

	// Update app
	updates := map[string]interface{}{
		"latest_version": meta.Version,
	}
	if iconPath != "" {
		updates["icon_path"] = iconPath
	}
	database.DB.Model(&app).Updates(updates)
	database.DB.First(&app, app.ID)

	return &app, &version, nil
}

func (h *VersionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version id"})
		return
	}

	var version models.Version
	if err := database.DB.First(&version, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	// Delete files
	os.Remove(version.FilePath)
	if version.PlistPath != "" {
		os.Remove(version.PlistPath)
	}

	database.DB.Delete(&version)

	// Update app's latest version
	var latest models.Version
	database.DB.Where("app_id = ?", version.AppID).Order("created_at DESC").First(&latest)
	database.DB.Model(&models.App{}).Where("id = ?", version.AppID).Update("latest_version", latest.Version)

	c.JSON(http.StatusOK, gin.H{"message": "version deleted"})
}
