package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	wxpayv3 "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayv3"
)

// MedRefundService 医保订单退款业务编排（对齐 Gpay service/medrefund.go，V3新版，直连与服务商同构）
// 医保部分(统筹/个账/其他)走 V3 POST /v3/med-ins/refunds/notify?mix_trade_no=；
// 现金部分(wechat_pay_cash_fee)走普通退款接口(服务商模式自动带sub_mchid)
// 现金部分由微信支付原路退回，医保部分由医保局退回；结果以退款查询/退款单同步为准
type MedRefundService struct {
	paymentRepo      repo.PaymentRepository
	medRefundRepo    repo.MedRefundRepository
	wechatPayService *WechatPayService
	log              logger.Logger
}

// NewMedRefundService 创建医保退款服务
func NewMedRefundService(paymentRepo repo.PaymentRepository, medRefundRepo repo.MedRefundRepository, wechatPayService *WechatPayService, log logger.Logger) *MedRefundService {
	return &MedRefundService{
		paymentRepo:      paymentRepo,
		medRefundRepo:    medRefundRepo,
		wechatPayService: wechatPayService,
		log:              log,
	}
}

// MedRefundInput 医保退款入参
type MedRefundInput struct {
	OutTradeNo      string // 本地订单号(必填，即hosp_out_trade_no)
	HospOutRefundNo string // 医院退款订单号(空则自动生成R+订单号+时间戳；有自费单时须与自费退款申请out_refund_no一致)
	PartRefundType  string // 部分退款类型(空=全退)
	CashRefundFee   int64  // 现金退款金额(分)
	InsRefundFee    int64  // 个账退款金额(分，仅INS_INDIVIDUAL_PART生效)
	RefReason       string // 退费原因(本地留档)
}

// MedRefundQueryResponse 医保退款状态(本地聚合查询结果)
// 领域层值对象：不带 json 标签，接口响应由 API Handler 映射到 dto/wx/response 的同名 DTO
type MedRefundQueryResponse struct {
	HospOutRefundNo       string
	HospOutTradeNo        string
	MedRefundID           string
	CashRefundID          string
	InsuranceRefundID     string
	CashRefundFee         int64
	InsuranceRefundFee    int64
	CashRefundStatus      string
	InsuranceRefundStatus string
	MedRefundState        string
	RefundEndTime         string
	ResponseContent       string
}

// Create 发起医保退款(直连与服务商同流程)
// 医保部分走V3退款通知，现金部分走普通退款接口；结果以查询为准
func (s *MedRefundService) Create(in *MedRefundInput) (*entity.MedRefundRecord, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	payment, err := s.paymentRepo.GetByOrderID(in.OutTradeNo)
	if err != nil {
		return nil, fmt.Errorf("订单不存在: %s", in.OutTradeNo)
	}
	if payment.OrderType != entity.PaymentOrderTypeMedical {
		return nil, fmt.Errorf("非医保订单请使用普通退款接口 /api/wx/refund/create")
	}
	if payment.Status != entity.PaymentStatusSuccess && payment.Status != entity.PaymentStatusRefunded {
		return nil, fmt.Errorf("订单当前状态不可退款: %s", payment.Status)
	}
	if payment.MedInfo == nil {
		return nil, fmt.Errorf("订单缺少医保信息")
	}
	if payment.MixTradeNo == "" {
		return nil, fmt.Errorf("订单缺少医保混合订单号(mix_trade_no)，无法发起医保退款")
	}

	med := payment.MedInfo
	mixType := med.MixPayType()

	// 服务端金额校验
	if in.PartRefundType == wxpayv3.MedRefundCashOnly && in.CashRefundFee <= 0 {
		return nil, fmt.Errorf("只退现金(CASH_ONLY)时现金退款金额(cashRefundFee)必须大于0")
	}
	if in.PartRefundType == wxpayv3.MedRefundInsIndividualPart {
		if in.InsRefundFee <= 0 {
			return nil, fmt.Errorf("退医保个账(INS_INDIVIDUAL_PART)时个账退款金额(insRefundFee)必须大于0")
		}
		if in.InsRefundFee > med.MedInsSelfFee {
			return nil, fmt.Errorf("个账退款金额(%d)超过可退个账金额(%d)", in.InsRefundFee, med.MedInsSelfFee)
		}
	}

	// 医保退款单号: 未传时自动生成(医院侧唯一)
	hospOutRefundNo := in.HospOutRefundNo
	if hospOutRefundNo == "" {
		hospOutRefundNo = fmt.Sprintf("R%s%s", in.OutTradeNo, time.Now().Format("150405"))
	}
	if _, err := s.medRefundRepo.GetByHospOutRefundNo(hospOutRefundNo); err == nil {
		return nil, fmt.Errorf("医院退款订单号已存在: %s，重试请沿用原单号", hospOutRefundNo)
	}

	// 按退款类型拆分医保部分与现金部分退款金额
	var govFee, selfFee, otherFee, cashFee int64
	switch in.PartRefundType {
	case wxpayv3.MedRefundCashOnly: // 只退现金
		cashFee = in.CashRefundFee
	case wxpayv3.MedRefundInsuranceOnly: // 全额退医保(统筹+个账+其他)
		govFee, selfFee, otherFee = med.MedInsGovFee, med.MedInsSelfFee, med.MedInsOtherFee
	case wxpayv3.MedRefundInsPlanPart: // 退医保统筹
		govFee = med.MedInsGovFee
	case wxpayv3.MedRefundInsIndividualPart: // 退医保个账(部分)
		selfFee = in.InsRefundFee
	default: // 全退(医保全额+现金按入参或全额)
		govFee, selfFee, otherFee = med.MedInsGovFee, med.MedInsSelfFee, med.MedInsOtherFee
		cashFee = in.CashRefundFee
		if cashFee == 0 && mixType != "INSURANCE_ONLY" {
			cashFee = med.WechatPayCashFee
		}
	}
	medTotal := govFee + selfFee + otherFee
	if medTotal == 0 && cashFee == 0 {
		return nil, fmt.Errorf("退款金额为0，请指定退款金额或退款类型")
	}
	if cashFee > med.WechatPayCashFee {
		return nil, fmt.Errorf("现金退款金额(%d)超过可退现金金额(%d)", cashFee, med.WechatPayCashFee)
	}

	rc, _ := json.Marshal(map[string]string{"ref_reason": in.RefReason})
	record := &entity.MedRefundRecord{
		HospOutRefundNo: hospOutRefundNo,
		OutTradeNo:      in.OutTradeNo,
		PartRefundType:  in.PartRefundType,
		CashRefundFee:   cashFee,
		InsRefundFee:    medTotal, // 医保部分退款总额(统筹+个账+其他)
		RequestContent:  string(rc),
		Reason:          in.RefReason,
	}

	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}

	// 1) 医保部分: V3退款通知(无应答业务体，受理成功即返回)
	if medTotal > 0 {
		req := &wxpayv3.MedInsRefundNotifyRequest{
			OutRefundNo:       hospOutRefundNo,
			RefundTime:        time.Now().In(cstZone).Format(time.RFC3339),
			MedRefundTotalFee: medTotal,
			MedRefundGovFee:   govFee,
			MedRefundSelfFee:  selfFee,
			MedRefundOtherFee: otherFee,
		}
		if err := client.MedInsRefundNotify(payment.MixTradeNo, req); err != nil {
			return nil, fmt.Errorf("医保部分退款失败: %w", err)
		}
		record.MedRefundState = wxpayv3.MedRefundStateRefunding
	}

	// 2) 现金部分: 普通退款接口
	if cashFee > 0 {
		refReq := &wxpayv3.RefundCreateRequest{
			OutTradeNo:  in.OutTradeNo,
			OutRefundNo: hospOutRefundNo,
			Reason:      in.RefReason,
			NotifyURL:   s.wechatPayService.GetRefundNotifyURL(),
			Amount:      &wxpayv3.RefundAmount{Refund: cashFee, Total: payment.Amount, Currency: "CNY"},
		}
		resp, err := client.CreateRefund(refReq)
		if err != nil {
			if record.MedRefundState != "" {
				// 医保部分已受理但现金退款失败，保留退款单供人工处理
				record.MedRefundState = wxpayv3.MedRefundStateRefunding
				if err := s.medRefundRepo.Create(record); err != nil {
					s.log.Error("保存医保退款单失败", "hosp_out_refund_no", hospOutRefundNo, "error", err)
				}
			}
			return nil, fmt.Errorf("现金部分退款失败: %w", err)
		}
		record.CashRefundID = resp.RefundID
		record.CashRefundStatus = resp.Status
	}

	if raw, e := json.Marshal(record); e == nil {
		record.LastResponse = string(raw)
	}
	if err := s.medRefundRepo.Create(record); err != nil {
		return nil, fmt.Errorf("保存医保退款单失败: %w", err)
	}
	s.log.Info("医保退款已受理", "hosp_out_refund_no", hospOutRefundNo, "医保", medTotal, "现金", cashFee)
	return record, nil
}

// Query 医保退款状态查询(按医院退款订单号或本地订单号)，同步退款单与订单状态
// 现金部分走V3退款查询；医保部分以医保订单查询(mix_pay_status=MIX_PAY_REFUND)为准
func (s *MedRefundService) Query(hospOutRefundNo, outTradeNo string) (*MedRefundQueryResponse, error) {
	if hospOutRefundNo == "" && outTradeNo == "" {
		return nil, fmt.Errorf("hosp_out_refund_no与out_trade_no至少传一个")
	}
	if hospOutRefundNo == "" {
		// 按本地订单号找最近的医保退款单
		found, err := s.medRefundRepo.LatestByOutTradeNo(outTradeNo)
		if err != nil {
			return nil, fmt.Errorf("该订单没有医保退款单: %s", outTradeNo)
		}
		hospOutRefundNo = found.HospOutRefundNo
	}
	return s.queryV3(hospOutRefundNo)
}

// List 医保退款单列表(分页，按创建时间倒序)
func (s *MedRefundService) List(page, pageSize int) ([]*entity.MedRefundRecord, int, error) {
	return s.medRefundRepo.List(page, pageSize)
}

// queryV3 退款状态聚合查询
// 现金部分以V3单笔退款查询为准；医保部分以医保订单查询(mix_pay_status=MIX_PAY_REFUND)为准
func (s *MedRefundService) queryV3(hospOutRefundNo string) (*MedRefundQueryResponse, error) {
	record, err := s.medRefundRepo.GetByHospOutRefundNo(hospOutRefundNo)
	if err != nil {
		return nil, fmt.Errorf("医保退款单不存在: %s", hospOutRefundNo)
	}
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}

	out := &MedRefundQueryResponse{
		HospOutRefundNo:       hospOutRefundNo,
		HospOutTradeNo:        record.OutTradeNo,
		MedRefundID:           record.MedRefundID,
		CashRefundID:          record.CashRefundID,
		InsuranceRefundID:     record.InsuranceRefundID,
		CashRefundFee:         record.CashRefundFee,
		InsuranceRefundFee:    record.InsRefundFee,
		CashRefundStatus:      record.CashRefundStatus,
		InsuranceRefundStatus: record.InsuranceRefundStatus,
		MedRefundState:        record.MedRefundState,
		ResponseContent:       record.ResponseContent,
	}

	// 1) 现金部分: V3退款查询(服务商模式自动带sub_mchid)
	if record.CashRefundID != "" {
		if resp, err := client.QueryRefund(hospOutRefundNo); err == nil && resp != nil {
			out.CashRefundStatus = resp.Status
		} else if err != nil {
			s.log.Warn("现金退款查询失败", "hosp_out_refund_no", hospOutRefundNo, "error", err)
		}
	}

	// 2) 医保部分: 查医保订单获取退款状态(含医保局透传内容)
	if payment, err := s.paymentRepo.GetByOrderID(record.OutTradeNo); err == nil && payment.MixTradeNo != "" {
		if resp, err := client.MedInsQueryByMixTradeNo(payment.MixTradeNo); err == nil && resp != nil {
			if resp.MixPayStatus == "MIX_PAY_REFUND" {
				out.InsuranceRefundStatus = wxpayv3.MedRefundStateSuccess
				out.MedRefundState = wxpayv3.MedRefundStateSuccess
			}
			if resp.PassthroughResponseContent != "" {
				out.ResponseContent = resp.PassthroughResponseContent
			}
		} else if err != nil {
			s.log.Warn("医保订单查询失败", "mix_trade_no", payment.MixTradeNo, "error", err)
		}
	}

	s.syncMedRefund(out)
	return out, nil
}

// syncMedRefund 根据退款查询结果同步本地退款单与订单状态
// 部分退款语义: 现金/医保部分退款成功金额分别累计至订单RefundedFee，累计退满才置订单为已退款
func (s *MedRefundService) syncMedRefund(resp *MedRefundQueryResponse) {
	record, err := s.medRefundRepo.GetByHospOutRefundNo(resp.HospOutRefundNo)
	if err != nil {
		s.log.Error("更新医保退款单失败（退款单不存在）", "hosp_out_refund_no", resp.HospOutRefundNo)
		return
	}

	if resp.MedRefundID != "" {
		record.MedRefundID = resp.MedRefundID
	}
	if resp.CashRefundID != "" {
		record.CashRefundID = resp.CashRefundID
	}
	if resp.InsuranceRefundID != "" {
		record.InsuranceRefundID = resp.InsuranceRefundID
	}
	if resp.CashRefundStatus != "" {
		record.CashRefundStatus = resp.CashRefundStatus
	}
	if resp.InsuranceRefundStatus != "" {
		record.InsuranceRefundStatus = resp.InsuranceRefundStatus
	}
	if resp.MedRefundState != "" {
		record.MedRefundState = resp.MedRefundState
	}
	if resp.RefundEndTime != "" {
		record.RefundEndTime = resp.RefundEndTime
	}
	if resp.ResponseContent != "" {
		record.ResponseContent = resp.ResponseContent
	}
	if raw, e := json.Marshal(resp); e == nil {
		record.LastResponse = string(raw)
	}

	// 现金部分首次成功: 累计现金退款金额(纯医保退时现金状态为空，不进此分支)
	if resp.CashRefundStatus == wxpayv3.MedRefundStateSuccess && !record.CashApplied && record.CashRefundFee > 0 {
		if s.applyRefundToOrder(resp.HospOutTradeNo, record.CashRefundFee) {
			record.CashApplied = true
		}
	}
	// 医保部分首次成功: 累计医保退款金额(纯现金退时医保状态为空，不进此分支)
	if resp.InsuranceRefundStatus == wxpayv3.MedRefundStateSuccess && !record.InsApplied && record.InsRefundFee > 0 {
		if s.applyRefundToOrder(resp.HospOutTradeNo, record.InsRefundFee) {
			record.InsApplied = true
		}
	}

	if err := s.medRefundRepo.Update(record); err != nil {
		s.log.Error("更新医保退款单失败", "hosp_out_refund_no", resp.HospOutRefundNo, "error", err)
	}
}

// applyRefundToOrder 累计退款金额至订单，累计退满置订单为已退款
func (s *MedRefundService) applyRefundToOrder(outTradeNo string, fee int64) bool {
	payment, err := s.paymentRepo.GetByOrderID(outTradeNo)
	if err != nil {
		s.log.Error("累计退款金额失败（订单不存在）", "out_trade_no", outTradeNo, "error", err)
		return false
	}
	payment.RefundedFee += fee
	if payment.RefundedFee >= payment.Amount {
		payment.Status = entity.PaymentStatusRefunded
		payment.TradeState = "REFUND"
	}
	if err := s.paymentRepo.Update(payment); err != nil {
		s.log.Error("累计退款金额失败", "out_trade_no", outTradeNo, "error", err)
		return false // 未累计成功不置Applied，下次查询可重试
	}
	return true
}
