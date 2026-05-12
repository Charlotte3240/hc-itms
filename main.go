package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Charlotte3240/hc-itms/config"
	"github.com/Charlotte3240/hc-itms/database"
	"github.com/Charlotte3240/hc-itms/handlers"
	"github.com/Charlotte3240/hc-itms/middleware"

	"github.com/gin-gonic/gin"
)

//go:embed all:web/dist
var staticFiles embed.FS

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// Set max multipart memory
	r.MaxMultipartMemory = cfg.Storage.MaxFileSize

	// Handlers
	userH := handlers.NewUserHandler(cfg)
	appH := handlers.NewAppHandler()
	versionH := handlers.NewVersionHandler(cfg)
	downloadH := handlers.NewDownloadHandler(cfg)

	// Auth routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", userH.Login)
		auth.POST("/register", userH.Register)
	}

	// Admin API routes
	api := r.Group("/api")
	api.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		api.GET("/apps", appH.List)
		api.GET("/apps/:id", appH.Get)
		api.PATCH("/apps/:id", appH.Update)
		api.DELETE("/apps/:id", appH.Delete)
		api.POST("/apps/:id/versions", versionH.Upload)
		api.DELETE("/versions/:id", versionH.Delete)
	}

	// Public download routes
	d := r.Group("/d")
	{
		d.GET("/:id", downloadH.DownloadPage)
		d.GET("/:id/latest", downloadH.GetLatest)
		d.GET("/:id/icon", downloadH.ServeIcon)
		d.GET("/:id/v/:vid/plist", downloadH.ServePlist)
		d.GET("/:id/v/:vid/ipa", downloadH.ServeIPA)
		d.GET("/:id/v/:vid/apk", downloadH.ServeAPK)
		d.GET("/:id/v/:vid/qrcode", downloadH.QRCode)
	}

	// Serve frontend (SPA)
	setupSPA(r)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupSPA(r *gin.Engine) {
	fsys, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		log.Printf("Warning: frontend not embedded, skipping SPA setup")
		return
	}

	fileServer := http.FileServer(http.FS(fsys))

	r.NoRoute(func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")

		// Check if file exists in embedded FS
		if path != "" {
			f, err := fsys.Open(path)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// SPA fallback - serve index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
