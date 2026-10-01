// API HTTP: residents, delivery registration/pickup and queries.
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
	bootstrap "github.com/Moreira-Henrique-Pedro/entregador/internal/providers"
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

	api, err := bootstrap.NewAPI(envs, logger)
	if err != nil {
		log.Fatalf("failed to initialize api: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + envs.HTTP.Port,
		Handler:           api.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", "addr", server.Addr, "version", version)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown HTTP server", "error", err.Error())
	}

	if err := api.Close(shutdownCtx); err != nil {
		logger.Error("Failed to close api dependencies", "error", err.Error())
	}

	logger.Info("Application shutdown completed")
}
