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

		// 跳过监控接口，不记录到操作日志表
		if strings.HasPrefix(path, "/api/admin/monitor") {
			return
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

	// API 路径格式: /api/admin/{module}/...
	if len(parts) >= 4 && parts[1] == "api" && parts[2] == "admin" {
		resource := parts[3]
		return getResourceModuleName(resource)
	}

	// 兼容 /v1/{module}/... 格式
	if len(parts) >= 3 && parts[1] == "v1" {
		resource := parts[2]
		return getResourceModuleName(resource)
	}

	// 其他情况
	if len(parts) >= 2 {
		return getResourceModuleName(parts[1])
	}

	return "系统"
}

// getResourceModuleName 将资源路径转换为中文模块名称
func getResourceModuleName(resource string) string {
	// 路径与模块名称映射表
	moduleMap := map[string]string{
		"user":            "用户管理",
		"role":            "角色管理",
		"menu":            "菜单管理",
		"btn-perms":       "按钮权限",
		"apis":            "API管理",
		"logs":            "操作日志",
		"login-log":       "登录日志",
		"permission-logs": "权限日志",
		"dept":            "部门管理",
		"position":        "岗位管理",
		"config":          "配置管理",
		"dict":            "字典管理",
		"notices":         "公告管理",
		"online-user":     "在线用户",
		"monitor":         "系统监控",
		"task":            "任务管理",
		"auth":            "认证",
	}

	if name, ok := moduleMap[resource]; ok {
		return name
	}

	// 如果映射表中不存在，返回资源名
	return resource
}

// getOperationDescription 构建操作描述
func getOperationDescription(method, path, module string) string {
	operationType := getOperationType(method, path)
	operationTypeName := ""
	action := ""

	switch operationType {
	case entity.OperationTypeCreate:
		operationTypeName = "创建"
		action = getActionDescription(path)
	case entity.OperationTypeUpdate:
		operationTypeName = "更新"
		action = getActionDescription(path)
	case entity.OperationTypeDelete:
		operationTypeName = "删除"
		action = getActionDescription(path)
	case entity.OperationTypeQuery:
		operationTypeName = "查询"
		action = getActionDescription(path)
	case entity.OperationTypeLogin:
		return "用户登录"
	case entity.OperationTypeLogout:
		return "用户登出"
	default:
		operationTypeName = ""
		action = getActionDescription(path)
	}

	// 优先使用具体操作描述，如 "创建用户"、"查询用户列表"
	if action != "" {
		return operationTypeName + action
	}

	// 降级使用模块名称，如 "查询用户管理"
	return operationTypeName + module
}

// getActionDescription 根据路径获取具体操作描述
func getActionDescription(path string) string {
	// 提取路径的最后一段
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}

	lastPart := parts[len(parts)-1]

	// 批量操作
	if strings.Contains(path, "/batch") || strings.Contains(path, "/bulk") {
		return "批量操作"
	}

	// 清空操作
	if strings.Contains(path, "/clear") {
		return "清空"
	}

	// 特定操作描述
	actionMap := map[string]string{
		"info":                 "详情",
		"detail":               "详情",
		"tree":                 "树形结构",
		"all":                  "全部数据",
		"statistics":           "统计",
		"password":             "密码",
		"reset-password":       "重置密码",
		"change-password":      "修改密码",
		"refresh-token":        "刷新令牌",
		"captcha":              "验证码",
		"register":             "注册",
		"login":                "登录",
		"logout":               "登出",
		"assign-users":         "分配用户",
		"assign-menus":         "分配菜单",
		"assign-api-routes":    "分配API",
		"switch-role":          "切换角色",
		"setUserAuthority":     "设置用户权限",
		"setUserAuthorities":   "设置用户权限",
		"update-user-settings": "更新用户设置",
		"get-user-settings":    "获取用户设置",
		"publish":              "发布",
		"read":                 "标记已读",
		"is-read":              "检查已读状态",
		"start":                "启动",
		"stop":                 "停止",
		"retry":                "重试",
		"execution-logs":       "执行日志",
		"group":                "设置分组",
		"active":               "启用配置",
		"by-role":              "角色权限",
		"by-menu":              "菜单权限",
		"api-routes":           "API路由",
		"for-route":            "路由权限",
	}

	if action, ok := actionMap[lastPart]; ok {
		return action
	}

	// 对于列表查询，返回"列表"
	if strings.Contains(path, "/list") || strings.Contains(path, "/page") || strings.Contains(path, "/search") {
		return "列表"
	}

	// 如果路径只有两个部分（如 /api/admin/user），可能是列表或详情
	if len(parts) <= 3 {
		// GET 请求通常是查询列表
		if strings.HasPrefix(path, "/api/admin") || strings.HasPrefix(path, "/v1") {
			return getResourceName(path)
		}
	}

	return ""
}

// getResourceName 获取资源名称（用于操作描述）
func getResourceName(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// 找到资源部分
	resourceIdx := -1
	for i, part := range parts {
		if part == "api" || part == "v1" {
			resourceIdx = i + 1
			break
		}
	}

	if resourceIdx >= 0 && resourceIdx < len(parts) {
		resource := parts[resourceIdx]
		// 转换为单数形式（简单处理）
		if strings.HasSuffix(resource, "s") && resource != "apis" {
			resource = resource[:len(resource)-1]
		}
		return resource
	}

	return ""
}
