package entities

type ResidentType string

const (
	// ResidentTypePrimary is the apartment's main resident: the first one registered.
	// Deliveries addressed to the "other" resident are notified to it.
	ResidentTypePrimary ResidentType = "resident-primary"
	// ResidentTypeSecondary is any other registered resident of the apartment.
	ResidentTypeSecondary ResidentType = "resident-secondary"
	// ResidentTypeOther is the placeholder every apartment has, used when the
	// recipient is not one of its residents or has not been registered yet.
	ResidentTypeOther ResidentType = "other"
)

const (
	otherResidentIDPrefix = "other-"
	otherResidentName     = "Outro"
)

// OtherResidentID is deterministic so that each apartment has exactly one "other" resident.
func OtherResidentID(apartment string) string {
	return otherResidentIDPrefix + apartment
}

func NewOtherResident(apartment string) *Resident {
	id := OtherResidentID(apartment)
	return &Resident{
		ID:         id,
		ResidentID: id,
		Apartment:  apartment,
		Name:       otherResidentName,
		Type:       ResidentTypeOther,
		Status:     ResidentStatusCreated,
	}
}
