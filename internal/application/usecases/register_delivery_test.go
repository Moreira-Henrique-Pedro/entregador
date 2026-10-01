package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	outmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

func TestRegisterDelivery(t *testing.T) {
	failure := errors.New("mongo down")
	otherID := domain.OtherResidentID("101")
	ana := &domain.Resident{ResidentID: "ana", Apartment: "101", Type: domain.ResidentTypePrimary}
	bob := &domain.Resident{ResidentID: "bob", Apartment: "202", Type: domain.ResidentTypePrimary}
	apartment101 := []*domain.Resident{ana, domain.NewOtherResident("101")}

	tests := []struct {
		name         string
		input        in.RegisterDeliveryInput
		setup        func(residents *outmocks.ResidentRepository)
		wantResident string
		insertErr    error
		schedule     bool
		scheduleErr  error
		wantErr      error
	}{
		{
			name:  "resident of the apartment receives the delivery",
			input: in.RegisterDeliveryInput{Apartment: "101", ResidentID: "ana", PackageType: "caixa", Urgency: "alta"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
			},
			wantResident: "ana",
			schedule:     true,
		},
		{
			name:  "no resident informed goes to the apartment other",
			input: in.RegisterDeliveryInput{Apartment: "101"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(nil).Once()
			},
			wantResident: otherID,
			schedule:     true,
		},
		{
			name:  "unknown resident goes to the apartment other",
			input: in.RegisterDeliveryInput{Apartment: "101", ResidentID: "ghost"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ghost").Return(nil, domain.ErrEntityNotFound).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(nil).Once()
			},
			wantResident: otherID,
			schedule:     true,
		},
		{
			name:  "resident from another apartment goes to the apartment other",
			input: in.RegisterDeliveryInput{Apartment: "101", ResidentID: "bob"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "bob").Return(bob, nil).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(nil).Once()
			},
			wantResident: otherID,
			schedule:     true,
		},
		{
			name:    "missing apartment is invalid",
			input:   in.RegisterDeliveryInput{ResidentID: "ana"},
			wantErr: domain.ErrInvalidDelivery,
		},
		{
			name:  "apartment without residents is rejected",
			input: in.RegisterDeliveryInput{Apartment: "303"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "303").Return(nil, nil).Once()
			},
			wantErr: domain.ErrNoResidentInApartment,
		},
		{
			name:  "apartment with only the other resident is rejected",
			input: in.RegisterDeliveryInput{Apartment: "404"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "404").
					Return([]*domain.Resident{domain.NewOtherResident("404")}, nil).Once()
			},
			wantErr: domain.ErrNoResidentInApartment,
		},
		{
			name:  "apartment residents lookup error is returned",
			input: in.RegisterDeliveryInput{Apartment: "101", ResidentID: "ana"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(nil, failure).Once()
			},
			wantErr: failure,
		},
		{
			name:  "resident lookup error is returned",
			input: in.RegisterDeliveryInput{Apartment: "101", ResidentID: "ana"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(nil, failure).Once()
			},
			wantErr: failure,
		},
		{
			name:  "ensure other error is returned",
			input: in.RegisterDeliveryInput{Apartment: "101"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(failure).Once()
			},
			wantErr: failure,
		},
		{
			name:  "insert error is returned without scheduling",
			input: in.RegisterDeliveryInput{Apartment: "101", ResidentID: "ana"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
			},
			wantResident: "ana",
			insertErr:    failure,
			wantErr:      failure,
		},
		{
			name:  "schedule error does not fail the already saved delivery",
			input: in.RegisterDeliveryInput{Apartment: "101", ResidentID: "ana"},
			setup: func(residents *outmocks.ResidentRepository) {
				residents.EXPECT().FindByApartment(mock.Anything, "101").Return(apartment101, nil).Once()
				residents.EXPECT().FindByResidentID(mock.Anything, "ana").Return(ana, nil).Once()
			},
			wantResident: "ana",
			schedule:     true,
			scheduleErr:  errors.New("kafka down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := outmocks.NewResidentRepository(t)
			deliveries := outmocks.NewDeliveryRepository(t)
			scheduler := outmocks.NewNotificationScheduler(t)
			if tt.setup != nil {
				tt.setup(residents)
			}
			want := &domain.Delivery{
				ID:          "d1",
				DeliveryID:  "d1",
				Apartment:   tt.input.Apartment,
				ResidentID:  tt.wantResident,
				PackageType: tt.input.PackageType,
				Urgency:     tt.input.Urgency,
				Status:      domain.DeliveryStatusPending,
			}
			if tt.wantResident != "" {
				deliveries.EXPECT().Insert(mock.Anything, want).Return(tt.insertErr).Once()
			}
			if tt.schedule {
				scheduler.EXPECT().Schedule(mock.Anything, "d1", domain.NotificationTypeDeliveryArrived).Return(tt.scheduleErr).Once()
			}

			useCase := NewRegisterDelivery(deliveries, residents, scheduler)
			useCase.newID = func() string { return "d1" }

			got, err := useCase.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}
