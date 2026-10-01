package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	pubsubmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub/mocks"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
)

const internalTopic = "delivery-internal.commands"

func TestProcessCreateDelivery(t *testing.T) {
	failure := errors.New("mongo down")
	otherID := entities.OtherResidentID("101")
	ana := &entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypePrimary}
	bob := &entities.Resident{ResidentID: "bob", Apartment: "202", Type: entities.ResidentTypePrimary}
	apartment101 := []*entities.Resident{ana, entities.NewOtherResident("101")}

	tests := []struct {
		name          string
		command       commands.ProcessCreateDeliveryCommand
		setup         func(residents *repomocks.ResidentRepositoryPort)
		wantResident  string
		insertErr     error
		publish       bool
		publishErr    error
		wantErr       error
		wantPermanent bool
	}{
		{
			name:    "resident of the apartment receives the delivery",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana", PackageType: "caixa", Urgency: "alta"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
			},
			wantResident: "ana",
			publish:      true,
		},
		{
			name:    "no resident informed goes to the apartment other",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(nil).Once()
			},
			wantResident: otherID,
			publish:      true,
		},
		{
			name:    "unknown resident goes to the apartment other",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ghost"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ghost").Return(nil, entities.ErrEntityNotFound).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(nil).Once()
			},
			wantResident: otherID,
			publish:      true,
		},
		{
			name:    "resident from another apartment goes to the apartment other",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "bob"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "bob").Return(bob, nil).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(nil).Once()
			},
			wantResident: otherID,
			publish:      true,
		},
		{
			name:    "apartment without residents is rejected without retry",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "303"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "303").Return(nil, nil).Once()
			},
			wantErr:       entities.ErrNoResidentInApartment,
			wantPermanent: true,
		},
		{
			name:    "apartment with only the other resident is rejected without retry",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "404"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "404").
					Return([]*entities.Resident{entities.NewOtherResident("404")}, nil).Once()
			},
			wantErr:       entities.ErrNoResidentInApartment,
			wantPermanent: true,
		},
		{
			name:    "apartment residents lookup error is returned for retry",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(nil, failure).Once()
			},
			wantErr: failure,
		},
		{
			name:    "command without apartment is discarded",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", ResidentID: "ana"},
		},
		{
			name:    "resident lookup error is returned",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(nil, failure).Once()
			},
			wantErr: failure,
		},
		{
			name:    "ensure other error is returned",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(failure).Once()
			},
			wantErr: failure,
		},
		{
			name:    "insert error is returned",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
			},
			wantResident: "ana",
			insertErr:    failure,
			wantErr:      failure,
		},
		{
			name:    "notify publish error is returned so the command is retried",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana"},
			setup: func(residents *repomocks.ResidentRepositoryPort) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
			},
			wantResident: "ana",
			publish:      true,
			publishErr:   failure,
			wantErr:      failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := repomocks.NewResidentRepositoryPort(t)
			deliveries := repomocks.NewDeliveryRepositoryPort(t)
			publisher := pubsubmocks.NewMessagePublisher[any](t)
			if tt.setup != nil {
				tt.setup(residents)
			}
			if tt.wantResident != "" {
				deliveries.EXPECT().Insert(mock.Anything, &entities.Delivery{
					ID:          tt.command.CommandID,
					DeliveryID:  tt.command.CommandID,
					Apartment:   tt.command.Apartment,
					ResidentID:  tt.wantResident,
					PackageType: tt.command.PackageType,
					Urgency:     tt.command.Urgency,
					Status:      entities.DeliveryStatusPending,
				}).Return(tt.insertErr).Once()
			}
			if tt.publish {
				expectNotifyCommand(t, publisher, "d1", notifier.NotificationTypeDeliveryArrived, tt.publishErr)
			}

			err := NewProcessCreateDelivery(deliveries, residents, publisher, internalTopic).Handle(context.Background(), &tt.command)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantPermanent, pkgEvents.IsPermanent(err))
		})
	}
}

func TestProcessCreateDelivery_RetryPublishesSameNotifyCommand(t *testing.T) {
	residents := repomocks.NewResidentRepositoryPort(t)
	deliveries := repomocks.NewDeliveryRepositoryPort(t)
	publisher := pubsubmocks.NewMessagePublisher[any](t)
	ana := &entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypePrimary}

	residents.EXPECT().FindByApartment(mock.Anything, "101").Return([]*entities.Resident{ana}, nil).Twice()
	residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(nil).Twice()
	deliveries.EXPECT().Insert(mock.Anything, mock.AnythingOfType("*entities.Delivery")).Return(nil).Twice()
	first := expectNotifyCommand(t, publisher, "d1", notifier.NotificationTypeDeliveryArrived, nil)
	second := expectNotifyCommand(t, publisher, "d1", notifier.NotificationTypeDeliveryArrived, nil)

	writer := NewProcessCreateDelivery(deliveries, residents, publisher, internalTopic)
	command := &commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101"}
	for i := 0; i < 2; i++ {
		require.NoError(t, writer.Handle(context.Background(), command))
	}

	require.NotEmpty(t, first.CommandID)
	assert.Equal(t, first.CommandID, second.CommandID)
}

func expectNotifyCommand(
	t *testing.T,
	publisher *pubsubmocks.MessagePublisher[any],
	deliveryID string,
	notificationType notifier.NotificationType,
	publishErr error,
) *commands.ProcessNotifyDeliveryCommand {
	t.Helper()

	published := &commands.ProcessNotifyDeliveryCommand{}
	publisher.EXPECT().Publish(mock.Anything, internalTopic, mock.Anything).
		Run(func(_ context.Context, _ string, messages ...*pubsub.Message[any]) {
			require.Len(t, messages, 1)
			message := messages[0]
			assert.Equal(t, commands.ProcessNotifyDeliveryCommandType, message.Headers.EventType)
			assert.Equal(t, deliveryID, message.Headers.Key)

			command, ok := message.Payload.Data.(*commands.ProcessNotifyDeliveryCommand)
			require.True(t, ok, "payload = %T, want *commands.ProcessNotifyDeliveryCommand", message.Payload.Data)
			assert.Equal(t, deliveryID, command.DeliveryID)
			assert.Equal(t, notificationType, command.NotificationType)
			assert.NotEmpty(t, command.CommandID)
			*published = *command
		}).
		Return(publishErr).Once()
	return published
}
