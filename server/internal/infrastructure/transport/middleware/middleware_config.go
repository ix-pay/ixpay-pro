package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	auth "github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	apperror "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/error"
)

// MiddlewareConfig 中间件配置中心
// 集中管理所有中间件，提供统一的中间件注册和使用机制
type MiddlewareConfig struct {
	// 依赖服务
	Auth   *auth.JWTAuth
	Logger logger.Logger
	Cache  cache.Cache

	// 中间件
	AuthMiddleware          gin.HandlerFunc
	PermissionMiddleware    gin.HandlerFunc
	ErrorMiddleware         gin.HandlerFunc
	LogMiddleware           gin.HandlerFunc
	ContextLoggerMiddleware gin.HandlerFunc
	RequestLogMiddleware    gin.HandlerFunc // 请求日志中间件
	AuditLogMiddleware      gin.HandlerFunc // 审计日志中间件
	CacheMiddleware         gin.HandlerFunc
	CORSMiddleware          gin.HandlerFunc
	CacheControlMiddleware  gin.HandlerFunc
	CaptchaMiddleware       gin.HandlerFunc
}

// SetupMiddlewareConfig 创建中间件配置中心
// 参数:
// - jwtAuth: JWT认证服务，用于身份验证
// - log: 日志记录器
// - cache: 缓存服务
// 返回值:
// - 配置好的中间件配置中心实例
func SetupMiddlewareConfig(
	auth *auth.JWTAuth,
	log logger.Logger,
	cache cache.Cache,
) *MiddlewareConfig {
	// 创建中间件配置实例
	mc := &MiddlewareConfig{
		Auth:   auth,
		Logger: log,
		Cache:  cache,
	}

	// 设置中间件
	mc.SetupLogMiddleware()
	mc.SetupContextLoggerMiddleware()
	mc.SetupRequestLogMiddleware()
	mc.SetupAuditLogMiddleware()
	mc.SetupErrorMiddleware()
	mc.SetupCorsMiddleware()
	mc.SetupCacheMiddleware(0) // 默认缓存过期时间为 0
	mc.SetupCacheControlMiddleware()
	mc.SetupAuthMiddleware()
	mc.SetupPermissionMiddleware()
	mc.SetupCaptchaMiddleware()

	return mc
}

// RegisterAllMiddlewares 注册所有中间件到路由引擎
// 参数:
// - router: Gin 路由引擎
// 作用:
// - 按照正确的顺序注册所有中间件
func (mc *MiddlewareConfig) RegisterAllMiddlewares(router *gin.Engine) {
	// 先注册全局中间件
	router.Use(mc.LogMiddleware)
	router.Use(mc.ContextLoggerMiddleware)
	router.Use(mc.RequestLogMiddleware)
	router.Use(mc.AuditLogMiddleware)
	router.Use(mc.ErrorMiddleware)
	router.Use(mc.CORSMiddleware)
	router.Use(mc.CacheControlMiddleware)
	// 如果设置了缓存中间件，注册为全局中间件
	// 注意：缓存中间件通常在路由组级别注册，而不是全局注册，这样可以更灵活地控制哪些路由需要缓存
	// if mc.CacheMiddleware != nil {
	//     router.Use(mc.CacheMiddleware)
	// }

	// 注意：认证和权限中间件通常在路由组级别注册，而不是全局注册
	// 这样可以更灵活地控制哪些路由需要验证
}

// SetPermissionMiddleware 设置权限中间件
// 参数:
// - permissionMiddleware: 权限中间件处理函数
// 作用:
// - 允许在应用程序初始化后设置权限中间件，因为它可能依赖于其他服务
func (mc *MiddlewareConfig) SetPermissionMiddleware(permissionMiddleware gin.HandlerFunc) error {
	if permissionMiddleware == nil {
		return fmt.Errorf("权限中间件不能为空")
	}

	mc.PermissionMiddleware = permissionMiddleware
	return nil
}

// SetCacheMiddleware 设置缓存中间件
// 参数:
// - cacheMiddleware: 缓存中间件处理函数
// 作用:
// - 允许在应用程序初始化后设置缓存中间件，因为它需要缓存服务
func (mc *MiddlewareConfig) SetCacheMiddleware(cacheMiddleware gin.HandlerFunc) error {
	if cacheMiddleware == nil {
		return fmt.Errorf("缓存中间件不能为空")
	}

	mc.CacheMiddleware = cacheMiddleware
	return nil
}

// GetAuthMiddleware 获取认证中间件
// 返回值:
// - 认证中间件处理函数
func (mc *MiddlewareConfig) GetAuthMiddleware() gin.HandlerFunc {
	return mc.AuthMiddleware
}

// GetPermissionMiddleware 获取权限中间件
// 返回值:
// - 权限中间件处理函数
// - 是否存在的布尔值
func (mc *MiddlewareConfig) GetPermissionMiddleware() (gin.HandlerFunc, bool) {
	return mc.PermissionMiddleware, mc.PermissionMiddleware != nil
}

// GetLogMiddleware 获取日志中间件
// 返回值:
// - 日志中间件处理函数
func (mc *MiddlewareConfig) GetLogMiddleware() gin.HandlerFunc {
	return mc.LogMiddleware
}

// SetupCacheMiddleware 创建并设置缓存中间件
// 参数:
// - expiration: 缓存过期时间
// 作用:
// - 创建缓存中间件并设置到中间件配置中心
func (mc *MiddlewareConfig) SetupCacheMiddleware(expiration time.Duration) {
	mc.CacheMiddleware = CacheMiddleware(mc.Cache, expiration)
}

// GetCORSMiddleware 获取CORS中间件
// 返回值:
// - CORS中间件处理函数
func (mc *MiddlewareConfig) GetCORSMiddleware() gin.HandlerFunc {
	return mc.CORSMiddleware
}

// GetErrorMiddleware 获取错误处理中间件
// 返回值:
// - 错误处理中间件处理函数
func (mc *MiddlewareConfig) GetErrorMiddleware() gin.HandlerFunc {
	return mc.ErrorMiddleware
}

// GetCacheMiddleware 获取缓存中间件
// 返回值:
// - 缓存中间件处理函数
// - 是否存在的布尔值
func (mc *MiddlewareConfig) GetCacheMiddleware() (gin.HandlerFunc, bool) {
	return mc.CacheMiddleware, mc.CacheMiddleware != nil
}

// SetupLogMiddleware 设置 HTTP 请求日志中间件
func (mc *MiddlewareConfig) SetupLogMiddleware() {
	mc.LogMiddleware = LogMiddleware(mc.Logger)
}

// SetupContextLoggerMiddleware 设置请求上下文日志中间件
func (mc *MiddlewareConfig) SetupContextLoggerMiddleware() {
	mc.ContextLoggerMiddleware = ContextLoggerMiddleware(mc.Logger)
}

// SetupRequestLogMiddleware 设置请求日志中间件（记录到独立的 request.log 文件）
func (mc *MiddlewareConfig) SetupRequestLogMiddleware() {
	mc.RequestLogMiddleware = RequestLogMiddleware()
}

// SetupAuditLogMiddleware 设置审计日志中间件（记录敏感操作到 audit.log）
func (mc *MiddlewareConfig) SetupAuditLogMiddleware() {
	mc.AuditLogMiddleware = AuditLogMiddleware()
}

// SetupErrorMiddleware 设置错误处理中间件
func (mc *MiddlewareConfig) SetupErrorMiddleware() {
	mc.ErrorMiddleware = ErrorMiddleware()
}

// SetupCorsMiddleware 设置CORS中间件
func (mc *MiddlewareConfig) SetupCorsMiddleware() {
	mc.CORSMiddleware = CORSMiddleware()
}

// SetupCacheControlMiddleware 设置缓存控制中间件
func (mc *MiddlewareConfig) SetupCacheControlMiddleware() {
	mc.CacheControlMiddleware = func(c *gin.Context) {
		// 默认缓存控制设置
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// SetupAuthMiddleware 设置认证中间件
func (mc *MiddlewareConfig) SetupAuthMiddleware() {
	if mc.Auth == nil {
		mc.AuthMiddleware = nil
		return
	}

	mc.AuthMiddleware = func(c *gin.Context) {
		// 从Authorization头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			_ = c.AbortWithError(http.StatusUnauthorized, apperror.Unauthorized("Authorization header is required", nil))
			return
		}

		// 检查令牌格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			_ = c.AbortWithError(http.StatusUnauthorized, apperror.Unauthorized("Authorization header format must be Bearer {token}", nil))
			return
		}

		// 解析和验证令牌
		claims, err := mc.Auth.ParseToken(parts[1])
		if err != nil {
			_ = c.AbortWithError(http.StatusUnauthorized, apperror.Unauthorized("Invalid or expired token", err))
			return
		}

		// 将用户信息添加到上下文
		c.Set("userID", claims.UserID)
		c.Set("userName", claims.Username)
		c.Set("role", claims.Role)
		c.Set("loginType", claims.LoginType)
		c.Set("claims", claims)

		c.Next()
	}
}

// SetupPermissionMiddleware 设置权限中间件
func (mc *MiddlewareConfig) SetupPermissionMiddleware() {
	// 默认权限中间件为空，需要在应用程序中单独设置
	mc.PermissionMiddleware = nil
}

// SetupCaptchaMiddleware 设置验证码中间件
func (mc *MiddlewareConfig) SetupCaptchaMiddleware() {
	// 默认验证码中间件为空，需要在应用程序中单独设置
	mc.CaptchaMiddleware = nil
}
