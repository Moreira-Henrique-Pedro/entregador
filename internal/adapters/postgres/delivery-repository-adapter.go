package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DeliveryRepository struct {
	postgresClient *Client
}

// NewDeliveryRepository creates a new instance of DeliveryRepository
func NewDeliveryRepository(postgresClient *Client) ports.DeliveryRepositoryPort {
	return &DeliveryRepository{
		postgresClient: postgresClient,
	}
}

func (repository *DeliveryRepository) CreateDelivery(ctx context.Context, delivery *entities.Delivery) (*entities.Delivery, error) {
	err := repository.postgresClient.DB.WithContext(ctx).Create(delivery).Error
	if err != nil {
		return nil, err
	}
	return delivery, nil
}

// DeleteDeliveryByID deletes a delivery record by its ID
func (repository *DeliveryRepository) DeleteDeliveryByID(ctx context.Context, id string) error {
	logger := logrus.New()
	result := repository.postgresClient.DB.WithContext(ctx).Delete(&entities.Delivery{}, "id = ?", id)
	if result.Error != nil {
		logger.Error("failed to delete delivery: ", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		//TODO: Criar tabela de erros
		err := fmt.Errorf("no delivery found with the given ID: %s", id)
		logger.Error(err)
		return err
	}
	return nil
}

// GetDeliveryByID retrieves a delivery record by its ID
func (repository *DeliveryRepository) GetDeliveryByID(ctx context.Context, id string) (*entities.Delivery, error) {
	logger := logrus.New()
	var delivery entities.Delivery
	err := repository.postgresClient.DB.WithContext(ctx).First(&delivery, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.WithField("delivery_id", id).Info("delivery not found")
			return nil, errors.New("delivery not found")
		}
		logger.WithError(err).WithField("delivery_id", id).Error("failed to get delivery")
		return nil, err
	}
	return &delivery, nil
}

// UpdateDelivery updates an existing delivery record in the database
func (repository *DeliveryRepository) UpdateDelivery(ctx context.Context, delivery *entities.Delivery) error {
	logger := logrus.New()

	result := repository.postgresClient.DB.WithContext(ctx).Save(delivery)
	if result.Error != nil {
		logger.Error("failed to update delivery: ", result.Error)
		return result.Error
	}
	return nil
}
