package entities

type DeliveryStatus string

const (
	// DeliveryStatusPending is a delivery waiting at the front desk to be picked up.
	DeliveryStatusPending DeliveryStatus = "pending"
	// DeliveryStatusDeleted is a delivery already picked up by the resident.
	DeliveryStatusDeleted DeliveryStatus = "deleted"
)

func (s DeliveryStatus) IsValid() bool {
	switch s {
	case DeliveryStatusPending, DeliveryStatusDeleted:
		return true
	}
	return false
}
