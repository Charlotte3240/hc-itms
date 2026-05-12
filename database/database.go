package database

import (
	"github.com/Charlotte3240/hc-itms/config"
	"github.com/Charlotte3240/hc-itms/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.DatabaseConfig) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(&models.App{}, &models.Version{}, &models.User{})
}
