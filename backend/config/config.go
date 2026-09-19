package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port        string
	Environment string
	MongoURI    string
	RedisURL    string
	JWTSecret   string
	FrontendURL string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", ":8080"),
		Environment: strings.ToLower(getEnv("APP_ENV", "development")),
		MongoURI:    getEnv("MONGODB_URI", "mongodb://localhost:27017/pulsevote"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:   getEnv("JWT_SECRET", "pulsevote-dev-secret"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}

	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET cannot be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func NewRedisClient(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
