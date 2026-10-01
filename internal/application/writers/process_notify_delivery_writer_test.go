package writers

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	notifiermocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier/mocks"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
)

var (
	other101  = entities.NewOtherResident("101")
	arrived   = notifier.NotificationTypeDeliveryArrived
	pickedUp  = notifier.NotificationTypeDeliveryPickedUp
	errFailed = errors.New("mongo down")
)

func TestProcessNotifyDelivery(t *testing.T) {

	ana := &entities.Resident{ResidentID: "ana", Name: "Ana", Apartment: "101", Phone: "111", Type: entities.ResidentTypePrimary}
	bia := &entities.Resident{ResidentID: "bia", Name: "Bia", Apartment: "101", Phone: "222", Type: entities.ResidentTypeSecondary}
	caio := &entities.Resident{ResidentID: "caio", Name: "Caio", Apartment: "101", Type: entities.ResidentTypeSecondary}
	apartment101 := []*entities.Resident{ana, bia, caio, other101}

	pendingFor := func(residentID string) *entities.Delivery {
		return &entities.Delivery{DeliveryID: "d1", Apartment: "101", ResidentID: residentID, PackageType: "caixa", Status: entities.DeliveryStatusPending}
	}
	deletedFor := func(residentID string) *entities.Delivery {
		delivery := pendingFor(residentID)
		delivery.Status = entities.DeliveryStatusDeleted
		return delivery
	}

	type mocks struct {
		residents *repomocks.ResidentRepositoryPort
		sender    *notifiermocks.NotifierPort
	}

	expectSend := func(m mocks, notificationType notifier.NotificationType, resident *entities.Resident, delivery *entities.Delivery, err error) {
		m.sender.EXPECT().Send(mock.Anything, buildNotification(notificationType, resident, delivery)).Return(err).Once()
	}

	tests := []struct {
		name             string
		notificationType notifier.NotificationType
		delivery         *entities.Delivery
		deliveryErr      error
		setup            func(m mocks, delivery *entities.Delivery)
		wantMarked       string
		markErr          error
		wantErr          error
	}{
		{
			name:             "arrival notifies the delivery resident",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
				expectSend(m, arrived, ana, delivery, nil)
			},
			wantMarked: "arrival",
		},
		{
			name:             "pickup notifies the delivery resident",
			notificationType: pickedUp,
			delivery:         deletedFor("ana"),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
				expectSend(m, pickedUp, ana, delivery, nil)
			},
			wantMarked: "pickup",
		},
		{
			name:             "other delivery notifies only the apartment primary",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				expectSend(m, arrived, ana, delivery, nil)
			},
			wantMarked: "arrival",
		},
		{
			name:             "other pickup notifies only the apartment primary",
			notificationType: pickedUp,
			delivery:         deletedFor(other101.ResidentID),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				expectSend(m, pickedUp, ana, delivery, nil)
			},
			wantMarked: "pickup",
		},
		{
			name:             "deleted delivery resident falls back to the apartment primary",
			notificationType: arrived,
			delivery:         pendingFor("gone"),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, "gone").Return(nil, entities.ErrEntityNotFound).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				expectSend(m, arrived, ana, delivery, nil)
			},
			wantMarked: "arrival",
		},
		{
			name:             "apartment without primary promotes the oldest resident and notifies it",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			setup: func(m mocks, delivery *entities.Delivery) {
				anaSecondary := &entities.Resident{ResidentID: "ana", Name: "Ana", Apartment: "101", Phone: "111", Type: entities.ResidentTypeSecondary}
				biaPrimary := &entities.Resident{ResidentID: "bia", Name: "Bia", Apartment: "101", Phone: "222", Type: entities.ResidentTypePrimary}
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				mock.InOrder(
					m.residents.EXPECT().FindByApartment(mock.Anything, "101").
						Return([]*entities.Resident{anaSecondary, bia, other101}, nil).Once(),
					m.residents.EXPECT().EnsurePrimaryResident(mock.Anything, "101").Return(nil).Once(),
					m.residents.EXPECT().FindByApartment(mock.Anything, "101").
						Return([]*entities.Resident{anaSecondary, biaPrimary, other101}, nil).Once(),
				)
				expectSend(m, arrived, biaPrimary, delivery, nil)
			},
			wantMarked: "arrival",
		},
		{
			name:             "primary without phone leaves the other delivery without recipients",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			setup: func(m mocks, _ *entities.Delivery) {
				caioPrimary := &entities.Resident{ResidentID: "caio", Name: "Caio", Apartment: "101", Type: entities.ResidentTypePrimary}
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").
					Return([]*entities.Resident{caioPrimary, bia, other101}, nil).Once()
			},
			wantMarked: "arrival",
		},
		{
			name:             "apartment with only the other resident has no recipients",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			setup: func(m mocks, _ *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").Return([]*entities.Resident{other101}, nil).Twice()
				m.residents.EXPECT().EnsurePrimaryResident(mock.Anything, "101").Return(nil).Once()
			},
			wantMarked: "arrival",
		},
		{
			name:             "ensure primary error is returned",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			setup: func(m mocks, _ *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").Return([]*entities.Resident{bia, other101}, nil).Once()
				m.residents.EXPECT().EnsurePrimaryResident(mock.Anything, "101").Return(errFailed).Once()
			},
			wantErr: errFailed,
		},
		{
			name:             "resident without phone is marked as notified without sending",
			notificationType: arrived,
			delivery:         pendingFor("caio"),
			setup: func(m mocks, _ *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, "caio").Return(caio, nil).Once()
			},
			wantMarked: "arrival",
		},
		{
			name:             "invalid recipient is skipped and the delivery is marked as notified",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				expectSend(m, arrived, ana, delivery, fmt.Errorf("twilio: %w", notifier.ErrInvalidRecipient))
			},
			wantMarked: "arrival",
		},
		{
			name:             "provider failure is returned and not marked so it is retried",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
				expectSend(m, arrived, ana, delivery, errFailed)
			},
			wantErr: errFailed,
		},
		{
			name:             "arrival already notified is skipped",
			notificationType: arrived,
			delivery:         &entities.Delivery{DeliveryID: "d1", Apartment: "101", ResidentID: "ana", Status: entities.DeliveryStatusPending, ArrivalNotifiedAt: time.Now()},
		},
		{
			name:             "arrival of a picked up delivery is skipped",
			notificationType: arrived,
			delivery:         deletedFor("ana"),
		},
		{
			name:             "pickup already notified is skipped",
			notificationType: pickedUp,
			delivery:         &entities.Delivery{DeliveryID: "d1", Apartment: "101", ResidentID: "ana", Status: entities.DeliveryStatusDeleted, PickupNotifiedAt: time.Now()},
		},
		{
			name:             "pickup of a pending delivery is skipped",
			notificationType: pickedUp,
			delivery:         pendingFor("ana"),
		},
		{

			name:             "invalid notification type is discarded",
			notificationType: "sms_marketing",
		},
		{
			name:             "unknown delivery is ignored",
			notificationType: arrived,
			deliveryErr:      entities.ErrEntityNotFound,
		},
		{
			name:             "delivery lookup error is returned",
			notificationType: arrived,
			deliveryErr:      errFailed,
			wantErr:          errFailed,
		},
		{
			name:             "resident lookup error is returned",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			setup: func(m mocks, _ *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(nil, errFailed).Once()
			},
			wantErr: errFailed,
		},
		{
			name:             "apartment lookup error is returned",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			setup: func(m mocks, _ *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, other101.ResidentID).Return(other101, nil).Once()
				m.residents.EXPECT().FindByApartment(mock.Anything, "101").Return(nil, errFailed).Once()
			},
			wantErr: errFailed,
		},
		{
			name:             "mark as notified error is returned",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			setup: func(m mocks, delivery *entities.Delivery) {
				m.residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
				expectSend(m, arrived, ana, delivery, nil)
			},
			wantMarked: "arrival",
			markErr:    errFailed,
			wantErr:    errFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveries := repomocks.NewDeliveryRepositoryPort(t)
			m := mocks{
				residents: repomocks.NewResidentRepositoryPort(t),
				sender:    notifiermocks.NewNotifierPort(t),
			}
			if tt.delivery != nil || tt.deliveryErr != nil {
				deliveries.EXPECT().FindByDeliveryID(mock.Anything, "d1").Return(tt.delivery, tt.deliveryErr).Once()
			}
			if tt.setup != nil {
				tt.setup(m, tt.delivery)
			}
			switch tt.wantMarked {
			case "arrival":
				deliveries.EXPECT().MarkArrivalAsNotified(mock.Anything, "d1").Return(tt.markErr).Once()
			case "pickup":
				deliveries.EXPECT().MarkPickupAsNotified(mock.Anything, "d1").Return(tt.markErr).Once()
			}

			err := NewProcessNotifyDelivery(deliveries, m.residents, m.sender).Handle(context.Background(), &commands.ProcessNotifyDeliveryCommand{
				DeliveryID:       "d1",
				NotificationType: tt.notificationType,
			})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBuildNotification(t *testing.T) {
	resident := &entities.Resident{Name: "Ana", Phone: "111"}

	tests := []struct {
		name             string
		notificationType notifier.NotificationType
		delivery         *entities.Delivery
		wantBody         string
		wantVariables    []string
	}{
		{
			name:             "arrival with package and urgency",
			notificationType: arrived,
			delivery:         &entities.Delivery{Apartment: "101", PackageType: "caixa", Urgency: "alta"},
			wantBody:         "Olá, Ana! Chegou uma entrega para o apartamento 101 (caixa). Retire na portaria. Urgência: alta.",
			wantVariables:    []string{"Ana", "101", "caixa"},
		},
		{
			name:             "arrival without package type uses the default label in the template",
			notificationType: arrived,
			delivery:         &entities.Delivery{Apartment: "101"},
			wantBody:         "Olá, Ana! Chegou uma entrega para o apartamento 101. Retire na portaria.",
			wantVariables:    []string{"Ana", "101", "encomenda"},
		},
		{
			name:             "pickup with package",
			notificationType: pickedUp,
			delivery:         &entities.Delivery{Apartment: "101", PackageType: "caixa", Urgency: "alta"},
			wantBody:         "Olá, Ana! A entrega (caixa) do apartamento 101 foi retirada na portaria.",
			wantVariables:    []string{"Ana", "101", "caixa"},
		},
		{
			name:             "pickup without package type",
			notificationType: pickedUp,
			delivery:         &entities.Delivery{Apartment: "101"},
			wantBody:         "Olá, Ana! A entrega do apartamento 101 foi retirada na portaria.",
			wantVariables:    []string{"Ana", "101", "encomenda"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notification := buildNotification(tt.notificationType, resident, tt.delivery)

			assert.Equal(t, tt.notificationType, notification.Type)
			assert.Equal(t, "111", notification.Phone)
			assert.Equal(t, tt.wantBody, notification.Body)
			assert.Equal(t, tt.wantVariables, notification.Variables)
		})
	}
}
