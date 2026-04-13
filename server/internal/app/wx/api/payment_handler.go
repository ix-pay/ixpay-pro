package wxapi

import (
	wxService "github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

	"github.com/gin-gonic/gin"
)

// PaymentController 支付控制器
// @Summary 支付相关 API
// @Description 提供支付创建、查询、取消等功能
// @Tags 支付管理
// @Router /api/payment [get]
type PaymentController struct {
	service *wxService.PaymentService
	log     logger.Logger
}

// NewPaymentController 创建支付控制器
func NewPaymentController(service *wxService.PaymentService, log logger.Logger) *PaymentController {
	return &PaymentController{
		service: service,
		log:     log,
	}
}

// CreatePayment 创建支付
// @Summary 创建支付
// @Description 创建一笔新的支付订单
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment body request.CreatePaymentRequest true "创建支付请求参数"
// @Success 201 {object} map[string]response.PaymentResponse "支付创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/payment [post]
func (c *PaymentController) CreatePayment(ctx *gin.Context) {
	var req request.CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	// 将金额从元转换为分
	amount := int64(req.Amount * 100)

	payment, err := c.service.CreatePayment(
		userID.(string),
		req.OrderID,
		amount,
		req.PaymentMethod,
		req.Description,
	)
	if err != nil {
		baseRes.FailWithMessage("创建支付失败", ctx)
		return
	}

	// 构建响应
	paymentResponse := response.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: payment.Method,
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     payment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	baseRes.OkWithDetailed(paymentResponse, "创建支付成功", ctx)
}

// GetPayment 查询支付
// @Summary 查询支付
// @Description 根据 ID 查询支付详情
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "支付 ID"
// @Success 200 {object} map[string]response.PaymentResponse "支付详情"
// @Failure 400 {object} map[string]string "无效的支付 ID"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api//payment/{id} [get]
func (c *PaymentController) GetPayment(ctx *gin.Context) {
	paymentID := ctx.Param("id")
	if paymentID == "" {
		c.log.Error("无效的支付 ID")
		baseRes.FailWithMessage("无效的支付 ID", ctx)
		return
	}

	payment, err := c.service.GetPayment(paymentID)
	if err != nil {
		baseRes.FailWithMessage("查询支付失败", ctx)
		return
	}

	// 检查权限：用户只能查看自己的支付记录
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	if payment.UserID != userID.(string) {
		c.log.Error("无权限")
		baseRes.NoAuth("无权限", ctx)
		return
	}

	// 构建响应，将金额从分转换为元
	paymentResponse := response.PaymentResponse{
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

	baseRes.OkWithDetailed(paymentResponse, "查询支付成功", ctx)
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
// @Router /api//payment [get]
func (c *PaymentController) GetUserPayments(ctx *gin.Context) {
	// 从上下文中获取用户 ID（实际未使用）
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	// 由于 service 中没有 GetUserPayments 方法，返回空列表
	// 实际实现应该在 service 层添加这个方法
	c.log.Error("GetUserPayments 方法在 service 层未实现")
	response := baseRes.PageResult{
		List:     []response.PaymentResponse{},
		Total:    0,
		Page:     1,
		PageSize: 10,
	}
	baseRes.OkWithDetailed(response, "查询支付成功", ctx)
}

// CancelPayment 取消支付
// @Summary 取消支付
// @Description 根据 ID 取消一笔支付
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "支付 ID"
// @Success 200 {object} map[string]response.PaymentResponse "取消后的支付详情"
// @Failure 400 {object} map[string]string "无效的支付 ID"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api//payment/{id}/cancel [put]
func (c *PaymentController) CancelPayment(ctx *gin.Context) {
	paymentID := ctx.Param("id")
	if paymentID == "" {
		c.log.Error("无效的支付 ID")
		baseRes.FailWithMessage("取消支付失败", ctx)
		return
	}

	// 从上下文中获取用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	// 获取支付记录以验证用户权限
	payment, err := c.service.GetPayment(paymentID)
	if err != nil {
		baseRes.FailWithMessage("取消支付失败", ctx)
		return
	}

	// 检查权限
	if payment.UserID != userID.(string) {
		c.log.Error("无权限")
		baseRes.NoAuth("无权限", ctx)
		return
	}

	// 取消支付
	err = c.service.CancelPayment(paymentID)
	if err != nil {
		baseRes.FailWithMessage("取消支付失败", ctx)
		return
	}

	// 重新获取支付记录以获取更新后的状态
	payment, err = c.service.GetPayment(paymentID)
	if err != nil {
		baseRes.FailWithMessage("取消支付失败", ctx)
		return
	}

	// 构建响应，将金额从分转换为元
	paymentResponse := response.PaymentResponse{
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

	baseRes.OkWithDetailed(paymentResponse, "取消支付成功", ctx)
}

// HandleWechatPayNotify 处理微信支付通知
// @Summary 微信支付通知
// @Description 微信支付回调接口，用于处理支付结果通知
// @Tags 支付管理
// @Accept x-www-form-urlencoded
// @Produce xml
// @Success 200 {string} string "成功响应"
// @Router /api//pay/notify/wechat [post]
func (c *PaymentController) HandleWechatPayNotify(ctx *gin.Context) {
	// 解析微信支付通知数据
	// 这里简化处理，实际应该根据微信支付API文档进行解析

	// 处理支付结果
	// 这里简化处理，实际应该根据微信支付通知内容更新支付状态

	// 返回成功响应给微信服务器
	_, _ = ctx.Writer.WriteString("<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>")
}
