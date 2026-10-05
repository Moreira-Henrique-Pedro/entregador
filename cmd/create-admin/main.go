package main

import (
	"context"
	"flag"
	"log"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/usecases"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/providers"
	appLogger "github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

func main() {
	email := flag.String("email", "", "admin email")
	name := flag.String("name", "", "admin name")
	password := flag.String("password", "", "admin password (at least 8 characters)")
	flag.Parse()

	envs, err := config.ReadEnvs()
	if err != nil {
		log.Fatalf("failed to read envs: %v", err)
	}

	logger, err := appLogger.NewLogrusLogger(envs.App.Name, envs.App.Env, envs.App.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	ctx := logger.AddToContext(context.Background(), logger)

	identityProvider, err := providers.NewIdentityProvider(ctx, envs)
	if err != nil {
		logger.Fatal("Failed to create identity provider", "error", err.Error())
	}

	admin := &entities.User{Email: *email, Name: *name, Role: entities.RoleAdmin}
	created, err := usecases.NewCreateUser(identityProvider).Execute(ctx, admin, *password)
	if err != nil {
		logger.Fatal("Failed to create admin", "error", err.Error())
	}

	logger.Info("Admin created", "user_id", created.ID, "email", created.Email)
}
