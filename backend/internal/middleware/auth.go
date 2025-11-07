package middleware

import (
	"net/http"
	"strings"

	"MLQueue/internal/database"
	"MLQueue/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates Bearer token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "需要身份验证",
				"code":    "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer TOKEN"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "无效的Authorization header格式",
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token against database
		var user models.User
		if err := database.DB.Where("api_key = ?", token).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "无效的Token",
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", user.ID)
		c.Set("user_tier", user.Tier)
		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// GetUserTier retrieves user tier from context
func GetUserTier(c *gin.Context) string {
	if tier, exists := c.Get("user_tier"); exists {
		return tier.(string)
	}
	return "standard"
}
