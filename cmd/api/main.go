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
	httpEntrypoint "github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/entrypoints/http"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/providers"
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

	readerProviders, err := providers.NewReaderProviders(envs)
	if err != nil {
		log.Fatalf("failed to create reader providers: %v", err)
	}

	residentHandler := httpEntrypoint.NewResidentHandler(
		readerProviders.GetResidentsByApartment,
		readerProviders.GetResidentsByPhone,
	)

	deliveryHandler := httpEntrypoint.NewDeliveryHandler(readerProviders.GetDeliveriesByApartment)

	server := &http.Server{
		Addr:              ":" + envs.HTTP.Port,
		Handler:           httpEntrypoint.NewRouter(residentHandler, deliveryHandler, logger),
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

	if err := readerProviders.Close(shutdownCtx); err != nil {
		logger.Error("Failed to close reader providers", "error", err.Error())
	}

	logger.Info("Application shutdown completed")
}
