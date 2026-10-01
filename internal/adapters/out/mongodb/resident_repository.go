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

// Timestamp fields have no bson tags, so the driver stores them lowercased.
const (
	createdAtField         = "createdat"
	updatedAtField         = "updatedat"
	deleteAtField          = "deleteat"
	arrivalNotifiedAtField = "arrivalnotifiedat"
	pickupNotifiedAtField  = "pickupnotifiedat"
)

const primaryPerApartmentIndex = "apartment_primary_unique"

type ResidentRepository struct {
	collection client.MongoClientCollectionPort
}

func NewResidentRepository(ctx context.Context, client client.MongoClientCollectionPort) (out.ResidentRepository, error) {
	if err := client.EnsureIndexes(ctx, residentIndexes()); err != nil {
		return nil, fmt.Errorf("ensure residents indexes: %w", err)
	}
	return &ResidentRepository{
		collection: client,
	}, nil
}

func residentIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "resident_id", Value: 1}}, Options: options.Index().SetUnique(true)},

		{Keys: bson.D{{Key: "apartment", Value: 1}, {Key: deleteAtField, Value: 1}}},
		{Keys: bson.D{{Key: "phone", Value: 1}, {Key: deleteAtField, Value: 1}}},
		// At most one active primary per apartment, even under concurrent promotions.
		{
			Keys: bson.D{{Key: "apartment", Value: 1}},
			Options: options.Index().
				SetName(primaryPerApartmentIndex).
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"type": string(domain.ResidentTypePrimary), deleteAtField: time.Time{}}),
		},
	}
}

func (r *ResidentRepository) Insert(ctx context.Context, resident *domain.Resident) error {
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
		logger.Warn("Duplicate key error while inserting resident", "resident_id", model.ResidentID, "error", err.Error())
		return nil
	}
	return err
}

func (r *ResidentRepository) EnsureOtherResident(ctx context.Context, apartment string) error {
	model := models.ResidentFromEntity(domain.NewOtherResident(apartment))
	now := time.Now().UTC()
	model.CreatedAt = now
	model.UpdatedAt = now

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"resident_id": model.ResidentID},
		bson.M{"$setOnInsert": model},
		options.Update().SetUpsert(true),
	)

	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (r *ResidentRepository) EnsurePrimaryResident(ctx context.Context, apartment string) error {
	err := r.collection.FindOne(ctx, activeResidentFilter(bson.M{
		"apartment": apartment,
		"type":      string(domain.ResidentTypePrimary),
	})).Err()
	if err == nil {
		return nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	var candidate models.Resident
	err = r.collection.FindOne(
		ctx,
		activeResidentFilter(bson.M{"apartment": apartment, "type": bson.M{"$ne": string(domain.ResidentTypeOther)}}),
		options.FindOne().SetSort(bson.D{{Key: createdAtField, Value: 1}, {Key: "resident_id", Value: 1}}),
	).Decode(&candidate)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		activeResidentFilter(bson.M{"resident_id": candidate.ResidentID}),
		bson.M{"$set": bson.M{"type": string(domain.ResidentTypePrimary), updatedAtField: time.Now().UTC()}},
	)

	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (r *ResidentRepository) Update(ctx context.Context, resident *domain.Resident) error {
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
	if resident.Type != "" {
		fields["type"] = string(resident.Type)
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

func (r *ResidentRepository) DeleteByResidentID(ctx context.Context, residentID string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{
		"status":       string(domain.ResidentStatusDeleted),
		deleteAtField:  now,
		updatedAtField: now,
	}}

	result, err := r.collection.UpdateOne(ctx, activeResidentFilter(bson.M{"resident_id": residentID}), update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return residentNotFound(residentID)
	}
	return nil
}

func (r *ResidentRepository) FindByResidentID(ctx context.Context, residentID string) (*domain.Resident, error) {
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

func (r *ResidentRepository) FindByApartment(ctx context.Context, apartment string) ([]*domain.Resident, error) {
	return r.find(ctx, activeResidentFilter(bson.M{"apartment": apartment}))
}

func (r *ResidentRepository) FindByPhone(ctx context.Context, phone string) ([]*domain.Resident, error) {
	return r.find(ctx, activeResidentFilter(bson.M{"phone": phone}))
}

func (r *ResidentRepository) find(ctx context.Context, filter bson.M) ([]*domain.Resident, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var residentModels []models.Resident
	if err := cursor.All(ctx, &residentModels); err != nil {
		return nil, err
	}

	residents := make([]*domain.Resident, 0, len(residentModels))
	for i := range residentModels {
		residents = append(residents, residentModels[i].ToEntity())
	}
	return residents, nil
}

func residentNotFound(residentID string) error {
	return fmt.Errorf("resident %s: %w", residentID, domain.ErrEntityNotFound)
}

func activeResidentFilter(filter bson.M) bson.M {
	return notDeletedFilter(filter)
}

func notDeletedFilter(filter bson.M) bson.M {
	filter[deleteAtField] = time.Time{}
	return filter
}
