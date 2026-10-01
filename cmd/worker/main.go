// Worker: consumes the internal commands topic and sends the delivery notifications.
package main

import (
	"context"
	"log"
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

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load configs: %v", err)
	}
	cfg.Envs.App.Version = version

	logger, err := appLogger.NewLogrusLogger(cfg.Envs.App.Name, cfg.Envs.App.Env, cfg.Envs.App.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	ctx = logger.AddToContext(ctx, logger)

	worker, err := bootstrap.NewWorker(cfg, logger)
	if err != nil {
		log.Fatalf("failed to initialize worker: %v", err)
	}

	logger.Info("Worker initialized",
		"version", version,
		"topic", cfg.SubscriberConfigs.Topic,
		"consumer_group", cfg.SubscriberConfigs.ConsumerGroup,
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Consumer.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
	case err := <-errCh:
		if err != nil {
			logger.Error("Worker stopped with error", "error", err.Error())
		}
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shutdownCtx = logger.AddToContext(shutdownCtx, logger)

	if err := worker.Close(shutdownCtx); err != nil {
		logger.Error("Worker shutdown finished with errors", "error", err.Error())
		return
	}

	logger.Info("Worker shutdown completed")
}
