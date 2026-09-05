package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/global-news/news-service/internal/ingestion"
	"github.com/global-news/news-service/internal/enrichment"
	"github.com/global-news/news-service/internal/ingestion/connectors"
	"github.com/global-news/news-service/internal/repository/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 1. Setup Database Connection
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "devuser"
	}
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "devpassword"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "globepulse"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		slog.Error("Failed to open database connection", "error", err)
	} else {
		slog.Info("Database connection opened")
	}

	// 2. Setup Scheduler
	repo := postgres.NewIngestionRepository(db)
	
	intervalStr := os.Getenv("INGESTION_INTERVAL")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil || interval == 0 {
		interval = 10 * time.Minute // Default safe interval
	}

	pipeline := enrichment.NewPipeline(
		enrichment.NewProvenanceEnricher(),
		enrichment.NewGeoEnricher(),
		enrichment.NewMetadataEnricher(),
	)

	scheduler := ingestion.NewScheduler(repo, pipeline, interval)
	scheduler.Register(connectors.NewGDELTConnector())
	scheduler.Register(connectors.NewUSGSConnector())
	scheduler.Register(connectors.NewReliefWebConnector())

	// 3. Start Scheduler in Background
	ctx, cancel := context.WithCancel(context.Background())
	go scheduler.Start(ctx)

	// 4. Setup HTTP Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ok"}`)
	})

	mux.HandleFunc("/health/ingestion", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		stats := scheduler.GetStats()
		
		status := "healthy"
		for _, stat := range stats {
			if s, ok := stat.(map[string]interface{}); ok {
				if s["consecutive_failures"].(int) > 3 {
					status = "degraded"
					break
				}
			}
		}

		response := map[string]interface{}{
			"status":     status,
			"connectors": stats,
		}

		json.NewEncoder(w).Encode(response)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// 5. Start Server
	go func() {
		slog.Info("Starting news-service", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	// 6. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down gracefully...")

	// Cancel scheduler context
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	// Close DB connection
	if db != nil {
		db.Close()
	}

	slog.Info("Shutdown complete")
}
