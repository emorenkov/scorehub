package email

import (
	"context"

	ckafka "github.com/emorenkov/scorehub/pkg/common/kafka"
)

type Consumer struct {
	c *ckafka.Consumer
}

// NewConsumerWithBrokers creates an email consumer for the notifications topic.
func NewConsumerWithBrokers(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{c: ckafka.NewConsumerWithBrokers(brokers, topic, groupID)}
}

func (c *Consumer) ReadMessage(ctx context.Context) (key []byte, value []byte, err error) {
	msg, err := c.c.ReadMessage(ctx)
	if err != nil {
		return nil, nil, err
	}
	return msg.Key, msg.Value, nil
}

func (c *Consumer) Close() error { return c.c.Close() }
