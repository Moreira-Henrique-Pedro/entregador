package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDeliveryUseCaseExecuteSuccess(t *testing.T) {
	mockDeliveryRepo := new(mocks.DeliveryRepositoryPort)
	mockResidentRepo := new(mocks.ResidentRepositoryPort)
	mockTwilio := new(mocks.TwilioPort)

	useCase := usecases.NewCreateDeliveryUseCase(
		mockDeliveryRepo,
		mockResidentRepo,
		mockTwilio,
	)

	ctx := context.Background()

	inputDelivery := entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	createdDelivery := &entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	resident := &entities.Resident{
		Apartamento: "101",
		Resident: []entities.ResidentInfo{
			{
				Nome:     "Maria Silva",
				Telefone: "+5511999998888",
			},
		},
	}

	expectedMessage := "Olá Maria Silva, você tem uma entrega aguardando na portaria"

	mockDeliveryRepo.On("CreateDelivery", ctx, mock.MatchedBy(func(d *entities.Delivery) bool {
		return d.ApNum == inputDelivery.ApNum &&
			d.PackageType == inputDelivery.PackageType
	})).Return(createdDelivery, nil)

	mockResidentRepo.On("GetByApartment", ctx, "101").Return(resident, nil)
	mockTwilio.On("SendWhatsAppMessage", ctx, "+5511999998888", expectedMessage).Return(nil)

	result, err := useCase.Execute(ctx, inputDelivery)

	assert.Equal(t, result.ApNum, createdDelivery.ApNum)
	assert.Equal(t, result.PackageType, createdDelivery.PackageType)
	assert.NoError(t, err, "Não deveria retornar erro em caso de sucesso")
	mockDeliveryRepo.AssertExpectations(t)
	mockResidentRepo.AssertExpectations(t)
	mockTwilio.AssertExpectations(t)
}

func TestCreateDeliveryUseCaseCreateDeliveryError(t *testing.T) {
	// Arrange
	mockDeliveryRepo := new(mocks.DeliveryRepositoryPort)
	mockResidentRepo := new(mocks.ResidentRepositoryPort)
	mockTwilio := new(mocks.TwilioPort)

	useCase := usecases.NewCreateDeliveryUseCase(
		mockDeliveryRepo,
		mockResidentRepo,
		mockTwilio,
	)

	ctx := context.Background()

	inputDelivery := entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	// Simular erro na criação da entrega
	mockErr := errors.New("erro de banco de dados")
	mockDeliveryRepo.On("CreateDelivery", ctx, mock.Anything).Return(nil, mockErr)

	// Act
	result, err := useCase.Execute(ctx, inputDelivery)

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err, "Deveria retornar erro quando falha ao criar entrega")
	assert.ErrorContains(t, err, "falha ao registrar entrega")
	mockDeliveryRepo.AssertExpectations(t)
	// Os outros mocks não devem ser chamados
	mockResidentRepo.AssertNotCalled(t, "GetByApartment")
	mockTwilio.AssertNotCalled(t, "SendWhatsAppMessage")
}

func TestCreateDeliveryUseCaseResidentNotFound(t *testing.T) {
	// Arrange
	mockDeliveryRepo := new(mocks.DeliveryRepositoryPort)
	mockResidentRepo := new(mocks.ResidentRepositoryPort)
	mockTwilio := new(mocks.TwilioPort)

	useCase := usecases.NewCreateDeliveryUseCase(
		mockDeliveryRepo,
		mockResidentRepo,
		mockTwilio,
	)

	ctx := context.Background()

	inputDelivery := entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	createdDelivery := &entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	mockDeliveryRepo.On("CreateDelivery", ctx, mock.Anything).Return(createdDelivery, nil)
	mockResidentRepo.On("GetByApartment", ctx, "101").Return(nil, nil)

	// Act
	result, err := useCase.Execute(ctx, inputDelivery)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, result)
	mockDeliveryRepo.AssertExpectations(t)
	mockResidentRepo.AssertExpectations(t)
	mockTwilio.AssertNotCalled(t, "SendWhatsAppMessage")
}

func TestCreateDeliveryUseCaseEmptyResidentList(t *testing.T) {
	// Arrange
	mockDeliveryRepo := new(mocks.DeliveryRepositoryPort)
	mockResidentRepo := new(mocks.ResidentRepositoryPort)
	mockTwilio := new(mocks.TwilioPort)

	useCase := usecases.NewCreateDeliveryUseCase(
		mockDeliveryRepo,
		mockResidentRepo,
		mockTwilio,
	)

	ctx := context.Background()

	inputDelivery := entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	createdDelivery := &entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	emptyResident := &entities.Resident{
		Apartamento: "101",
		Resident:    []entities.ResidentInfo{},
	}

	mockDeliveryRepo.On("CreateDelivery", ctx, mock.Anything).Return(createdDelivery, nil)
	mockResidentRepo.On("GetByApartment", ctx, "101").Return(emptyResident, nil)

	// Act
	result, err := useCase.Execute(ctx, inputDelivery)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	mockDeliveryRepo.AssertExpectations(t)
	mockResidentRepo.AssertExpectations(t)
	mockTwilio.AssertNotCalled(t, "SendWhatsAppMessage")
}

func TestCreateDeliveryUseCaseSendMessageError(t *testing.T) {
	// Arrange
	mockDeliveryRepo := new(mocks.DeliveryRepositoryPort)
	mockResidentRepo := new(mocks.ResidentRepositoryPort)
	mockTwilio := new(mocks.TwilioPort)

	useCase := usecases.NewCreateDeliveryUseCase(
		mockDeliveryRepo,
		mockResidentRepo,
		mockTwilio,
	)

	ctx := context.Background()

	inputDelivery := entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	createdDelivery := &entities.Delivery{
		ApNum:       "101",
		PackageType: "Envelope",
	}

	resident := &entities.Resident{
		Apartamento: "101",
		Resident: []entities.ResidentInfo{
			{
				Nome:     "Maria Silva",
				Telefone: "+5511999998888",
			},
		},
	}

	expectedMessage := "Olá Maria Silva, você tem uma entrega aguardando na portaria"

	mockDeliveryRepo.On("CreateDelivery", ctx, mock.Anything).Return(createdDelivery, nil)
	mockResidentRepo.On("GetByApartment", ctx, "101").Return(resident, nil)

	mockErr := errors.New("erro na API do Twilio")
	mockTwilio.On("SendWhatsAppMessage", ctx, "+5511999998888", expectedMessage).Return(mockErr)

	// Act
	result, err := useCase.Execute(ctx, inputDelivery)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	mockDeliveryRepo.AssertExpectations(t)
	mockResidentRepo.AssertExpectations(t)
	mockTwilio.AssertExpectations(t)
}
