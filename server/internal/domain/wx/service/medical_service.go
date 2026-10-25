package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	wxpayv3 "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayv3"
)

// MedicalService 微信医保支付编排（对齐 Gpay service/medical.go，仅微信部分：
// 自费JSAPI下单、医保自费混合下单、查单、收款回调；医保授权/费用明细上传由业务系统前置完成）
// 本地订单存储于 wx_payments（Payment 实体，order_type=medical，医保信息 JSON 存 med_info 列）
type MedicalService struct {
	paymentRepo      repo.PaymentRepository
	wechatPayService *WechatPayService
	log              logger.Logger
}

// NewMedicalService 创建医保支付服务
func NewMedicalService(paymentRepo repo.PaymentRepository, wechatPayService *WechatPayService, log logger.Logger) *MedicalService {
	return &MedicalService{paymentRepo: paymentRepo, wechatPayService: wechatPayService, log: log}
}

// MedAuthURL 生成医保支付授权链接(免密授权页)
// payForRelatives=true时为代亲人支付授权(familyId=MD5(亲属姓名+身份证后4位))
// 用户微信内打开完成授权后回跳 redirectBase?retCode=0&authCode=xxx
func (s *MedicalService) MedAuthURL(payForRelatives bool, relName, relIDCard, redirectBase string) (string, error) {
	cfg := s.wechatPayService.GetConfig()
	if cfg.MedPartnerID == "" || cfg.MedPartnerSecret == "" {
		return "", fmt.Errorf("医保授权未配置: wechat_med_partner_id / wechat_med_partner_secret")
	}
	if cfg.MedOrgChnlCrtfCodg == "" || cfg.MedOrgAppID == "" {
		return "", fmt.Errorf("医保授权未配置: wechat_med_org_chnl_crtf_codg / wechat_med_org_app_id (来源于国家局医保移动支付反馈单)")
	}
	if redirectBase == "" {
		redirectBase = cfg.FrontendURL + "/"
	}
	familyID := ""
	if payForRelatives {
		if relName == "" || relIDCard == "" {
			return "", fmt.Errorf("代亲人支付授权需提供亲属姓名与身份证号")
		}
		familyID = wxpayv3.FamilyIDDigest(relName, relIDCard)
	}
	return wxpayv3.BuildMedAuthLink(&wxpayv3.MedAuthLinkInput{
		AppID:           cfg.AppID,
		CityCode:        cfg.MedCityID,
		Channel:         cfg.MedChannelNo,
		OrgChnlCrtfCodg: cfg.MedOrgChnlCrtfCodg,
		OrgCodg:         cfg.MedInstNo,
		OrgAppID:        cfg.MedOrgAppID,
		RedirectURL:     redirectBase,
		FamilyID:        familyID,
		TestEnv:         cfg.MedTestEnv,
	}), nil
}

// MedQueryAuth 授权查询: 用授权回跳的authCode换取payAuthNo与用户经纬度
// 亲情支付授权码同样通过此接口查询(返回的是代亲属支付的payAuthNo)
func (s *MedicalService) MedQueryAuth(authCode, openid string) (*wxpayv3.MedUserQueryResponse, error) {
	if authCode == "" || openid == "" {
		return nil, fmt.Errorf("授权码(auth_code)与openid不能为空")
	}
	cfg := s.wechatPayService.GetConfig()
	if cfg.MedPartnerID == "" || cfg.MedPartnerSecret == "" {
		return nil, fmt.Errorf("医保授权未配置: wechat_med_partner_id / wechat_med_partner_secret")
	}
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	return client.MedUserQuery(cfg.MedPartnerID, cfg.MedPartnerSecret, cfg.MedTestEnv, authCode, openid)
}

// MedCreateOrder 创建本地医保订单（wx_payments，order_type=medical）
// payAuthNo/payOrderId/serialNo/geoLocation及费用明细由业务系统(医院HIS)前置流程产出后传入
func (s *MedicalService) MedCreateOrder(userID int64, description string, med *entity.MedOrderInfo) (*entity.Payment, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	if med == nil {
		return nil, fmt.Errorf("医保订单信息不能为空")
	}
	if med.SerialNo == "" {
		return nil, fmt.Errorf("医疗机构订单号(serial_no)不能为空")
	}
	if med.PayerName == "" || med.PayerIDCard == "" {
		return nil, fmt.Errorf("支付人姓名与身份证号不能为空")
	}
	mixType := med.MixPayType()
	if mixType != "CASH_ONLY" {
		if med.PayOrderID == "" || med.PayAuthNo == "" {
			return nil, fmt.Errorf("含医保支付时必须传入医保局支付单ID(pay_order_id)与支付授权码(pay_auth_no)")
		}
		if med.GeoLocation == "" {
			return nil, fmt.Errorf("含医保支付时必须传入用户定位(geo_location，格式: 经度,纬度)")
		}
	}
	if mixType != "INSURANCE_ONLY" && med.WechatPayCashFee <= 0 {
		return nil, fmt.Errorf("含自费支付时微信支付金额(wechat_pay_cash_fee)必须大于0")
	}
	// 易错点: med_ins_cash_fee>0时须满足 cashFee+Σ现金补充 == wechatPayCashFee+Σ现金减免
	if med.MedInsCashFee > 0 {
		cashAdd, cashReduce := int64(0), int64(0)
		for _, a := range med.CashAddDetail {
			cashAdd += a.CashAddFee
		}
		for _, r := range med.CashReduceDetail {
			cashReduce += r.CashReduceFee
		}
		if med.MedInsCashFee+cashAdd != med.WechatPayCashFee+cashReduce {
			return nil, fmt.Errorf("金额校验失败: 医保自费金额(%d)+现金补充(%d) 必须等于 微信支付金额(%d)+现金减免(%d)",
				med.MedInsCashFee, cashAdd, med.WechatPayCashFee, cashReduce)
		}
	}
	if description == "" {
		description = "医疗费用"
	}
	if med.OrderType == "" {
		med.OrderType = "DIAG_PAY" // 默认诊间支付
	}
	// 快照医疗机构名称，供前端待结算详情页展示
	cfg := s.wechatPayService.GetConfig()
	if med.MedInstName == "" {
		med.MedInstName = cfg.MedInstName
	}

	// 医保单号即商户订单号（out_trade_no）
	orderID := genOrderID()
	payment := &entity.Payment{
		OrderID:     orderID,
		UserID:      userID,
		Amount:      med.WechatPayCashFee, // 自费部分金额
		Currency:    "CNY",
		Method:      "wechat",
		Status:      entity.PaymentStatusPending,
		Description: description,
		OrderType:   entity.PaymentOrderTypeMedical,
		MedInfo:     med,
	}
	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, fmt.Errorf("保存医保订单失败: %w", err)
	}
	s.log.Info("医保订单创建成功", "orderID", orderID, "mixPayType", mixType, "totalFee", med.TotalFee())
	return payment, nil
}

// MedPrepay 医保自费混合下单，返回前端调起医保支付参数
// 流程: 含自费部分时先JSAPI下单获取prepay_id(与混合下单使用同一out_trade_no)，再调用医保混合下单
func (s *MedicalService) MedPrepay(userID int64, openID, outTradeNo string) (*wxpayv3.MedInsPayParams, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	payment, err := s.paymentRepo.GetByOrderID(outTradeNo)
	if err != nil {
		return nil, fmt.Errorf("医保订单不存在: %s", outTradeNo)
	}
	if payment.OrderType != entity.PaymentOrderTypeMedical {
		return nil, fmt.Errorf("非医保订单，请使用普通支付下单接口")
	}
	if payment.Status == entity.PaymentStatusSuccess {
		return nil, fmt.Errorf("订单已支付，请勿重复支付")
	}
	if payment.MedInfo == nil {
		return nil, fmt.Errorf("订单缺少医保信息")
	}
	if payment.UserID != userID {
		return nil, fmt.Errorf("无权支付该订单")
	}
	if payment.Openid != "" && payment.Openid != openID {
		return nil, fmt.Errorf("下单账号与支付账号不一致")
	}
	med := payment.MedInfo
	cfg := s.wechatPayService.GetConfig()
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}

	mixType := med.MixPayType()

	// 1) 自费部分: 先JSAPI下单拿prepay_id
	var prepayID string
	var cashParams *wxpayv3.JSAPIPayParams
	if mixType != "INSURANCE_ONLY" {
		appID := cfg.AppID
		if appID == "" {
			return nil, fmt.Errorf("微信 AppID 未配置")
		}
		prepayID, err = client.JSAPIPrepay(appID, openID, outTradeNo, payment.Description, cfg.NotifyURL, payment.Amount)
		if err != nil {
			return nil, fmt.Errorf("自费部分JSAPI下单失败: %w", err)
		}
		cashParams, err = client.BuildJSAPIPayParams(prepayID, appID)
		if err != nil {
			return nil, err
		}
		// 保存自费部分调起参数
		if err := s.saveMedPrepayInfo(payment.ID, appID, prepayID, cashParams); err != nil {
			s.log.Warn("保存自费支付参数失败", "error", err)
		}
	}

	// 2) 加密支付人身份信息（官方工具类 OAEP + 身份证 MD5 摘要）
	payer, err := client.BuildMedInsPayer(med.PayerName, med.PayerIDCard)
	if err != nil {
		return nil, err
	}

	// 3) 组装医保混合下单请求
	req := &wxpayv3.MedInsPrepayRequest{
		MixPayType:    mixType,
		OrderType:     med.OrderType,
		AppID:         cfg.AppID,
		Openid:        openID,
		Payer:         payer,
		OutTradeNo:    outTradeNo,
		SerialNo:      med.SerialNo,
		CityID:        cfg.MedCityID,
		MedInstName:   med.MedInstName,
		MedInstNo:     cfg.MedInstNo,
		TotalFee:      med.TotalFee(),
		CallbackURL:   cfg.MedNotifyURL,
		ChannelNo:     med.ChannelNo,
		MedInsTestEnv: cfg.MedTestEnv,
	}
	// 服务商模式: appid=服务商公众号、sub_appid=医疗机构公众号、sub_mchid=子商户号；
	// openid主体跟随调起公众号(wechat_appid为sub_appid时以sub_openid下单)
	if cfg.Mode == "partner" {
		req.AppID = cfg.SpAppID
		req.SubAppID = cfg.SubAppID
		req.SubMchID = cfg.SubMchID
		if cfg.AppID == cfg.SubAppID {
			req.SubOpenid = openID
			req.Openid = ""
		}
	}

	if mixType != "CASH_ONLY" {
		req.PayOrderID = med.PayOrderID
		req.PayAuthNo = med.PayAuthNo
		req.GeoLocation = med.GeoLocation
		req.MedInsOrderCreateTime = time.Now().In(cstZone).Format("2006-01-02T15:04:05+08:00")
		req.MedInsGovFee = &med.MedInsGovFee
		req.MedInsSelfFee = &med.MedInsSelfFee
		req.MedInsOtherFee = &med.MedInsOtherFee
		req.MedInsCashFee = &med.MedInsCashFee
	}
	if mixType != "INSURANCE_ONLY" {
		req.WechatPayCashFee = &med.WechatPayCashFee
		req.PrepayID = prepayID
	}
	if len(med.CashAddDetail) > 0 {
		req.CashAddDetail = toV3CashAdd(med.CashAddDetail)
	}
	if len(med.CashReduceDetail) > 0 {
		req.CashReduceDetail = toV3CashReduce(med.CashReduceDetail)
	}
	// 医疗机构透传给医保的数据(选填)
	if med.PassthroughRequestContent != "" {
		req.PassthroughRequestContent = med.PassthroughRequestContent
	}
	// 代亲属支付: 本人信息放payer/其余字段，亲属(就诊人)信息放relative与extends
	if med.PayForRelatives {
		if med.RelativeName == "" || med.RelativeIDCard == "" {
			return nil, fmt.Errorf("代亲属支付需提供亲属姓名与身份证号")
		}
		req.PayForRelatives = true
		req.Extends = wxpayv3.BuildMedExtends(med.RelativeName, med.RelativeIDCard)
		relative, err := client.BuildMedInsPayer(med.RelativeName, med.RelativeIDCard)
		if err != nil {
			return nil, err
		}
		req.Relative = relative
	}

	// 4) 调用医保混合下单
	resp, err := client.MedInsPrepay(req)
	if err != nil {
		// 含自费时已先完成JSAPI下单，医保下单失败须关单，避免微信侧残留可支付预付单
		if prepayID != "" {
			if cerr := client.CloseOrder(outTradeNo); cerr != nil {
				s.log.Error("医保下单失败后关单失败(需人工核查)", "out_trade_no", outTradeNo, "error", cerr)
			}
		}
		return nil, fmt.Errorf("医保混合下单失败: %w", err)
	}

	// 5) 更新本地订单
	payment.Openid = openID
	payment.MixTradeNo = resp.MixTradeNo
	payment.TradeState = resp.MixPayStatus
	if raw, e := json.Marshal(resp); e == nil {
		payment.WxLastResponse = string(raw)
	}
	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, err
	}

	// 6) 组装前端调起参数
	payParams := &wxpayv3.MedInsPayParams{
		MixTradeNo: resp.MixTradeNo,
		AppID:      req.AppID,
	}
	if cashParams != nil {
		payParams.TimeStamp = cashParams.TimeStamp
		payParams.NonceStr = cashParams.NonceStr
		payParams.Package = cashParams.Package
		payParams.SignType = cashParams.SignType
		payParams.PaySign = cashParams.PaySign
	}
	s.log.Info("医保混合下单成功", "out_trade_no", outTradeNo, "mix_trade_no", resp.MixTradeNo, "mix_pay_type", mixType)
	return payParams, nil
}

// saveMedPrepayInfo 保存自费部分调起参数到微信支付信息
func (s *MedicalService) saveMedPrepayInfo(paymentID int64, appID, prepayID string, params *wxpayv3.JSAPIPayParams) error {
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return err
	}
	payment.WechatPayInfo = &entity.WechatPayInfo{
		PaymentID: paymentID,
		AppID:     appID,
		MCHID:     s.wechatPayService.GetMCHID(),
		PrepayID:  prepayID,
		NonceStr:  params.NonceStr,
		Timestamp: params.TimeStamp,
		Package:   params.Package,
		PaySign:   params.PaySign,
		Sign:      params.PaySign,
	}
	return s.paymentRepo.Update(payment)
}

// MedQueryByMixTradeNo 按医保混合订单号查单（同步本地订单状态）
func (s *MedicalService) MedQueryByMixTradeNo(mixTradeNo string) (*wxpayv3.MedInsOrderResponse, error) {
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	resp, err := client.MedInsQueryByMixTradeNo(mixTradeNo)
	if err != nil {
		return nil, fmt.Errorf("查询医保订单失败: %w", err)
	}
	if err := s.syncOrderFromMedQuery(resp); err != nil {
		s.log.Error("同步医保订单状态失败", "mix_trade_no", mixTradeNo, "error", err)
	}
	return resp, nil
}

// MedQueryByOutTradeNo 按商户订单号查医保订单（同步本地订单状态）
func (s *MedicalService) MedQueryByOutTradeNo(outTradeNo string) (*wxpayv3.MedInsOrderResponse, error) {
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	resp, err := client.MedInsQueryByOutTradeNo(outTradeNo)
	if err != nil {
		return nil, fmt.Errorf("查询医保订单失败: %w", err)
	}
	if err := s.syncOrderFromMedQuery(resp); err != nil {
		s.log.Error("同步医保订单状态失败", "out_trade_no", outTradeNo, "error", err)
	}
	return resp, nil
}

// syncOrderFromMedQuery 根据医保查单结果更新本地订单
func (s *MedicalService) syncOrderFromMedQuery(resp *wxpayv3.MedInsOrderResponse) error {
	outTradeNo := resp.OutTradeNo
	if outTradeNo == "" && resp.MixTradeNo != "" {
		if p, err := s.paymentRepo.GetByMixTradeNo(resp.MixTradeNo); err == nil {
			outTradeNo = p.OrderID
		}
	}
	if outTradeNo == "" {
		return fmt.Errorf("医保查单结果缺少订单号(mix_trade_no=%s)", resp.MixTradeNo)
	}
	payment, err := s.paymentRepo.GetByOrderID(outTradeNo)
	if err != nil {
		return err
	}
	payment.MixTradeNo = resp.MixTradeNo
	payment.TradeState = resp.MixPayStatus
	if raw, e := json.Marshal(resp); e == nil {
		payment.WxLastResponse = string(raw)
	}
	switch resp.MixPayStatus {
	case "MIX_PAY_SUCCESS":
		payment.Status = entity.PaymentStatusSuccess
		if t, e := time.Parse(time.RFC3339, resp.PaidTime); e == nil {
			payment.PaidAt = &t
		}
	case "MIX_PAY_FAIL":
		payment.Status = entity.PaymentStatusFailed
	case "MIX_PAY_REFUND":
		payment.Status = entity.PaymentStatusRefunded
	}
	return s.paymentRepo.Update(payment)
}

// HandleNotify 处理医保混合收款成功回调（MEDICAL_INSURANCE.SUCCESS，同步处理）
// ctx 用于回调租户定位（按 Wechatpay-Serial 查主库映射得租户后到租户库验签，见设计文档 01 §九）
func (s *MedicalService) HandleNotify(ctx context.Context, header http.Header, body []byte) (*entity.Payment, error) {
	notification, plaintext, err := s.wechatPayService.ParseNotify(ctx, header, body)
	if err != nil {
		s.log.Error("医保回调验签/解密失败", "error", err)
		return nil, err
	}
	s.log.Info("收到医保回调", "event_type", notification.EventType, "id", notification.ID)

	if notification.EventType != "MEDICAL_INSURANCE.SUCCESS" {
		s.log.Info("忽略医保回调事件", "event_type", notification.EventType)
		return nil, nil
	}

	var orderResp wxpayv3.MedInsOrderResponse
	if err := json.Unmarshal(plaintext, &orderResp); err != nil || orderResp.OutTradeNo == "" {
		s.log.Error("解析医保回调业务数据失败", "error", err)
		return nil, fmt.Errorf("解析医保回调业务数据失败: %w", err)
	}
	if err := s.syncOrderFromMedQuery(&orderResp); err != nil {
		return nil, err
	}
	return s.paymentRepo.GetByOrderID(orderResp.OutTradeNo)
}

// toV3CashAdd 实体现金补充项转协议模型
func toV3CashAdd(items []entity.MedCashAddItem) []wxpayv3.CashAddItem {
	out := make([]wxpayv3.CashAddItem, len(items))
	for i, it := range items {
		out[i] = wxpayv3.CashAddItem{CashAddFee: it.CashAddFee, CashAddType: it.CashAddType}
	}
	return out
}

// toV3CashReduce 实体现金减免项转协议模型
func toV3CashReduce(items []entity.MedCashReduceItem) []wxpayv3.CashReduceItem {
	out := make([]wxpayv3.CashReduceItem, len(items))
	for i, it := range items {
		out[i] = wxpayv3.CashReduceItem{CashReduceFee: it.CashReduceFee, CashReduceType: it.CashReduceType}
	}
	return out
}
