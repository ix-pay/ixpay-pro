package middleware

import (
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"

	"github.com/gin-gonic/gin"
)

// LogMiddleware 定义日志中间件
func LogMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 请求方式
		requestMethod := c.Request.Method
		// 请求路由
		requestURI := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		// 错误信息
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 构建日志字段
		fields := []interface{}{
			"status_code", statusCode,
			"latency", latencyTime,
			"client_ip", clientIP,
			"method", requestMethod,
			"uri", requestURI,
			"error", errorMessage,
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error("请求失败", fields...)
		case statusCode >= 400:
			logger.Warn("请求异常", fields...)
		default:
			logger.Info("请求成功", fields...)
		}
	}
}

// ContextLoggerMiddleware 为请求上下文添加日志记录器
func ContextLoggerMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		// 设置请求ID到响应头
		c.Header("X-Request-ID", requestID)
		// 创建带有请求ID的日志记录器
		ctxLogger := logger.With("request_id", requestID)
		// 将日志记录器设置到上下文
		c.Set("logger", ctxLogger)
		// 处理请求
		c.Next()
	}
}

// GetLoggerFromContext 从上下文获取日志记录器
func GetLoggerFromContext(c *gin.Context) logger.Logger {
	if logInstance, exists := c.Get("logger"); exists {
		if l, ok := logInstance.(logger.Logger); ok {
			return l
		}
	}
	return nil
}
