package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"

	http "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"
)

// CircuitBreakerMiddleware 熔断中间件
func CircuitBreakerMiddleware(name string, settings gobreaker.Settings) gin.HandlerFunc {
	// 创建断路器
	cb := gobreaker.NewCircuitBreaker(settings)

	return func(c *gin.Context) {
		// 使用断路器包装请求处理
		result, err := cb.Execute(func() (interface{}, error) {
			// 继续处理请求
			c.Next()

			// 检查是否有错误
			if len(c.Errors) > 0 {
				return nil, c.Errors.Last()
			}

			return nil, nil
		})

		// 处理断路器错误
		if err != nil {
			switch err {
			case gobreaker.ErrOpenState:
				// 断路器打开状态，拒绝请求
				http.ServiceUnavailableResponse(c, fmt.Sprintf("服务暂时不可用，请稍后再试: %s", err.Error()))
				c.Abort()
				return
			case gobreaker.ErrTooManyRequests:
				// 请求过多，拒绝请求
				http.TooManyRequestsResponse(c, fmt.Sprintf("请求过多，请稍后再试: %s", err.Error()))
				c.Abort()
				return
			default:
				// 其他错误，返回内部服务器错误
				http.InternalServerErrorResponse(c, fmt.Sprintf("服务内部错误: %s", err.Error()))
				c.Abort()
				return
			}
		}

		// 设置结果到上下文（如果有）
		if result != nil {
			c.Set("breaker_result", result)
		}
	}
}

// DefaultCircuitBreakerSettings 默认的断路器设置
func DefaultCircuitBreakerSettings() gobreaker.Settings {
	return gobreaker.Settings{
		Name:        "GinCircuitBreaker",
		MaxRequests: 3,
		Interval:    60000,
		Timeout:     30000,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 失败率超过50%时跳闸
			total := counts.Requests
			if total < 5 {
				return false
			}
			return float64(counts.TotalFailures)/float64(total) >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// 可以在这里添加状态变化的日志
			fmt.Printf("CircuitBreaker %s state changed from %s to %s\n", name, from, to)
		},
	}
}
