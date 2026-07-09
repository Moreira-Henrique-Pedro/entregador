package watermill

import "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"

func NewWatermillMarshaler() kafka.Marshaler {
	return kafka.DefaultMarshaler{}
}
