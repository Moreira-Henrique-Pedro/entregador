package events

const (
	UpdateResidentEventType = "UpdateResident"
)

type UpdateResident struct {
	ResidentID string `json:"resident_id"`
	Name       string `json:"name"`
	Apartment  string `json:"apartment"`
	Phone      string `json:"phone"`
}
