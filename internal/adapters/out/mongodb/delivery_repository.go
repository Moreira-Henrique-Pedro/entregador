package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/out/mongodb/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/out/mongodb/models"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeliveryRepository struct {
	collection client.MongoClientCollectionPort
}

func NewDeliveryRepository(ctx context.Context, client client.MongoClientCollectionPort) (out.DeliveryRepository, error) {
	if err := client.EnsureIndexes(ctx, deliveryIndexes()); err != nil {
		return nil, fmt.Errorf("ensure deliveries indexes: %w", err)
	}
	return &DeliveryRepository{
		collection: client,
	}, nil
}

func deliveryIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "delivery_id", Value: 1}}, Options: options.Index().SetUnique(true)},

		{Keys: bson.D{{Key: "apartment", Value: 1}, {Key: createdAtField, Value: -1}}},
		{Keys: bson.D{{Key: "apartment", Value: 1}, {Key: "status", Value: 1}, {Key: createdAtField, Value: -1}}},
	}
}

func (r *DeliveryRepository) Insert(ctx context.Context, delivery *domain.Delivery) error {
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
		logger.Warn("Duplicate key error while inserting delivery", "delivery_id", model.DeliveryID, "error", err.Error())
		return nil
	}
	if err != nil {
		return err
	}

	// The API returns the registered delivery, so it carries the persisted timestamps.
	delivery.CreatedAt = model.CreatedAt
	delivery.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *DeliveryRepository) FindByDeliveryID(ctx context.Context, deliveryID string) (*domain.Delivery, error) {
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

func (r *DeliveryRepository) FindByApartment(ctx context.Context, apartment string, status *domain.DeliveryStatus) ([]*domain.Delivery, error) {
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

	deliveries := make([]*domain.Delivery, 0, len(deliveryModels))
	for i := range deliveryModels {
		deliveries = append(deliveries, deliveryModels[i].ToEntity())
	}
	return deliveries, nil
}

func (r *DeliveryRepository) MarkAsDeleted(ctx context.Context, deliveryID string) error {
	now := time.Now().UTC()
	filter := notDeletedFilter(bson.M{
		"delivery_id": deliveryID,
		"status":      string(domain.DeliveryStatusPending),
	})
	update := bson.M{"$set": bson.M{
		"status":       string(domain.DeliveryStatusDeleted),
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

func (r *DeliveryRepository) MarkArrivalAsNotified(ctx context.Context, deliveryID string) error {
	return r.markAsNotified(ctx, deliveryID, arrivalNotifiedAtField)
}

func (r *DeliveryRepository) MarkPickupAsNotified(ctx context.Context, deliveryID string) error {
	return r.markAsNotified(ctx, deliveryID, pickupNotifiedAtField)
}

func (r *DeliveryRepository) markAsNotified(ctx context.Context, deliveryID, field string) error {
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
	return fmt.Errorf("delivery %s: %w", deliveryID, domain.ErrEntityNotFound)
}
