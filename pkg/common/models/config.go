package models

import (
	"os"
	"strconv"
)

type RedisConfig struct {
	RateLimitBurst int
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
}

func LoadRedisConfig() *RedisConfig {
	return &RedisConfig{
		RateLimitBurst: GetEnvAsInt("RATE_LIMIT_BURST", 20),
		RedisAddr:      GetEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  GetEnv("REDIS_PASSWORD", ""),
		RedisDB:        GetEnvAsInt("REDIS_DB", 0),
	}
}

type PostgresConfig struct {
	Host     string
	User     string
	Password string
	DB       string
	Port     int
}

func LoadPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:     GetEnv("POSTGRES_HOST", "localhost"),
		Port:     GetEnvAsInt("POSTGRES_PORT", 5432),
		User:     GetEnv("POSTGRES_USER", "scorehub_user"),
		Password: GetEnv("POSTGRES_PASSWORD", "postgres"),
		DB:       GetEnv("POSTGRES_DB", "scorehub"),
	}
}

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
