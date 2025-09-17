package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// bodyLogWriter 包装响应写入器，用于捕获响应体

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，捕获响应体
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		tstart := time.Now()

		// 获取请求信息
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		// 捕获请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := c.GetRawData()
			requestBody = string(bodyBytes)
			// 重新设置请求体，以便后续处理
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 捕获响应体
		w := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 计算请求处理时间
		duration := time.Since(tstart)

		// 获取状态码和响应体
		statusCode := c.Writer.Status()
		responseBody := w.body.String()

		// 获取用户信息（如果已认证）
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		role, _ := c.Get("role")

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			log.Error("API request",
				"path", path,
				"method", method,
				"ip", ip,
				"status", statusCode,
				"duration", duration,
				"userID", userID,
				"username", username,
				"role", role,
				"requestBody", requestBody,
				"responseBody", responseBody,
			)
		} else if statusCode >= 400 {
			log.Warn("API request",
				"path", path,
				"method", method,
				"ip", ip,
				"status", statusCode,
				"duration", duration,
				"userID", userID,
				"username", username,
				"role", role,
				"requestBody", requestBody,
				"responseBody", responseBody,
			)
		} else {
			log.Info("API request",
				"path", path,
				"method", method,
				"ip", ip,
				"status", statusCode,
				"duration", duration,
				"userID", userID,
				"username", username,
				"role", role,
				// 对于成功的请求，通常不需要记录请求体和响应体，除非特别需要
			)
		}
	}
}
