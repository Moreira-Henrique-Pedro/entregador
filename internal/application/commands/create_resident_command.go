package commands

const (
	ProcessCreateResidentCommandType = "ProcessCreateResident"
)

type ProcessCreateResidentCommand struct {
	CommandID string `json:"command_id"`
	Name      string `json:"name"`
	Apartment string `json:"apartment"`
	Phone     string `json:"phone"`
}
