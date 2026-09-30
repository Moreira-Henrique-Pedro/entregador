package events

const (
	CreateDeliveryEventType = "CreateDelivery"
)

type CreateDelivery struct {
	Apartment string `json:"apartment"`
	// ResidentID is optional; when empty or unknown the delivery goes to the apartment's "other" resident.
	ResidentID  string `json:"resident_id"`
	PackageType string `json:"package_type"`
	Urgency     string `json:"urgency"`
}
