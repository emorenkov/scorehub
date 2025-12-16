package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ckafka.EnsureTopic(ctx, brokers, topic, 1, 1); err != nil {
		return nil, fmt.Errorf("ensure topic %s: %w", topic, err)
	}

	return &KafkaPublisher{
		producer: ckafka.NewProducerWithBrokers(brokers, topic),
	}, nil
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
