package commands

const (
	ProcessDeleteDeliveryCommandType = "ProcessDeleteDelivery"
)

type ProcessDeleteDeliveryCommand struct {
	CommandID  string `json:"command_id"`
	DeliveryID string `json:"delivery_id"`
}
