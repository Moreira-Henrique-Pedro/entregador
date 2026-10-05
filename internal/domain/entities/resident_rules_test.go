package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResident(t *testing.T) {
	assert.Equal(t, &Resident{
		ID: "r1", ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "1",
		Type: ResidentTypeSecondary, Status: ResidentStatusCreated,
	}, NewResident("r1", "Ana", "101", "1"))
}

func TestResidentCanBeNotified(t *testing.T) {
	assert.True(t, (&Resident{Phone: "1", Type: ResidentTypePrimary}).CanBeNotified())
	assert.False(t, (&Resident{Type: ResidentTypePrimary}).CanBeNotified())
	assert.False(t, (&Resident{Phone: "1", Type: ResidentTypeOther}).CanBeNotified())
}

func TestResidentMovesTo(t *testing.T) {
	resident := &Resident{Apartment: "101"}

	assert.True(t, resident.MovesTo("202"))
	assert.False(t, resident.MovesTo("101"))
	assert.False(t, resident.MovesTo(""))
	assert.True(t, resident.LivesIn("101"))
}

func TestResidentEnsureEditable(t *testing.T) {
	assert.NoError(t, (&Resident{Type: ResidentTypePrimary}).EnsureEditable())
	assert.ErrorIs(t, NewOtherResident("101").EnsureEditable(), ErrOtherResidentReadOnly)
}

func TestResidentValidateForCreate(t *testing.T) {
	assert.NoError(t, (&Resident{Name: "Ana", Apartment: "101", Phone: "1"}).ValidateForCreate())
	assert.EqualError(t, (&Resident{Apartment: "101", Phone: "1"}).ValidateForCreate(), "invalid resident: name is required")
	assert.EqualError(t, (&Resident{Name: "Ana", Phone: "1"}).ValidateForCreate(), "invalid resident: apartment is required")
	assert.EqualError(t, (&Resident{Name: "Ana", Apartment: "101"}).ValidateForCreate(), "invalid resident: phone is required")
}

func TestResidentValidateForUpdate(t *testing.T) {
	assert.NoError(t, (&Resident{ResidentID: "r1", Phone: "1"}).ValidateForUpdate())
	assert.EqualError(t, (&Resident{Name: "Ana"}).ValidateForUpdate(), "invalid resident: resident_id is required")
	assert.ErrorIs(t, (&Resident{ResidentID: "r1"}).ValidateForUpdate(), ErrInvalidResident)
}

func TestResidentFieldValidations(t *testing.T) {
	assert.NoError(t, ValidateResidentID("r1"))
	assert.NoError(t, ValidateApartment("101"))
	assert.NoError(t, ValidatePhone("1"))
	assert.ErrorIs(t, ValidateResidentID(""), ErrInvalidResident)
	assert.ErrorIs(t, ValidateApartment(""), ErrInvalidResident)
	assert.ErrorIs(t, ValidatePhone(""), ErrInvalidResident)
}

func TestHasResidentToReceive(t *testing.T) {
	ana := &Resident{ResidentID: "ana", Type: ResidentTypePrimary}

	assert.True(t, HasResidentToReceive([]*Resident{NewOtherResident("101"), ana}))
	assert.False(t, HasResidentToReceive([]*Resident{NewOtherResident("101")}))
	assert.False(t, HasResidentToReceive(nil))
}

func TestFindPrimary(t *testing.T) {
	ana := &Resident{ResidentID: "ana", Type: ResidentTypePrimary}
	bia := &Resident{ResidentID: "bia", Type: ResidentTypeSecondary}

	assert.Same(t, ana, FindPrimary([]*Resident{bia, ana}))
	assert.Nil(t, FindPrimary([]*Resident{bia}))
}
