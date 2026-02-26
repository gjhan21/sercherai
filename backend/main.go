package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/platform/config"
	"sercherai/backend/router"
)

func main() {
	cfg := config.Load()
	r := gin.Default()
	uploadDir := strings.TrimSpace(cfg.AttachmentUploadDir)
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	r.Static("/uploads", uploadDir)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.Register(r)

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
