package commands

const (
	ProcessUpdateResidentCommandType = "ProcessUpdateResident"
)

type ProcessUpdateResidentCommand struct {
	CommandID  string `json:"command_id"`
	ResidentID string `json:"resident_id"`
	Name       string `json:"name"`
	Apartment  string `json:"apartment"`
	Phone      string `json:"phone"`
}
