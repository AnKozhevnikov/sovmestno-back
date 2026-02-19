package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func UserSwaggerHandler(c *gin.Context) {
	path := c.Param("any")

	if strings.HasSuffix(path, "swagger.json") || path == "/doc.json" {
		serveModifiedSwaggerSpec(c)
		return
	}

	serviceURL := os.Getenv("USER_SERVICE_URL")
	if serviceURL == "" {
		c.JSON(503, gin.H{"error": "User service not configured"})
		return
	}

	targetURL := serviceURL + "/swagger" + path

	resp, err := http.Get(targetURL)
	if err != nil {
		c.JSON(503, gin.H{"error": "Failed to fetch from user service"})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func serveModifiedSwaggerSpec(c *gin.Context) {
	serviceURL := os.Getenv("USER_SERVICE_URL")
	if serviceURL == "" {
		c.JSON(503, gin.H{"error": "User service not configured"})
		return
	}

	resp, err := http.Get(serviceURL + "/swagger/doc.json")
	if err != nil {
		c.JSON(503, gin.H{"error": "Failed to fetch swagger spec"})
		return
	}
	defer resp.Body.Close()

	var spec map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse swagger spec"})
		return
	}

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = c.Request.Host // Use current request host if not set
	}
	spec["host"] = apiHost
	spec["basePath"] = "/"

	if paths, ok := spec["paths"].(map[string]interface{}); ok {
		newPaths := make(map[string]interface{})
		for path, value := range paths {
			newPath := "/api/user" + path
			newPaths[newPath] = value
		}
		spec["paths"] = newPaths
	}

	c.JSON(200, spec)
}

func EventSwaggerHandler(c *gin.Context) {
	path := c.Param("any")

	if strings.HasSuffix(path, "swagger.json") || path == "/doc.json" {
		serveModifiedEventSwaggerSpec(c)
		return
	}

	serviceURL := os.Getenv("EVENT_SERVICE_URL")
	if serviceURL == "" {
		c.JSON(503, gin.H{"error": "Event service not configured"})
		return
	}

	targetURL := serviceURL + "/swagger" + path

	resp, err := http.Get(targetURL)
	if err != nil {
		c.JSON(503, gin.H{"error": "Failed to fetch from event service"})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func serveModifiedEventSwaggerSpec(c *gin.Context) {
	serviceURL := os.Getenv("EVENT_SERVICE_URL")
	if serviceURL == "" {
		c.JSON(503, gin.H{"error": "Event service not configured"})
		return
	}

	resp, err := http.Get(serviceURL + "/swagger/doc.json")
	if err != nil {
		c.JSON(503, gin.H{"error": "Failed to fetch swagger spec"})
		return
	}
	defer resp.Body.Close()

	var spec map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse swagger spec"})
		return
	}

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = c.Request.Host // Use current request host if not set
	}
	spec["host"] = apiHost
	spec["basePath"] = "/"

	if paths, ok := spec["paths"].(map[string]interface{}); ok {
		newPaths := make(map[string]interface{})
		for path, value := range paths {
			newPath := "/api/event" + path
			newPaths[newPath] = value
		}
		spec["paths"] = newPaths
	}

	c.JSON(200, spec)
}
