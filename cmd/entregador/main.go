package main

import (
	"log"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/postgres"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio"
	twilioClient "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/client"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	logger := logrus.New()
	logger.Info("About to connect to DB")

	twilioClient := twilioClient.NewTwilioClient()
	twilio.NewTwilioPort(&twilioClient)

	postgresClient, err := postgres.NewClient()
	if err != nil {
		log.Fatalf("error creating postgres client: %v", err)
	}

	deliveryRepo := postgres.NewDeliveryRepository(postgresClient)
	logger.Info("Delivery repository created: %v", deliveryRepo)
}
