package response

// PaymentResponse 支付响应 DTO
// ID 字段使用 int64 + json:",string" 标签，避免 JS 精度丢失
type PaymentResponse struct {
	ID            int64   `json:"id,string"`
	OrderID       string  `json:"orderId"`
	UserID        int64   `json:"userId,string"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"paymentMethod"`
	Status        string  `json:"status"`
	TransactionID string  `json:"transactionId"`
	Description   string  `json:"description"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	// 微信支付相关参数
	WechatPayParams map[string]interface{} `json:"wechatPayParams,omitempty"`
}

// PaymentListResponse 支付列表响应 DTO
type PaymentListResponse struct {
	List     []PaymentResponse `json:"list"`
	Total    int64             `json:"total"`
	Page     int64             `json:"page"`
	PageSize int64             `json:"pageSize"`
}
