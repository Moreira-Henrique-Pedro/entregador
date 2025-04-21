package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
)

type CreateDeliveryUseCasePort interface {
	Execute(ctx context.Context, delivery entities.Delivery) error
}

type CreateDeliveryUseCase struct {
	deliveryRepo ports.DeliveryRepositoryPort
	residentRepo ports.ResidentRepositoryPort
	messaging    ports.TwilioPort
}

func NewCreateDeliveryUseCase(
	deliveryRepo ports.DeliveryRepositoryPort,
	residentRepo ports.ResidentRepositoryPort,
	messaging ports.TwilioPort,
) CreateDeliveryUseCasePort {
	return &CreateDeliveryUseCase{
		deliveryRepo: deliveryRepo,
		residentRepo: residentRepo,
		messaging:    messaging,
	}
}

func (usecase *CreateDeliveryUseCase) Execute(ctx context.Context, delivery entities.Delivery) error {
	log := logrus.WithFields(logrus.Fields{
		"apNum":       delivery.ApNum,
		"packageType": delivery.PackageType,
	})
	log.Info("Iniciando criação de entrega")

	createdDelivery, err := usecase.handleCreateDelivery(ctx, delivery)
	if err != nil {
		return err
	}

	resident, err := usecase.handleResident(ctx, *createdDelivery)
	if err != nil {
		log.WithError(err).Warn("Erro ao buscar residente, continuando sem notificação")
	}

	if resident != nil && len(resident.Resident) > 0 {
		if err := usecase.handleSendMessage(ctx, *resident); err != nil {
			log.WithError(err).Warn("Erro ao enviar mensagem, entrega já foi registrada")
		}
	}

	log.Info("Entrega registrada com sucesso")
	return nil
}

func (usecase *CreateDeliveryUseCase) handleCreateDelivery(ctx context.Context, delivery entities.Delivery) (*entities.Delivery, error) {
	log := logrus.WithField("deliveryID", delivery.ID)

	log.Info("Iniciando criação de entrega")

	createdDelivery, err := usecase.deliveryRepo.CreateDelivery(ctx, &delivery)
	if err != nil {
		log.WithError(err).Error("Erro ao registrar entrega no banco de dados")
		return nil, fmt.Errorf("falha ao registrar entrega: %w", err)
	}

	log.Info("Entrega registrada com sucesso")
	return createdDelivery, nil
}

func (usecase *CreateDeliveryUseCase) handleResident(ctx context.Context, delivery entities.Delivery) (*entities.Resident, error) {
	log := logrus.WithField("apNum", delivery.ApNum)

	resident, err := usecase.residentRepo.GetByApartment(ctx, delivery.ApNum)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar morador no banco de dados")
		return nil, nil
	}

	if resident == nil || len(resident.Resident) == 0 {
		log.Warn("Morador não encontrado")
		return nil, nil
	}

	residentInfo := resident.Resident[0]
	log.WithFields(logrus.Fields{
		"name":  residentInfo.Nome,
		"phone": residentInfo.Telefone,
	}).Info("Morador encontrado, preparando mensagem")

	return resident, nil
}

func (usecase *CreateDeliveryUseCase) handleSendMessage(ctx context.Context, resident entities.Resident) error {
	log := logrus.WithField("Send Message To", resident.Resident[0].Telefone)

	// TODO: Criar constantes com mensagens
	message := fmt.Sprintf(
		"Olá %s, você tem uma entrega aguardando na portaria",
		resident.Resident[0].Nome,
	)

	err := usecase.messaging.SendWhatsAppMessage(ctx, resident.Resident[0].Telefone, message)
	if err != nil {
		log.WithError(err).Error("Erro ao enviar mensagem via Twilio")
		return nil
	}

	log.Info("Mensagem enviada com sucesso")
	return nil
}
