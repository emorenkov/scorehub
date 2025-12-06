package config

import (
	"os"
	"strings"

	"github.com/emorenkov/scorehub/pkg/common/models"
)

type Config struct {
	ServiceName        string
	HTTPPort           string
	APIKey             string
	KafkaBrokers       []string
	KafkaGroupID       string
	ScoreEventsTopic   string
	NotificationsTopic string
	DbConfig           *models.PostgresConfig
}

func Load() *Config {
	return &Config{
		ServiceName:        getEnv("SERVICE_NAME", "notification-service"),
		HTTPPort:           getEnv("HTTP_PORT", "8083"),
		APIKey:             getEnv("API_KEY", ""),
		KafkaBrokers:       splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaGroupID:       getEnv("KAFKA_GROUP_ID", "scorehub-group"),
		ScoreEventsTopic:   getEnv("SCORE_EVENTS_TOPIC", "score_events"),
		NotificationsTopic: getEnv("NOTIFICATIONS_TOPIC", "notifications"),
		DbConfig:           models.LoadPostgresConfig(),
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
