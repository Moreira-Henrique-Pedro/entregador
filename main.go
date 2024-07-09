package main

import (
	"log"

	"github.com/Moreira-Henrique-Pedro/entregador/src/controller"
	"github.com/Moreira-Henrique-Pedro/entregador/src/infra"
	"github.com/Moreira-Henrique-Pedro/entregador/src/service"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	db := infra.CreateConnection()
	stockService := service.NewBoxService(db)
	stockController := controller.NewBoxController(stockService)

	stockController.InitRoutes()

}
