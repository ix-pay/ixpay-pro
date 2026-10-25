package request

// 医保支付请求参数（金额单位：分，对齐微信 V3 医保接口）

// MedAuthURLRequest 医保免密授权链接请求参数
type MedAuthURLRequest struct {
	PayForRelatives bool   `json:"payForRelatives" form:"payForRelatives"` // 是否代亲人支付授权
	RelName         string `json:"relName" form:"relName"`                 // 亲属姓名(代亲人时必填)
	RelCardNo       string `json:"relCardNo" form:"relCardNo"`             // 亲属身份证号(代亲人时必填)
	Redirect        string `json:"redirect" form:"redirect"`               // 授权后回跳地址(空则用前端地址)
}

// MedAuthQueryRequest 医保授权查询请求参数(授权回跳 authCode 换 payAuthNo)
type MedAuthQueryRequest struct {
	AuthCode string `json:"authCode" binding:"required"`
}

// MedCashAddItem 现金补充项
type MedCashAddItem struct {
	CashAddFee  int64  `json:"cashAddFee"`
	CashAddType string `json:"cashAddType"`
}

// MedCashReduceItem 现金减免项
type MedCashReduceItem struct {
	CashReduceFee  int64  `json:"cashReduceFee"`
	CashReduceType string `json:"cashReduceType"`
}

// MedOrderInfoRequest 医保订单信息
type MedOrderInfoRequest struct {
	OrderType                 string              `json:"orderType"`   // REG_PAY挂号/DIAG_PAY诊间/IN_HOSP_PAY住院等
	SerialNo                  string              `json:"serialNo"`    // 医疗机构订单号(HIS订单号)
	PayOrderID                string              `json:"payOrderId"`  // 医保局支付单ID(纯自费禁止填)
	PayAuthNo                 string              `json:"payAuthNo"`   // 医保局支付授权码(纯自费禁止填)
	GeoLocation               string              `json:"geoLocation"` // 用户定位 经度,纬度(含医保必填)
	PassthroughRequestContent string              `json:"passthroughRequestContent"`
	PayerName                 string              `json:"payerName"`   // 支付人真实姓名
	PayerIDCard               string              `json:"payerIdCard"` // 支付人身份证号
	PayForRelatives           bool                `json:"payForRelatives"`
	RelativeName              string              `json:"relativeName"`
	RelativeIDCard            string              `json:"relativeIdCard"`
	MedInsGovFee              int64               `json:"medInsGovFee"`     // 医保统筹(分)
	MedInsSelfFee             int64               `json:"medInsSelfFee"`    // 医保个账(分)
	MedInsOtherFee            int64               `json:"medInsOtherFee"`   // 医保其他(分)
	MedInsCashFee             int64               `json:"medInsCashFee"`    // 结算后需自费(分)
	WechatPayCashFee          int64               `json:"wechatPayCashFee"` // 实际微信支付金额(分)
	CashAddDetail             []MedCashAddItem    `json:"cashAddDetail"`
	CashReduceDetail          []MedCashReduceItem `json:"cashReduceDetail"`
	ChannelNo                 string              `json:"channelNo"`
}

// CreateMedOrderRequest 创建医保订单请求参数
type CreateMedOrderRequest struct {
	Description string              `json:"description"`
	MedInfo     MedOrderInfoRequest `json:"medInfo" binding:"required"`
}

// MedPrepayRequest 医保下单请求参数
type MedPrepayRequest struct {
	OutTradeNo string `json:"outTradeNo" binding:"required"`
}

// MedQueryRequest 医保订单查询请求参数(二选一)
type MedQueryRequest struct {
	OutTradeNo string `json:"outTradeNo" form:"outTradeNo"`
	MixTradeNo string `json:"mixTradeNo" form:"mixTradeNo"`
}

// MedRefundCreateRequest 医保退款请求参数(金额单位：分)
type MedRefundCreateRequest struct {
	OutTradeNo      string `json:"outTradeNo" binding:"required"`
	HospOutRefundNo string `json:"hospOutRefundNo"`
	PartRefundType  string `json:"partRefundType" binding:"omitempty,oneof=CASH_ONLY INS_ONLY INS_PLAN_PART INS_INDIVIDUAL_PART"`
	CashRefundFee   int64  `json:"cashRefundFee"`
	InsRefundFee    int64  `json:"insRefundFee"`
	RefReason       string `json:"refReason"`
}

// MedRefundQueryRequest 医保退款查询请求参数(二选一)
type MedRefundQueryRequest struct {
	HospOutRefundNo string `json:"hospOutRefundNo" form:"hospOutRefundNo"`
	OutTradeNo      string `json:"outTradeNo" form:"outTradeNo"`
}
