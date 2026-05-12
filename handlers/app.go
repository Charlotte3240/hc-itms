package handlers

import (
	"net/http"
	"strconv"

	"github.com/Charlotte3240/hc-itms/database"
	"github.com/Charlotte3240/hc-itms/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppHandler struct{}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

func (h *AppHandler) List(c *gin.Context) {
	var apps []models.App
	platform := c.Query("platform")

	query := database.DB.Order("updated_at desc")
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	if err := query.Find(&apps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch apps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"apps": apps})
}

func (h *AppHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid app id"})
		return
	}

	var app models.App
	if err := database.DB.Preload("Versions", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"app": app})
}

func (h *AppHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid app id"})
		return
	}

	var app models.App
	if err := database.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	var req struct {
		Name         *string `json:"name"`
		IsEnterprise *bool   `json:"is_enterprise"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IsEnterprise != nil {
		updates["is_enterprise"] = *req.IsEnterprise
	}

	database.DB.Model(&app).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"app": app})
}

func (h *AppHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid app id"})
		return
	}

	var app models.App
	if err := database.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	// Delete associated versions
	database.DB.Where("app_id = ?", id).Delete(&models.Version{})
	database.DB.Delete(&app)

	c.JSON(http.StatusOK, gin.H{"message": "app deleted"})
}
