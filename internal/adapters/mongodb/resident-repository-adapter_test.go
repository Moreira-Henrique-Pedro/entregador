package mongodb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/mongodb"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	mockClient "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
)

// MockSingleResult implementa a interface SingleResultPort para testes
type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(val interface{}) error {
	args := m.Called(val)
	if args.Get(0) != nil {
		// Verifica se há um resultado esperado e copia para o valor de destino
		resident, ok := args.Get(0).(*entities.Resident)
		if ok && val != nil {
			*val.(*entities.Resident) = *resident
		}
	}
	return args.Error(1)
}

func TestCreateResidentSuccess(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	resident := &entities.Resident{
		Apartamento: "63",
		Resident: []entities.ResidentInfo{
			{
				Nome:     "Henrique",
				Telefone: "123456789",
			},
		},
	}

	mockCollection.On("InsertOne", mock.Anything, resident).Return(nil, nil)

	err := repo.Create(context.Background(), resident)

	assert.NoError(t, err)
	mockCollection.AssertExpectations(t)
}

func TestCreateResidentError(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	resident := &entities.Resident{
		Apartamento: "63",
		Resident: []entities.ResidentInfo{
			{
				Nome:     "Henrique",
				Telefone: "123456789",
			},
		},
	}

	expectedErr := errors.New("insert failed")

	mockCollection.On("InsertOne", mock.Anything, resident).Return(nil, expectedErr)

	err := repo.Create(context.Background(), resident)

	assert.EqualError(t, err, expectedErr.Error())
	mockCollection.AssertExpectations(t)
}

func TestUpdateResidentSuccess(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	resident := &entities.Resident{
		Apartamento: "63",
		Resident: []entities.ResidentInfo{
			{Nome: "Henrique", Telefone: "999999999"},
		},
	}

	mockCollection.On(
		"UpdateOne",
		mock.Anything,
		bson.M{"apartamento": resident.Apartamento},
		bson.M{"$set": resident},
	).Return(nil, nil)

	err := repo.Update(context.Background(), resident)

	assert.NoError(t, err)
	mockCollection.AssertExpectations(t)
}

func TestUpdateResidentError(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	resident := &entities.Resident{
		Apartamento: "63",
		Resident: []entities.ResidentInfo{
			{Nome: "Henrique", Telefone: "999999999"},
		},
	}

	expectedErr := errors.New("update failed")

	mockCollection.On(
		"UpdateOne",
		mock.Anything,
		bson.M{"apartamento": resident.Apartamento},
		bson.M{"$set": resident},
	).Return(nil, expectedErr)

	err := repo.Update(context.Background(), resident)

	assert.EqualError(t, err, expectedErr.Error())
	mockCollection.AssertExpectations(t)
}

func TestDeleteResidentSuccess(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	mockCollection.On(
		"DeleteOne",
		mock.Anything,
		bson.M{"apartamento": "63"},
	).Return(nil, nil)

	err := repo.Delete(context.Background(), "63")

	assert.NoError(t, err)
	mockCollection.AssertExpectations(t)
}

func TestDeleteResidentError(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	expectedErr := errors.New("delete failed")

	mockCollection.On(
		"DeleteOne",
		mock.Anything,
		bson.M{"apartamento": "63"},
	).Return(nil, expectedErr)

	err := repo.Delete(context.Background(), "63")

	assert.EqualError(t, err, expectedErr.Error())
	mockCollection.AssertExpectations(t)
}

func TestGetResidentByApartmentSuccess(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	expectedResident := &entities.Resident{
		Apartamento: "63",
		Resident: []entities.ResidentInfo{
			{
				Nome:     "Henrique",
				Telefone: "123456789",
			},
		},
	}

	// Criando o mock para o SingleResult
	mockResult := new(MockSingleResult)

	// Configurando o mock do Decode - aqui estamos passando o expectedResident e nil para erro
	mockResult.On("Decode", mock.AnythingOfType("*entities.Resident")).Return(expectedResident, nil)

	// Mock do FindOne - retorna o mockResult que implementa SingleResultPort
	mockCollection.On("FindOne", mock.Anything, bson.M{"apartamento": "63"}, mock.Anything).Return(mockResult)

	// Chama o repositório para pegar o residente
	resident, err := repo.GetByApartment(context.Background(), "63")

	// Verifica os resultados
	assert.NoError(t, err)
	assert.Equal(t, expectedResident.Apartamento, resident.Apartamento)
	assert.Equal(t, expectedResident.Resident, resident.Resident)
	mockCollection.AssertExpectations(t)
	mockResult.AssertExpectations(t)
}

func TestGetResidentByApartmentError(t *testing.T) {
	mockCollection := new(mockClient.MongoCollectionPort)
	repo := mongodb.NewResidentRepository(mockCollection)

	// Simulando o erro no Decode
	mockResult := new(MockSingleResult)
	expectedErr := errors.New("not found")

	mockResult.On("Decode", mock.AnythingOfType("*entities.Resident")).Return(nil, expectedErr)

	// Mock do FindOne
	mockCollection.On("FindOne", mock.Anything, bson.M{"apartamento": "63"}, mock.Anything).Return(mockResult)

	resident, err := repo.GetByApartment(context.Background(), "63")

	// Assertando o erro
	assert.Nil(t, resident)
	assert.EqualError(t, err, expectedErr.Error())
	mockCollection.AssertExpectations(t)
	mockResult.AssertExpectations(t)
}
