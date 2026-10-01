package events

const (
	CreateDeliveryEventType = "CreateDelivery"
)

type CreateDelivery struct {
	Apartment   string `json:"apartment"`
	ResidentID  string `json:"resident_id"`
	PackageType string `json:"package_type"`
	Urgency     string `json:"urgency"`
}
