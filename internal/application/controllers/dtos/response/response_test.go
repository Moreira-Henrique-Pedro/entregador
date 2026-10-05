package response

import (
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestNewDeliveries(t *testing.T) {
	deletedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	deliveries := []*entities.Delivery{
		{DeliveryID: "d1", Apartment: "101", Status: entities.DeliveryStatusPending},
		{DeliveryID: "d2", Apartment: "101", Status: entities.DeliveryStatusDeleted, DeleteAt: deletedAt},
	}

	got := NewDeliveries(deliveries)

	assert.Equal(t, []Delivery{
		{DeliveryID: "d1", Apartment: "101", Status: "pending"},
		{DeliveryID: "d2", Apartment: "101", Status: "deleted", DeletedAt: &deletedAt},
	}, got)
	assert.NotNil(t, NewDeliveries(nil), "an empty list must encode as [] and not null")
}

func TestNewResidents(t *testing.T) {
	residents := []*entities.Resident{{
		ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
		Type: entities.ResidentTypePrimary, Status: entities.ResidentStatusCreated,
	}}

	assert.Equal(t, []Resident{{
		ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
		Type: "resident-primary", Status: "created",
	}}, NewResidents(residents))
	assert.NotNil(t, NewResidents(nil), "an empty list must encode as [] and not null")
}
