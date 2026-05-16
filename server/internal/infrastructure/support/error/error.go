package error

import (
	"net/http"
)

// ErrorCode 业务错误码类型
type ErrorCode int

// 业务错误码定义
const (
	// 系统错误
	ErrorCodeInternal          ErrorCode = 10001 // 内部服务器错误
	ErrorCodeDatabase          ErrorCode = 10002 // 数据库错误
	ErrorCodeRedis             ErrorCode = 10003 // Redis错误
	ErrorCodeCache             ErrorCode = 10004 // 缓存错误
	ErrorCodeThirdParty        ErrorCode = 10005 // 第三方服务错误
	ErrorCodeConfig            ErrorCode = 10006 // 配置错误
	ErrorCodeResourceExhausted ErrorCode = 10007 // 资源耗尽

	// 认证授权错误
	ErrorCodeUnauthorized      ErrorCode = 20001 // 未授权
	ErrorCodeForbidden         ErrorCode = 20002 // 禁止访问
	ErrorCodeTokenInvalid      ErrorCode = 20003 // Token无效
	ErrorCodeTokenExpired      ErrorCode = 20004 // Token过期
	ErrorCodePasswordIncorrect ErrorCode = 20005 // 密码错误
	ErrorCodeAccountLocked     ErrorCode = 20006 // 账户锁定
	ErrorCodeAccountExpired    ErrorCode = 20007 // 账户过期

	// 参数错误
	ErrorCodeBadRequest   ErrorCode = 30001 // 请求参数错误
	ErrorCodeValidation   ErrorCode = 30002 // 参数验证失败
	ErrorCodeMissingParam ErrorCode = 30003 // 缺少必填参数
	ErrorCodeInvalidParam ErrorCode = 30004 // 参数格式无效

	// 业务逻辑错误
	ErrorCodeNotFound        ErrorCode = 40001 // 资源不存在
	ErrorCodeAlreadyExists   ErrorCode = 40002 // 资源已存在
	ErrorCodeOperationFailed ErrorCode = 40003 // 操作失败
	ErrorCodeBusinessRule    ErrorCode = 40004 // 业务规则违反
	ErrorCodeDataIntegrity   ErrorCode = 40005 // 数据完整性错误

	// 权限错误
	ErrorCodePermissionDenied   ErrorCode = 50001 // 权限不足
	ErrorCodeRoleInvalid        ErrorCode = 50002 // 角色无效
	ErrorCodeMenuAccessDenied   ErrorCode = 50003 // 菜单访问被拒绝
	ErrorCodeButtonAccessDenied ErrorCode = 50004 // 按钮访问被拒绝
	ErrorCodeDataAccessDenied   ErrorCode = 50005 // 数据访问被拒绝
)

// AppError 应用错误结构
type AppError struct {
	Code         int         `json:"code"`              // HTTP状态码
	BusinessCode ErrorCode   `json:"businessCode"`      // 业务错误码
	Message      string      `json:"message"`           // 错误消息
	Details      interface{} `json:"details,omitempty"` // 错误详情
}

// Error 实现error接口
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError 创建新的应用错误
func NewAppError(code int, businessCode ErrorCode, message string, details interface{}) *AppError {
	return &AppError{
		Code:         code,
		BusinessCode: businessCode,
		Message:      message,
		Details:      details,
	}
}

// BadRequest 创建400错误
func BadRequest(message string, details interface{}) *AppError {
	return NewAppError(http.StatusBadRequest, ErrorCodeBadRequest, message, details)
}

// Unauthorized 创建401错误
func Unauthorized(message string, details interface{}) *AppError {
	return NewAppError(http.StatusUnauthorized, ErrorCodeUnauthorized, message, details)
}

// Forbidden 创建403错误
func Forbidden(message string, details interface{}) *AppError {
	return NewAppError(http.StatusForbidden, ErrorCodeForbidden, message, details)
}

// NotFound 创建404错误
func NotFound(message string, details interface{}) *AppError {
	return NewAppError(http.StatusNotFound, ErrorCodeNotFound, message, details)
}

// InternalServerError 创建500错误
func InternalServerError(message string, details interface{}) *AppError {
	return NewAppError(http.StatusInternalServerError, ErrorCodeInternal, message, details)
}

// ValidationError 创建参数验证错误
func ValidationError(message string, details interface{}) *AppError {
	return NewAppError(http.StatusBadRequest, ErrorCodeValidation, message, details)
}

// DatabaseError 创建数据库错误
func DatabaseError(message string, details interface{}) *AppError {
	return NewAppError(http.StatusInternalServerError, ErrorCodeDatabase, message, details)
}

// TokenInvalid 创建Token无效错误
func TokenInvalid(message string, details interface{}) *AppError {
	return NewAppError(http.StatusUnauthorized, ErrorCodeTokenInvalid, message, details)
}

// TokenExpired 创建Token过期错误
func TokenExpired(message string, details interface{}) *AppError {
	return NewAppError(http.StatusUnauthorized, ErrorCodeTokenExpired, message, details)
}

// PermissionDenied 创建权限不足错误
func PermissionDenied(message string, details interface{}) *AppError {
	return NewAppError(http.StatusForbidden, ErrorCodePermissionDenied, message, details)
}

// AlreadyExists 创建资源已存在错误
func AlreadyExists(message string, details interface{}) *AppError {
	return NewAppError(http.StatusBadRequest, ErrorCodeAlreadyExists, message, details)
}

// OperationFailed 创建操作失败错误
func OperationFailed(message string, details interface{}) *AppError {
	return NewAppError(http.StatusBadRequest, ErrorCodeOperationFailed, message, details)
}
