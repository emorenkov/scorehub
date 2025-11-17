package config

import (
	"os"
	"strconv"
)

type Config struct {
	PostgresHost       string
	PostgresPort       int
	PostgresUser       string
	PostgresPassword   string
	PostgresDB         string
	KafkaBrokers       []string
	KafkaGroupID       string
	ScoreEventsTopic   string
	NotificationsTopic string
	GRPCPort           string
	HTTPPort           string
	GatewayPort        string
	ServiceName        string
	UserServiceAddr    string
}

func Load() *Config {
	return &Config{
		PostgresHost:       getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:       getEnvAsInt("POSTGRES_PORT", 5432),
		PostgresUser:       getEnv("POSTGRES_USER", "scorehub"),
		PostgresPassword:   getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:         getEnv("POSTGRES_DB", "scorehub"),
		KafkaBrokers:       []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaGroupID:       getEnv("KAFKA_GROUP_ID", "scorehub-group"),
		ScoreEventsTopic:   getEnv("SCORE_EVENTS_TOPIC", "score_events"),
		NotificationsTopic: getEnv("NOTIFICATIONS_TOPIC", "notifications"),
		GRPCPort:           getEnv("GRPC_PORT", "50051"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		GatewayPort:        getEnv("GATEWAY_PORT", "8081"),
		ServiceName:        getEnv("SERVICE_NAME", "unknown-service"),
		UserServiceAddr:    getEnv("USER_SERVICE_ADDR", "localhost:50051"),
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
