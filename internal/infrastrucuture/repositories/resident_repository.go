package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	client "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/repositories/models"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBResidentRepository struct {
	collection client.MongoClientCollectionPort
}

func NewMongoDBResidentRepository(client client.MongoClientCollectionPort) interfaces.ResidentRepositoryPort {
	_ = client.EnsureUniqueIndex(map[string]interface{}{"resident_id": 1})
	return &MongoDBResidentRepository{
		collection: client,
	}
}

func (r *MongoDBResidentRepository) Insert(ctx context.Context, resident *entities.Resident) error {
	if resident == nil {
		return errors.New("resident is nil")
	}

	model := models.ResidentFromEntity(resident)
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now().UTC()
	}
	model.UpdatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, model)
	if mongo.IsDuplicateKeyError(err) {
		logger := logger.GetLoggerFromContext(ctx)
		logger.Warn("Duplicate key error while inserting resident", model.ResidentID, "error", err)
		return nil
	}
	return err
}
