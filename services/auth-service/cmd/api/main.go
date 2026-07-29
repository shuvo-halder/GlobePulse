package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/global-news/auth-service/internal/config"
	http_handler "github.com/global-news/auth-service/internal/handler/http"
	"github.com/global-news/auth-service/internal/handler/http/middleware"
	"github.com/global-news/auth-service/internal/repository/postgres"
	"github.com/global-news/auth-service/internal/repository/redis"
	"github.com/global-news/auth-service/internal/service"
	"github.com/global-news/auth-service/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	logger.InitLogger(cfg.AppEnv)
	defer logger.Log.Sync()

	logger.Log.Info("Starting Auth Service...")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		logger.Log.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	sessionRepo := redis.NewSessionRepository(cfg.RedisAddr, cfg.RedisPass)
	userRepo := postgres.NewUserRepository(db)
	auditRepo := postgres.NewAuditRepository(db)

	authService := service.NewAuthService(userRepo, sessionRepo, auditRepo, cfg)

	router := http_handler.NewRouter(cfg)
	authMiddleware := middleware.AuthMiddleware(cfg, sessionRepo)
	http_handler.NewAuthHandler(router, authService, authMiddleware)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

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
