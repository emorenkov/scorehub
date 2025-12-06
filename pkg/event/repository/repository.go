package repository

import (
	"context"
	"strconv"

	ckafka "github.com/emorenkov/scorehub/pkg/common/kafka"
	"github.com/emorenkov/scorehub/pkg/event"
)

type Publisher interface {
	PublishScoreEvent(ctx context.Context, ev *event.ScoreEvent) error
	Close() error
}

type KafkaPublisher struct {
	producer *ckafka.Producer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		producer: ckafka.NewProducerWithBrokers(brokers, topic),
	}
}

func (p *KafkaPublisher) PublishScoreEvent(ctx context.Context, ev *event.ScoreEvent) error {
	return p.producer.SendMessage(ctx, keyForUser(ev.UserID), ev)
}

func (p *KafkaPublisher) Close() error {
	return p.producer.Close()
}

func keyForUser(userID int64) string {
	return "user:" + strconv.FormatInt(userID, 10)
}
