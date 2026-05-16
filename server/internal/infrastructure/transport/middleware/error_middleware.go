package middleware

import (
	"log"
	"net/http"

	apperror "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/error"
	httpresponse "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"

	"github.com/gin-gonic/gin"
)

// ErrorMiddleware 统一错误处理中间件
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last().Err

			// 判断是否为自定义错误
			if appErr, ok := err.(*apperror.AppError); ok {
				// 自定义错误，返回标准化响应
				httpresponse.ErrorResponse(c, appErr.Code, appErr.Message, appErr)
				// 记录错误日志
				log.Printf("[ERROR] Business Error - HTTP Code: %d, Business Code: %d, Message: %s, Details: %+v",
					appErr.Code, appErr.BusinessCode, appErr.Message, appErr.Details)
				return
			}

			// 非自定义错误，转换为内部服务器错误
			// 创建默认的系统错误
			sysErr := apperror.NewAppError(
				http.StatusInternalServerError,
				apperror.ErrorCodeInternal,
				"Internal Server Error",
				err.Error(),
			)
			httpresponse.ErrorResponse(c, sysErr.Code, sysErr.Message, sysErr)
			// 记录系统错误日志
			log.Printf("[ERROR] System Error - URL: %s, Method: %s, Error: %+v",
				c.Request.URL.Path, c.Request.Method, err)
			return
		}
	}
}
