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

func NewMongoDBDeliveryRepository(ctx context.Context, client client.MongoClientCollectionPort) (interfaces.DeliveryRepositoryPort, error) {
	if err := client.EnsureIndexes(ctx, deliveryIndexes()); err != nil {
		return nil, fmt.Errorf("ensure deliveries indexes: %w", err)
	}
	return &MongoDBDeliveryRepository{
		collection: client,
	}, nil
}

func deliveryIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "delivery_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		// FindByApartment sorts by creation date, with and without a status filter.
		{Keys: bson.D{{Key: "apartment", Value: 1}, {Key: createdAtField, Value: -1}}},
		{Keys: bson.D{{Key: "apartment", Value: 1}, {Key: "status", Value: 1}, {Key: createdAtField, Value: -1}}},
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

func (r *MongoDBDeliveryRepository) FindByDeliveryID(ctx context.Context, deliveryID string) (*entities.Delivery, error) {
	var model models.Delivery
	err := r.collection.FindOne(ctx, bson.M{"delivery_id": deliveryID}).Decode(&model)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, deliveryNotFound(deliveryID)
	}
	if err != nil {
		return nil, err
	}
	return model.ToEntity(), nil
}

func (r *MongoDBDeliveryRepository) FindByApartment(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	filter := bson.M{"apartment": apartment}
	if status != nil {
		filter["status"] = string(*status)
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: createdAtField, Value: -1}}))
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
		return deliveryNotFound(deliveryID)
	}
	return nil
}

func (r *MongoDBDeliveryRepository) MarkArrivalAsNotified(ctx context.Context, deliveryID string) error {
	return r.markAsNotified(ctx, deliveryID, arrivalNotifiedAtField)
}

func (r *MongoDBDeliveryRepository) MarkPickupAsNotified(ctx context.Context, deliveryID string) error {
	return r.markAsNotified(ctx, deliveryID, pickupNotifiedAtField)
}

func (r *MongoDBDeliveryRepository) markAsNotified(ctx context.Context, deliveryID, field string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{field: now, updatedAtField: now}}

	result, err := r.collection.UpdateOne(ctx, bson.M{"delivery_id": deliveryID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return deliveryNotFound(deliveryID)
	}
	return nil
}

func deliveryNotFound(deliveryID string) error {
	return fmt.Errorf("delivery %s: %w", deliveryID, entities.ErrEntityNotFound)
}
