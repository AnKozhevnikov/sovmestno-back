package middleware

import (
	"event-service/internal/apperror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ExtractUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Missing user context"))
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Invalid user ID"))
			c.Abort()
			return
		}

		role := c.GetHeader("X-User-Role")
		if role == "" {
			c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Missing user role"))
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Role not found in context"))
			c.Abort()
			return
		}

		roleStr := role.(string)
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "Insufficient permissions"))
		c.Abort()
	}
}
