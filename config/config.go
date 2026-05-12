package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
	Mode    string `yaml:"mode"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type StorageConfig struct {
	UploadDir   string `yaml:"upload_dir"`
	IconDir     string `yaml:"icon_dir"`
	MaxFileSize int64  `yaml:"max_file_size"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8080,
			BaseURL: "http://localhost:8080",
			Mode:    "debug",
		},
		Database: DatabaseConfig{
			Path: "./data.db",
		},
		Storage: StorageConfig{
			UploadDir:   "./storage/uploads",
			IconDir:     "./storage/icons",
			MaxFileSize: 524288000,
		},
		JWT: JWTConfig{
			Secret:      "hc-itms-secret-change-in-production",
			ExpireHours: 72,
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
