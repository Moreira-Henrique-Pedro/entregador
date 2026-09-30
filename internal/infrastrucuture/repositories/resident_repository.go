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

// Timestamp fields have no bson tags, so the driver stores them lowercased.
const (
	createdAtField         = "createdat"
	updatedAtField         = "updatedat"
	deleteAtField          = "deleteat"
	arrivalNotifiedAtField = "arrivalnotifiedat"
	pickupNotifiedAtField  = "pickupnotifiedat"
)

type MongoDBResidentRepository struct {
	collection client.MongoClientCollectionPort
}

func NewMongoDBResidentRepository(ctx context.Context, client client.MongoClientCollectionPort) (interfaces.ResidentRepositoryPort, error) {
	if err := client.EnsureIndexes(ctx, residentIndexes()); err != nil {
		return nil, fmt.Errorf("ensure residents indexes: %w", err)
	}
	return &MongoDBResidentRepository{
		collection: client,
	}, nil
}

func residentIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "resident_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		// FindByApartment / FindByPhone always filter out soft-deleted residents.
		{Keys: bson.D{{Key: "apartment", Value: 1}, {Key: deleteAtField, Value: 1}}},
		{Keys: bson.D{{Key: "phone", Value: 1}, {Key: deleteAtField, Value: 1}}},
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

func (r *MongoDBResidentRepository) EnsureOtherResident(ctx context.Context, apartment string) error {
	model := models.ResidentFromEntity(entities.NewOtherResident(apartment))
	now := time.Now().UTC()
	model.CreatedAt = now
	model.UpdatedAt = now

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"resident_id": model.ResidentID},
		bson.M{"$setOnInsert": model},
		options.Update().SetUpsert(true),
	)
	// Two concurrent upserts may race on the unique index; the other one already created it.
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (r *MongoDBResidentRepository) Update(ctx context.Context, resident *entities.Resident) error {
	if resident == nil {
		return errors.New("resident is nil")
	}

	fields := bson.M{updatedAtField: time.Now().UTC()}
	if resident.Name != "" {
		fields["name"] = resident.Name
	}
	if resident.Apartment != "" {
		fields["apartment"] = resident.Apartment
	}
	if resident.Phone != "" {
		fields["phone"] = resident.Phone
	}

	result, err := r.collection.UpdateOne(ctx, activeResidentFilter(bson.M{"resident_id": resident.ResidentID}), bson.M{"$set": fields})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return residentNotFound(resident.ResidentID)
	}
	return nil
}

func (r *MongoDBResidentRepository) DeleteByResidentID(ctx context.Context, residentID string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{deleteAtField: now, updatedAtField: now}}

	result, err := r.collection.UpdateOne(ctx, activeResidentFilter(bson.M{"resident_id": residentID}), update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return residentNotFound(residentID)
	}
	return nil
}

func (r *MongoDBResidentRepository) FindByResidentID(ctx context.Context, residentID string) (*entities.Resident, error) {
	var model models.Resident
	err := r.collection.FindOne(ctx, activeResidentFilter(bson.M{"resident_id": residentID})).Decode(&model)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, residentNotFound(residentID)
	}
	if err != nil {
		return nil, err
	}
	return model.ToEntity(), nil
}

func (r *MongoDBResidentRepository) FindByApartment(ctx context.Context, apartment string) ([]*entities.Resident, error) {
	return r.find(ctx, activeResidentFilter(bson.M{"apartment": apartment}))
}

func (r *MongoDBResidentRepository) FindByPhone(ctx context.Context, phone string) ([]*entities.Resident, error) {
	return r.find(ctx, activeResidentFilter(bson.M{"phone": phone}))
}

func (r *MongoDBResidentRepository) find(ctx context.Context, filter bson.M) ([]*entities.Resident, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var residentModels []models.Resident
	if err := cursor.All(ctx, &residentModels); err != nil {
		return nil, err
	}

	residents := make([]*entities.Resident, 0, len(residentModels))
	for i := range residentModels {
		residents = append(residents, residentModels[i].ToEntity())
	}
	return residents, nil
}

func residentNotFound(residentID string) error {
	return fmt.Errorf("resident %s: %w", residentID, entities.ErrEntityNotFound)
}

// activeResidentFilter ignores soft-deleted residents (deleteat set).
func activeResidentFilter(filter bson.M) bson.M {
	return notDeletedFilter(filter)
}

// notDeletedFilter ignores soft-deleted documents (deleteat set).
func notDeletedFilter(filter bson.M) bson.M {
	filter[deleteAtField] = time.Time{}
	return filter
}
