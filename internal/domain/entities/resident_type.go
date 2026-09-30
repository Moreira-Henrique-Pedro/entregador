package entities

type ResidentType string

const (
	// ResidentTypeResident is a registered resident of the apartment.
	ResidentTypeResident ResidentType = "resident"
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
	}
}
