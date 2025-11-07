package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	Queue     QueueConfig
	Webhook   WebhookConfig
}

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type RateLimitConfig struct {
	Standard int
	Premium  int
	Batch    int
}

type QueueConfig struct {
	WorkerCount int
	MaxSize     int
}

type WebhookConfig struct {
	TimeoutSeconds int
	RetryCount     int
}

var AppConfig *Config

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "lingxi"),
			Password:     getEnv("DB_PASSWORD", "test_password"),
			DBName:       getEnv("DB_NAME", "lingxi"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", "test_password"),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 100),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "default-secret-change-me"),
			ExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		},
		RateLimit: RateLimitConfig{
			Standard: getEnvAsInt("RATE_LIMIT_STANDARD", 100),
			Premium:  getEnvAsInt("RATE_LIMIT_PREMIUM", 1000),
			Batch:    getEnvAsInt("RATE_LIMIT_BATCH", 10),
		},
		Queue: QueueConfig{
			WorkerCount: getEnvAsInt("QUEUE_WORKER_COUNT", 10),
			MaxSize:     getEnvAsInt("QUEUE_MAX_SIZE", 10000),
		},
		Webhook: WebhookConfig{
			TimeoutSeconds: getEnvAsInt("WEBHOOK_TIMEOUT_SECONDS", 30),
			RetryCount:     getEnvAsInt("WEBHOOK_RETRY_COUNT", 3),
		},
	}

	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
