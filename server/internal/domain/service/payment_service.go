package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/task"
)

// PaymentService 实现支付领域服务接口
type PaymentService struct {
	repo        model.PaymentRepository
	log         logger.Logger
	wechatSvc   *WechatService
	taskManager *task.TaskManager
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(repo model.PaymentRepository, log logger.Logger, wechatSvc *WechatService, taskManager *task.TaskManager) model.PaymentService {
	return &PaymentService{
		repo:        repo,
		log:         log,
		wechatSvc:   wechatSvc,
		taskManager: taskManager,
	}
}

// CreatePayment 创建支付记录
func (s *PaymentService) CreatePayment(userID uint, orderID string, amount int64, method string, description string) (*model.Payment, error) {
	// 检查订单ID是否已存在
	_, err := s.repo.GetByOrderID(orderID)
	if err == nil {
		return nil, errors.New("order ID already exists")
	}

	// 创建支付记录
	payment := &model.Payment{
		OrderID:     orderID,
		UserID:      userID,
		Amount:      amount,
		Method:      method,
		Status:      model.PaymentStatusPending,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存支付记录
	if err := s.repo.Create(payment); err != nil {
		s.log.Error("Failed to create payment", "error", err)
		return nil, err
	}

	// 添加支付超时检查任务（5分钟后）
	timeoutTask := &PaymentTimeoutCheckTask{
		paymentID: payment.ID,
		service:   s,
		log:       s.log,
	}
	s.taskManager.AddOneTimeTask(timeoutTask, 5*time.Minute)

	s.log.Info("Payment created successfully", "paymentID", payment.ID, "orderID", orderID)
	return payment, nil
}

// GetPayment 获取支付记录
func (s *PaymentService) GetPayment(paymentID uint) (*model.Payment, error) {
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("Failed to get payment", "error", err)
		return nil, err
	}
	return payment, nil
}

// GetPaymentByOrderID 根据订单ID获取支付记录
func (s *PaymentService) GetPaymentByOrderID(orderID string) (*model.Payment, error) {
	payment, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		s.log.Error("Failed to get payment by order ID", "error", err)
		return nil, err
	}
	return payment, nil
}

// UpdatePaymentStatus 更新支付状态
func (s *PaymentService) UpdatePaymentStatus(paymentID uint, status model.PaymentStatus) error {
	// 获取支付记录
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("Failed to get payment", "error", err)
		return err
	}

	// 更新状态
	payment.Status = status
	payment.UpdatedAt = time.Now()

	// 如果是支付成功，记录支付时间
	if status == model.PaymentStatusSuccess && payment.PaidAt == nil {
		now := time.Now()
		payment.PaidAt = &now
	}

	// 保存更新
	if err := s.repo.Update(payment); err != nil {
		s.log.Error("Failed to update payment status", "error", err)
		return err
	}

	s.log.Info("Payment status updated", "paymentID", paymentID, "status", status)
	return nil
}

// CreateWechatPayment 创建微信支付
func (s *PaymentService) CreateWechatPayment(userID uint, orderID string, amount int64, description string) (*model.Payment, error) {
	// 创建支付记录
	payment, err := s.CreatePayment(userID, orderID, amount, "wechat", description)
	if err != nil {
		return nil, err
	}

	// 生成微信支付参数
	payParams, err := s.wechatSvc.GenerateWechatPayParams(orderID, amount, description)
	if err != nil {
		s.log.Error("Failed to generate WeChat pay params", "error", err)
		return nil, err
	}

	// 创建微信支付信息
	wechatPayInfo := &model.WechatPayInfo{
		PaymentID: payment.ID,
		AppID:     payParams["appId"].(string),
		NonceStr:  payParams["nonceStr"].(string),
		Sign:      payParams["paySign"].(string),
		Timestamp: payParams["timeStamp"].(string),
		Package:   payParams["package"].(string),
		PaySign:   payParams["paySign"].(string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 关联微信支付信息
	payment.WechatPayInfo = wechatPayInfo

	// 保存更新
	if err := s.repo.Update(payment); err != nil {
		s.log.Error("Failed to update payment with WeChat info", "error", err)
		return nil, err
	}

	s.log.Info("WeChat payment created successfully", "paymentID", payment.ID)
	return payment, nil
}

// HandleWechatPayNotify 处理微信支付通知
func (s *PaymentService) HandleWechatPayNotify(notifyData []byte) (*model.Payment, error) {
	// 处理微信支付通知
	response, err := s.wechatSvc.HandleWechatPayNotify(notifyData)
	if err != nil {
		s.log.Error("Failed to handle WeChat pay notify", "error", err)
		return nil, err
	}

	// 检查响应
	if response["return_code"] != "SUCCESS" {
		s.log.Error("WeChat pay notify failed", "return_msg", response["return_msg"])
		return nil, fmt.Errorf("wechat pay notify failed: %s", response["return_msg"])
	}

	// 这里应该从通知数据中解析出订单信息和支付结果
	// 简化处理，假设通知数据中包含订单ID和交易ID

	// 更新支付状态为成功
	// 实际实现需要解析通知数据，获取支付ID和状态

	// 返回模拟的支付记录
	payment := &model.Payment{
		Status:        model.PaymentStatusSuccess,
		TransactionID: "wx1234567890",
	}

	s.log.Info("WeChat pay notify handled successfully")
	return payment, nil
}

// CheckPaymentStatus 检查支付状态
func (s *PaymentService) CheckPaymentStatus(paymentID uint) (model.PaymentStatus, error) {
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("Failed to get payment", "error", err)
		return "", err
	}
	return payment.Status, nil
}

// CancelPayment 取消支付
func (s *PaymentService) CancelPayment(paymentID uint) error {
	return s.UpdatePaymentStatus(paymentID, model.PaymentStatusCancelled)
}

// RefundPayment 退款
func (s *PaymentService) RefundPayment(paymentID uint, refundAmount int64) error {
	// 获取支付记录
	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		s.log.Error("Failed to get payment", "error", err)
		return err
	}

	// 检查支付状态
	if payment.Status != model.PaymentStatusSuccess {
		s.log.Error("Cannot refund payment with status", "status", payment.Status)
		return errors.New("only successful payments can be refunded")
	}

	// 检查退款金额
	if refundAmount <= 0 || refundAmount > payment.Amount {
		s.log.Error("Invalid refund amount", "refundAmount", refundAmount, "paymentAmount", payment.Amount)
		return errors.New("invalid refund amount")
	}

	// 这里应该实现退款逻辑
	// 实际实现需要调用支付渠道的退款API

	// 更新支付状态为已退款
	return s.UpdatePaymentStatus(paymentID, model.PaymentStatusRefunded)
}

// PaymentTimeoutCheckTask 支付超时检查任务
type PaymentTimeoutCheckTask struct {
	paymentID uint
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
		t.log.Error("Failed to check payment status", "error", err)
		return err
	}

	// 如果支付仍然是待支付状态，将其标记为失败
	if status == model.PaymentStatusPending {
		if err := t.service.UpdatePaymentStatus(t.paymentID, model.PaymentStatusFailed); err != nil {
			t.log.Error("Failed to update payment status to failed", "error", err)
			return err
		}
		t.log.Info("Payment marked as failed due to timeout", "paymentID", t.paymentID)
	} else {
		t.log.Info("Payment is not pending, no action needed", "paymentID", t.paymentID, "status", status)
	}

	return nil
}
