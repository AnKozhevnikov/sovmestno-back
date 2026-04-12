package middleware

import (
	"application-service/internal/apperror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ExtractUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		role := c.GetHeader("X-User-Role")

		if userIDStr == "" || role == "" {
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

		c.Set("user_id", userID)
		c.Set("role", role)

		c.Next()
	}
}
