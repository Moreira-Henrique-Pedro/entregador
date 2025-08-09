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

	resident, err := usecase.handleResident(ctx, deliveryID)
	if err != nil {
		log.WithError(err).Warn("Erro ao buscar residente, continuando sem notificação")
	}

	err = usecase.deliveryRepo.DeleteDeliveryByID(ctx, deliveryID)
	if err != nil {
		log.WithError(err).Error("Erro ao excluir entrega")
		return fmt.Errorf("erro ao excluir entrega: %w", err)
	}

	if resident != nil && len(resident.Resident) > 0 {
		if err := usecase.handleSendMessage(ctx, *resident); err != nil {
			log.WithError(err).Warn("Erro ao enviar mensagem, entrega já foi registrada")
		}
	}

	return nil
}

func (usecase *DeleteDeliveryUseCase) handleResident(ctx context.Context, deliveryID string) (*entities.Resident, error) {
	log := logrus.WithField("deliveryID", deliveryID)

	resident, err := usecase.residentRepo.GetByDeliveryID(ctx, deliveryID)
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

func (usecase *DeleteDeliveryUseCase) handleSendMessage(ctx context.Context, resident entities.Resident) error {
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
