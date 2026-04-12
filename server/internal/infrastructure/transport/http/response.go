package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperror "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/error"
)

// Response 统一API响应格式
type Response struct {
	Code         int         `json:"code"`         // HTTP状态码
	BusinessCode int         `json:"businessCode"` // 业务错误码
	Message      string      `json:"message"`      // 响应消息
	Data         interface{} `json:"data"`         // 响应数据
	Error        string      `json:"error"`        // 错误信息（仅在错误响应中使用）
}

// SuccessResponse 返回成功响应
// 参数:
// - c: Gin上下文
// - code: HTTP状态码
// - message: 响应消息
// - data: 响应数据
func SuccessResponse(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Code:         code,
		BusinessCode: 0, // 0表示成功
		Message:      message,
		Data:         data,
		Error:        "",
	})
}

// ErrorResponse 返回错误响应
// 参数:
// - c: Gin上下文
// - code: HTTP状态码
// - message: 响应消息
// - err: 错误信息
func ErrorResponse(c *gin.Context, code int, message string, err error) {
	errorMsg := ""
	businessCode := 0

	if err != nil {
		errorMsg = err.Error()

		// 检查是否为AppError类型
		if appErr, ok := err.(*apperror.AppError); ok {
			businessCode = int(appErr.BusinessCode)
			// 如果AppError中包含HTTP状态码，使用它覆盖传入的code
			if appErr.Code != 0 {
				code = appErr.Code
			}
			// 如果AppError中包含消息，使用它覆盖传入的message
			if appErr.Message != "" {
				message = appErr.Message
			}
		}
	}

	c.JSON(code, Response{
		Code:         code,
		BusinessCode: businessCode,
		Message:      message,
		Data:         nil,
		Error:        errorMsg,
	})
}

// Ok 返回200 OK响应
// 参数:
// - c: Gin上下文
// - data: 响应数据
func Ok(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusOK, "success", data)
}

// Created 返回201 Created响应
// 参数:
// - c: Gin上下文
// - data: 响应数据
func Created(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusCreated, "created", data)
}

// BadRequest 返回400 Bad Request响应
// 参数:
// - c: Gin上下文
// - message: 响应消息
// - err: 错误信息
func BadRequest(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusBadRequest, message, err)
}

// Unauthorized 返回401 Unauthorized响应
// 参数:
// - c: Gin上下文
// - message: 响应消息
// - err: 错误信息
func Unauthorized(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusUnauthorized, message, err)
}

// Forbidden 返回403 Forbidden响应
// 参数:
// - c: Gin上下文
// - message: 响应消息
// - err: 错误信息
func Forbidden(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusForbidden, message, err)
}

// NotFound 返回404 Not Found响应
// 参数:
// - c: Gin上下文
// - message: 响应消息
// - err: 错误信息
func NotFound(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusNotFound, message, err)
}

// InternalServerError 返回500 Internal Server Error响应
// 参数:
// - c: Gin上下文
// - message: 响应消息
// - err: 错误信息
func InternalServerError(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusInternalServerError, message, err)
}

// ServiceUnavailable 返回503 Service Unavailable响应
// 参数:
// - c: Gin上下文
// - message: 响应消息
// - err: 错误信息
func ServiceUnavailable(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusServiceUnavailable, message, err)
}

// TooManyRequests 返回429 Too Many Requests响应
// 参数:
// - c: Gin上下文
// - message: 响应消息
// - err: 错误信息
func TooManyRequests(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusTooManyRequests, message, err)
}

// 以下是带下划线Response后缀的函数别名，用于兼容现有代码

// UnauthorizedResponse 是Unauthorized的别名
func UnauthorizedResponse(c *gin.Context, message string) {
	Unauthorized(c, message, nil)
}

// InternalServerErrorResponse 是InternalServerError的别名
func InternalServerErrorResponse(c *gin.Context, message string) {
	InternalServerError(c, message, nil)
}

// ForbiddenResponse 是Forbidden的别名
func ForbiddenResponse(c *gin.Context, message string) {
	Forbidden(c, message, nil)
}

// BadRequestResponse 是BadRequest的别名
func BadRequestResponse(c *gin.Context, message string) {
	BadRequest(c, message, nil)
}

// ServiceUnavailableResponse 是ServiceUnavailable的别名
func ServiceUnavailableResponse(c *gin.Context, message string) {
	ServiceUnavailable(c, message, nil)
}

// TooManyRequestsResponse 是TooManyRequests的别名
func TooManyRequestsResponse(c *gin.Context, message string) {
	TooManyRequests(c, message, nil)
}
