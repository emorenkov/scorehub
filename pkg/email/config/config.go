package config

import (
	"os"
	"strings"
)

type Config struct {
	ServiceName        string
	HTTPPort           string
	APIKey             string
	KafkaBrokers       []string
	KafkaGroupID       string
	NotificationsTopic string
}

func Load() *Config {
	return &Config{
		ServiceName:        getEnv("SERVICE_NAME", "email-service"),
		HTTPPort:           getEnv("HTTP_PORT", "8084"),
		APIKey:             getEnv("API_KEY", ""),
		KafkaBrokers:       splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaGroupID:       getEnv("KAFKA_GROUP_ID", "scorehub-group"),
		NotificationsTopic: getEnv("NOTIFICATIONS_TOPIC", "notifications"),
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
