package request

import (
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDelivery_FromDTO(t *testing.T) {
	dto := RegisterDelivery{Apartment: "101", ResidentID: "ana", PackageType: "caixa", Urgency: "alta"}

	assert.Equal(t, &entities.Delivery{Apartment: "101", ResidentID: "ana", PackageType: "caixa", Urgency: "alta"}, dto.FromDTO())
}

func TestListDeliveries_FromDTO(t *testing.T) {
	apartment, status := (&ListDeliveries{Apartment: "101"}).FromDTO()
	assert.Equal(t, "101", apartment)
	assert.Nil(t, status)

	_, status = (&ListDeliveries{Apartment: "101", Status: "pending"}).FromDTO()
	require.NotNil(t, status)
	assert.Equal(t, entities.DeliveryStatusPending, *status)
}

func TestCreateResident_FromDTO(t *testing.T) {
	dto := CreateResident{Name: "Ana", Apartment: "101", Phone: "11999999999"}

	assert.Equal(t, &entities.Resident{Name: "Ana", Apartment: "101", Phone: "11999999999"}, dto.FromDTO())
}

func TestUpdateResident_FromDTO(t *testing.T) {
	dto := UpdateResident{Name: "Ana"}

	assert.Equal(t, &entities.Resident{ResidentID: "r1", Name: "Ana"}, dto.FromDTO("r1"))
}

func TestListResidents_ByApartment(t *testing.T) {
	assert.True(t, (&ListResidents{Apartment: "101"}).ByApartment())
	assert.False(t, (&ListResidents{Phone: "1"}).ByApartment())
}

func TestPubSubPush(t *testing.T) {
	push := PubSubPush{Message: PubSubMessage{
		Data:       []byte(`{"delivery_id":"d1"}`),
		Attributes: map[string]string{commands.AttributeEventType: "NotifyDelivery", commands.AttributeKey: "d1"},
	}}

	var command commands.NotifyDeliveryCommand
	require.NoError(t, push.DecodeData(&command))

	assert.Equal(t, "NotifyDelivery", push.EventType())
	assert.Equal(t, "d1", push.Key())
	assert.Equal(t, "d1", command.DeliveryID)
}
