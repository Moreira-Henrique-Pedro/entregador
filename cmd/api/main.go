package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/providers"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/server"
	appLogger "github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	envs, err := config.ReadEnvs()
	if err != nil {
		log.Fatalf("failed to read envs: %v", err)
	}

	logger, err := appLogger.NewLogrusLogger(envs.App.Name, envs.App.Env, envs.App.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	server.SetupGin(envs.App.Env == config.EnvironmentDevelopment, logger)

	api, err := providers.NewAPI(envs, logger)
	if err != nil {
		logger.Fatal("Failed to initialize api", "error", err.Error())
	}

	httpServer := &http.Server{
		Addr:              ":" + envs.HTTP.Port,
		Handler:           api.Handler,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          appLogger.NewStdLogger(logger),
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", "addr", httpServer.Addr, "version", version)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
	case err := <-errCh:
		logger.Error("HTTP server stopped with error", "error", err.Error())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown HTTP server", "error", err.Error())
	}

	if err := api.Close(shutdownCtx); err != nil {
		logger.Error("Failed to close api dependencies", "error", err.Error())
	}

	logger.Info("Application shutdown completed")
}
