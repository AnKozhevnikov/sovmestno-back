package handlers

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status   string          `json:"status"`
	Gateway  string          `json:"gateway"`
	Services []ServiceHealth `json:"services"`
}

// HealthHandler godoc
// @Summary      Health check
// @Description  Проверяет состояние gateway и всех подключенных микросервисов
// @Tags         health
// @Produce      json
// @Success      200 {object} HealthResponse "Все сервисы работают"
// @Failure      503 {object} HealthResponse "Один или несколько сервисов недоступны"
// @Router       /health [get]
func HealthHandler(c *gin.Context) {
	services := map[string]string{
		"user-service":        os.Getenv("USER_SERVICE_URL"),
		"event-service":       os.Getenv("EVENT_SERVICE_URL"),
		"application-service": os.Getenv("APPLICATION_SERVICE_URL"),
	}

	var wg sync.WaitGroup
	healthResults := make([]ServiceHealth, 0, len(services))
	resultsMutex := sync.Mutex{}
	overallHealthy := true

	for name, url := range services {
		if url == "" {
			continue
		}

		wg.Add(1)
		go func(serviceName, serviceURL string) {
			defer wg.Done()

			status := "healthy"
			errMsg := ""

			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(serviceURL + "/health")

			if err != nil {
				status = "unhealthy"
				errMsg = err.Error()
				overallHealthy = false
			} else {
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					status = "unhealthy"
					errMsg = "non-200 status code"
					overallHealthy = false
				}
			}

			resultsMutex.Lock()
			healthResults = append(healthResults, ServiceHealth{
				Name:   serviceName,
				Status: status,
				Error:  errMsg,
			})
			resultsMutex.Unlock()
		}(name, url)
	}

	wg.Wait()

	statusCode := 200
	overallStatus := "healthy"
	if !overallHealthy {
		statusCode = 503
		overallStatus = "degraded"
	}

	c.JSON(statusCode, gin.H{
		"status":   overallStatus,
		"gateway":  "healthy",
		"services": healthResults,
	})
}
