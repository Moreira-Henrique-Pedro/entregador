package watermill

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	appLogger "github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

type WatermillPublisher[T any] struct {
	publisher message.Publisher
}

func NewWatermillPublisher[T any](brokers []string, logger appLogger.Logger) (pubsub.MessagePublisher[T], error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: NewWatermillMarshaler(),
		},
		NewWatermillLoggerFromLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	return &WatermillPublisher[T]{
		publisher: publisher,
	}, nil
}

func (w *WatermillPublisher[T]) Publish(ctx context.Context, topic string, pubsubMessages ...*pubsub.Message[T]) error {
	logger := appLogger.GetLoggerFromContext(ctx)

	waterMillMessages := make([]*message.Message, 0, len(pubsubMessages))
	for _, pubsubMessage := range pubsubMessages {
		waterMillMessage, err := ConvertPubsubToWatermill(pubsubMessage, logger)
		if err != nil {
			logger.Error("Failed to convert pubsub message to watermill message",
				"error", err.Error(),
				"topic", topic,
				"key", pubsubMessage.Headers.Key,
			)
			return err
		}
		waterMillMessages = append(waterMillMessages, waterMillMessage)
	}

	err := w.publisher.Publish(topic, waterMillMessages...)
	eventTypeKeyArr := getEventTypeKeyArray(waterMillMessages)

	if err != nil {
		logger.Error("Failed to publish messages to Kafka",
			"error", err.Error(),
			"topic", topic,
			"events", eventTypeKeyArr,
		)
		return err
	}

	logger.Debug("Messages published successfully to Kafka",
		"topic", topic,
		"events", eventTypeKeyArr,
	)

	return nil
}

func getEventTypeKeyArray(messages []*message.Message) []map[string]string {
	arr := make([]map[string]string, 0, len(messages))
	for _, msg := range messages {
		eventType := msg.Metadata.Get(pubsub.EventTypeHeader)
		key := msg.Metadata.Get(pubsub.KeyHeader)
		arr = append(arr, map[string]string{
			"eventType": eventType,
			"key":       key,
		})
	}
	return arr
}

func (w *WatermillPublisher[T]) Close(ctx context.Context) error {
	return w.publisher.Close()
}
