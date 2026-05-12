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
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type DownloadHandler struct {
	cfg *config.Config
}

func NewDownloadHandler(cfg *config.Config) *DownloadHandler {
	return &DownloadHandler{cfg: cfg}
}

func (h *DownloadHandler) GetLatest(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid app id"})
		return
	}

	var app models.App
	if err := database.DB.First(&app, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	var version models.Version
	database.DB.Where("app_id = ?", appID).Order("created_at DESC").First(&version)

	c.JSON(http.StatusOK, gin.H{
		"app":     app,
		"version": version,
	})
}

func (h *DownloadHandler) ServePlist(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	versionID, _ := strconv.ParseUint(c.Param("vid"), 10, 32)

	var version models.Version
	if err := database.DB.Where("id = ? AND app_id = ?", versionID, appID).First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	var app models.App
	database.DB.First(&app, appID)

	// Generate plist on-the-fly
	plistContent := services.GeneratePlist(
		h.cfg.Server.BaseURL, uint(appID), uint(versionID),
		app.BundleID, version.Version, app.Name,
	)

	c.Header("Content-Type", "application/x-plist")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_%s.plist\"", app.BundleID, version.Version))
	c.String(http.StatusOK, plistContent)
}

func (h *DownloadHandler) ServeIPA(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	versionID, _ := strconv.ParseUint(c.Param("vid"), 10, 32)

	var version models.Version
	if err := database.DB.Where("id = ? AND app_id = ?", versionID, appID).First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	if _, err := os.Stat(version.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	// Increment install count
	database.DB.Model(&models.App{}).Where("id = ?", appID).UpdateColumn("install_count", gorm.Expr("install_count + 1"))

	var app models.App
	database.DB.First(&app, appID)

	filename := fmt.Sprintf("%s_%s.ipa", app.Name, version.Version)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.File(version.FilePath)
}

func (h *DownloadHandler) ServeAPK(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	versionID, _ := strconv.ParseUint(c.Param("vid"), 10, 32)

	var version models.Version
	if err := database.DB.Where("id = ? AND app_id = ?", versionID, appID).First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	if _, err := os.Stat(version.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	// Increment install count
	database.DB.Model(&models.App{}).Where("id = ?", appID).UpdateColumn("install_count", gorm.Expr("install_count + 1"))

	var app models.App
	database.DB.First(&app, appID)

	filename := fmt.Sprintf("%s_%s.apk", app.Name, version.Version)
	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.File(version.FilePath)
}

func (h *DownloadHandler) ServeIcon(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var app models.App
	if err := database.DB.First(&app, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Access-Control-Allow-Origin", "*")

	if app.IconPath == "" {
		data := services.CreatePlaceholderIcon()
		c.Data(http.StatusOK, "image/jpeg", data)
		return
	}

	iconPath := filepath.FromSlash(filepath.Clean(app.IconPath))

	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		data := services.CreatePlaceholderIcon()
		c.Data(http.StatusOK, "image/jpeg", data)
		return
	}

	// Read file manually and serve with explicit headers
	data, err := os.ReadFile(iconPath)
	if err != nil {
		data = services.CreatePlaceholderIcon()
		c.Data(http.StatusOK, "image/jpeg", data)
		return
	}

	ext := strings.ToLower(filepath.Ext(iconPath))
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Data(http.StatusOK, contentType, data)
}

func (h *DownloadHandler) QRCode(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	versionID, _ := strconv.ParseUint(c.Param("vid"), 10, 32)

	var app models.App
	if err := database.DB.First(&app, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	var url string
	if app.Platform == "ios" {
		url = fmt.Sprintf("%s/d/%d", h.cfg.Server.BaseURL, appID)
	} else {
		url = fmt.Sprintf("%s/d/%d/v/%d/apk", h.cfg.Server.BaseURL, appID, versionID)
	}

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR code"})
		return
	}

	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "image/png", png)
}

func (h *DownloadHandler) DownloadPage(c *gin.Context) {
	// The SPA frontend handles this route via the NoRoute handler in main.go
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "<!-- SPA will render -->")
}
