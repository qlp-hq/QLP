package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"QLP/internal/config"
	"QLP/internal/logger"
)

const (
	defaultTopic = "qlp-events"
)

// KafkaEventManager manages producing and consuming events from Kafka.
type KafkaEventManager struct {
	writer *kafka.Writer
	reader *kafka.Reader
	log    *zap.Logger
}

// NewKafkaEventManager creates a new manager for Kafka eventing.
// It requires KAFKA_BROKERS to be set in the environment.
func NewKafkaEventManager() (*KafkaEventManager, error) {
	brokers := config.GetKafkaBrokers()
	if len(brokers) == 0 {
		return nil, fmt.Errorf("KAFKA_BROKERS environment variable not set")
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    defaultTopic,
		Balancer: &kafka.LeastBytes{},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    defaultTopic,
		GroupID:  "qlp-orchestrator-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  2 * time.Second,
	})

	return &KafkaEventManager{
		writer: writer,
		reader: reader,
		log:    logger.Logger.With(zap.String("component", "kafka-event-manager")),
	}, nil
}

// Publish sends an event to the Kafka topic.
func (k *KafkaEventManager) Publish(ctx context.Context, event Event) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		k.log.Error("Failed to marshal event for Kafka", zap.Error(err), zap.String("event_id", event.ID))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = k.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(event.ID),
			Value: eventBytes,
		},
	)
	if err != nil {
		k.log.Error("Failed to write message to Kafka", zap.Error(err))
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	k.log.Info("Published event to Kafka", zap.String("event_type", string(event.Type)), zap.String("event_id", event.ID))
	return nil
}

// Subscribe listens for events on the Kafka topic and calls the appropriate handler.
// This implementation is non-blocking and runs the listener in a goroutine.
func (k *KafkaEventManager) Subscribe(ctx context.Context, eventType EventType, handler EventHandler) error {
	k.log.Info("Subscribing to event type", zap.String("event_type", string(eventType)))
	go func() {
		for {
			select {
			case <-ctx.Done():
				k.log.Info("Subscription stopped due to context cancellation.", zap.String("event_type", string(eventType)))
				return
			default:
				msg, err := k.reader.FetchMessage(ctx)
				if err != nil {
					k.log.Warn("Could not fetch message from Kafka", zap.Error(err))
					continue
				}

				var event Event
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					k.log.Error("Failed to unmarshal event from Kafka", zap.Error(err))
					// We commit the message even if it's unmarshalable to avoid getting stuck on a bad message.
					k.reader.CommitMessages(ctx, msg)
					continue
				}

				// If the event is the one we are subscribed to, handle it.
				if event.Type == eventType {
					if err := handler(ctx, event); err != nil {
						k.log.Error("Handler failed for event", zap.Error(err), zap.String("event_id", event.ID))
						// Depending on the error, we might not want to commit the message, allowing for a retry.
						// For now, we'll log the error and move on.
					}
				}

				// Commit the message to mark it as processed.
				if err := k.reader.CommitMessages(ctx, msg); err != nil {
					k.log.Error("Failed to commit message", zap.Error(err))
				}
			}
		}
	}()
	return nil
}

// Close cleans up the Kafka connections.
func (k *KafkaEventManager) Close() error {
	var firstErr error
	if err := k.writer.Close(); err != nil {
		k.log.Error("Failed to close Kafka writer", zap.Error(err))
		firstErr = err
	}
	if err := k.reader.Close(); err != nil {
		k.log.Error("Failed to close Kafka reader", zap.Error(err))
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
