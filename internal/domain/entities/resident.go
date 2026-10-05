package entities

import (
	"fmt"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/validation"
)

type Resident struct {
	ID         string       `bson:"_id"`
	ResidentID string       `bson:"resident_id"`
	Apartment  string       `bson:"apartment"`
	Name       string       `bson:"name"`
	Phone      string       `bson:"phone"`
	Type       ResidentType `bson:"type"`
	Status     ResidentStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeleteAt   time.Time
}

func NewResident(id, name, apartment, phone string) *Resident {
	return &Resident{
		ID:         id,
		ResidentID: id,
		Apartment:  apartment,
		Name:       name,
		Phone:      phone,
		Type:       ResidentTypeSecondary,
		Status:     ResidentStatusCreated,
	}
}

func (r *Resident) IsOther() bool {
	return r.Type == ResidentTypeOther
}

func (r *Resident) IsPrimary() bool {
	return r.Type == ResidentTypePrimary
}

func (r *Resident) CanBeNotified() bool {
	return !r.IsOther() && r.Phone != ""
}

func (r *Resident) MovesTo(apartment string) bool {
	return apartment != "" && apartment != r.Apartment
}

func (r *Resident) LivesIn(apartment string) bool {
	return r.Apartment == apartment
}

func (r *Resident) EnsureEditable() error {
	if r.IsOther() {
		return fmt.Errorf("resident %s: %w", r.ResidentID, ErrOtherResidentReadOnly)
	}
	return nil
}

func (r *Resident) ValidateForCreate() error {
	return validation.First(
		validation.Required(ErrInvalidResident, "name", r.Name),
		validation.Required(ErrInvalidResident, "apartment", r.Apartment),
		ValidatePhone(r.Phone),
	)
}

func (r *Resident) ValidateForUpdate() error {
	return validation.First(
		ValidateResidentID(r.ResidentID),
		r.validateHasChanges(),
	)
}

func (r *Resident) validateHasChanges() error {
	if r.Name == "" && r.Apartment == "" && r.Phone == "" {
		return fmt.Errorf("%w: at least one of name, apartment or phone is required", ErrInvalidResident)
	}
	return nil
}

func ValidateResidentID(residentID string) error {
	return validation.Required(ErrInvalidResident, "resident_id", residentID)
}

func ValidateApartment(apartment string) error {
	return validation.Required(ErrInvalidResident, "apartment", apartment)
}

func ValidatePhone(phone string) error {
	return validation.Required(ErrInvalidResident, "phone", phone)
}

func HasResidentToReceive(residents []*Resident) bool {
	for _, resident := range residents {
		if !resident.IsOther() {
			return true
		}
	}
	return false
}

func FindPrimary(residents []*Resident) *Resident {
	for _, resident := range residents {
		if resident.IsPrimary() {
			return resident
		}
	}
	return nil
}
