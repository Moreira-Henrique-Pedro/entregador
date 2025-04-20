package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/mongodb"
	mongodbClient "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/mongodb/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/postgres"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio"
	twilioClient "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/client"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	logger := logrus.New()
	logger.Info("Starting application...")

	// Twilio setup
	twilioClient := twilioClient.NewTwilioClient()
	twilio.NewTwilioPort(&twilioClient)

	// Postgres setup
	postgresClient, err := postgres.NewClient()
	if err != nil {
		log.Fatalf("error creating postgres client: %v", err)
	}
	deliveryRepo := postgres.NewDeliveryRepository(postgresClient)
	logger.Info("Delivery repository created", deliveryRepo)

	// MongoDB setup
	mongoURI := os.Getenv("MONGO_URI")
	mongoDBName := os.Getenv("MONGO_DB_NAME")
	mongoClient, err := mongodbClient.NewMongoClient(mongoURI)
	if err != nil {
		log.Fatalf("error creating mongo client: %v", err)
	}
	logger.Infof("Connected to MongoDB")

	// MongoDB repository setup
	residentRepo := mongodb.NewResidentRepository(mongoClient, mongoDBName)
	logger.Info("Resident repository created", residentRepo)

	// Aqui você pode seguir com suas handlers/usecases se necessário
}
