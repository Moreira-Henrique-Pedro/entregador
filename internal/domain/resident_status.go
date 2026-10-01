package domain

type ResidentStatus string

const (
	ResidentStatusCreated ResidentStatus = "created"
	ResidentStatusDeleted ResidentStatus = "deleted"
)

func (s ResidentStatus) IsValid() bool {
	switch s {
	case ResidentStatusCreated, ResidentStatusDeleted:
		return true
	}
	return false
}
