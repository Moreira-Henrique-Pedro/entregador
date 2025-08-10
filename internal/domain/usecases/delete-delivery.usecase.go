// package usecases implementa o caso de uso de criação de entrega
package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
)

// DeleteDeliveryUseCasePort é a interface que define os métodos para o caso de uso de exclusão de entrega
type DeleteDeliveryUseCasePort interface {
	Execute(ctx context.Context, deliveryID string) error
}

// DeleteDeliveryUseCase é a estrutura que implementa o caso de uso de exclusão de entrega
type DeleteDeliveryUseCase struct {
	deliveryRepo ports.DeliveryRepositoryPort
	residentRepo ports.ResidentRepositoryPort
	messaging    ports.TwilioPort
}

// NewDeleteDeliveryUseCase cria uma nova instância do caso de uso de exclusão de entrega
func NewDeleteDeliveryUseCase(
	deliveryRepo ports.DeliveryRepositoryPort,
	residentRepo ports.ResidentRepositoryPort,
	messaging ports.TwilioPort,
) DeleteDeliveryUseCasePort {
	return &DeleteDeliveryUseCase{
		deliveryRepo: deliveryRepo,
		residentRepo: residentRepo,
		messaging:    messaging,
	}
}

// Execute executa o caso de uso de exclusão de entrega
func (usecase *DeleteDeliveryUseCase) Execute(ctx context.Context, deliveryID string) error {
	log := logrus.WithField("deliveryID", deliveryID)
	log.Info("Iniciando exclusão de entrega")

	if deliveryID == "" {
		return fmt.Errorf("ID da entrega não pode ser vazio")
	}

	delivery, err := usecase.getDelivery(ctx, deliveryID)
	if err != nil {
		log.WithError(err).Warn("Erro ao buscar entrega")
		return err
	}

	if delivery == nil {
		log.Warn("Entrega não encontrada")
		return err
	}

	appNum := delivery.ApNum

	if err := usecase.deleteDelivery(ctx, deliveryID); err != nil {
		log.WithError(err).Warn("Erro ao excluir entrega")
		return err
	}

	if err := usecase.sendMessage(ctx, appNum); err != nil {
		log.WithError(err).WithField("apartment", appNum).Warn("Erro ao enviar mensagem para o apartamento")
	}

	return nil
}

func (usecase *DeleteDeliveryUseCase) deleteDelivery(ctx context.Context, deliveryID string) error {
	log := logrus.WithField("deliveryID", deliveryID)
	err := usecase.deliveryRepo.DeleteDeliveryByID(ctx, deliveryID)
	if err != nil {
		log.WithError(err).Error("Erro ao excluir entrega")
		return fmt.Errorf("erro ao excluir entrega: %w", err)
	}
	return nil
}

func (usecase *DeleteDeliveryUseCase) getDelivery(ctx context.Context, deliveryID string) (*entities.Delivery, error) {
	log := logrus.WithField("deliveryID", deliveryID)

	delivery, err := usecase.deliveryRepo.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar entrega no banco de dados")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"delivery":    delivery.ID,
		"apNum":       delivery.ApNum,
		"packageType": delivery.PackageType,
		"urgency":     delivery.Urgency,
	}).Info("Entrega encontrada, buscando residente")

	return delivery, nil
}

func (usecase *DeleteDeliveryUseCase) sendMessage(ctx context.Context, appNum string) error {
	log := logrus.WithField("Send Message To", appNum)

	resident, err := usecase.residentRepo.GetByApartment(ctx, appNum)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar residente pelo número do apartamento")
		return fmt.Errorf("erro ao buscar residente: %w", err)
	}

	// TODO: Criar constantes com mensagens
	message := fmt.Sprintf(
		"Olá %s, você tem uma entrega aguardando na portaria",
		resident.Resident[0].Nome,
	)

	err = usecase.messaging.SendWhatsAppMessage(ctx, resident.Resident[0].Telefone, message)
	if err != nil {
		log.WithError(err).Error("Erro ao enviar mensagem via Twilio")
		return nil
	}

	log.Info("Mensagem enviada com sucesso")
	return nil
}
