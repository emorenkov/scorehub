package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

// NewConsumerWithBrokers constructs a consumer with explicit brokers, topic, and groupID.
func NewConsumerWithBrokers(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
	}
}

// Deprecated: NewConsumer required common/config. Prefer NewConsumerWithBrokers.
// Kept for backward compatibility until all call sites are migrated.
// func NewConsumer(cfg *config.Config, topic, groupID string) *Consumer {
// 	return NewConsumerWithBrokers(cfg.KafkaBrokers, topic, groupID)
// }

func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
