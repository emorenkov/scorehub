package config

import (
	"os"
	"strconv"

	"github.com/emorenkov/scorehub/pkg/common/db"
)

type UserConfig struct {
	GRPCPort    string
	HTTPPort    string
	GatewayPort string
	ServiceName string
	// For grpc-gateway to reach the gRPC server; default is local user-service gRPC port
	UserServiceAddr string
	DbConfig        *db.PostgresConfig
}

func Load() *UserConfig {
	return &UserConfig{
		DbConfig: &db.PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "scorehub_user"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DB:       getEnv("POSTGRES_DB", "scorehub"),
		},
		GRPCPort:        getEnv("GRPC_PORT", "50051"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		GatewayPort:     getEnv("GATEWAY_PORT", "8081"),
		ServiceName:     getEnv("SERVICE_NAME", "user-service"),
		UserServiceAddr: getEnv("USER_SERVICE_ADDR", "localhost:50051"),
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
