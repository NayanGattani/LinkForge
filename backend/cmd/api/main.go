package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/makeitshort/backend/internal/config"
	"github.com/makeitshort/backend/internal/logger"
	"github.com/makeitshort/backend/internal/server"
	"github.com/makeitshort/backend/internal/shortid"
	"github.com/makeitshort/backend/internal/store"
)

func main() {
	// Automatically load the .env file from the root directory
	_ = godotenv.Load("../.env")

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.AppEnv)

	if err := shortid.Init(1); err != nil {
		log.Error("failed to initialize shortid generator", "error", err)
		os.Exit(1)
	}

	supabaseClient := store.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseKey)

	redisClient, err := store.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Error("failed to initialize redis", "error", err)
		os.Exit(1)
	}

	apiServer := server.New(cfg, log, supabaseClient, redisClient)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- apiServer.Start()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Error("server exited with error", "error", err)
			os.Exit(1)
		}
	case sig := <-signalCh:
		log.Info("shutdown signal received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	log.Info("server shutdown complete")
}
