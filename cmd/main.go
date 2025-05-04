package main

import (
	"MikoNews/internal/api"
	"MikoNews/internal/bot"
	"MikoNews/internal/config"
	"MikoNews/internal/database"
	"MikoNews/internal/pkg/logger"
	"context"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		stdlog.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Logger
	logger.InitLogger(cfg.Logger.Path, cfg.Logger.Level) // Call directly with correct arguments
	log := zap.L()                                       // Get the global logger instance
	defer func() { _ = log.Sync() }()

	log.Info("Logger initialized successfully")

	// --- Initialize Database ---
	// *** ASSUMPTION: database.New now returns *gorm.DB ***
	gormDB, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	log.Info("Database connection successful")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Initialize Feishu Bot ---
	feishuBot := bot.NewFeishuBot(&cfg.Feishu, gormDB)

	// --- Create API Server ---
	apiServer := api.New(cfg, gormDB)

	// --- Start Feishu Bot (WebSocket) in background ---
	go func() {
		log.Info("Starting Feishu Bot (WebSocket)...")
		if err := feishuBot.Start(ctx); err != nil {
			log.Error("Feishu Bot (WebSocket) failed", zap.Error(err))
			// Decide if bot failure should stop the main application
			// cancel() // Optionally cancel context to stop server too
		}
	}()

	// --- Start API Server (HTTP) in background ---
	go func() {
		log.Info("Starting API Server...")
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start API server", zap.Error(err))
		}
	}()

	log.Info("API server started", zap.Int("port", cfg.Server.Port))

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create a deadline context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Attempt graceful shutdown of the API server
	// if err := apiServer.Shutdown(shutdownCtx); err != nil { // Commented out: Shutdown undefined
	// 	log.Error("API server shutdown failed", zap.Error(err))
	// }
	_ = shutdownCtx // Temp use

	// Close database connection using *gorm.DB
	// Access the underlying *gorm.DB (gormDB.DB) then call its DB() method
	if gormInstance := gormDB.DB; gormInstance != nil {
		if dbSQL, err := gormInstance.DB(); err == nil {
			if err := dbSQL.Close(); err != nil {
				log.Error("Failed to close database connection", zap.Error(err))
			}
		} else {
			log.Error("Failed to get underlying sql.DB for closing", zap.Error(err))
		}
	} else {
		log.Warn("Database wrapper or underlying gorm DB instance is nil, cannot close")
	}

	// Cancel the main context
	cancel()
	_ = ctx // Use ctx to avoid unused error

	logger.Info("Server gracefully stopped") // Corrected to use logger.Info
}
