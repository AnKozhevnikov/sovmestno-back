package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ExtractUserContext извлекает user context из headers, установленных gateway
func ExtractUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Gateway передает user context через headers
		userIDHeader := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if userIDHeader != "" {
			userID, err := strconv.Atoi(userIDHeader)
			if err == nil {
				c.Set("user_id", userID)
			}
		}

		if userRole != "" {
			c.Set("role", userRole)
		}

		c.Next()
	}
}

// RequireAuth проверяет наличие user_id в контексте
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists || userID == nil {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID - helper для получения user_id из контекста
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// GetUserRole - helper для получения role из контекста
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	return roleStr, ok
}
