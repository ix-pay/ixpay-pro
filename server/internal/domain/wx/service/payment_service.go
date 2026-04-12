package service

import (
	"context"
	"errors"
	"fmt"
	"time"

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
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(repo repo.PaymentRepository, taskManager *task.TaskManager, baseConfigRepository baseRepo.ConfigRepository, log logger.Logger) *PaymentService {
	return &PaymentService{
		repo:                 repo,
		log:                  log,
		taskManager:          taskManager,
		baseConfigRepository: baseConfigRepository,
	}
}

// CreatePayment 创建支付记录
func (s *PaymentService) CreatePayment(userID string, orderID string, amount int64, method string, description string) (*entity.Payment, error) {
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
func (s *PaymentService) GetPayment(paymentID string) (*entity.Payment, error) {
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

// UpdatePaymentStatus 更新支付状态
func (s *PaymentService) UpdatePaymentStatus(paymentID string, status entity.PaymentStatus) error {
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
func (s *PaymentService) CreateWechatPayment(userID string, orderID string, amount int64, description string) (*entity.Payment, error) {
	// 创建支付记录
	payment, err := s.CreatePayment(userID, orderID, amount, "wechat", description)
	if err != nil {
		return nil, err
	}

	// 生成微信支付参数
	payParams, err := s.GenerateWXAuthPayParams(orderID, amount, description)
	if err != nil {
		s.log.Error("生成微信支付参数失败", "error", err)
		return nil, err
	}

	// 创建微信支付信息
	wechatPayInfo := &entity.WechatPayInfo{
		PaymentID: payment.ID,
		AppID:     payParams["appId"].(string),
		NonceStr:  payParams["nonceStr"].(string),
		Sign:      payParams["paySign"].(string),
		Timestamp: payParams["timeStamp"].(string),
		Package:   payParams["package"].(string),
		PaySign:   payParams["paySign"].(string),
	}

	// 关联微信支付信息
	payment.WechatPayInfo = wechatPayInfo

	// 保存更新
	if err := s.repo.Update(payment); err != nil {
		s.log.Error("更新微信支付信息失败", "error", err)
		return nil, err
	}

	s.log.Info("微信支付创建成功", "paymentID", payment.ID)
	return payment, nil
}

// GenerateWXAuthPayParams 生成微信支付参数
func (s *PaymentService) GenerateWXAuthPayParams(orderID string, amount int64, description string) (map[string]interface{}, error) {
	// 从配置读取服务获取微信 AppID
	appID, err := s.baseConfigRepository.GetByKey("wechat_app_id")
	if err != nil {
		s.log.Error("获取微信 AppID 配置失败", "error", err)
		return nil, fmt.Errorf("获取微信配置失败：%w", err)
	}

	// 这里应该实现微信支付参数的生成逻辑
	// 实际实现需要调用微信支付 API，生成签名等

	// 简化返回模拟数据
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
	// 处理微信支付通知
	response, err := s.HandleWXAuthPayNotify(notifyData)
	if err != nil {
		s.log.Error("处理微信支付通知失败", "error", err)
		return nil, err
	}

	// 检查响应
	if response["return_code"] != "SUCCESS" {
		s.log.Error("微信支付通知失败", "return_msg", response["return_msg"])
		return nil, fmt.Errorf("微信支付通知失败：%s", response["return_msg"])
	}

	// 这里应该从通知数据中解析出订单信息和支付结果
	// 简化处理，假设通知数据中包含订单ID和交易ID

	// 更新支付状态为成功
	// 实际实现需要解析通知数据，获取支付 ID 和状态

	// 返回模拟的支付记录
	payment := &entity.Payment{
		Status:        entity.PaymentStatusSuccess,
		TransactionID: "wx1234567890",
	}

	s.log.Info("微信支付通知处理成功")
	return payment, nil
}

// HandleWXAuthPayNotify 处理微信支付通知
func (s *PaymentService) HandleWXAuthPayNotify(notifyData []byte) (map[string]interface{}, error) {
	// 这里应该实现微信支付通知的处理逻辑
	// 实际实现需要验证签名，解析通知数据等

	// 简化返回成功响应
	return map[string]interface{}{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	}, nil
}

// CheckPaymentStatus 检查支付状态
func (s *PaymentService) CheckPaymentStatus(paymentID string) (entity.PaymentStatus, error) {
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("获取支付记录失败", "error", err)
		return "", err
	}
	return payment.Status, nil
}

// CancelPayment 取消支付
func (s *PaymentService) CancelPayment(paymentID string) error {
	return s.UpdatePaymentStatus(paymentID, entity.PaymentStatusCancelled)
}

// RefundPayment 退款
func (s *PaymentService) RefundPayment(paymentID string, refundAmount int64) error {
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
	// 实际实现需要调用支付渠道的退款API

	// 更新支付状态为已退款
	return s.UpdatePaymentStatus(paymentID, entity.PaymentStatusRefunded)
}

// PaymentTimeoutCheckTask 支付超时检查任务
type PaymentTimeoutCheckTask struct {
	paymentID string
	service   *PaymentService
	log       logger.Logger
}

// GetName 获取任务名称
func (t *PaymentTimeoutCheckTask) GetName() string {
	return fmt.Sprintf("payment_timeout_check_%s", t.paymentID)
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
