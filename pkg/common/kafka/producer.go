package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

// NewProducerWithBrokers constructs a producer with explicit broker list and topic.
func NewProducerWithBrokers(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Deprecated: NewProducer requires common/config. Prefer NewProducerWithBrokers.
// Kept for backward compatibility until all call sites are migrated.
// func NewProducer(cfg *config.Config, topic string) *Producer {
// 	return NewProducerWithBrokers(cfg.KafkaBrokers, topic)
// }

func (p *Producer) SendMessage(ctx context.Context, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
