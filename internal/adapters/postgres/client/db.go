package infra

import (
	"fmt"
	"log"
	"os"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/entrypoints/v1/presenters"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CreateConnection() *gorm.DB {
	fmt.Println("About to connect to DB")

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_DBNAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")
	timezone := os.Getenv("DB_TIMEZONE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host,
		user,
		password,
		dbname,
		port,
		sslmode,
		timezone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to DB ", err)
	}

	// Orientando o GORM a criar a tabela que corresponda a struct model.Package
	err = db.AutoMigrate(&presenters.DeliveryDTO{})
	if err != nil {
		log.Fatal("failed to migrate DB ", err)
	}

	fmt.Println("Successfully connected!")

	return db
}
