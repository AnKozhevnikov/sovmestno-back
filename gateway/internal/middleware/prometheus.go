package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"service", "method", "endpoint"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served",
		},
	)
)

// PrometheusMiddleware собирает метрики для каждого HTTP запроса
func PrometheusMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Увеличиваем счетчик активных запросов
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// Обрабатываем запрос
		c.Next()

		// Вычисляем длительность
		duration := time.Since(start).Seconds()

		// Получаем endpoint (путь без query параметров)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Записываем метрики
		status := strconv.Itoa(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(serviceName, c.Request.Method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(serviceName, c.Request.Method, endpoint).Observe(duration)
	}
}
