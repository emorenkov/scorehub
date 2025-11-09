package config

import (
	"os"
	"strings"
)

type Config struct {
	ServiceName      string
	GRPCPort         string
	KafkaBrokers     []string
	KafkaGroupID     string
	ScoreEventsTopic string
}

func Load() *Config {
	return &Config{
		ServiceName:      getEnv("SERVICE_NAME", "event-service"),
		GRPCPort:         getEnv("GRPC_PORT", "50052"),
		KafkaBrokers:     splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaGroupID:     getEnv("KAFKA_GROUP_ID", "scorehub-group"),
		ScoreEventsTopic: getEnv("SCORE_EVENTS_TOPIC", "score_events"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			res = append(res, v)
		}
	}
	return res
}
