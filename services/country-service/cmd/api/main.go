package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/global-news/country-service/internal/config"
	http_handler "github.com/global-news/country-service/internal/handler/http"
	"github.com/global-news/country-service/internal/repository/postgres"
	"github.com/global-news/country-service/internal/repository/redis"
	"github.com/global-news/country-service/internal/service"
	"github.com/global-news/country-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	logger.InitLogger(cfg.AppEnv)
	defer logger.Log.Sync()

	logger.Log.Info("Starting Country Service...")

	// Init DB
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	// Init Redis Cache
	cacheRepo := redis.NewCacheRepository(cfg.RedisAddr, cfg.RedisPass)

	// Init Repos
	countryRepo := postgres.NewCountryRepository(db)

	// Init Services
	countryService := service.NewCountryService(countryRepo, cacheRepo, cfg.CacheTTLMinutes)

	// Init Router
	router := http_handler.NewRouter(cfg)
	http_handler.NewCountryHandler(router, countryService)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Log.Info("Server listening", zap.String("port", cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}
