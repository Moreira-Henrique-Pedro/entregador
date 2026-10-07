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
	email := flag.String("email", "", "user email")
	name := flag.String("name", "", "user name")
	password := flag.String("password", "", "user password (at least 8 characters)")
	role := flag.String("role", string(entities.RoleAdmin), "user role: admin or doorman")
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

	user := &entities.User{Email: *email, Name: *name, Role: entities.Role(*role)}
	created, err := usecases.NewCreateUser(identityProvider).Execute(ctx, user, *password)
	if err != nil {
		logger.Fatal("Failed to create user", "error", err.Error())
	}

	logger.Info("User created", "user_id", created.ID, "email", created.Email, "role", string(created.Role))
}
