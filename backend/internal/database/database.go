package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"MLQueue/internal/config"
	"MLQueue/internal/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
)

// InitDB initializes PostgreSQL connection with connection pooling
func InitDB(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: true, // Cache prepared statements
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings for high concurrency
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate V1 models
	if err := models.AutoMigrate(DB); err != nil {
		return fmt.Errorf("failed to migrate V1 models: %w", err)
	}

	// Auto migrate V2 models (Python客户端驱动架构)
	if err := models.AutoMigrateV2(DB); err != nil {
		return fmt.Errorf("failed to migrate V2 models: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// InitRedis initializes Redis connection with connection pooling
func InitRedis(cfg *config.Config) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: 10,
		MaxRetries:   3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

// Close closes database connections
func Close() {
	if sqlDB, err := DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			return
		}
	}
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			return
		}
	}
}
