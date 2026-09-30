package events

const (
	DeleteResidentEventType = "DeleteResident"
)

type DeleteResident struct {
	ResidentID string `json:"resident_id"`
}
