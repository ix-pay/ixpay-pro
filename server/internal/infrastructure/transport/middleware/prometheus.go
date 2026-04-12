package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics 定义Prometheus指标
var (
	requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "The total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "The HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	requestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "The HTTP request size in bytes.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "The HTTP response size in bytes.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

// PrometheusMiddleware 创建一个Prometheus中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(startTime)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		statusStr := fmt.Sprintf("%d", status)

		// 记录请求大小（如果有）
		if c.Request.ContentLength > 0 {
			requestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))
		}

		// 记录响应大小
		responseSize.WithLabelValues(method, path, statusStr).Observe(float64(c.Writer.Size()))

		// 记录请求计数和持续时间
		requestCount.WithLabelValues(method, path, statusStr).Inc()
		requestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
	}
}

// PrometheusHandler 返回一个HTTP处理函数，用于暴露Prometheus指标
func PrometheusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}
}
