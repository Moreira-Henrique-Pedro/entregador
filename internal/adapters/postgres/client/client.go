package postgres

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/entrypoints/v1/presenters"
)

type Client struct {
	DB *gorm.DB
}

func NewClient() (*Client, error) {
	logger := logrus.New()
	logger.Info("About to connect to DB")

	dsn := buildDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Errorf("Failed to connect to DB: %v", err)
		return nil, err
	}

	logger.Info("Successfully connected to DB")

	if err := db.AutoMigrate(&presenters.DeliveryDTO{}); err != nil {
		logger.Errorf("Failed to migrate DB: %v", err)
		return nil, err
	}

	logger.Info("Database migrated successfully")

	return &Client{DB: db}, nil
}

func buildDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_DBNAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
	)
}
