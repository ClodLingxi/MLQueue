package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"MLQueue/internal/config"
	"MLQueue/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware implements token bucket rate limiting
func RateLimitMiddleware(isBatch bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		tier := GetUserTier(c)

		// Get rate limit based on tier and operation type
		var limit int
		if isBatch {
			limit = config.AppConfig.RateLimit.Batch
		} else if tier == "premium" {
			limit = config.AppConfig.RateLimit.Premium
		} else {
			limit = config.AppConfig.RateLimit.Standard
		}

		// Check rate limit using Redis
		allowed, err := checkRateLimit(userID, limit, isBatch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "速率限制检查失败",
				"code":    "INTERNAL_ERROR",
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "请求频率超限",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit uses Redis to implement sliding window rate limiting
func checkRateLimit(userID string, limit int, isBatch bool) (bool, error) {
	ctx := context.Background()
	now := time.Now()
	window := time.Minute

	key := fmt.Sprintf("ratelimit:%s", userID)
	if isBatch {
		key = fmt.Sprintf("ratelimit:batch:%s", userID)
	}

	// Remove old entries outside the window
	minScore := now.Add(-window).Unix()
	database.RedisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", minScore))

	// Count current requests in window
	count, err := database.RedisClient.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if int(count) >= limit {
		return false, nil
	}

	// Add current request
	member := fmt.Sprintf("%d", now.UnixNano())
	database.RedisClient.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.Unix()),
		Member: member,
	})

	// Set expiry on key
	database.RedisClient.Expire(ctx, key, window+time.Minute)

	return true, nil
}

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
