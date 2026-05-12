package models

import "time"

type App struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"`
	Platform      string    `json:"platform" gorm:"not null;uniqueIndex:idx_bundle_platform"`
	BundleID      string    `json:"bundle_id" gorm:"uniqueIndex:idx_bundle_platform"`
	IconPath      string    `json:"icon_path"`
	LatestVersion string    `json:"latest_version"`
	IsEnterprise  bool      `json:"is_enterprise"`
	InstallCount  int64     `json:"install_count" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Versions      []Version `json:"versions,omitempty" gorm:"foreignKey:AppID"`
}

type Version struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AppID        uint      `json:"app_id" gorm:"index;uniqueIndex:idx_app_version_build"`
	Version      string    `json:"version" gorm:"uniqueIndex:idx_app_version_build"`
	BuildNumber  string    `json:"build_number" gorm:"uniqueIndex:idx_app_version_build"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	PlistPath    string    `json:"plist_path"`
	MinOSVersion string    `json:"min_os_version"`
	Changelog    string    `json:"changelog"`
	CreatedAt    time.Time `json:"created_at"`
}
