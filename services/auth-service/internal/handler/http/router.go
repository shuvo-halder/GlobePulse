package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/global-news/auth-service/internal/config"
)

func NewRouter(cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
