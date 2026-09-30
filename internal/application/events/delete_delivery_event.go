package events

const (
	DeleteDeliveryEventType = "DeleteDelivery"
)

type DeleteDelivery struct {
	DeliveryID string `json:"delivery_id"`
}
