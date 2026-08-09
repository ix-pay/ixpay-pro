package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/dictconst"
	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
)

// PaymentService 实现支付领域服务接口
type PaymentService struct {
	repo                 repo.PaymentRepository
	log                  logger.Logger
	taskManager          *task.TaskManager
	baseConfigRepository baseRepo.ConfigRepository
	wechatPayService     *WechatPayService
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(repo repo.PaymentRepository, taskManager *task.TaskManager, baseConfigRepository baseRepo.ConfigRepository, log logger.Logger, wechatPayService *WechatPayService) *PaymentService {
	return &PaymentService{
		repo:                 repo,
		log:                  log,
		taskManager:          taskManager,
		baseConfigRepository: baseConfigRepository,
		wechatPayService:     wechatPayService,
	}
}

// CreatePayment 创建支付记录
func (s *PaymentService) CreatePayment(userID int64, orderID string, amount int64, method string, description string) (*entity.Payment, error) {
	// 检查订单 ID 是否已存在
	_, err := s.repo.GetByOrderID(orderID)
	if err == nil {
		return nil, errors.New("订单 ID 已存在")
	}

	// 创建支付记录
	payment := &entity.Payment{
		OrderID:     orderID,
		UserID:      userID,
		Amount:      amount,
		Method:      method,
		Status:      entity.PaymentStatusPending,
		Description: description,
	}

	// 保存支付记录
	if err := s.repo.Create(payment); err != nil {
		s.log.Error("创建支付记录失败", "error", err)
		return nil, err
	}

	// 添加支付超时检查任务（5 分钟后）
	timeoutTask := &PaymentTimeoutCheckTask{
		paymentID: payment.ID,
		service:   s,
		log:       s.log,
	}
	s.taskManager.AddOneTimeTask(timeoutTask, 5*time.Minute)

	s.log.Info("支付记录创建成功", "paymentID", payment.ID, "orderID", orderID)
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

// GetUserPayments 获取指定用户的所有支付记录（分页）
func (s *PaymentService) GetUserPayments(userID int64, page, pageSize int) ([]*entity.Payment, int, error) {
	payments, total, err := s.repo.ListByUser(userID, page, pageSize)
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

// CreateWechatPayment 创建微信支付
// openID 是用户的微信 openid，clientIP 是客户端 IP
func (s *PaymentService) CreateWechatPayment(userID int64, openID string, orderID string, amount int64, description string, clientIP string) (*entity.Payment, *JSAPIParams, error) {
	// 创建支付前，先从数据库加载微信配置（如果数据库有配置，覆盖 YAML 配置）
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}
	// 创建支付记录
	payment, err := s.CreatePayment(userID, orderID, amount, dictconst.UserTypeWechat, description)
	if err != nil {
		return nil, nil, err
	}

	// 调用微信统一下单 API
	prepayID, err := s.wechatPayService.UnifiedOrder(openID, orderID, description, clientIP, amount)
	if err != nil {
		s.log.Error("微信统一下单失败", "error", err)
		// 更新支付状态为失败
		_ = s.UpdatePaymentStatus(payment.ID, entity.PaymentStatusFailed)
		return nil, nil, err
	}

	// 生成 JSAPI 调起支付参数
	jsapiParams := s.wechatPayService.GenerateJSAPIParams(prepayID)

	// 创建微信支付信息
	wechatPayInfo := &entity.WechatPayInfo{
		PaymentID: payment.ID,
		AppID:     jsapiParams.AppID,
		MCHID:     s.wechatPayService.GetMCHID(),
		NonceStr:  jsapiParams.NonceStr,
		PrepayID:  prepayID,
		Timestamp: jsapiParams.TimeStamp,
		Package:   jsapiParams.Package,
		PaySign:   jsapiParams.PaySign,
		Sign:      jsapiParams.PaySign,
	}

	// 关联微信支付信息
	payment.WechatPayInfo = wechatPayInfo

	// 保存更新
	if err := s.repo.Update(payment); err != nil {
		s.log.Error("更新微信支付信息失败", "error", err)
		return nil, nil, err
	}

	s.log.Info("微信支付创建成功", "paymentID", payment.ID)
	return payment, jsapiParams, nil
}

// GenerateWXAuthPayParams 生成微信支付参数（保留兼容旧接口）
func (s *PaymentService) GenerateWXAuthPayParams(orderID string, amount int64, description string) (map[string]interface{}, error) {
	// 获取支付记录
	payment, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		s.log.Error("获取支付记录失败", "error", err)
		return nil, err
	}

	// 从配置读取服务获取微信 AppID
	appID, err := s.baseConfigRepository.GetByKey("wechat_app_id")
	if err != nil {
		s.log.Error("获取微信 AppID 配置失败", "error", err)
		return nil, fmt.Errorf("获取微信配置失败：%w", err)
	}

	// 如果已有 WechatPayInfo 则直接返回
	if payment.WechatPayInfo != nil {
		return map[string]interface{}{
			"appId":     payment.WechatPayInfo.AppID,
			"timeStamp": payment.WechatPayInfo.Timestamp,
			"nonceStr":  payment.WechatPayInfo.NonceStr,
			"package":   payment.WechatPayInfo.Package,
			"signType":  "MD5",
			"paySign":   payment.WechatPayInfo.PaySign,
		}, nil
	}

	// 降级：返回包含 appId 的简单信息（实际应走 CreateWechatPayment 流程）
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := generateNonceStr()

	return map[string]interface{}{
		"appId":     appID,
		"timeStamp": timeStamp,
		"nonceStr":  nonceStr,
		"package":   "prepay_id=wx1234567890",
		"signType":  "MD5",
		"paySign":   "1234567890abcdef",
	}, nil
}

// HandleWechatPayNotify 处理微信支付通知
func (s *PaymentService) HandleWechatPayNotify(notifyData []byte) (*entity.Payment, error) {
	// 验证通知签名并解析数据
	notify, err := s.wechatPayService.VerifyNotify(notifyData)
	if err != nil {
		s.log.Error("微信支付通知验证失败", "error", err)
		return nil, err
	}

	// 检查通信返回码
	if notify.ReturnCode != "SUCCESS" {
		s.log.Error("微信支付通知通信失败", "return_code", notify.ReturnCode, "return_msg", notify.ReturnMsg)
		return nil, fmt.Errorf("微信支付通知通信失败：%s", notify.ReturnMsg)
	}

	// 根据订单号查询支付记录
	payment, err := s.repo.GetByOrderID(notify.OutTradeNo)
	if err != nil {
		s.log.Error("根据订单号查询支付记录失败", "out_trade_no", notify.OutTradeNo, "error", err)
		return nil, err
	}

	// 检查业务结果
	if notify.ResultCode == "SUCCESS" {
		// 支付成功，更新状态
		payment.MarkAsPaid(notify.TransactionID, time.Now())

		// 更新微信支付信息
		if payment.WechatPayInfo == nil {
			payment.WechatPayInfo = &entity.WechatPayInfo{
				PaymentID:     payment.ID,
				TransactionID: notify.TransactionID,
				ReturnCode:    notify.ReturnCode,
				ReturnMsg:     notify.ReturnMsg,
				ResultCode:    notify.ResultCode,
				OpenID:        notify.OpenID,
				BankType:      notify.BankType,
				TotalFee:      notify.TotalFee,
				CashFee:       notify.CashFee,
				FeeType:       notify.FeeType,
				NotifyData:    string(notifyData),
			}
		} else {
			payment.WechatPayInfo.TransactionID = notify.TransactionID
			payment.WechatPayInfo.ReturnCode = notify.ReturnCode
			payment.WechatPayInfo.ReturnMsg = notify.ReturnMsg
			payment.WechatPayInfo.ResultCode = notify.ResultCode
			payment.WechatPayInfo.OpenID = notify.OpenID
			payment.WechatPayInfo.BankType = notify.BankType
			payment.WechatPayInfo.TotalFee = notify.TotalFee
			payment.WechatPayInfo.CashFee = notify.CashFee
			payment.WechatPayInfo.FeeType = notify.FeeType
			payment.WechatPayInfo.NotifyData = string(notifyData)
		}

		// 更新支付记录
		if err := s.repo.Update(payment); err != nil {
			s.log.Error("更新支付记录失败", "error", err)
			return nil, err
		}

		s.log.Info("微信支付成功", "out_trade_no", notify.OutTradeNo, "transaction_id", notify.TransactionID)
	} else {
		// 支付失败
		payment.MarkAsFailed()
		if err := s.repo.Update(payment); err != nil {
			s.log.Error("更新支付记录失败", "error", err)
			return nil, err
		}

		s.log.Warn("微信支付失败", "out_trade_no", notify.OutTradeNo, "err_code", notify.ErrCode, "err_code_des", notify.ErrCodeDes)
	}

	return payment, nil
}

// HandleWXAuthPayNotify 处理微信支付通知（旧接口，保留兼容）
func (s *PaymentService) HandleWXAuthPayNotify(notifyData []byte) (map[string]interface{}, error) {
	// 验证通知
	notify, err := s.wechatPayService.VerifyNotify(notifyData)
	if err != nil {
		return map[string]interface{}{
			"return_code": "FAIL",
			"return_msg":  err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"return_code": notify.ReturnCode,
		"return_msg":  notify.ReturnMsg,
	}, nil
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

// CancelPayment 取消支付
func (s *PaymentService) CancelPayment(paymentID int64) error {
	return s.UpdatePaymentStatus(paymentID, entity.PaymentStatusCancelled)
}

// GetJSAPIConfig 获取微信 JS-SDK 配置（用于 wx.config()）
// 返回 appId、timestamp、nonceStr、signature，供前端初始化 JS-SDK
func (s *PaymentService) GetJSAPIConfig() (*JSAPIConfig, error) {
	// 加载微信配置
	if err := s.wechatPayService.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}

	appID := s.wechatPayService.GetAppID()
	if appID == "" {
		return nil, fmt.Errorf("微信 AppID 未配置")
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := generateNonceStr()

	// 构建用于签名的字符串（JS-SDK 签名需要 jsapi_ticket，此处先返回基础信息）
	// TODO: 实际生产环境需要获取 jsapi_ticket 并生成签名
	signature := s.wechatPayService.GenerateSign(map[string]string{
		"appId":     appID,
		"timestamp": timestamp,
		"nonceStr":  nonceStr,
	})

	return &JSAPIConfig{
		AppID:     appID,
		Timestamp: timestamp,
		NonceStr:  nonceStr,
		Signature: signature,
	}, nil
}

// JSAPIConfig 微信 JS-SDK 配置
type JSAPIConfig struct {
	AppID     string `json:"appId"`
	Timestamp string `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Signature string `json:"signature"`
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

	// 这里应该实现退款逻辑
	// 实际实现需要调用支付渠道的退款 API

	// 更新支付状态为已退款
	return s.UpdatePaymentStatus(paymentID, entity.PaymentStatusRefunded)
}

// PaymentTimeoutCheckTask 支付超时检查任务
type PaymentTimeoutCheckTask struct {
	paymentID int64
	service   *PaymentService
	log       logger.Logger
}

// GetName 获取任务名称
func (t *PaymentTimeoutCheckTask) GetName() string {
	return fmt.Sprintf("payment_timeout_check_%d", t.paymentID)
}

// Run 运行任务
func (t *PaymentTimeoutCheckTask) Run(ctx context.Context) error {
	// 检查支付状态
	status, err := t.service.CheckPaymentStatus(t.paymentID)
	if err != nil {
		t.log.Error("检查支付状态失败", "error", err)
		return err
	}

	// 如果支付仍然是待支付状态，将其标记为失败
	if status == entity.PaymentStatusPending {
		if err := t.service.UpdatePaymentStatus(t.paymentID, entity.PaymentStatusFailed); err != nil {
			t.log.Error("更新支付状态为失败失败", "error", err)
			return err
		}
		t.log.Info("支付超时，标记为失败", "paymentID", t.paymentID)
	} else {
		t.log.Info("支付非待支付状态，无需操作", "paymentID", t.paymentID, "status", status)
	}

	return nil
}