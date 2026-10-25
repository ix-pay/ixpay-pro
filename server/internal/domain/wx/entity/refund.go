package entity

import "time"

// 退款单领域实体（对齐 Gpay model.RefundRecord）

// RefundStatus 退款状态
const (
	RefundStatusSuccess    = "SUCCESS"    // 退款成功
	RefundStatusClosed     = "CLOSED"     // 退款关闭(终态)
	RefundStatusProcessing = "PROCESSING" // 退款处理中
	RefundStatusAbnormal   = "ABNORMAL"   // 退款异常(需发起异常退款)
)

// RefundRecord 本地退款单（普通 JSAPI/Native 订单）
type RefundRecord struct {
	ID                  int64     // 退款单 ID
	OutRefundNo         string    // 商户退款单号
	OutTradeNo          string    // 关联商户订单号
	RefundID            string    // 微信退款单号
	Reason              string    // 退款原因
	RefundFee           int64     // 退款金额(分)
	TotalFee            int64     // 原订单金额(分)
	Status              string    // SUCCESS/CLOSED/PROCESSING/ABNORMAL
	Applied             bool      // 退款金额已累计进订单RefundedFee(防并发双计)
	UserReceivedAccount string    // 退款接收账户
	SuccessTime         string    // 退款成功时间
	WxLastResponse      string    // 微信侧原始应答
	CreatedAt           time.Time // 创建时间
	UpdatedAt           time.Time // 更新时间
}

// MedRefundRecord 本地医保退款单（对齐 Gpay model.MedRefundRecord）
type MedRefundRecord struct {
	ID                    int64     // 退款单 ID
	HospOutRefundNo       string    // 主键: 医院退款订单号
	OutTradeNo            string    // 本地订单号(hosp_out_trade_no)
	PartRefundType        string    // 部分退款类型
	CashRefundFee         int64     // 现金退款金额(分)
	InsRefundFee          int64     // 医保退款金额(分，统筹+个账+其他)
	RequestContent        string    // 请求内容留档
	Reason                string    // 退费原因
	MedRefundID           string    // 微信退款医疗订单号(预留)
	CashRefundID          string    // 微信支付退款单号(现金部分)
	InsuranceRefundID     string    // 医保退款单号(预留)
	MedRefundState        string    // 退款单总状态 SUCCESS/REFUNDING
	CashRefundStatus      string    // 现金部分状态 SUCCESS/REFUNDING/CHANGE/FAIL
	InsuranceRefundStatus string    // 医保部分状态 SUCCESS/REFUNDING/FAIL
	RefundEndTime         string    // 退款完成时间
	ResponseContent       string    // 医保局返回内容串
	CashApplied           bool      // 现金部分退款金额已累计进订单(防并发双计)
	InsApplied            bool      // 医保部分退款金额已累计进订单(防并发双计)
	LastResponse          string    // 最近一次微信应答/本地状态原文
	CreatedAt             time.Time // 创建时间
	UpdatedAt             time.Time // 更新时间
}
