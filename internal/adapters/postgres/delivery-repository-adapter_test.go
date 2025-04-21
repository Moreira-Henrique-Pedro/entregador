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
	postgresClientGorm "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/postgres/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
)

type DeliveryRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo ports.DeliveryRepositoryPort
	ctx  context.Context
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
	client := &postgresClientGorm.Client{DB: suite.db}
	suite.repo = postgresGorm.NewDeliveryRepository(client)

	// Contexto que será usado nos testes
	suite.ctx = context.Background()

}

func TestDeliveryRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(DeliveryRepositoryTestSuite))
}

func (suite *DeliveryRepositoryTestSuite) TestCreateDeliverySuccess() {
	// Arrange
	now := time.Now()
	delivery := &entities.Delivery{
		ID:          "teste-id",
		ApNum:       "63",
		PackageType: "box",
		Urgency:     "high",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO \"deliveries\"").
		WithArgs(
			delivery.ID,
			delivery.ApNum,
			delivery.PackageType,
			delivery.Urgency,
			delivery.Status,
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt (soft delete)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	// Act
	result, err := suite.repo.CreateDelivery(suite.ctx, delivery)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), delivery.ID, result.ID)
	assert.Equal(suite.T(), delivery.ApNum, result.ApNum)
	assert.Equal(suite.T(), delivery.PackageType, result.PackageType)
	assert.Equal(suite.T(), delivery.Urgency, result.Urgency)
	assert.Equal(suite.T(), delivery.Status, result.Status)
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

func (suite *DeliveryRepositoryTestSuite) TestGetDeliveryByIDSuccess() {
	// Arrange
	deliveryID := "123"
	expectedDelivery := &entities.Delivery{
		ID:          deliveryID,
		ApNum:       "63",
		PackageType: "box",
		Urgency:     "high",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.mock.ExpectQuery(`SELECT (.+) FROM "deliveries" WHERE id = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(deliveryID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "ap_num", "package_type", "urgency", "status", "created_at", "updated_at"}).
			AddRow(expectedDelivery.ID, expectedDelivery.ApNum, expectedDelivery.PackageType, expectedDelivery.Urgency, expectedDelivery.Status, expectedDelivery.CreatedAt, expectedDelivery.UpdatedAt))

	// Act
	result, err := suite.repo.GetDeliveryByID(suite.ctx, deliveryID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedDelivery.ID, result.ID)
	assert.Equal(suite.T(), expectedDelivery.ApNum, result.ApNum)
	assert.Equal(suite.T(), expectedDelivery.PackageType, result.PackageType)
	assert.Equal(suite.T(), expectedDelivery.Urgency, result.Urgency)
	assert.Equal(suite.T(), expectedDelivery.Status, result.Status)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestGetDeliveryByIDNotFound() {
	// Arrange
	deliveryID := "456"

	suite.mock.ExpectQuery(`SELECT (.+) FROM "deliveries" WHERE id = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(deliveryID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Act
	result, err := suite.repo.GetDeliveryByID(suite.ctx, deliveryID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "delivery not found", err.Error())
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestGetDeliveryByIDDatabaseError() {
	// Arrange
	deliveryID := "789"
	expectedError := errors.New("database connection error")

	suite.mock.ExpectQuery(`SELECT (.+) FROM "deliveries" WHERE id = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(deliveryID, 1).
		WillReturnError(expectedError)

	// Act
	result, err := suite.repo.GetDeliveryByID(suite.ctx, deliveryID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), expectedError, err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestUpdateDeliverySuccess() {
	// Arrange
	now := time.Now()
	delivery := &entities.Delivery{
		ID:          "test-id-123",
		ApNum:       "63",
		PackageType: "box",
		Urgency:     "high",
		Status:      "delivered",
		CreatedAt:   now.Add(-24 * time.Hour),
		UpdatedAt:   now,
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "deliveries" SET`).
		WithArgs(
			delivery.ApNum,
			delivery.PackageType,
			delivery.Urgency,
			delivery.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			delivery.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	suite.mock.ExpectCommit()

	// Act
	err := suite.repo.UpdateDelivery(suite.ctx, delivery)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DeliveryRepositoryTestSuite) TestUpdateDeliveryError() {
	// Arrange
	now := time.Now()
	delivery := &entities.Delivery{
		ID:          "test-id-456",
		ApNum:       "64",
		PackageType: "envelope",
		Urgency:     "low",
		Status:      "delivered",
		CreatedAt:   now.Add(-24 * time.Hour),
		UpdatedAt:   now,
	}

	expectedError := errors.New("database error")

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(`UPDATE "deliveries" SET`).
		WithArgs(
			delivery.ApNum,       // ap_num
			delivery.PackageType, // package_type
			delivery.Urgency,     // urgency
			delivery.Status,      // status
			sqlmock.AnyArg(),     // created_at
			sqlmock.AnyArg(),     // updated_at
			sqlmock.AnyArg(),     // delete_at (pode ser nulo)
			delivery.ID,          // ID (para a cláusula WHERE)
		).
		WillReturnError(expectedError)
	suite.mock.ExpectRollback()

	// Act
	err := suite.repo.UpdateDelivery(suite.ctx, delivery)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}
