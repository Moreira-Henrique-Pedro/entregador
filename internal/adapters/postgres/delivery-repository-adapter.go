// package postgres contem implementações específicas para o banco de dados Postgres, incluindo repositórios e clientes.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	postgresClient "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/postgres/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DeliveryRepository é a estrutura que representa o repositório de entregas
type DeliveryRepository struct {
	postgresClient *postgresClient.Client
}

// NewDeliveryRepository cria uma nova instância de DeliveryRepository
func NewDeliveryRepository(postgresClient *postgresClient.Client) ports.DeliveryRepositoryPort {
	return &DeliveryRepository{
		postgresClient: postgresClient,
	}
}

// CreateDelivery insere um novo registro de entrega no banco de dados
func (repository *DeliveryRepository) CreateDelivery(ctx context.Context, delivery *entities.Delivery) (*entities.Delivery, error) {
	logger := logrus.New()
	if delivery.CreatedAt.IsZero() {
		delivery.CreatedAt = time.Now()
	}
	if delivery.UpdatedAt.IsZero() {
		delivery.UpdatedAt = time.Now()
	}
	err := repository.postgresClient.DB.WithContext(ctx).Create(delivery).Error
	if err != nil {
		logger.Error("Error to create delivery", err)
		return nil, err
	}
	return delivery, nil
}

// DeleteDeliveryByID deleta um registro de entrega pelo ID
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

// GetDeliveryByID busca um registro de entrega pelo ID
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

// UpdateDelivery atualiza um registro de entrega
func (repository *DeliveryRepository) UpdateDelivery(ctx context.Context, delivery *entities.Delivery) error {
	logger := logrus.New()

	result := repository.postgresClient.DB.WithContext(ctx).Save(delivery)
	if result.Error != nil {
		logger.Error("failed to update delivery: ", result.Error)
		return result.Error
	}

	// Verificar se algum registro foi atualizado
	if result.RowsAffected == 0 {
		logger.WithField("delivery_id", delivery.ID).Info("no delivery found to update")
		return errors.New("delivery not found")
	}

	return nil
}
