package wxpayv3

import "time"

// 微信支付协议模型（与 Gpay 项目 model 包对齐，供协议层与服务层/DTO 共用）

// CashAddItem 现金补充项(运费等)
type CashAddItem struct {
	CashAddFee  int64  `json:"cash_add_fee"`
	CashAddType string `json:"cash_add_type"` // DEFAULT_ADD_TYPE/FREIGHT/OTHER_MEDICAL_EXPENSES
}

// CashReduceItem 现金减免项(预缴金等)
type CashReduceItem struct {
	CashReduceFee  int64  `json:"cash_reduce_fee"`
	CashReduceType string `json:"cash_reduce_type"` // DEFAULT_REDUCE_TYPE/HOSPITAL_REDUCE/PRE_PAYMENT等
}

// ==================== 医保支付 ====================

// MedInsPrepayRequest 医保自费混合收款下单请求 POST /v3/med-ins/orders
// 直连与服务商URL相同；服务商模式差异: appid=服务商公众号、必填sub_appid(医疗机构公众号)+sub_mchid，
// openid/sub_openid二选一(分别为服务商/医疗机构公众号下的openid)
type MedInsPrepayRequest struct {
	MixPayType      string       `json:"mix_pay_type"` // CASH_ONLY/INSURANCE_ONLY/CASH_AND_INSURANCE
	OrderType       string       `json:"order_type"`   // REG_PAY/DIAG_PAY/IN_HOSP_PAY等
	AppID           string       `json:"appid"`
	SubAppID        string       `json:"sub_appid,omitempty"`  // 服务商模式: 医疗机构公众号APPID
	Openid          string       `json:"openid,omitempty"`     // 服务商公众号openid(服务商模式)
	SubOpenid       string       `json:"sub_openid,omitempty"` // 医疗机构公众号openid(服务商模式)
	SubMchID        string       `json:"sub_mchid,omitempty"`  // 服务商模式: 子商户号
	Payer           *MedInsPayer `json:"payer"`
	PayForRelatives bool         `json:"pay_for_relatives,omitempty"`
	Relative        *MedInsPayer `json:"relative,omitempty"`
	OutTradeNo      string       `json:"out_trade_no"`
	SerialNo        string       `json:"serial_no"`
	PayOrderID      string       `json:"pay_order_id,omitempty"`
	PayAuthNo       string       `json:"pay_auth_no,omitempty"`
	GeoLocation     string       `json:"geo_location,omitempty"`
	CityID          string       `json:"city_id"`
	MedInstName     string       `json:"med_inst_name"`
	MedInstNo       string       `json:"med_inst_no"`

	MedInsOrderCreateTime string `json:"med_ins_order_create_time,omitempty"` // rfc3339
	TotalFee              int64  `json:"total_fee"`

	MedInsGovFee     *int64 `json:"med_ins_gov_fee,omitempty"`
	MedInsSelfFee    *int64 `json:"med_ins_self_fee,omitempty"`
	MedInsOtherFee   *int64 `json:"med_ins_other_fee,omitempty"`
	MedInsCashFee    *int64 `json:"med_ins_cash_fee,omitempty"`
	WechatPayCashFee *int64 `json:"wechat_pay_cash_fee,omitempty"`

	CashAddDetail    []CashAddItem    `json:"cash_add_detail,omitempty"`
	CashReduceDetail []CashReduceItem `json:"cash_reduce_detail,omitempty"`

	CallbackURL string `json:"callback_url"`
	PrepayID    string `json:"prepay_id,omitempty"` // 有自费部分时必填(JSAPI下单产物)

	PassthroughRequestContent string `json:"passthrough_request_content,omitempty"` // 医疗机构透传给医保的数据
	ChannelNo                 string `json:"channel_no,omitempty"`
	Attach                    string `json:"attach,omitempty"`
	Extends                   string `json:"extends,omitempty"` // 亲情移动支付: {"rel_user_name_md5":"...","rel_user_card_no_md5":"..."}
	MedInsTestEnv             bool   `json:"med_ins_test_env,omitempty"`
}

// MedInsPayer 医保支付人身份信息(敏感字段需使用微信支付公钥加密后传入)
type MedInsPayer struct {
	Name     string `json:"name"`      // RSA-OAEP-SHA1加密后的base64
	IDDigest string `json:"id_digest"` // 身份证MD5摘要(小写hex)再RSA-OAEP加密后的base64
	CardType string `json:"card_type,omitempty"`
}

// MedUserQueryResponse wecity医保授权查询接口应答(授权码换取payAuthNo)
type MedUserQueryResponse struct {
	Code        int          `json:"code"` // 0成功 小于0系统错误 大于0逻辑错误
	Message     string       `json:"message"`
	UserName    string       `json:"user_name,omitempty"` // 用户姓名
	CityID      string       `json:"city_id,omitempty"`   // 用户参保地代码
	PayAuthNo   string       `json:"pay_auth_no"`         // 医保线上核验payAuthNo
	GeoLocation *MedGeoPoint `json:"user_longitude_latitude"`
}

// MedGeoPoint 用户经纬度
type MedGeoPoint struct {
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}

// MedInsOrderResponse 医保下单/查单/回调业务数据(新版V3，直连与服务商同构，服务商多sub_字段)
type MedInsOrderResponse struct {
	MixTradeNo      string `json:"mix_trade_no"`
	MixPayStatus    string `json:"mix_pay_status"`     // MIX_PAY_CREATED/MIX_PAY_SUCCESS/MIX_PAY_REFUND/MIX_PAY_FAIL
	SelfPayStatus   string `json:"self_pay_status"`    // SELF_PAY_CREATED/SELF_PAY_SUCCESS/SELF_PAY_REFUND/SELF_PAY_FAIL/NO_SELF_PAY
	MedInsPayStatus string `json:"med_ins_pay_status"` // MED_INS_PAY_CREATED/MED_INS_PAY_SUCCESS/MED_INS_PAY_REFUND/MED_INS_PAY_FAIL/NO_MED_INS_PAY
	PaidTime        string `json:"paid_time,omitempty"`

	PassthroughResponseContent string `json:"passthrough_response_content,omitempty"`

	MixPayType  string `json:"mix_pay_type,omitempty"`
	OrderType   string `json:"order_type,omitempty"`
	AppID       string `json:"appid,omitempty"`
	SubAppID    string `json:"sub_appid,omitempty"`
	Openid      string `json:"openid,omitempty"`
	SubOpenid   string `json:"sub_openid,omitempty"`
	OutTradeNo  string `json:"out_trade_no,omitempty"`
	SubMchID    string `json:"sub_mchid,omitempty"`
	SerialNo    string `json:"serial_no,omitempty"`
	PayOrderID  string `json:"pay_order_id,omitempty"`
	PayAuthNo   string `json:"pay_auth_no,omitempty"`
	GeoLocation string `json:"geo_location,omitempty"`
	CityID      string `json:"city_id,omitempty"`

	MedInstName           string `json:"med_inst_name,omitempty"`
	MedInstNo             string `json:"med_inst_no,omitempty"`
	MedInsOrderCreateTime string `json:"med_ins_order_create_time,omitempty"`

	TotalFee         int64  `json:"total_fee,omitempty"`
	MedInsGovFee     int64  `json:"med_ins_gov_fee,omitempty"`
	MedInsSelfFee    int64  `json:"med_ins_self_fee,omitempty"`
	MedInsOtherFee   int64  `json:"med_ins_other_fee,omitempty"`
	MedInsCashFee    int64  `json:"med_ins_cash_fee,omitempty"`
	WechatPayCashFee int64  `json:"wechat_pay_cash_fee,omitempty"`
	MedInsFailReason string `json:"med_ins_fail_reason,omitempty"` // 仅MED_INS_PAY_FAIL时返回

	CashAddDetail    []CashAddItem    `json:"cash_add_detail,omitempty"`
	CashReduceDetail []CashReduceItem `json:"cash_reduce_detail,omitempty"`

	CallbackURL string `json:"callback_url,omitempty"`
	PrepayID    string `json:"prepay_id,omitempty"`

	PassthroughRequestContent string `json:"passthrough_request_content,omitempty"`
	Extends                   string `json:"extends,omitempty"`
	Attach                    string `json:"attach,omitempty"`
	ChannelNo                 string `json:"channel_no,omitempty"`
	MedInsTestEnv             bool   `json:"med_ins_test_env,omitempty"`
}

// MedInsPayParams 前端调起医保支付参数
type MedInsPayParams struct {
	MixTradeNo string `json:"mixTradeNo"`
	AppID      string `json:"appid"`
	// 以下自费部分必填
	TimeStamp string `json:"timeStamp,omitempty"`
	NonceStr  string `json:"nonceStr,omitempty"`
	Package   string `json:"package,omitempty"`
	SignType  string `json:"signType,omitempty"`
	PaySign   string `json:"paySign,omitempty"`
}

// ==================== 退款 ====================

// 退款状态
const (
	RefundStatusSuccess    = "SUCCESS"    // 退款成功
	RefundStatusClosed     = "CLOSED"     // 退款关闭(终态)
	RefundStatusProcessing = "PROCESSING" // 退款处理中
	RefundStatusAbnormal   = "ABNORMAL"   // 退款异常(需发起异常退款)
)

// RefundCreateRequest 退款申请请求 POST /v3/refund/domestic/refunds
type RefundCreateRequest struct {
	OutTradeNo    string        `json:"out_trade_no"`             // 商户订单号(与transaction_id二选一)
	TransactionID string        `json:"transaction_id,omitempty"` // 微信支付订单号
	OutRefundNo   string        `json:"out_refund_no"`            // 商户退款单号
	Reason        string        `json:"reason,omitempty"`         // 退款原因
	NotifyURL     string        `json:"notify_url"`               // 退款结果回调地址
	SubMchID      string        `json:"sub_mchid,omitempty"`      // 服务商模式: 子商户号(必填)
	Amount        *RefundAmount `json:"amount"`
}

// RefundAmount 退款金额信息
type RefundAmount struct {
	Refund   int64  `json:"refund"`             // 退款金额(分)
	Total    int64  `json:"total"`              // 原订单总金额(分)
	Currency string `json:"currency,omitempty"` // CNY
}

// RefundResponse 退款申请/查询退款应答(退款结果回调解密后结构一致)
type RefundResponse struct {
	RefundID            string              `json:"refund_id"`
	OutRefundNo         string              `json:"out_refund_no"`
	TransactionID       string              `json:"transaction_id"`
	OutTradeNo          string              `json:"out_trade_no"`
	Channel             string              `json:"channel,omitempty"`
	UserReceivedAccount string              `json:"user_received_account,omitempty"`
	SuccessTime         string              `json:"success_time,omitempty"`
	CreateTime          string              `json:"create_time,omitempty"`
	Status              string              `json:"status"` // SUCCESS/CLOSED/PROCESSING/ABNORMAL
	FundsAccount        string              `json:"funds_account,omitempty"`
	Amount              *RefundAmountDetail `json:"amount,omitempty"`
}

// RefundAmountDetail 退款应答金额明细
type RefundAmountDetail struct {
	Total            int64  `json:"total"`
	Refund           int64  `json:"refund"`
	PayerTotal       int64  `json:"payer_total,omitempty"`
	PayerRefund      int64  `json:"payer_refund,omitempty"`
	SettlementTotal  int64  `json:"settlement_total,omitempty"`
	SettlementRefund int64  `json:"settlement_refund,omitempty"`
	DiscountRefund   int64  `json:"discount_refund,omitempty"`
	Currency         string `json:"currency"`
}

// RefundNotifyData 退款结果回调(REFUND.SUCCESS)解密后业务数据
type RefundNotifyData struct {
	MchID               string              `json:"mchid"`
	OutTradeNo          string              `json:"out_trade_no"`
	TransactionID       string              `json:"transaction_id"`
	OutRefundNo         string              `json:"out_refund_no"`
	RefundID            string              `json:"refund_id"`
	RefundStatus        string              `json:"refund_status"` // SUCCESS/CLOSED/ABNORMAL
	SuccessTime         string              `json:"success_time,omitempty"`
	Amount              *RefundAmountDetail `json:"amount,omitempty"`
	UserReceivedAccount string              `json:"user_received_account,omitempty"`
}

// AbnormalRefundRequest 发起异常退款请求 POST /v3/refund/domestic/refunds/{refund_id}/apply-abnormal-refund
type AbnormalRefundRequest struct {
	OutRefundNo string `json:"out_refund_no"`          // 商户退款单号
	Type        string `json:"type"`                   // USER_BANK_CARD退用户银行卡/MERCHANT_BANK_CARD退商户银行账户
	BankType    string `json:"bank_type,omitempty"`    // 开户银行(退用户必填, 如ICBC_DEBIT)
	BankAccount string `json:"bank_account,omitempty"` // 收款银行卡号(敏感,需微信支付公钥加密;退用户必填)
	RealName    string `json:"real_name,omitempty"`    // 收款用户姓名(敏感,需加密;退用户必填)
	SubMchID    string `json:"sub_mchid,omitempty"`    // 服务商模式: 子商户号(必填)
}

// 异常退款处理方式
const (
	AbnormalRefundToUser     = "USER_BANK_CARD"     // 退款到用户银行卡
	AbnormalRefundToMerchant = "MERCHANT_BANK_CARD" // 退款至交易商户银行账户
)

// ==================== 医保退款 ====================

// 医保部分退款类型(part_refund_type，服务端退款拆分决策用)
const (
	MedRefundAll               = ""                    // 不填则全退
	MedRefundCashOnly          = "CASH_ONLY"           // 只退现金部分
	MedRefundInsuranceOnly     = "INS_ONLY"            // 全额退医保部分(统筹+个账+其他)
	MedRefundInsPlanPart       = "INS_PLAN_PART"       // 只退医保统筹部分(仅支持一次性退款)
	MedRefundInsIndividualPart = "INS_INDIVIDUAL_PART" // 只退医保个账部分(支持多次退款)
)

// 医保退款状态
const (
	MedRefundStateSuccess   = "SUCCESS"   // 退款成功
	MedRefundStateRefunding = "REFUNDING" // 退款处理中
	MedRefundStateFail      = "FAIL"      // 退款失败(医保退款)
	MedRefundStateChange    = "CHANGE"    // 转入代发(现金退款，需商户人工干预)
)

// MedInsRefundNotifyRequest 医保退款通知 POST /v3/med-ins/refunds/notify?mix_trade_no={mix_trade_no}
// 仅退医保部分(统筹/个账/其他)；现金部分(wechat_pay_cash_fee)走CreateRefund普通退款接口
// 规则: med_refund_total_fee = med_refund_gov_fee + med_refund_self_fee + med_refund_other_fee
type MedInsRefundNotifyRequest struct {
	SubMchID          string `json:"sub_mchid,omitempty"`  // 子商户号(服务商模式必填，直连不传)
	OutRefundNo       string `json:"out_refund_no"`        // 医疗机构退款单号
	RefundTime        string `json:"refund_time"`          // 医保退款成功时间 rfc3339(必填)
	MedRefundTotalFee int64  `json:"med_refund_total_fee"` // 医保退款总金额(分，必填)
	MedRefundGovFee   int64  `json:"med_refund_gov_fee"`   // 医保统筹退款金额(分，必填，0也需传)
	MedRefundSelfFee  int64  `json:"med_refund_self_fee"`  // 医保个账退款金额(分，必填，0也需传)
	MedRefundOtherFee int64  `json:"med_refund_other_fee"` // 医保其他退款金额(分，必填，0也需传)
}

// ==================== 账单 ====================

// TradeBillRequest 申请交易账单请求 GET /v3/bill/tradebill
type TradeBillRequest struct {
	BillDate string `json:"bill_date"` // 账单日期 yyyy-MM-DD(三个月内,非当天)
	BillType string `json:"bill_type"` // ALL/SUCCESS/REFUND/RECHARGE_REFUND
}

// FundFlowBillRequest 申请资金账单请求 GET /v3/bill/fundflowbill
type FundFlowBillRequest struct {
	BillDate    string `json:"bill_date"`    // 账单日期 yyyy-MM-DD
	AccountType string `json:"account_type"` // BASIC基本账户/OPERATION运营账户/FEES手续费账户
}

// BillResponse 申请账单应答(返回下载链接，不返回账单数据)
type BillResponse struct {
	HashType    string `json:"hash_type"`    // SHA1
	HashValue   string `json:"hash_value"`   // 账单文件摘要(下载后校验)
	DownloadURL string `json:"download_url"` // 下载链接(5分钟有效)
}

// NotifyResponse 回调应答(失败时返回)
type NotifyResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ExpireAtNow 医保单有效期固定2小时(本地对账判断用)
func ExpireAtNow() time.Time {
	return time.Now().Add(2 * time.Hour)
}
