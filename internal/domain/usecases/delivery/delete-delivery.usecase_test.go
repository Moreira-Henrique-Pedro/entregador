package usecases_test

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports/mocks"
	usecases "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/usecases/delivery"
)

func setupDeleteDeliveryUseCase() (
	ctx context.Context,
	mockDeliveryRepo *mocks.DeliveryRepositoryPort,
	mockResidentRepo *mocks.ResidentRepositoryPort,
	mockTwilio *mocks.TwilioPort,
	useCase usecases.DeleteDeliveryUseCasePort,
) {
	ctx = context.Background()
	mockDeliveryRepo = new(mocks.DeliveryRepositoryPort)
	mockResidentRepo = new(mocks.ResidentRepositoryPort)
	mockTwilio = new(mocks.TwilioPort)
	useCase = usecases.NewDeleteDeliveryUseCase(
		mockDeliveryRepo,
		mockResidentRepo,
		mockTwilio,
	)
	return
}
