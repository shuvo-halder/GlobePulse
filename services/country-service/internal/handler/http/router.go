package http

import (
	"github.com/gin-gonic/gin"
	"github.com/global-news/country-service/internal/config"
)

func NewRouter(cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	return r
}
