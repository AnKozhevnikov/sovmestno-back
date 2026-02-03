package middleware

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(c *gin.Context) {
	path := c.Request.URL.Path
	if isPublicRoute(path) {
		c.Next()
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(401, gin.H{"error": "Unauthorized: missing token"})
		c.Abort()
		return
	}

	claims, err := validateToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
		c.Abort()
		return
	}

	c.Set("user_id", claims.UserID)
	c.Set("role", claims.Role)
	c.Next()
}

func isPublicRoute(path string) bool {
	publicRoutes := map[string]bool{
		"/health":              true,
		"/swagger":             true,
		"/swagger/index.html":  true,
	}

	if publicRoutes[path] {
		return true
	}

	allowedPrefixes := []string{
		"/api/auth/",
		"/api/user/auth/",
		"/swagger/",
		"/swagger-user/",
		"/swagger-event/",
		"/swagger-application/",
	}

	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	if path == "/api/auth" || path == "/api/user/auth" {
		return true
	}

	return false
}

func validateToken(tokenString string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if len(secretKey) == 0 {
		return nil, errors.New("JWT_SECRET not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
