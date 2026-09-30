package commands

const (
	ProcessCreateDeliveryCommandType = "ProcessCreateDelivery"
)

type ProcessCreateDeliveryCommand struct {
	CommandID   string `json:"command_id"`
	Apartment   string `json:"apartment"`
	ResidentID  string `json:"resident_id"`
	PackageType string `json:"package_type"`
	Urgency     string `json:"urgency"`
}
