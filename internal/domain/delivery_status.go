package domain

type DeliveryStatus string

const (
	DeliveryStatusPending DeliveryStatus = "pending"
	DeliveryStatusDeleted DeliveryStatus = "deleted"
)

func (s DeliveryStatus) IsValid() bool {
	switch s {
	case DeliveryStatusPending, DeliveryStatusDeleted:
		return true
	}
	return false
}
