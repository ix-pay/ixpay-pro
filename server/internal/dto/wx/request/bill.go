package request

// 账单请求参数（对齐微信 V3 账单接口）

// TradeBillRequest 申请交易账单请求参数
type TradeBillRequest struct {
	BillDate string `json:"billDate" binding:"required"` // yyyy-MM-DD(三个月内,非当天)
	BillType string `json:"billType" binding:"omitempty,oneof=ALL SUCCESS REFUND RECHARGE_REFUND"`
}

// FundFlowBillRequest 申请资金账单请求参数
type FundFlowBillRequest struct {
	BillDate    string `json:"billDate" binding:"required"` // yyyy-MM-DD
	AccountType string `json:"accountType" binding:"omitempty,oneof=BASIC OPERATION FEES"`
}

// DownloadBillRequest 下载账单请求参数
type DownloadBillRequest struct {
	DownloadURL string `json:"downloadUrl" binding:"required"` // 申请账单应答返回的下载链接
	HashValue   string `json:"hashValue"`                      // 申请账单应答返回的SHA1摘要(校验用)
}
