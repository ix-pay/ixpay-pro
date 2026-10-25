package entity

// 医保订单信息（对齐 Gpay model.MedOrderInfo）
// 创建时由业务方（医院 HIS 前置流程）传入，以 JSON 存储于 wx_payments.med_info
// 注意：json 标签仅用于 med_info 列的内部序列化/反序列化（读写对称），对外 API 契约由 dto 层定义

// MedOrderInfo 医保订单信息(创建时由业务方传入)
type MedOrderInfo struct {
	OrderType   string `json:"orderType"`   // 订单类型: REG_PAY挂号/DIAG_PAY诊间/IN_HOSP_PAY住院等
	SerialNo    string `json:"serialNo"`    // 医疗机构订单号(如HIS订单号，需与6201费用明细上传medOrgOrd一致)
	PayOrderID  string `json:"payOrderId"`  // 医保局返回的支付单ID(纯自费禁止填)
	PayAuthNo   string `json:"payAuthNo"`   // 医保局返回的支付授权码(纯自费禁止填)
	GeoLocation string `json:"geoLocation"` // 用户定位 经度,纬度(含医保必填)

	// 医疗机构透传给医保的数据(沿用地方医保局定义，不含payAuthNo/payOrderId/setlLatlnt)
	PassthroughRequestContent string `json:"passthroughRequestContent,omitempty"`

	MedInstName string `json:"medInstName"` // 医疗机构名称(服务端从医保配置快照，用于待结算详情页展示)

	// 支付人身份(本人)
	PayerName   string `json:"payerName"`   // 真实姓名
	PayerIDCard string `json:"payerIdCard"` // 身份证号
	// 代亲属支付
	PayForRelatives bool   `json:"payForRelatives"`
	RelativeName    string `json:"relativeName,omitempty"`
	RelativeIDCard  string `json:"relativeIdCard,omitempty"`

	// 医保费用明细(单位: 分)
	MedInsGovFee     int64 `json:"medInsGovFee"`     // 医保统筹支付金额
	MedInsSelfFee    int64 `json:"medInsSelfFee"`    // 医保个账支付金额
	MedInsOtherFee   int64 `json:"medInsOtherFee"`   // 医保其他支付金额
	MedInsCashFee    int64 `json:"medInsCashFee"`    // 医保结算后需自费金额
	WechatPayCashFee int64 `json:"wechatPayCashFee"` // 实际微信支付金额(有自费必填)

	// 现金补充/减免
	CashAddDetail    []MedCashAddItem    `json:"cashAddDetail,omitempty"`
	CashReduceDetail []MedCashReduceItem `json:"cashReduceDetail,omitempty"`

	// 渠道号(如有)
	ChannelNo string `json:"channelNo,omitempty"`
}

// MedCashAddItem 现金补充项(运费等)
type MedCashAddItem struct {
	CashAddFee  int64  `json:"cashAddFee"`
	CashAddType string `json:"cashAddType"` // DEFAULT_ADD_TYPE/FREIGHT/OTHER_MEDICAL_EXPENSES
}

// MedCashReduceItem 现金减免项(预缴金等)
type MedCashReduceItem struct {
	CashReduceFee  int64  `json:"cashReduceFee"`
	CashReduceType string `json:"cashReduceType"` // DEFAULT_REDUCE_TYPE/HOSPITAL_REDUCE/PRE_PAYMENT等
}

// TotalFee 医保订单总金额(分)
// totalFee = 统筹 + 个账 + 其他 + 微信支付金额 + 现金减免
func (m *MedOrderInfo) TotalFee() int64 {
	total := m.MedInsGovFee + m.MedInsSelfFee + m.MedInsOtherFee + m.WechatPayCashFee
	for _, r := range m.CashReduceDetail {
		total += r.CashReduceFee
	}
	return total
}

// MixPayType 混合支付类型: CASH_ONLY纯自费/INSURANCE_ONLY纯医保/CASH_AND_INSURANCE混合
func (m *MedOrderInfo) MixPayType() string {
	hasIns := m.MedInsGovFee+m.MedInsSelfFee+m.MedInsOtherFee+m.MedInsCashFee > 0
	hasCash := m.WechatPayCashFee > 0
	switch {
	case hasIns && hasCash:
		return "CASH_AND_INSURANCE"
	case hasIns:
		return "INSURANCE_ONLY"
	default:
		return "CASH_ONLY"
	}
}
