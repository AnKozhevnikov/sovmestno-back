package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func ApplicationHandler(c *gin.Context) {
	capturedPath := c.Param("path")
	c.Request.URL.Path = capturedPath

	serviceURL := os.Getenv("APPLICATION_SERVICE_URL")
	if serviceURL == "" {
		c.JSON(503, gin.H{"error": "Application service not configured"})
		return
	}

	targetURL, err := url.Parse(serviceURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Invalid target URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       30 * time.Second,
	}

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(int); ok {
				req.Header.Set("X-User-ID", strconv.Itoa(id))
			}
		}
		if role, exists := c.Get("role"); exists {
			if roleStr, ok := role.(string); ok {
				req.Header.Set("X-User-Role", roleStr)
			}
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		c.JSON(503, gin.H{
			"error":   "Service temporarily unavailable",
			"details": err.Error(),
		})
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func ApplicationSwaggerHandler(c *gin.Context) {
	serviceURL := os.Getenv("APPLICATION_SERVICE_URL")
	if serviceURL == "" {
		c.JSON(503, gin.H{"error": "Application service not configured"})
		return
	}

	targetURL, err := url.Parse(serviceURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Invalid target URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	c.Request.URL.Path = "/swagger" + c.Param("any")
	proxy.ServeHTTP(c.Writer, c.Request)
}
