package infra

import (
	"fmt"
	"log"
	"os"

	"github.com/Moreira-Henrique-Pedro/entregador/src/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CreateConnection() *gorm.DB {
	fmt.Println("About to connect to DB")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PSW"),
		os.Getenv("DB_DBNAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to DB ", err)
	}

	// Orientando o GORM a criar a tabela que corresponda a struct model.Package
	err = db.AutoMigrate(&model.Package{})
	if err != nil {
		log.Fatal("failed to migrate DB ", err)
	}
	return db
}
