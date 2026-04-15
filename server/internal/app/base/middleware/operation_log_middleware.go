package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	auth "github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
)

// bodyLogWriter 用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// OperationLogMiddleware 操作日志中间件
func OperationLogMiddleware(operationLogService *service.OperationLogService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 记录请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 解析请求参数
		var params string
		if method == http.MethodGet || method == http.MethodDelete {
			params = c.Request.URL.RawQuery
		} else {
			// 读取请求体
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Error("读取请求体失败", "error", err)
			} else {
				params = string(bodyBytes)
				// 重置请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 记录响应体
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 执行请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)

		// 获取响应信息
		statusCode := c.Writer.Status()
		responseBody := blw.body.String()

		// 判断请求是否成功
		isSuccess := statusCode < 400

		// 获取用户信息
		userIDStr := ""
		userName := ""
		nickname := ""

		// 从上下文中获取用户信息（从认证中间件设置的 claims 中获取）
		if claims, exists := c.Get("claims"); exists {
			if c, ok := claims.(*auth.Claims); ok {
				userIDStr = c.UserID
				userName = c.Username
				nickname = c.Nickname // 从 claims 中获取准确的 nickname
			}
		}

		// 将 userID 从 string 转换为 int64
		var userID int64
		if userIDStr != "" {
			userID, _ = strconv.ParseInt(userIDStr, 10, 64)
		}

		// 获取操作类型和模块
		operationType := getOperationType(method, path)
		module := getModule(path)

		// 构建操作描述
		description := getOperationDescription(method, path, module)

		// 构建错误信息
		errorMessage := ""
		if !isSuccess {
			// 尝试解析错误信息
			var errResponse map[string]interface{}
			if err := json.Unmarshal([]byte(responseBody), &errResponse); err == nil {
				if msg, ok := errResponse["message"].(string); ok {
					errorMessage = msg
				}
				if data, ok := errResponse["data"]; ok {
					if dataStr, ok := data.(string); ok {
						errorMessage += ": " + dataStr
					}
				}
			} else {
				// 如果解析失败，直接使用响应体的前100个字符
				if len(responseBody) > 100 {
					errorMessage = responseBody[:100] + "..."
				} else {
					errorMessage = responseBody
				}
			}
		}

		// 异步记录操作日志
		go func() {
			// 构建操作日志
			operationLog := &entity.OperationLog{
				UserID:        userID,
				Username:      userName,
				Nickname:      nickname,
				OperationType: operationType,
				Module:        module,
				Description:   description,
				Method:        method,
				Path:          path,
				Params:        params,
				ClientIP:      clientIP,
				UserAgent:     userAgent,
				StatusCode:    statusCode,
				Result:        responseBody,
				Duration:      duration.Milliseconds(),
				ErrorMessage:  errorMessage,
				IsSuccess:     isSuccess,
			}

			// 记录操作日志
			if err := operationLogService.CreateLog(operationLog); err != nil {
				log.Error("记录操作日志失败", "error", err)
			}
		}()
	}
}

// getOperationType 根据请求方法和路径获取操作类型
func getOperationType(method, path string) entity.OperationType {
	// 特殊路径处理
	if strings.Contains(path, "/login") {
		return entity.OperationTypeLogin
	}
	if strings.Contains(path, "/logout") {
		return entity.OperationTypeLogout
	}

	// 根据方法判断
	switch method {
	case http.MethodPost:
		if strings.Contains(path, "/batch") || strings.Contains(path, "/bulk") {
			return entity.OperationTypeOther
		}
		return entity.OperationTypeCreate
	case http.MethodPut, http.MethodPatch:
		return entity.OperationTypeUpdate
	case http.MethodDelete:
		return entity.OperationTypeDelete
	case http.MethodGet, http.MethodHead:
		// 如果是详情查询，也算查询操作
		if strings.Contains(path, "/detail") || strings.Contains(path, "/info") {
			return entity.OperationTypeQuery
		}
		// 如果是列表查询，也算查询操作
		if strings.Contains(path, "/list") || strings.Contains(path, "/page") || strings.Contains(path, "/search") {
			return entity.OperationTypeQuery
		}
		return entity.OperationTypeQuery
	default:
		return entity.OperationTypeOther
	}
}

// getModule 根据路径获取模块名称
func getModule(path string) string {
	// 分割路径
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return "system"
	}

	// 通常API路径格式为 //module/...
	if parts[1] == "v1" && len(parts) >= 3 {
		return parts[2]
	}

	// 其他情况
	if len(parts) >= 2 {
		return parts[1]
	}

	return "system"
}

// getOperationDescription 构建操作描述
func getOperationDescription(method, path, module string) string {
	operationType := getOperationType(method, path)
	operationTypeName := ""

	switch operationType {
	case entity.OperationTypeCreate:
		operationTypeName = "创建"
	case entity.OperationTypeUpdate:
		operationTypeName = "更新"
	case entity.OperationTypeDelete:
		operationTypeName = "删除"
	case entity.OperationTypeQuery:
		operationTypeName = "查询"
	case entity.OperationTypeLogin:
		operationTypeName = "登录"
	case entity.OperationTypeLogout:
		operationTypeName = "登出"
	default:
		operationTypeName = "操作"
	}

	return operationTypeName + module
}
