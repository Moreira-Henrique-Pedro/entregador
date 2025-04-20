package mongodb

import (
	"context"
	"time"

	mongoClient "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/mongodb/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type ResidentRepository struct {
	collection mongoClient.MongoCollectionPort
	logger     *logrus.Logger
}

func NewResidentRepository(client mongoClient.MongoCollectionPort) ports.ResidentRepositoryPort {
	return &ResidentRepository{
		collection: client,
		logger:     logrus.New(),
	}
}

func (r *ResidentRepository) Create(ctx context.Context, resident *entities.Resident) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, resident)
	if err != nil {
		r.logger.Error("Failed to create resident: ", err)
		return err
	}

	r.logger.Info("Resident created successfully: ", resident.Apartamento)
	return nil
}

func (r *ResidentRepository) GetByApartment(ctx context.Context, apartamento string) (*entities.Resident, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result entities.Resident
	err := r.collection.FindOne(ctx, bson.M{"apartamento": apartamento}).Decode(&result)
	if err != nil {
		r.logger.Error("Failed to get resident: ", err)
		return nil, err
	}

	r.logger.Info("Resident found: ", apartamento)
	return &result, nil
}

func (r *ResidentRepository) Update(ctx context.Context, resident *entities.Resident) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"apartamento": resident.Apartamento}
	update := bson.M{"$set": resident}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update resident: ", err)
		return err
	}

	r.logger.Info("Resident updated: ", resident.Apartamento)
	return nil
}

func (r *ResidentRepository) Delete(ctx context.Context, apartamento string) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"apartamento": apartamento})
	if err != nil {
		r.logger.Error("Failed to delete resident: ", err)
		return err
	}

	r.logger.Info("Resident deleted: ", apartamento)
	return nil
}
