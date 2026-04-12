package request

// CreatePaymentRequest 创建支付请求参数
type CreatePaymentRequest struct {
	OrderID       string  `json:"order_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,oneof=CNY USD"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=wechat_alipay"`
	Description   string  `json:"description"`
}

// GetPaymentRequest 获取支付详情请求参数
type GetPaymentRequest struct {
	ID string `json:"id" form:"id" binding:"required"`
}

// CancelPaymentRequest 取消支付请求参数
type CancelPaymentRequest struct {
	ID     string `json:"id" binding:"required"`
	Reason string `json:"reason"`
}

// GetPaymentListRequest 获取支付列表请求参数
type GetPaymentListRequest struct {
	Page      int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize  int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	StartTime string `json:"startTime" form:"startTime" binding:"omitempty"`
	EndTime   string `json:"endTime" form:"endTime" binding:"omitempty"`
	Status    string `json:"status" form:"status" binding:"omitempty"`
}
