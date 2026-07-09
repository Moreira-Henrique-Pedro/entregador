package events

const (
	CreateResidentEventType = "CreateResident"
)

type CreateResident struct {
	Name      string `json:"name"`
	Apartment string `json:"apartment"`
	Phone     string `json:"phone"`
}
