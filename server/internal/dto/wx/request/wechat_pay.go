package request

// CreateUnifiedOrderRequest 微信统一下单请求参数
type CreateUnifiedOrderRequest struct {
	OrderID     string  `json:"order_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"required"`
}