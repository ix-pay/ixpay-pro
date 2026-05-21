package middleware

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	loggerKey contextKey = "logger"
)

func LogMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查日志
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		requestMethod := c.Request.Method
		requestURI := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		reqLogger := logger.WithContext(c.Request.Context())

		fields := []interface{}{
			"status_code", statusCode,
			"latency", latencyTime,
			"client_ip", clientIP,
			"method", requestMethod,
			"uri", requestURI,
			"error", errorMessage,
		}

		switch {
		case statusCode >= 500:
			reqLogger.Error("请求失败", fields...)
		case statusCode >= 400:
			reqLogger.Warn("请求异常", fields...)
		default:
			reqLogger.Info("请求成功", fields...)
		}
	}
}

func ContextLoggerMiddleware(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = logger.NewTraceID()
		}

		c.Header("X-Trace-ID", traceID)
		c.Header("X-Request-ID", uuid.New().String())

		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		reqLogger := l.With("trace_id", traceID)
		c.Set(loggerKey, reqLogger)

		c.Next()
	}
}

func GetLoggerFromContext(c *gin.Context) logger.Logger {
	if logInstance, exists := c.Get(loggerKey); exists {
		if l, ok := logInstance.(logger.Logger); ok {
			return l
		}
	}
	return logger.GetGlobalLogger(logger.RequestLogger)
}
