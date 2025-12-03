package config

import (
	"os"
	"strconv"

	"github.com/emorenkov/scorehub/pkg/common/models"
)

type UserConfig struct {
	GRPCPort    string
	HTTPPort    string
	GatewayPort string
	ServiceName string
	// For grpc-gateway to reach the gRPC server; default is local user-service gRPC port
	UserServiceAddr string
	APIKey          string
	RateLimitRPS    int
	RateLimitBurst  int
	RedisConfig     *models.RedisConfig
	DbConfig        *models.PostgresConfig
}

func Load() *UserConfig {
	return &UserConfig{
		DbConfig:        models.LoadPostgresConfig(),
		RedisConfig:     models.LoadRedisConfig(),
		GRPCPort:        getEnv("GRPC_PORT", "50051"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		GatewayPort:     getEnv("GATEWAY_PORT", "8081"),
		ServiceName:     getEnv("SERVICE_NAME", "user-service"),
		UserServiceAddr: getEnv("USER_SERVICE_ADDR", "localhost:50051"),
		APIKey:          getEnv("API_KEY", ""),
		RateLimitRPS:    getEnvAsInt("RATE_LIMIT_RPS", 10),
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
