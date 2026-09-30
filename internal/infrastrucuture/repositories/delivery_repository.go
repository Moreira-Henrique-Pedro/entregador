package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	client "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/repositories/models"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBDeliveryRepository struct {
	collection client.MongoClientCollectionPort
}

func NewMongoDBDeliveryRepository(client client.MongoClientCollectionPort) interfaces.DeliveryRepositoryPort {
	_ = client.EnsureUniqueIndex(map[string]interface{}{"delivery_id": 1})
	return &MongoDBDeliveryRepository{
		collection: client,
	}
}

func (r *MongoDBDeliveryRepository) Insert(ctx context.Context, delivery *entities.Delivery) error {
	if delivery == nil {
		return errors.New("delivery is nil")
	}

	model := models.DeliveryFromEntity(delivery)
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now().UTC()
	}
	model.UpdatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, model)
	if mongo.IsDuplicateKeyError(err) {
		logger := logger.GetLoggerFromContext(ctx)
		logger.Warn("Duplicate key error while inserting delivery", model.DeliveryID, "error", err)
		return nil
	}
	return err
}

func (r *MongoDBDeliveryRepository) FindByApartment(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	filter := bson.M{"apartment": apartment}
	if status != nil {
		filter["status"] = string(*status)
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var deliveryModels []models.Delivery
	if err := cursor.All(ctx, &deliveryModels); err != nil {
		return nil, err
	}

	deliveries := make([]*entities.Delivery, 0, len(deliveryModels))
	for i := range deliveryModels {
		deliveries = append(deliveries, deliveryModels[i].ToEntity())
	}
	return deliveries, nil
}

func (r *MongoDBDeliveryRepository) MarkAsDeleted(ctx context.Context, deliveryID string) error {
	now := time.Now().UTC()
	filter := notDeletedFilter(bson.M{
		"delivery_id": deliveryID,
		"status":      string(entities.DeliveryStatusPending),
	})
	update := bson.M{"$set": bson.M{
		"status":       string(entities.DeliveryStatusDeleted),
		deleteAtField:  now,
		updatedAtField: now,
	}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("delivery %s: %w", deliveryID, entities.ErrEntityNotFound)
	}
	return nil
}
