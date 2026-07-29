package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/global-news/analytics-service/internal/config"
	"github.com/global-news/analytics-service/internal/event"
	http_handler "github.com/global-news/analytics-service/internal/handler/http"
	"github.com/global-news/analytics-service/internal/repository/postgres"
	"github.com/global-news/analytics-service/internal/repository/redis"
	"github.com/global-news/analytics-service/internal/service"
	"github.com/global-news/analytics-service/pkg/logger"
	"github.com/global-news/analytics-service/pkg/rabbitmq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	logger.InitLogger(cfg.AppEnv)
	defer logger.Log.Sync()

	logger.Log.Info("Starting Analytics Service...")

	// Init DB
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	// Init Redis Cache
	cacheRepo := redis.NewCacheRepository(cfg.RedisAddr, cfg.RedisPass)

	// Init Repos
	analyticsRepo := postgres.NewAnalyticsRepository(db)

	// Init Services
	analyticsService := service.NewAnalyticsService(analyticsRepo, cacheRepo, cfg.CacheTTLMinutes)

	// Init RabbitMQ
	rmqClient, err := rabbitmq.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		logger.Log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rmqClient.Close()

	// Init Consumer
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	consumer := event.NewConsumer(rmqClient, analyticsService)
	if err := consumer.StartConsuming(consumerCtx, "analytics_events_queue"); err != nil {
		logger.Log.Fatal("Failed to start consuming", zap.Error(err))
	}

	// Init Background Scheduler for Aggregation
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-schedulerCtx.Done():
				return
			case <-ticker.C:
				date := time.Now().Add(-24 * time.Hour) // Aggregate yesterday's data
				if err := analyticsRepo.AggregateDailyMetrics(context.Background(), date); err != nil {
					logger.Log.Error("Failed to aggregate daily metrics", zap.Error(err))
				} else {
					logger.Log.Info("Successfully aggregated daily metrics", zap.Time("date", date))
				}
			}
		}
	}()

	// Init Router
	router := http_handler.NewRouter(cfg)
	http_handler.NewAnalyticsHandler(router, analyticsService)

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

	// Cancel background tasks
	consumerCancel()
	schedulerCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}
