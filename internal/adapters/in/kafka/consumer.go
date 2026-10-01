package kafka

import (
	"context"
	"fmt"
	"time"

	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
	appLogger "github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	appWatermill "github.com/Moreira-Henrique-Pedro/entregador/pkg/watermill"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
)

type ConsumerConfig struct {
	Topic    string
	DLQTopic string
	Timeout  time.Duration
	Retry    RetryPolicy
}

// Consumer reads a Kafka topic and dispatches each message, by its EventType header, to the
// handler registered for it. Failed messages are retried and then sent to the DLQ.
type Consumer struct {
	subscriber   watermillMessage.Subscriber
	dlqPublisher pubsub.MessagePublisher[any]
	eventBus     *pkgEvents.EventBus
	config       ConsumerConfig
	logger       appLogger.Logger
}

func NewConsumer(
	subscriber watermillMessage.Subscriber,
	dlqPublisher pubsub.MessagePublisher[any],
	registry *pkgEvents.EventHandlerRegistry,
	config ConsumerConfig,
	logger appLogger.Logger,
) *Consumer {
	return &Consumer{
		subscriber:   subscriber,
		dlqPublisher: dlqPublisher,
		eventBus:     pkgEvents.NewEventBus(pkgEvents.EventBusDependencies{EventHandlerRegistry: registry}),
		config:       config,
		logger:       logger,
	}
}

// Run blocks until ctx is cancelled or the subscription is closed.
func (c *Consumer) Run(ctx context.Context) error {
	messages, err := c.subscriber.Subscribe(ctx, c.config.Topic)
	if err != nil {
		return fmt.Errorf("subscribe to topic %s: %w", c.config.Topic, err)
	}

	c.logger.Info("Kafka consumer started", "topic", c.config.Topic)

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-messages:
			if !ok {
				return nil
			}

			c.handleMessage(ctx, msg)
		}
	}
}

func (c *Consumer) Close() error {
	return c.subscriber.Close()
}

func (c *Consumer) handleMessage(ctx context.Context, msg *watermillMessage.Message) {
	err := c.config.Retry.run(ctx, func() error {
		return c.processMessage(ctx, msg)
	}, func(attempt int, err error, wait time.Duration) {
		c.logger.Warn("Retrying Kafka message",
			"error", err.Error(),
			"message_uuid", msg.UUID,
			"attempt", attempt,
			"wait", wait.String(),
		)
	})
	if err == nil {
		msg.Ack()
		return
	}

	if ctx.Err() != nil {
		msg.Nack()
		return
	}

	c.logger.Error("Failed to process Kafka message, sending to DLQ",
		"error", err.Error(),
		"message_uuid", msg.UUID,
		"permanent", pkgEvents.IsPermanent(err),
		"dlq_topic", c.config.DLQTopic,
	)

	if dlqErr := c.publishToDLQ(ctx, msg, err); dlqErr != nil {
		c.logger.Error("Failed to publish message to DLQ",
			"error", dlqErr.Error(),
			"message_uuid", msg.UUID,
		)
		msg.Nack()
		return
	}

	msg.Ack()
}

func (c *Consumer) processMessage(ctx context.Context, kafkaMessage *watermillMessage.Message) error {
	messageCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	pubsubMessage, err := appWatermill.ConvertWatermillToPubsub(kafkaMessage, nil)
	if err != nil {
		return pkgEvents.Permanent(fmt.Errorf("convert kafka message: %w", err))
	}

	messageLogger := c.logger.With(
		"message_uuid", kafkaMessage.UUID,
		"event_type", pubsubMessage.Headers.EventType,
		"message_key", pubsubMessage.Headers.Key,
		"topic", c.config.Topic,
	)
	messageCtx = messageLogger.AddToContext(messageCtx, messageLogger)

	messageLogger.Info("Processing Kafka message")

	if err := c.eventBus.Handle(messageCtx, pubsubMessage); err != nil {
		return fmt.Errorf("handle event bus message: %w", err)
	}

	messageLogger.Info("Kafka message processed successfully")
	return nil
}

func (c *Consumer) publishToDLQ(ctx context.Context, msg *watermillMessage.Message, processErr error) error {
	dlqMessage := appWatermill.BuildRawDLQMessage(msg, processErr)
	originalTopic := c.config.Topic
	dlqMessage.Headers.OriginalTopic = &originalTopic

	return c.dlqPublisher.Publish(ctx, c.config.DLQTopic, dlqMessage)
}
