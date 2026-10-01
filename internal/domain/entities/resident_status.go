package entities

type ResidentStatus string

const (
	// ResidentStatusCreated is an active resident, the status every resident starts with.
	ResidentStatusCreated ResidentStatus = "created"
	// ResidentStatusDeleted is a resident removed from the apartment (soft delete).
	ResidentStatusDeleted ResidentStatus = "deleted"
)

func (s ResidentStatus) IsValid() bool {
	switch s {
	case ResidentStatusCreated, ResidentStatusDeleted:
		return true
	}
	return false
}
