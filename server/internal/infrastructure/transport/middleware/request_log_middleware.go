package middleware

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"

	"github.com/gin-gonic/gin"
)

// RequestLogMiddleware 请求日志中间件（记录到独立的 request.log 文件）
func RequestLogMiddleware() gin.HandlerFunc {
	// 获取请求日志记录器
	requestLogger := logger.GetGlobalLogger(logger.RequestLogger)

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行时间（毫秒）
		latencyMs := endTime.Sub(startTime).Milliseconds()
		// 请求方式
		requestMethod := c.Request.Method
		// 请求路由
		requestURI := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求 IP
		clientIP := c.ClientIP()
		// User-Agent
		userAgent := c.Request.UserAgent()
		// 错误信息
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 构建日志字段
		fields := []interface{}{
			"status_code", statusCode,
			"latency_ms", latencyMs,
			"client_ip", clientIP,
			"method", requestMethod,
			"uri", requestURI,
			"user_agent", userAgent,
		}

		// 如果有错误，添加错误信息
		if errorMessage != "" {
			fields = append(fields, "error", errorMessage)
		}

		// 记录请求日志（只记录到 request.log）
		if requestLogger != nil {
			requestLogger.Info("HTTP 请求", fields...)
		}
	}
}

// AuditLogMiddleware 审计日志中间件（记录敏感操作到 audit.log）
func AuditLogMiddleware() gin.HandlerFunc {
	// 获取审计日志记录器
	auditLogger := logger.GetGlobalLogger(logger.AuditLogger)

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 只对特定路径记录审计日志
		shouldAudit := shouldAuditRequest(c.Request.RequestURI, c.Request.Method)
		if !shouldAudit {
			return
		}

		// 执行时间（毫秒）
		latencyMs := time.Since(startTime).Milliseconds()
		// 状态码
		statusCode := c.Writer.Status()
		// 请求 IP
		clientIP := c.ClientIP()
		// 请求方式
		requestMethod := c.Request.Method
		// 请求路由
		requestURI := c.Request.RequestURI

		// 构建日志字段
		fields := []interface{}{
			"status_code", statusCode,
			"latency_ms", latencyMs,
			"client_ip", clientIP,
			"method", requestMethod,
			"uri", requestURI,
		}

		// 记录审计日志
		if auditLogger != nil {
			if statusCode >= 400 {
				auditLogger.Warn("敏感操作失败", fields...)
			} else {
				auditLogger.Info("敏感操作成功", fields...)
			}
		}
	}
}

// shouldAuditRequest 判断是否需要记录审计日志
func shouldAuditRequest(uri, method string) bool {
	// 需要审计的敏感操作路径
	sensitivePaths := []string{
		"/api//user",
		"/api//role",
		"/api//menu",
		"/api//permission",
		"/api//config",
		"/api//dict",
		"/api//department",
		"/api//position",
	}

	// 只记录写操作（POST, PUT, DELETE, PATCH）
	writeMethods := map[string]bool{
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
		"PATCH":  true,
	}

	if !writeMethods[method] {
		return false
	}

	// 检查路径是否匹配敏感操作
	for _, path := range sensitivePaths {
		if len(uri) >= len(path) && uri[:len(path)] == path {
			return true
		}
	}

	return false
}
