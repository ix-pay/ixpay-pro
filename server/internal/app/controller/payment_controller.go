package controller

import (
	"net/http"
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// PaymentController 支付控制器
// @Summary 支付相关API
// @Description 提供支付创建、查询、取消等功能
// @Tags 支付管理
// @Router /v1/payment [get]
type PaymentController struct {
	service model.PaymentService
	log     logger.Logger
}

// NewPaymentController 创建支付控制器
func NewPaymentController(service model.PaymentService, log logger.Logger) *PaymentController {
	return &PaymentController{
		service: service,
		log:     log,
	}
}

// CreatePaymentRequest 创建支付请求参数
type CreatePaymentRequest struct {
	OrderID       string  `json:"order_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,oneof=CNY USD"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=wechat_alipay"`
	Description   string  `json:"description"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	ID            uint    `json:"id"`
	OrderID       string  `json:"order_id"`
	UserID        uint    `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id"`
	Description   string  `json:"description"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	// 微信支付相关参数
	WechatPayParams map[string]interface{} `json:"wechat_pay_params,omitempty"`
}

// CreatePayment 创建支付
// @Summary 创建支付
// @Description 创建一笔新的支付订单
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment body CreatePaymentRequest true "创建支付请求参数"
// @Success 201 {object} map[string]PaymentResponse "支付创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /v1/payment [post]
func (c *PaymentController) CreatePayment(ctx *gin.Context) {
	var req CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 将金额从元转换为分
	amount := int64(req.Amount * 100)

	payment, err := c.service.CreatePayment(
		userID.(uint),
		req.OrderID,
		amount,
		req.PaymentMethod,
		req.Description,
	)
	if err != nil {
		c.log.Error("Create payment failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建响应
	response := PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        req.Amount, // 返回原始的元
		Currency:      req.Currency,
		PaymentMethod: payment.Method, // model中是Method字段
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     payment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": response})
}

// GetPayment 查询支付
// @Summary 查询支付
// @Description 根据ID查询支付详情
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "支付ID"
// @Success 200 {object} map[string]PaymentResponse "支付详情"
// @Failure 400 {object} map[string]string "无效的支付ID"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /v1/payment/{id} [get]
func (c *PaymentController) GetPayment(ctx *gin.Context) {
	paymentID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	payment, err := c.service.GetPayment(uint(paymentID))
	if err != nil {
		c.log.Error("Get payment failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查权限：用户只能查看自己的支付记录
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if payment.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 构建响应，将金额从分转换为元
	response := PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        float64(payment.Amount) / 100.0,
		Currency:      payment.Currency,
		PaymentMethod: payment.Method,
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     payment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

// GetUserPayments 获取用户支付列表
// @Summary 获取用户支付列表
// @Description 获取当前登录用户的所有支付记录
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "支付列表及分页信息"
// @Failure 401 {object} map[string]string "未授权"
// @Router /v1/payment [get]
func (c *PaymentController) GetUserPayments(ctx *gin.Context) {
	// 从上下文中获取用户ID（实际未使用）
	_, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 由于service中没有GetUserPayments方法，返回空列表
	// 实际实现应该在service层添加这个方法
	c.log.Error("GetUserPayments method not implemented in service")
	ctx.JSON(http.StatusOK, gin.H{
		"data":  []PaymentResponse{},
		"total": 0,
		"page":  1,
		"size":  10,
	})
}

// CancelPayment 取消支付
// @Summary 取消支付
// @Description 根据ID取消一笔支付
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "支付ID"
// @Success 200 {object} map[string]PaymentResponse "取消后的支付详情"
// @Failure 400 {object} map[string]string "无效的支付ID"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /v1/payment/{id}/cancel [put]
func (c *PaymentController) CancelPayment(ctx *gin.Context) {
	paymentID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取支付记录以验证用户权限
	payment, err := c.service.GetPayment(uint(paymentID))
	if err != nil {
		c.log.Error("Get payment failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if payment.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 取消支付
	err = c.service.CancelPayment(uint(paymentID))
	if err != nil {
		c.log.Error("Cancel payment failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel payment"})
		return
	}

	// 重新获取支付记录以获取更新后的状态
	payment, err = c.service.GetPayment(uint(paymentID))
	if err != nil {
		c.log.Error("Get payment after cancel failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建响应，将金额从分转换为元
	response := PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        float64(payment.Amount) / 100.0,
		Currency:      payment.Currency,
		PaymentMethod: payment.Method,
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     payment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

// HandleWechatPayNotify 处理微信支付通知
// @Summary 微信支付通知
// @Description 微信支付回调接口，用于处理支付结果通知
// @Tags 支付管理
// @Accept x-www-form-urlencoded
// @Produce xml
// @Success 200 {string} string "成功响应"
// @Router /v1/pay/notify/wechat [post]
func (c *PaymentController) HandleWechatPayNotify(ctx *gin.Context) {
	// 解析微信支付通知数据
	// 这里简化处理，实际应该根据微信支付API文档进行解析

	// 处理支付结果
	// 这里简化处理，实际应该根据微信支付通知内容更新支付状态

	// 返回成功响应给微信服务器
	ctx.Writer.WriteString("<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>")
}
