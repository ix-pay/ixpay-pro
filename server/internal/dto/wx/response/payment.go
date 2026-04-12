package response

// PaymentResponse 支付响应 DTO
type PaymentResponse struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id"`
	Description   string  `json:"description"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	// 微信支付相关参数
	WechatPayParams map[string]interface{} `json:"wechat_pay_params,omitempty"`
}

// PaymentListResponse 支付列表响应 DTO
type PaymentListResponse struct {
	List     []PaymentResponse `json:"list"`
	Total    int64             `json:"total"`
	Page     int64             `json:"page"`
	PageSize int64             `json:"pageSize"`
}
