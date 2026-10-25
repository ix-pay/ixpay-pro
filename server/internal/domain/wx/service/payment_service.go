package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/dictconst"
	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	wxpayv3 "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayv3"
)

// PaymentService 实现支付领域服务接口
// 支付流程参考 Gpay 项目（Prepay 语义）：
//  1. 创建待支付记录（POST /payment）→ 出现在"待支付列表"
//  2. 明细页查看（GET /payment/:id）
//  3. 下单支付（POST /payment/unified-order）→ V3 JSAPI 下单 → 前端 wx.requestPayment/chooseWXPay 调起
//  4. 微信回调（V3 验签+解密）驱动支付单终态
type PaymentService struct {
	repo                 repo.PaymentRepository
	wechatPayInfoRepo    repo.WechatPayInfoRepository
	log                  logger.Logger
	baseConfigRepository baseRepo.ConfigRepository
	wechatPayService     *WechatPayService
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(repo repo.PaymentRepository, wechatPayInfoRepo repo.WechatPayInfoRepository, baseConfigRepository baseRepo.ConfigRepository, log logger.Logger, wechatPayService *WechatPayService) *PaymentService {
	return &PaymentService{
		repo:                 repo,
		wechatPayInfoRepo:    wechatPayInfoRepo,
		log:                  log,
		baseConfigRepository: baseConfigRepository,
		wechatPayService:     wechatPayService,
	}
}

// genOrderID 生成商户订单号（out_trade_no 仅允许数字/字母/|_-，长度 6-32）
func genOrderID() string {
	return "WX" + time.Now().Format("20060102150405") + generateNonceStr()[:8]
}

// CreatePayment 创建支付记录（待支付）
// orderID 为空时服务端自动生成；防重复：同一订单号已存在时返回错误
func (s *PaymentService) CreatePayment(userID int64, orderID string, amount int64, method string, description string) (*entity.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("订单金额必须大于0(单位:分)")
	}
	if orderID == "" {
		orderID = genOrderID()
	}
	if method == "" {
		method = dictconst.UserTypeWechat
	}

	// 检查订单 ID 是否已存在
	if _, err := s.repo.GetByOrderID(orderID); err == nil {
		return nil, errors.New("订单 ID 已存在")
	}

	// 创建支付记录（待支付）
	payment := &entity.Payment{
		OrderID:     orderID,
		UserID:      userID,
		Amount:      amount,
		Currency:    "CNY",
		Method:      method,
		Status:      entity.PaymentStatusPending,
		Description: description,
	}
	if err := s.repo.Create(payment); err != nil {
		s.log.Error("创建支付记录失败", "error", err)
		return nil, err
	}

	s.log.Info("支付记录创建成功", "paymentID", payment.ID, "orderID", payment.OrderID)
	return payment, nil
}

// GetPayment 获取支付记录
func (s *PaymentService) GetPayment(paymentID int64) (*entity.Payment, error) {
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("获取支付记录失败", "error", err)
		return nil, err
	}
	return payment, nil
}

// GetPaymentByOrderID 根据订单 ID 获取支付记录
func (s *PaymentService) GetPaymentByOrderID(orderID string) (*entity.Payment, error) {
	payment, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		s.log.Error("根据订单 ID 获取支付记录失败", "error", err)
		return nil, err
	}
	return payment, nil
}

// GetUserPayments 获取指定用户的支付记录（分页，支持状态过滤；ctx 决定查询哪个租户库）
// status 为空查全部；传 "pending" 查待支付列表
func (s *PaymentService) GetUserPayments(ctx context.Context, userID int64, status string, page, pageSize int) ([]*entity.Payment, int, error) {
	var (
		payments []*entity.Payment
		total    int
		err      error
	)
	if status != "" {
		payments, total, err = s.repo.List(ctx, page, pageSize, map[string]interface{}{
			"user_id": userID,
			"status":  status,
		})
	} else {
		payments, total, err = s.repo.ListByUser(ctx, userID, page, pageSize)
	}
	if err != nil {
		s.log.Error("获取用户支付列表失败", "error", err, "userID", userID)
		return nil, 0, err
	}
	return payments, total, nil
}

// UpdatePaymentStatus 更新支付状态
func (s *PaymentService) UpdatePaymentStatus(paymentID int64, status entity.PaymentStatus) error {
	// 获取支付记录
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("获取支付记录失败", "error", err)
		return err
	}

	// 更新状态
	payment.Status = status

	// 如果是支付成功，记录支付时间
	if status == entity.PaymentStatusSuccess && payment.PaidAt == nil {
		now := time.Now()
		payment.PaidAt = &now
	}

	// 保存更新
	if err := s.repo.Update(payment); err != nil {
		s.log.Error("更新支付状态失败", "error", err)
		return err
	}

	s.log.Info("支付状态更新成功", "paymentID", paymentID, "status", status)
	return nil
}

// CreateWechatPayment 微信下单（Prepay 语义，多渠道版）
// 渠道路由：appKey 定位应用（服务号/小程序），mchID 定位商户（空则用应用默认商户）；
// openid 必须为该 appid 下的用户标识。订单存在则校验归属与状态（待支付可重复下单刷新 prepay_id）
func (s *PaymentService) CreateWechatPayment(ctx context.Context, userID int64, openID string, appKey string, mchID string, orderID string, amount int64, description string) (*entity.Payment, *JSAPIParams, error) {
	// 1. 解析支付渠道（应用+商户+协议客户端）
	channel, err := s.wechatPayService.ResolveChannel(ctx, appKey, mchID)
	if err != nil {
		return nil, nil, err
	}

	// 2. 查找或创建支付记录
	payment, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		// 不存在 → 创建待支付记录
		payment, err = s.CreatePayment(userID, orderID, amount, dictconst.UserTypeWechat, description)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// 已存在 → 校验归属与状态（防越权支付与重复支付）
		if payment.UserID != userID {
			return nil, nil, errors.New("无权支付该订单")
		}
		if payment.Status == entity.PaymentStatusSuccess {
			return nil, nil, errors.New("订单已支付，请勿重复支付")
		}
		if payment.Status == entity.PaymentStatusCancelled {
			return nil, nil, errors.New("订单已取消，无法支付")
		}
		if payment.Status == entity.PaymentStatusRefunded {
			return nil, nil, errors.New("订单已退款，无法支付")
		}
	}

	appID := channel.App.AppID
	if appID == "" {
		return nil, nil, errors.New("应用未配置微信 AppID")
	}

	// 3. V3 JSAPI 下单（该商户的协议客户端，openid 与 appid 匹配）
	prepayID, err := channel.Client.JSAPIPrepay(appID, openID, payment.OrderID, payment.Description, channel.PayNotifyURL, payment.Amount)
	if err != nil {
		s.log.Error("微信 JSAPI 下单失败", "error", err, "mch_id", channel.Merchant.MchID)
		return nil, nil, err
	}

	// 4. 生成调起支付参数（该商户私钥签名）
	v3Params, err := channel.Client.BuildJSAPIPayParams(prepayID, appID)
	if err != nil {
		return nil, nil, err
	}
	jsapiParams := toServiceJSAPIParams(v3Params)

	// 5. 保存微信支付信息（持久化到 wx_wechat_pay_infos）
	if err := s.saveWechatPayInfo(payment.ID, appID, channel.Merchant.MchID, prepayID, jsapiParams); err != nil {
		s.log.Error("保存微信支付信息失败", "error", err)
	}

	payment.WechatPayInfo = &entity.WechatPayInfo{
		PaymentID: payment.ID,
		AppID:     appID,
		MCHID:     channel.Merchant.MchID,
		PrepayID:  prepayID,
		NonceStr:  jsapiParams.NonceStr,
		Timestamp: jsapiParams.TimeStamp,
		Package:   jsapiParams.Package,
		PaySign:   jsapiParams.PaySign,
		Sign:      jsapiParams.PaySign,
	}

	s.log.Info("微信支付下单成功", "paymentID", payment.ID, "orderID", payment.OrderID, "app_key", channel.App.AppKey, "mch_id", channel.Merchant.MchID)
	return payment, jsapiParams, nil
}

// toServiceJSAPIParams 协议层调起参数转服务层模型
func toServiceJSAPIParams(p *wxpayv3.JSAPIPayParams) *JSAPIParams {
	if p == nil {
		return nil
	}
	return &JSAPIParams{
		AppID:     p.AppID,
		TimeStamp: p.TimeStamp,
		NonceStr:  p.NonceStr,
		Package:   p.Package,
		SignType:  p.SignType,
		PaySign:   p.PaySign,
	}
}

// saveWechatPayInfo 保存/更新微信支付信息（按 payment_id 幂等）
func (s *PaymentService) saveWechatPayInfo(paymentID int64, appID, mchID, prepayID string, params *JSAPIParams) error {
	infos, _, err := s.wechatPayInfoRepo.List(1, 1, map[string]interface{}{"payment_id": paymentID})
	if err == nil && len(infos) > 0 {
		info := infos[0]
		info.AppID = appID
		info.MCHID = mchID
		info.PrepayID = prepayID
		info.NonceStr = params.NonceStr
		info.Timestamp = params.TimeStamp
		info.Package = params.Package
		info.PaySign = params.PaySign
		info.Sign = params.PaySign
		return s.wechatPayInfoRepo.Update(info)
	}
	return s.wechatPayInfoRepo.Create(&entity.WechatPayInfo{
		PaymentID: paymentID,
		AppID:     appID,
		MCHID:     mchID,
		PrepayID:  prepayID,
		NonceStr:  params.NonceStr,
		Timestamp: params.TimeStamp,
		Package:   params.Package,
		PaySign:   params.PaySign,
		Sign:      params.PaySign,
	})
}

// HandleWechatPayNotify 处理微信支付回调通知（V3）
// 验签（微信支付公钥）+ AES-GCM 解密后同步支付单状态
// ctx 用于回调租户定位（按 Wechatpay-Serial 查主库映射得租户后到租户库验签，见设计文档 01 §九）
func (s *PaymentService) HandleWechatPayNotify(ctx context.Context, header http.Header, body []byte) (*entity.Payment, error) {
	notification, plaintext, err := s.wechatPayService.ParseNotify(ctx, header, body)
	if err != nil {
		s.log.Error("微信支付回调验签/解密失败", "error", err)
		return nil, err
	}
	s.log.Info("收到微信支付回调", "event_type", notification.EventType, "id", notification.ID)

	if notification.EventType != "TRANSACTION.SUCCESS" {
		// 非支付成功事件忽略（微信要求非成功应答须返回非 2xx，这里按忽略处理返回成功应答）
		s.log.Info("忽略支付回调事件", "event_type", notification.EventType)
		return nil, nil
	}

	// 解析业务数据：合单回调报文含 combine_out_trade_no（按子单逐笔通知），走合单分支
	if bytes.Contains(plaintext, []byte(`"combine_out_trade_no"`)) {
		if err := s.HandleCombineNotify(plaintext); err != nil {
			s.log.Error("处理合单回调失败", "error", err)
			return nil, err
		}
		return nil, nil
	}

	var txn wxpayv3.QueryOrderResult
	if err := json.Unmarshal(plaintext, &txn); err != nil {
		s.log.Error("解析回调业务数据失败", "error", err)
		return nil, fmt.Errorf("解析回调业务数据失败: %w", err)
	}
	if txn.OutTradeNo == "" {
		return nil, errors.New("回调数据缺少 out_trade_no")
	}

	// 查询支付记录并更新状态
	payment, err := s.repo.GetByOrderID(txn.OutTradeNo)
	if err != nil {
		s.log.Error("根据订单号查询支付记录失败", "out_trade_no", txn.OutTradeNo, "error", err)
		return nil, err
	}

	// 幂等：已支付直接返回
	if payment.Status == entity.PaymentStatusSuccess {
		return payment, nil
	}

	paidAt := time.Now()
	if txn.SuccessTime != "" {
		// V3 时间格式 RFC3339，如 2015-05-20T13:29:35+08:00
		if t, e := time.Parse(time.RFC3339, txn.SuccessTime); e == nil {
			paidAt = t
		}
	}
	payment.MarkAsPaid(txn.TransactionID, paidAt)

	// 保存微信侧回执快照
	if payment.WechatPayInfo == nil {
		payment.WechatPayInfo = &entity.WechatPayInfo{PaymentID: payment.ID}
	}
	payment.WechatPayInfo.TransactionID = txn.TransactionID
	payment.WechatPayInfo.ResultCode = txn.TradeState
	payment.WechatPayInfo.ReturnCode = "SUCCESS"
	payment.WechatPayInfo.NotifyData = string(plaintext)
	if txn.Amount != nil {
		payment.WechatPayInfo.TotalFee = txn.Amount.Total
		payment.WechatPayInfo.CashFee = txn.Amount.PayerTotal
		payment.WechatPayInfo.FeeType = txn.Amount.Currency
	}

	if err := s.repo.Update(payment); err != nil {
		s.log.Error("更新支付记录失败", "error", err)
		return nil, err
	}

	s.log.Info("微信支付成功", "out_trade_no", txn.OutTradeNo, "transaction_id", txn.TransactionID)
	return payment, nil
}

// CheckPaymentStatus 检查支付状态
func (s *PaymentService) CheckPaymentStatus(paymentID int64) (entity.PaymentStatus, error) {
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("获取支付记录失败", "error", err)
		return "", err
	}
	return payment.Status, nil
}

// SyncPaymentStatus 主动查单同步支付状态（V3 查单）
func (s *PaymentService) SyncPaymentStatus(orderID string) (*entity.Payment, error) {
	result, err := s.wechatPayService.QueryOrderByOutTradeNo(orderID)
	if err != nil {
		s.log.Error("微信查单失败", "orderID", orderID, "error", err)
		return nil, err
	}

	payment, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	switch result.TradeState {
	case "SUCCESS":
		if payment.Status != entity.PaymentStatusSuccess {
			paidAt := time.Now()
			if result.SuccessTime != "" {
				if t, e := time.Parse(time.RFC3339, result.SuccessTime); e == nil {
					paidAt = t
				}
			}
			payment.MarkAsPaid(result.TransactionID, paidAt)
			if err := s.repo.Update(payment); err != nil {
				return nil, err
			}
		}
	case "CLOSED":
		if payment.Status == entity.PaymentStatusPending {
			payment.MarkAsCancelled()
			if err := s.repo.Update(payment); err != nil {
				return nil, err
			}
		}
	case "PAYERROR":
		if payment.Status == entity.PaymentStatusPending {
			payment.MarkAsFailed()
			if err := s.repo.Update(payment); err != nil {
				return nil, err
			}
		}
	}
	return payment, nil
}

// CancelPayment 取消支付（本地置为已取消，并同步关闭微信侧订单）
func (s *PaymentService) CancelPayment(paymentID int64) error {
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		return err
	}
	if payment.Status == entity.PaymentStatusSuccess {
		return errors.New("订单已支付，不可取消")
	}
	// 微信侧已下单的订单同步关单（失败不阻断本地取消）
	if err := s.wechatPayService.CloseOrder(payment.OrderID); err != nil {
		s.log.Warn("微信侧关单失败（可能未下单或已关闭）", "orderID", payment.OrderID, "error", err)
	}
	return s.UpdatePaymentStatus(paymentID, entity.PaymentStatusCancelled)
}

// GetJSAPIConfig 获取微信 JS-SDK 配置（用于 wx.config()）
// pageURL 为当前网页 URL（不含 # 及其后部分）
func (s *PaymentService) GetJSAPIConfig(pageURL string) (*JSAPIConfig, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	return s.wechatPayService.GetJSAPIConfig(pageURL)
}

// NativePrepay Native扫码下单（多渠道版）
// 订单不存在时创建待支付记录；已支付拒绝；每次调用重新下单刷新 code_url（有效期2小时）
// 返回二维码链接 code_url，由前端渲染二维码（服务端可经 /native/qrcode 生成 PNG）
func (s *PaymentService) NativePrepay(ctx context.Context, userID int64, appKey string, mchID string, orderID string, amount int64, description string) (*entity.Payment, string, error) {
	// 1. 解析支付渠道
	channel, err := s.wechatPayService.ResolveChannel(ctx, appKey, mchID)
	if err != nil {
		return nil, "", err
	}

	// 2. 查找或创建支付记录（与 JSAPI Prepay 语义一致）
	payment, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		payment, err = s.CreatePayment(userID, orderID, amount, dictconst.UserTypeWechat, description)
		if err != nil {
			return nil, "", err
		}
	} else {
		if payment.UserID != userID {
			return nil, "", errors.New("无权操作该订单")
		}
		if payment.Status == entity.PaymentStatusSuccess {
			return nil, "", errors.New("订单已支付，请勿重复支付")
		}
		if payment.Status == entity.PaymentStatusCancelled {
			return nil, "", errors.New("订单已取消，无法支付")
		}
	}

	// 3. Native 下单（无需 payer，扫码用户支付；appid 须与商户号绑定）
	appID := channel.App.AppID
	if appID == "" {
		return nil, "", errors.New("应用未配置微信 AppID")
	}
	codeURL, err := channel.Client.NativePrepay(appID, payment.OrderID, payment.Description, channel.PayNotifyURL, payment.Amount)
	if err != nil {
		s.log.Error("Native下单失败", "orderID", payment.OrderID, "error", err)
		return nil, "", fmt.Errorf("Native下单失败: %w", err)
	}

	// 4. 回写 code_url 与订单类型
	payment.CodeURL = codeURL
	payment.OrderType = entity.PaymentOrderTypeNative
	if err := s.repo.Update(payment); err != nil {
		return nil, "", err
	}

	s.log.Info("Native下单成功", "orderID", payment.OrderID, "mch_id", channel.Merchant.MchID)
	return payment, codeURL, nil
}

// NativeCodeURL 读取Native订单二维码链接(需已下单)
func (s *PaymentService) NativeCodeURL(orderID string) (string, error) {
	payment, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		return "", fmt.Errorf("订单不存在: %s", orderID)
	}
	if payment.OrderType != entity.PaymentOrderTypeNative {
		return "", errors.New("非Native扫码订单")
	}
	if payment.CodeURL == "" {
		return "", errors.New("订单尚未下单，无法生成二维码")
	}
	return payment.CodeURL, nil
}

// WxQuery 微信侧订单查询（商户订单号或微信支付订单号二选一，直连/服务商自动切换）
func (s *PaymentService) WxQuery(outTradeNo, transactionID string) (*wxpayv3.QueryOrderResult, error) {
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}

	var result *wxpayv3.QueryOrderResult
	if transactionID != "" {
		result, err = client.QueryOrderByTransactionID(transactionID)
	} else {
		result, err = client.QueryOrderByOutTradeNo(outTradeNo)
	}
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	// 同步本地订单状态
	syncNo := outTradeNo
	if syncNo == "" {
		syncNo = result.OutTradeNo
	}
	if syncNo != "" {
		if payment, perr := s.repo.GetByOrderID(syncNo); perr == nil {
			switch result.TradeState {
			case "SUCCESS":
				if payment.Status != entity.PaymentStatusSuccess {
					paidAt := time.Now()
					if t, e := time.Parse(time.RFC3339, result.SuccessTime); e == nil {
						paidAt = t
					}
					payment.MarkAsPaid(result.TransactionID, paidAt)
					if raw, je := json.Marshal(result); je == nil {
						payment.WxLastResponse = string(raw)
					}
					_ = s.repo.Update(payment)
				}
			case "CLOSED":
				if payment.Status == entity.PaymentStatusPending {
					payment.TradeState = "CLOSED"
					payment.MarkAsCancelled()
					_ = s.repo.Update(payment)
				}
			case "PAYERROR":
				if payment.Status == entity.PaymentStatusPending {
					payment.TradeState = "PAYERROR"
					payment.MarkAsFailed()
					_ = s.repo.Update(payment)
				}
			default:
				payment.TradeState = result.TradeState
				_ = s.repo.Update(payment)
			}
		}
	}
	return result, nil
}

// RefundPayment 退款
func (s *PaymentService) RefundPayment(paymentID int64, refundAmount int64) error {
	// 获取支付记录
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("获取支付记录失败", "error", err)
		return err
	}

	// 检查支付状态
	if payment.Status != entity.PaymentStatusSuccess {
		s.log.Error("无法退款，支付状态不符合要求", "status", payment.Status)
		return errors.New("仅成功支付的订单可以退款")
	}

	// 检查退款金额
	if refundAmount <= 0 || refundAmount > payment.Amount {
		s.log.Error("退款金额无效", "refundAmount", refundAmount, "paymentAmount", payment.Amount)
		return errors.New("退款金额无效")
	}

	// 这里应该实现退款逻辑（V3 /v3/refund/domestic/refunds）
	// 实际实现需要调用支付渠道的退款 API

	// 更新支付状态为已退款
	return s.UpdatePaymentStatus(paymentID, entity.PaymentStatusRefunded)
}

// CombinePrepay JSAPI 合单支付下单（官方文档 4012556926，多渠道版）
// 将 2-10 笔待支付订单合并为一笔合单订单统一下单（仅普通商户/直连模式支持）。
// 渠道路由：appKey+mchID 解析应用与商户；本地创建一条合单主单（order_type=combine）；
// 子单保持待支付，由合单支付成功回调（按子单逐笔通知）驱动子单与主单终态。
func (s *PaymentService) CombinePrepay(ctx context.Context, userID int64, openID string, appKey string, mchID string, paymentIDs []int64, clientIP string) (*entity.Payment, *JSAPIParams, error) {
	// 1. 解析支付渠道（应用+商户+协议客户端）
	channel, err := s.wechatPayService.ResolveChannel(ctx, appKey, mchID)
	if err != nil {
		return nil, nil, err
	}
	// 合单支付仅普通商户（直连）模式支持
	if channel.Client.IsPartner() {
		return nil, nil, errors.New("合单支付仅支持普通商户(直连)模式，服务商模式不可用")
	}
	appID := channel.App.AppID
	if appID == "" {
		return nil, nil, errors.New("应用未配置微信 AppID")
	}

	// 1. 数量校验（官方：2-10 笔）
	if len(paymentIDs) < 2 || len(paymentIDs) > 10 {
		return nil, nil, errors.New("合单支付订单数量须为2-10笔")
	}

	// 2. 逐笔校验子单（同用户、待支付、订单号存在、非医保/Native/合单类型）
	subs := make([]*entity.Payment, 0, len(paymentIDs))
	seen := make(map[int64]bool, len(paymentIDs))
	for _, id := range paymentIDs {
		if seen[id] {
			return nil, nil, errors.New("合单支付订单列表存在重复项")
		}
		seen[id] = true
		p, err := s.repo.GetByID(id)
		if err != nil {
			return nil, nil, fmt.Errorf("支付订单不存在: %d", id)
		}
		if p.UserID != userID {
			return nil, nil, errors.New("存在无权操作的订单")
		}
		if p.Status != entity.PaymentStatusPending {
			return nil, nil, fmt.Errorf("订单 %s 当前状态不可支付: %s", p.OrderID, p.Status)
		}
		if p.OrderType != entity.PaymentOrderTypeJSAPI && p.OrderType != "" {
			return nil, nil, fmt.Errorf("订单 %s 类型不支持合单支付", p.OrderID)
		}
		subs = append(subs, p)
	}

	// 4. 创建合单主单（本地）
	mchIDStr := channel.Merchant.MchID
	combineOutTradeNo := "WC" + time.Now().Format("20060102150405") + generateNonceStr()[:8]
	var total int64
	subIDs := make([]string, 0, len(subs))
	for _, p := range subs {
		total += p.Amount
		subIDs = append(subIDs, p.OrderID)
	}
	snapshot, _ := json.Marshal(map[string]interface{}{
		"combine_out_trade_no": combineOutTradeNo,
		"sub_order_ids":        subIDs,
	})
	combineOrder := &entity.Payment{
		OrderID:     combineOutTradeNo,
		UserID:      userID,
		Amount:      total,
		Currency:    "CNY",
		Method:      dictconst.UserTypeWechat,
		Status:      entity.PaymentStatusPending,
		Description: fmt.Sprintf("合单支付(%d笔)", len(subs)),
		OrderType:   entity.PaymentOrderTypeCombine,
		Openid:      openID,
		// WxLastResponse 暂存子单关联快照，回调核验用
		WxLastResponse: string(snapshot),
	}
	if err := s.repo.Create(combineOrder); err != nil {
		return nil, nil, fmt.Errorf("创建合单主单失败: %w", err)
	}

	// 5. 组装合单下单请求（官方字段：attach 子单必填）
	subOrders := make([]*wxpayv3.CombineSubOrder, 0, len(subs))
	for _, p := range subs {
		subOrders = append(subOrders, &wxpayv3.CombineSubOrder{
			Mchid:       mchIDStr,
			Attach:      p.OrderID, // 合单子单 attach 必填，回传子单商户订单号便于核验
			Amount:      &wxpayv3.CombineAmount{TotalAmount: p.Amount, Currency: "CNY"},
			OutTradeNo:  p.OrderID,
			Description: p.Description,
			SettleInfo:  &wxpayv3.SettleInfo{ProfitSharing: false},
		})
	}
	req := &wxpayv3.CombinePrepayRequest{
		CombineAppid:      appID,
		CombineMchid:      mchIDStr,
		CombineOutTradeNo: combineOutTradeNo,
		CombinePayerInfo:  &wxpayv3.CombinePayerInfo{Openid: openID},
		SubOrders:         subOrders,
		NotifyURL:         channel.PayNotifyURL,
	}
	if clientIP != "" {
		req.SceneInfo = &wxpayv3.CombineSceneInfo{PayerClientIP: clientIP}
	}

	// 6. 调用合单下单（该商户的协议客户端）
	prepayID, err := channel.Client.CombineJSAPIPrepay(req)
	if err != nil {
		// 下单失败：合单主单置失败，子单保持待支付可重新支付
		combineOrder.MarkAsFailed()
		_ = s.repo.Update(combineOrder)
		s.log.Error("合单下单失败", "combine_out_trade_no", combineOutTradeNo, "error", err)
		return nil, nil, fmt.Errorf("合单下单失败: %w", err)
	}

	// 7. 生成调起支付参数并保存主单微信支付信息（该商户私钥签名）
	v3Params, err := channel.Client.BuildJSAPIPayParams(prepayID, appID)
	if err != nil {
		return nil, nil, err
	}
	jsapiParams := toServiceJSAPIParams(v3Params)
	combineOrder.WechatPayInfo = &entity.WechatPayInfo{
		PaymentID: combineOrder.ID,
		AppID:     appID,
		MCHID:     mchIDStr,
		PrepayID:  prepayID,
		NonceStr:  jsapiParams.NonceStr,
		Timestamp: jsapiParams.TimeStamp,
		Package:   jsapiParams.Package,
		PaySign:   jsapiParams.PaySign,
		Sign:      jsapiParams.PaySign,
	}
	if err := s.repo.Update(combineOrder); err != nil {
		return nil, nil, err
	}
	if err := s.saveWechatPayInfo(combineOrder.ID, appID, mchIDStr, prepayID, jsapiParams); err != nil {
		s.log.Error("保存合单支付信息失败", "error", err)
	}

	s.log.Info("合单下单成功", "combine_out_trade_no", combineOutTradeNo, "子单数", len(subs), "总金额", total)
	return combineOrder, jsapiParams, nil
}

// HandleCombineNotify 处理合单支付成功回调（TRANSACTION.SUCCESS 且解密报文含 combine_out_trade_no）
// 合单回调按子单逐笔通知：更新对应子单支付状态；全部子单成功后置合单主单为已支付。
func (s *PaymentService) HandleCombineNotify(plaintext []byte) error {
	var combine wxpayv3.CombineNotifyResult
	if err := json.Unmarshal(plaintext, &combine); err != nil {
		return fmt.Errorf("解析合单回调业务数据失败: %w", err)
	}
	if combine.CombineOutTradeNo == "" || len(combine.SubOrders) == 0 {
		return errors.New("合单回调数据缺少 combine_out_trade_no 或 sub_orders")
	}

	allSuccess := true
	for _, sub := range combine.SubOrders {
		if sub.OutTradeNo == "" {
			continue
		}
		payment, err := s.repo.GetByOrderID(sub.OutTradeNo)
		if err != nil {
			s.log.Error("合单回调子单不存在", "out_trade_no", sub.OutTradeNo)
			allSuccess = false
			continue
		}
		switch sub.TradeState {
		case "SUCCESS":
			if payment.Status != entity.PaymentStatusSuccess {
				paidAt := time.Now()
				if sub.SuccessTime != "" {
					if t, e := time.Parse(time.RFC3339, sub.SuccessTime); e == nil {
						paidAt = t
					}
				}
				payment.MarkAsPaid(sub.TransactionID, paidAt)
				payment.TradeState = "SUCCESS"
				if err := s.repo.Update(payment); err != nil {
					return err
				}
				s.log.Info("合单子单支付成功", "out_trade_no", sub.OutTradeNo, "transaction_id", sub.TransactionID)
			}
		default:
			allSuccess = false
			payment.TradeState = sub.TradeState
			_ = s.repo.Update(payment)
		}
	}

	// 全部子单成功 → 合单主单置已支付
	if allSuccess {
		if combineOrder, err := s.repo.GetByOrderID(combine.CombineOutTradeNo); err == nil &&
			combineOrder.OrderType == entity.PaymentOrderTypeCombine &&
			combineOrder.Status != entity.PaymentStatusSuccess {
			combineOrder.MarkAsPaid("", time.Now())
			combineOrder.TradeState = "SUCCESS"
			if raw, e := json.Marshal(combine); e == nil {
				combineOrder.WxLastResponse = string(raw)
			}
			if err := s.repo.Update(combineOrder); err != nil {
				return err
			}
			s.log.Info("合单支付全部成功", "combine_out_trade_no", combine.CombineOutTradeNo)
		}
	}
	return nil
}
