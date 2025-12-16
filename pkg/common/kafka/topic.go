package kafka

import (
	"context"
	"errors"
	"strings"

	"github.com/segmentio/kafka-go"
)

// EnsureTopic creates a topic if it does not already exist.
func EnsureTopic(ctx context.Context, brokers []string, topic string, partitions, replicationFactor int) error {
	if len(brokers) == 0 {
		return errors.New("no kafka brokers provided")
	}
	if partitions <= 0 {
		partitions = 1
	}
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	})
	if err != nil {
		// Kafka returns an error if the topic already exists; treat that as success.
		if strings.Contains(err.Error(), "Topic with this name already exists") {
			return nil
		}
	}
	return err
}
