package producer

import (
	"context"
	"strconv"

	ckafka "github.com/emorenkov/scorehub/pkg/common/kafka"
	"github.com/emorenkov/scorehub/pkg/notification"
)

type Publisher interface {
	Publish(ctx context.Context, msg *notification.NotificationMessage) error
	Close() error
}

type KafkaPublisher struct {
	writer *ckafka.Producer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: ckafka.NewProducerWithBrokers(brokers, topic),
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, msg *notification.NotificationMessage) error {
	return p.writer.SendMessage(ctx, keyForUser(msg.UserID), msg)
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}

func keyForUser(userID int64) string {
	return "user:" + strconv.FormatInt(userID, 10)
}
