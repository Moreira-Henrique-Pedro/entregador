package commands

const (
	ProcessDeleteResidentCommandType = "ProcessDeleteResident"
)

type ProcessDeleteResidentCommand struct {
	CommandID  string `json:"command_id"`
	ResidentID string `json:"resident_id"`
}
