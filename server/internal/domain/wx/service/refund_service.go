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

// RefundService 退款业务编排（对齐 Gpay service/refund.go，普通 JSAPI/Native 订单）
// 医保订单退款需走医保退款接口（MedRefundService）
type RefundService struct {
	paymentRepo      repo.PaymentRepository
	refundRepo       repo.RefundRepository
	wechatPayService *WechatPayService
	log              logger.Logger
}

// NewRefundService 创建退款服务
func NewRefundService(paymentRepo repo.PaymentRepository, refundRepo repo.RefundRepository, wechatPayService *WechatPayService, log logger.Logger) *RefundService {
	return &RefundService{
		paymentRepo:      paymentRepo,
		refundRepo:       refundRepo,
		wechatPayService: wechatPayService,
		log:              log,
	}
}

// genOutRefundNo 生成商户退款单号
func genOutRefundNo() string {
	return "RF" + time.Now().Format("20060102150405") + generateNonceStr()[:8]
}

// Create 发起退款(支持部分退款，累计退款金额不超过订单金额)
// 返回的退款单状态为受理结果，最终以退款回调/查询为准
func (s *RefundService) Create(outTradeNo string, refundFee int64, reason string) (*entity.RefundRecord, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}

	payment, err := s.paymentRepo.GetByOrderID(outTradeNo)
	if err != nil {
		return nil, fmt.Errorf("订单不存在: %s", outTradeNo)
	}
	if payment.OrderType == entity.PaymentOrderTypeMedical {
		return nil, fmt.Errorf("医保订单退款需走医保退款接口 /api/wx/refund/med")
	}
	if payment.Status != entity.PaymentStatusSuccess && payment.Status != entity.PaymentStatusRefunded {
		return nil, fmt.Errorf("订单当前状态不可退款: %s", payment.Status)
	}
	if refundFee <= 0 {
		return nil, fmt.Errorf("退款金额必须大于0(单位:分)")
	}

	// 累计退款金额(不含已关闭的退款单)不得超过订单金额
	records, err := s.refundRepo.ListByOutTradeNo(outTradeNo)
	if err != nil {
		return nil, err
	}
	var refunded int64
	for _, r := range records {
		if r.Status != entity.RefundStatusClosed {
			refunded += r.RefundFee
		}
	}
	if refunded+refundFee > payment.Amount {
		return nil, fmt.Errorf("退款金额超出可退金额: 已退%d分，订单金额%d分，最多再退%d分",
			refunded, payment.Amount, payment.Amount-refunded)
	}

	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	req := &wxpayv3.RefundCreateRequest{
		OutTradeNo:  outTradeNo,
		OutRefundNo: genOutRefundNo(),
		Reason:      reason,
		NotifyURL:   s.wechatPayService.GetRefundNotifyURL(),
		Amount:      &wxpayv3.RefundAmount{Refund: refundFee, Total: payment.Amount, Currency: "CNY"},
	}
	resp, err := client.CreateRefund(req)
	if err != nil {
		return nil, fmt.Errorf("申请退款失败: %w", err)
	}

	record := &entity.RefundRecord{
		OutRefundNo: req.OutRefundNo,
		OutTradeNo:  outTradeNo,
		RefundID:    resp.RefundID,
		Reason:      reason,
		RefundFee:   refundFee,
		TotalFee:    payment.Amount,
		Status:      resp.Status,
	}
	if raw, e := json.Marshal(resp); e == nil {
		record.WxLastResponse = string(raw)
	}
	if err := s.refundRepo.Create(record); err != nil {
		return nil, fmt.Errorf("保存退款单失败: %w", err)
	}
	s.log.Info("退款申请已受理", "out_refund_no", req.OutRefundNo, "out_trade_no", outTradeNo, "refund_fee", refundFee)
	return record, nil
}

// Query 查询单笔退款并同步本地退款单与订单状态
func (s *RefundService) Query(outRefundNo string) (*entity.RefundRecord, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	resp, err := client.QueryRefund(outRefundNo)
	if err != nil {
		return nil, fmt.Errorf("查询退款失败: %w", err)
	}
	if err := s.syncRefund(resp); err != nil {
		s.log.Error("同步退款单失败", "out_refund_no", outRefundNo, "error", err)
	}

	if record, err := s.refundRepo.GetByOutRefundNo(outRefundNo); err == nil {
		return record, nil
	}
	// 非本服务发起的退款单，返回微信侧应答快照
	record := &entity.RefundRecord{
		OutRefundNo: resp.OutRefundNo,
		OutTradeNo:  resp.OutTradeNo,
		RefundID:    resp.RefundID,
		Status:      resp.Status,
	}
	if resp.Amount != nil {
		record.RefundFee = resp.Amount.Refund
		record.TotalFee = resp.Amount.Total
	}
	if raw, e := json.Marshal(resp); e == nil {
		record.WxLastResponse = string(raw)
	}
	return record, nil
}

// List 退款单列表(分页，按创建时间倒序)
func (s *RefundService) List(page, pageSize int) ([]*entity.RefundRecord, int, error) {
	return s.refundRepo.List(page, pageSize)
}

// syncRefund 根据微信退款应答同步本地退款单与订单状态
// 部分退款语义: 退款成功金额累计至订单RefundedFee，累计退满才置订单为已退款
func (s *RefundService) syncRefund(resp *wxpayv3.RefundResponse) error {
	record, err := s.refundRepo.GetByOutRefundNo(resp.OutRefundNo)
	if err != nil {
		// 非本服务发起的退款单，落库快照
		newRecord := &entity.RefundRecord{
			OutRefundNo: resp.OutRefundNo,
			OutTradeNo:  resp.OutTradeNo,
			RefundID:    resp.RefundID,
			Status:      resp.Status,
		}
		if resp.Amount != nil {
			newRecord.RefundFee = resp.Amount.Refund
			newRecord.TotalFee = resp.Amount.Total
		}
		if raw, e := json.Marshal(resp); e == nil {
			newRecord.WxLastResponse = string(raw)
		}
		return s.refundRepo.Create(newRecord)
	}

	record.RefundID = resp.RefundID
	record.Status = resp.Status
	record.SuccessTime = resp.SuccessTime
	record.UserReceivedAccount = resp.UserReceivedAccount
	if raw, e := json.Marshal(resp); e == nil {
		record.WxLastResponse = string(raw)
	}

	// 首次成功标记防回调重推与主动查询并发双计
	if resp.Status == entity.RefundStatusSuccess && !record.Applied {
		fee := record.RefundFee
		if resp.Amount != nil && resp.Amount.Refund > 0 {
			fee = resp.Amount.Refund
		}
		if payment, err := s.paymentRepo.GetByOrderID(resp.OutTradeNo); err == nil {
			payment.RefundedFee += fee
			if payment.RefundedFee >= payment.Amount {
				payment.Status = entity.PaymentStatusRefunded
				payment.TradeState = "REFUND"
			}
			if err := s.paymentRepo.Update(payment); err != nil {
				s.log.Error("更新订单退款状态失败", "out_trade_no", resp.OutTradeNo, "error", err)
				// 未累计成功不置Applied，下次查询/回调可重试
				if uerr := s.refundRepo.Update(record); uerr != nil {
					return uerr
				}
				return nil
			}
			record.Applied = true
		}
	}
	return s.refundRepo.Update(record)
}

// ApplyAbnormal 发起异常退款(退款单状态为ABNORMAL时)
// handleType: USER_BANK_CARD退用户银行卡(需银行类型/卡号/姓名，服务端OAEP加密)；MERCHANT_BANK_CARD退交易商户银行账户
func (s *RefundService) ApplyAbnormal(outRefundNo, handleType, bankType, bankAccount, realName string) (*entity.RefundRecord, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	record, err := s.refundRepo.GetByOutRefundNo(outRefundNo)
	if err != nil {
		return nil, fmt.Errorf("退款单不存在: %s", outRefundNo)
	}
	if record.Status != entity.RefundStatusAbnormal {
		return nil, fmt.Errorf("退款单状态不是ABNORMAL，无需异常退款处理，当前: %s", record.Status)
	}
	if record.RefundID == "" {
		return nil, fmt.Errorf("退款单缺少微信退款单号，请先查询退款刷新状态")
	}

	req := &wxpayv3.AbnormalRefundRequest{
		OutRefundNo: outRefundNo,
		Type:        handleType,
		BankType:    bankType,
	}
	if handleType == wxpayv3.AbnormalRefundToUser {
		if bankType == "" || bankAccount == "" || realName == "" {
			return nil, fmt.Errorf("退款至用户银行卡需提供开户银行(bank_type)、银行卡号与用户姓名")
		}
		client, err := s.wechatPayService.GetV3Client()
		if err != nil {
			return nil, err
		}
		// 敏感字段服务端 OAEP 加密后才出网（官方工具类）
		encAccount, err := wxpayv3.EncryptOAEP(bankAccount, client.GetPublicKey())
		if err != nil {
			return nil, fmt.Errorf("银行卡号加密失败: %w", err)
		}
		encName, err := wxpayv3.EncryptOAEP(realName, client.GetPublicKey())
		if err != nil {
			return nil, fmt.Errorf("用户姓名加密失败: %w", err)
		}
		req.BankAccount = encAccount
		req.RealName = encName
		if err := client.ApplyAbnormalRefund(record.RefundID, req); err != nil {
			return nil, fmt.Errorf("发起异常退款失败: %w", err)
		}
	} else {
		client, err := s.wechatPayService.GetV3Client()
		if err != nil {
			return nil, err
		}
		if err := client.ApplyAbnormalRefund(record.RefundID, req); err != nil {
			return nil, fmt.Errorf("发起异常退款失败: %w", err)
		}
	}

	// 受理成功后置为处理中，最终状态由退款回调/查询刷新
	record.Status = entity.RefundStatusProcessing
	if err := s.refundRepo.Update(record); err != nil {
		return nil, err
	}
	s.log.Info("异常退款已受理", "out_refund_no", outRefundNo, "type", handleType)
	return record, nil
}

// HandleNotify 处理退款结果回调（REFUND.SUCCESS/ABNORMAL/CLOSED，同步处理）
// ctx 用于回调租户定位（按 Wechatpay-Serial 查主库映射得租户后到租户库验签，见设计文档 01 §九）
func (s *RefundService) HandleNotify(ctx context.Context, header http.Header, body []byte) (*entity.RefundRecord, error) {
	notification, plaintext, err := s.wechatPayService.ParseNotify(ctx, header, body)
	if err != nil {
		s.log.Error("退款回调验签/解密失败", "error", err)
		return nil, err
	}
	s.log.Info("收到退款回调", "event_type", notification.EventType, "id", notification.ID)

	if len(notification.EventType) < 7 || notification.EventType[:7] != "REFUND." {
		s.log.Info("忽略退款回调事件", "event_type", notification.EventType)
		return nil, nil
	}

	var data wxpayv3.RefundNotifyData
	if err := json.Unmarshal(plaintext, &data); err != nil || data.OutRefundNo == "" {
		s.log.Error("解析退款回调业务数据失败", "error", err)
		return nil, fmt.Errorf("解析退款回调业务数据失败: %w", err)
	}

	resp := &wxpayv3.RefundResponse{
		RefundID:            data.RefundID,
		OutRefundNo:         data.OutRefundNo,
		OutTradeNo:          data.OutTradeNo,
		TransactionID:       data.TransactionID,
		Status:              data.RefundStatus,
		SuccessTime:         data.SuccessTime,
		UserReceivedAccount: data.UserReceivedAccount,
		Amount:              data.Amount,
	}
	if err := s.syncRefund(resp); err != nil {
		return nil, err
	}
	return s.refundRepo.GetByOutRefundNo(data.OutRefundNo)
}
