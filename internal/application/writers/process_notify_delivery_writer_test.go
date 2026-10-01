package writers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
)

var (
	other101  = entities.NewOtherResident("101")
	arrived   = notifier.NotificationTypeDeliveryArrived
	pickedUp  = notifier.NotificationTypeDeliveryPickedUp
	errFailed = errors.New("mongo down")
)

// notifyResidents returns fresh residents for each case, since EnsurePrimaryResident
// changes their type. Apartment 101: ana is primary, bia and caio (no phone) secondary.
func notifyResidents() []*entities.Resident {
	return []*entities.Resident{
		{ResidentID: "ana", Name: "Ana", Apartment: "101", Phone: "111", Type: entities.ResidentTypePrimary},
		{ResidentID: "bia", Name: "Bia", Apartment: "101", Phone: "222", Type: entities.ResidentTypeSecondary},
		{ResidentID: "caio", Name: "Caio", Apartment: "101", Type: entities.ResidentTypeSecondary},
		{ResidentID: "davi", Name: "Davi", Apartment: "202", Phone: "333", Type: entities.ResidentTypePrimary},
		entities.NewOtherResident("101"),
	}
}

func TestProcessNotifyDelivery(t *testing.T) {
	now := time.Now()
	pendingFor := func(residentID string) *entities.Delivery {
		return &entities.Delivery{DeliveryID: "d1", Apartment: "101", ResidentID: residentID, PackageType: "caixa", Status: entities.DeliveryStatusPending}
	}
	deletedFor := func(residentID string) *entities.Delivery {
		delivery := pendingFor(residentID)
		delivery.Status = entities.DeliveryStatusDeleted
		return delivery
	}

	tests := []struct {
		name             string
		notificationType notifier.NotificationType
		delivery         *entities.Delivery
		residents        []*entities.Resident // defaults to notifyResidents()
		residentFindErr  error
		apartmentErr     error
		ensurePrimaryErr error
		notifierErrs     map[string]error
		markErr          error
		wantErr          error
		wantPhones       []string
		wantMarked       string // "arrival", "pickup" or "" for not marked
	}{
		{
			name:             "arrival notifies the delivery resident",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			wantPhones:       []string{"111"},
			wantMarked:       "arrival",
		},
		{
			name:             "pickup notifies the delivery resident",
			notificationType: pickedUp,
			delivery:         deletedFor("ana"),
			wantPhones:       []string{"111"},
			wantMarked:       "pickup",
		},
		{
			name:             "other delivery notifies only the apartment primary",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			wantPhones:       []string{"111"},
			wantMarked:       "arrival",
		},
		{
			name:             "other pickup notifies only the apartment primary",
			notificationType: pickedUp,
			delivery:         deletedFor(other101.ResidentID),
			wantPhones:       []string{"111"},
			wantMarked:       "pickup",
		},
		{
			name:             "deleted delivery resident falls back to the apartment primary",
			notificationType: arrived,
			delivery:         pendingFor("gone"),
			wantPhones:       []string{"111"},
			wantMarked:       "arrival",
		},
		{
			name:             "apartment without primary promotes the oldest resident and notifies it",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			residents: []*entities.Resident{
				{ResidentID: "ana", Apartment: "101", Phone: "111", Type: entities.ResidentTypeSecondary, CreatedAt: now},
				{ResidentID: "bia", Apartment: "101", Phone: "222", Type: entities.ResidentTypeSecondary, CreatedAt: now.Add(-time.Hour)},
				entities.NewOtherResident("101"),
			},
			wantPhones: []string{"222"},
			wantMarked: "arrival",
		},
		{
			name:             "primary without phone leaves the other delivery without recipients",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			residents: []*entities.Resident{
				{ResidentID: "caio", Apartment: "101", Type: entities.ResidentTypePrimary},
				{ResidentID: "bia", Apartment: "101", Phone: "222", Type: entities.ResidentTypeSecondary},
				entities.NewOtherResident("101"),
			},
			wantMarked: "arrival",
		},
		{
			name:             "apartment with only the other resident has no recipients",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			residents:        []*entities.Resident{entities.NewOtherResident("101")},
			wantMarked:       "arrival",
		},
		{
			name:             "ensure primary error is returned",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			residents: []*entities.Resident{
				{ResidentID: "bia", Apartment: "101", Phone: "222", Type: entities.ResidentTypeSecondary},
				entities.NewOtherResident("101"),
			},
			ensurePrimaryErr: errFailed,
			wantErr:          errFailed,
		},
		{
			name:             "resident without phone is marked as notified without sending",
			notificationType: arrived,
			delivery:         pendingFor("caio"),
			wantMarked:       "arrival",
		},
		{
			name:             "invalid recipient is skipped and the delivery is marked as notified",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			notifierErrs:     map[string]error{"111": fmt.Errorf("twilio: %w", notifier.ErrInvalidRecipient)},
			wantMarked:       "arrival",
		},
		{
			name:             "provider failure is returned and not marked so it is retried",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			notifierErrs:     map[string]error{"111": errFailed},
			wantErr:          errFailed,
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
			delivery:         pendingFor("ana"),
		},
		{
			name:             "unknown delivery is ignored",
			notificationType: arrived,
		},
		{
			name:             "resident lookup error is returned",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			residentFindErr:  errFailed,
			wantErr:          errFailed,
		},
		{
			name:             "apartment lookup error is returned",
			notificationType: arrived,
			delivery:         pendingFor(other101.ResidentID),
			apartmentErr:     errFailed,
			wantErr:          errFailed,
		},
		{
			name:             "mark as notified error is returned",
			notificationType: arrived,
			delivery:         pendingFor("ana"),
			markErr:          errFailed,
			wantErr:          errFailed,
			wantPhones:       []string{"111"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveries := newFakeDeliveryRepository()
			if tt.delivery != nil {
				deliveries = newFakeDeliveryRepository(tt.delivery)
			}
			deliveries.markNotifiedErr = tt.markErr
			fixture := tt.residents
			if fixture == nil {
				fixture = notifyResidents()
			}
			residents := newFakeResidentRepository(fixture...)
			residents.findErr = tt.residentFindErr
			residents.findApartmentErr = tt.apartmentErr
			residents.ensurePrimaryErr = tt.ensurePrimaryErr
			sender := &fakeNotifier{errByPhone: tt.notifierErrs}

			err := NewProcessNotifyDelivery(deliveries, residents, sender).Handle(context.Background(), &commands.ProcessNotifyDeliveryCommand{
				DeliveryID:       "d1",
				NotificationType: tt.notificationType,
			})

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if phones := slices.Sorted(slices.Values(sender.phones())); !equalStrings(phones, tt.wantPhones) {
				t.Errorf("notified phones = %v, want %v", phones, tt.wantPhones)
			}
			for _, notification := range sender.sent {
				if notification.Type != tt.notificationType {
					t.Errorf("notification type = %q, want %q", notification.Type, tt.notificationType)
				}
			}

			gotMarked := ""
			switch {
			case len(deliveries.arrivalNotifiedIDs) == 1 && len(deliveries.pickupNotifiedIDs) == 0:
				gotMarked = "arrival"
			case len(deliveries.pickupNotifiedIDs) == 1 && len(deliveries.arrivalNotifiedIDs) == 0:
				gotMarked = "pickup"
			case len(deliveries.arrivalNotifiedIDs)+len(deliveries.pickupNotifiedIDs) > 0:
				gotMarked = "both"
			}
			if gotMarked != tt.wantMarked {
				t.Errorf("marked = %q, want %q", gotMarked, tt.wantMarked)
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

			if notification.Type != tt.notificationType || notification.Phone != "111" {
				t.Errorf("notification = %+v, want type %s to 111", notification, tt.notificationType)
			}
			if notification.Body != tt.wantBody {
				t.Errorf("body = %q, want %q", notification.Body, tt.wantBody)
			}
			if !equalStrings(notification.Variables, tt.wantVariables) {
				t.Errorf("variables = %v, want %v", notification.Variables, tt.wantVariables)
			}
		})
	}
}
