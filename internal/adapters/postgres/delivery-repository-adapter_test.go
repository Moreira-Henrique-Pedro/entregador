package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	postgresGorm "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/postgres"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
)

type DeliveryRepositoryTestSuite struct {
	suite.Suite
	db      *gorm.DB
	mock    sqlmock.Sqlmock
	repo    ports.DeliveryRepositoryPort
	ctx     context.Context
	cleanup func()
}

func (suite *DeliveryRepositoryTestSuite) SetupTest() {
	var err error
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(suite.T(), err)
	suite.mock = mock

	// Abre a conexão com o banco GORM usando o mock
	suite.db, err = gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Adicionando log para visualização das queries
	})
	assert.NoError(suite.T(), err)

	// Configura o repositório de entregas com a conexão GORM
	client := &postgresGorm.Client{DB: suite.db}
	suite.repo = postgresGorm.NewDeliveryRepository(client)

	// Contexto que será usado nos testes
	suite.ctx = context.Background()

}

func TestDeliveryRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(DeliveryRepositoryTestSuite))
}

func (suite *DeliveryRepositoryTestSuite) TestCreateDeliverySuccess() {
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO \"deliveries\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	now := time.Now()
	delivery := &entities.Delivery{
		ApNum:       "63",
		PackageType: "box",
		Urgency:     "high",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := suite.repo.CreateDelivery(suite.ctx, delivery)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)

	assert.Equal(suite.T(), "63", result.ApNum)
	assert.Equal(suite.T(), "box", result.PackageType)
	assert.Equal(suite.T(), "high", result.Urgency)
	assert.Equal(suite.T(), "pending", result.Status)
	assert.WithinDuration(suite.T(), now, result.CreatedAt, time.Second)
	assert.WithinDuration(suite.T(), now, result.UpdatedAt, time.Second)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestCreateDeliveryError() {
	expectedError := errors.New("some error")
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO \"deliveries\"").WillReturnError(expectedError)
	suite.mock.ExpectRollback()

	now := time.Now()

	delivery := &entities.Delivery{
		ApNum:       "63",
		PackageType: "box",
		Urgency:     "high",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := suite.repo.CreateDelivery(suite.ctx, delivery)
	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestDeleteDeliveryByIDSuccess() {
	// Arrange
	deliveryID := "123"
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("DELETE FROM \"deliveries\"").
		WithArgs(deliveryID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repo.DeleteDeliveryByID(suite.ctx, deliveryID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestDeleteDeliveryByIDError() {
	// Arrange
	deliveryID := "456"
	expectedError := errors.New("database error")
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("DELETE FROM \"deliveries\"").
		WithArgs(deliveryID).
		WillReturnError(expectedError)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repo.DeleteDeliveryByID(suite.ctx, deliveryID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestDeleteDeliveryByIDNotFound() {
	// Arrange
	deliveryID := "789"
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("DELETE FROM \"deliveries\"").
		WithArgs(deliveryID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repo.DeleteDeliveryByID(suite.ctx, deliveryID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "no delivery found with the given ID: 789", err.Error())
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}
