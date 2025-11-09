package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServiceName        string
	PostgresHost       string
	PostgresPort       int
	PostgresUser       string
	PostgresPassword   string
	PostgresDB         string
	KafkaBrokers       []string
	KafkaGroupID       string
	ScoreEventsTopic   string
	NotificationsTopic string
}

func Load() *Config {
	return &Config{
		ServiceName:        getEnv("SERVICE_NAME", "notification-service"),
		PostgresHost:       getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:       getEnvAsInt("POSTGRES_PORT", 5432),
		PostgresUser:       getEnv("POSTGRES_USER", "scorehub_user"),
		PostgresPassword:   getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:         getEnv("POSTGRES_DB", "scorehub"),
		KafkaBrokers:       splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaGroupID:       getEnv("KAFKA_GROUP_ID", "scorehub-group"),
		ScoreEventsTopic:   getEnv("SCORE_EVENTS_TOPIC", "score_events"),
		NotificationsTopic: getEnv("NOTIFICATIONS_TOPIC", "notifications"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
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
