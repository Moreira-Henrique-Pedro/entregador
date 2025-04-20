package mongodb

import (
	"context"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ResidentRepository struct {
	collection *mongo.Collection
}

func NewResidentRepository(client *mongo.Client, dbName string) ports.ResidentRepositoryPort {
	collection := client.Database(dbName).Collection("residents")
	return &ResidentRepository{collection: collection}
}

func (r *ResidentRepository) Create(ctx context.Context, resident *entities.Resident) error {
	logger := logrus.New()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, resident)
	if err != nil {
		logger.Error("Failed to create resident: ", err)
		return err
	}

	logger.Info("Resident created successfully: ", resident.Apartamento)
	return nil
}

func (r *ResidentRepository) GetByApartment(ctx context.Context, apartamento string) (*entities.Resident, error) {
	logger := logrus.New()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result entities.Resident
	err := r.collection.FindOne(ctx, bson.M{"apartamento": apartamento}).Decode(&result)
	if err != nil {
		logger.Error("Failed to get resident: ", err)
		return nil, err
	}

	logger.Info("Resident found: ", apartamento)
	return &result, nil
}

func (r *ResidentRepository) Update(ctx context.Context, resident *entities.Resident) error {
	logger := logrus.New()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"apartamento": resident.Apartamento}
	update := bson.M{"$set": resident}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Failed to update resident: ", err)
		return err
	}

	logger.Info("Resident updated: ", resident.Apartamento)
	return nil
}

func (r *ResidentRepository) Delete(ctx context.Context, apartamento string) error {
	logger := logrus.New()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"apartamento": apartamento})
	if err != nil {
		logger.Error("Failed to delete resident: ", err)
		return err
	}

	logger.Info("Resident deleted: ", apartamento)
	return nil
}
