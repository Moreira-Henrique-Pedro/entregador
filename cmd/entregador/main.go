package main

import (
	"log"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio"
	twilioClient "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/client"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	twilioClient := twilioClient.NewTwilioClient()
	twilio.NewTwilioPort(twilioClient)

}
