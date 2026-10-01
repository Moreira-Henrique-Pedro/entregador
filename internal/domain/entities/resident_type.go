package entities

type ResidentType string

const (
	ResidentTypePrimary   ResidentType = "resident-primary"
	ResidentTypeSecondary ResidentType = "resident-secondary"
	ResidentTypeOther     ResidentType = "other"
)

const (
	otherResidentIDPrefix = "other-"
	otherResidentName     = "Outro"
)

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
