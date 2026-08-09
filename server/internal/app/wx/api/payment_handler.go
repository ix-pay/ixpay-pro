package wxapi

import (
	"fmt"
	"io"
	"strconv"

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

	// 将 userID 从 string 转换为 int64
	userIDInt, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		c.log.Error("用户 ID 格式错误", "error", err)
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	// 将金额从元转换为分
	amount := int64(req.Amount * 100)

	payment, err := c.service.CreatePayment(
		userIDInt,
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
		ID:            fmt.Sprintf("%d", payment.ID),
		OrderID:       payment.OrderID,
		UserID:        fmt.Sprintf("%d", payment.UserID),
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

// CreateUnifiedOrder 创建微信统一下单
// @Summary 微信统一下单
// @Description 创建微信支付统一下单，返回 JSAPI 调起支付参数
// @Tags 微信支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CreateUnifiedOrderRequest true "统一下单请求参数"
// @Success 200 {object} baseRes.Response{data=response.UnifiedOrderResponse,msg=string} "统一下单成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/wx/pay/unified-order [post]
func (c *PaymentController) CreateUnifiedOrder(ctx *gin.Context) {
	var req request.CreateUnifiedOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户 ID 和 openID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	userIDInt, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		c.log.Error("用户 ID 格式错误", "error", err)
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	// 从上下文中获取 openID（由认证中间件设置）
	openID, exists := ctx.Get("openID")
	if !exists {
		c.log.Error("缺少微信 openID")
		baseRes.FailWithMessage("缺少微信 openID", ctx)
		return
	}
	openIDStr := openID.(string)

	// 获取客户端 IP
	clientIP := ctx.ClientIP()

	// 将金额从元转换为分
	amount := int64(req.Amount * 100)

	// 调用服务层创建微信支付
	payment, jsapiParams, err := c.service.CreateWechatPayment(
		userIDInt,
		openIDStr,
		req.OrderID,
		amount,
		req.Description,
		clientIP,
	)
	if err != nil {
		c.log.Error("创建微信支付失败", "error", err)
		baseRes.FailWithMessage("创建微信支付失败", ctx)
		return
	}

	// 构建响应
	resp := response.UnifiedOrderResponse{
		PrepayID: payment.WechatPayInfo.PrepayID,
		JSAPIParams: response.JSAPIParamsResponse{
			AppID:     jsapiParams.AppID,
			TimeStamp: jsapiParams.TimeStamp,
			NonceStr:  jsapiParams.NonceStr,
			Package:   jsapiParams.Package,
			SignType:  jsapiParams.SignType,
			PaySign:   jsapiParams.PaySign,
		},
	}

	baseRes.OkWithDetailed(resp, "创建微信支付成功", ctx)
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

	// 将 paymentID 从 string 转换为 int64
	paymentIDInt, err := strconv.ParseInt(paymentID, 10, 64)
	if err != nil {
		c.log.Error("支付 ID 格式错误", "error", err)
		baseRes.FailWithMessage("支付 ID 格式错误", ctx)
		return
	}

	payment, err := c.service.GetPayment(paymentIDInt)
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

	userIDInt, _ := strconv.ParseInt(userID.(string), 10, 64)
	if payment.UserID != userIDInt {
		c.log.Error("无权限")
		baseRes.NoAuth("无权限", ctx)
		return
	}

	// 构建响应，将金额从分转换为元
	paymentResponse := response.PaymentResponse{
		ID:            fmt.Sprintf("%d", payment.ID),
		OrderID:       payment.OrderID,
		UserID:        fmt.Sprintf("%d", payment.UserID),
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
// @Description 获取当前登录用户的所有支付记录（按用户ID过滤）
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码（默认1）"
// @Param pageSize query int false "每页条数（默认10）"
// @Success 200 {object} map[string]interface{} "支付列表及分页信息"
// @Failure 401 {object} map[string]string "未授权"
// @Router /api/wx/payment [get]
func (c *PaymentController) GetUserPayments(ctx *gin.Context) {
	// 从上下文中获取用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	userIDInt, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		c.log.Error("用户 ID 格式错误", "error", err)
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 调用服务层按用户ID查询支付记录
	payments, total, err := c.service.GetUserPayments(userIDInt, page, pageSize)
	if err != nil {
		c.log.Error("获取用户支付列表失败", "error", err)
		baseRes.FailWithMessage("获取支付列表失败", ctx)
		return
	}

	// 构建响应
	paymentResponses := make([]response.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResponses = append(paymentResponses, response.PaymentResponse{
			ID:            fmt.Sprintf("%d", payment.ID),
			OrderID:       payment.OrderID,
			UserID:        fmt.Sprintf("%d", payment.UserID),
			Amount:        float64(payment.Amount) / 100.0,
			Currency:      payment.Currency,
			PaymentMethod: payment.Method,
			Status:        string(payment.Status),
			TransactionID: payment.TransactionID,
			Description:   payment.Description,
			CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     payment.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	result := baseRes.PageResult{
		List:     paymentResponses,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}
	baseRes.OkWithDetailed(result, "查询支付成功", ctx)
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

	// 将 paymentID 从 string 转换为 int64
	paymentIDInt, err := strconv.ParseInt(paymentID, 10, 64)
	if err != nil {
		c.log.Error("支付 ID 格式错误", "error", err)
		baseRes.FailWithMessage("支付 ID 格式错误", ctx)
		return
	}

	// 将 userID 从 string 转换为 int64
	userIDInt, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		c.log.Error("用户 ID 格式错误", "error", err)
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	// 获取支付记录以验证用户权限
	payment, err := c.service.GetPayment(paymentIDInt)
	if err != nil {
		baseRes.FailWithMessage("取消支付失败", ctx)
		return
	}

	// 检查权限
	if payment.UserID != userIDInt {
		c.log.Error("无权限")
		baseRes.NoAuth("无权限", ctx)
		return
	}

	// 取消支付
	err = c.service.CancelPayment(paymentIDInt)
	if err != nil {
		baseRes.FailWithMessage("取消支付失败", ctx)
		return
	}

	// 重新获取支付记录以获取更新后的状态
	payment, err = c.service.GetPayment(paymentIDInt)
	if err != nil {
		baseRes.FailWithMessage("取消支付失败", ctx)
		return
	}

	// 构建响应，将金额从分转换为元
	paymentResponse := response.PaymentResponse{
		ID:            fmt.Sprintf("%d", payment.ID),
		OrderID:       payment.OrderID,
		UserID:        fmt.Sprintf("%d", payment.UserID),
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
// @Accept xml
// @Produce xml
// @Success 200 {string} string "成功响应"
// @Router /api/wx/pay/notify/wechat [post]
func (c *PaymentController) HandleWechatPayNotify(ctx *gin.Context) {
	// 读取 XML 请求体
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		c.log.Error("读取通知请求体失败", "error", err)
		ctx.Data(200, "application/xml", wxService.BuildFailXMLResponse("读取请求体失败"))
		return
	}

	if len(body) == 0 {
		c.log.Error("通知请求体为空")
		ctx.Data(200, "application/xml", wxService.BuildFailXMLResponse("请求体为空"))
		return
	}

	// 调用服务层处理通知
	_, err = c.service.HandleWechatPayNotify(body)
	if err != nil {
		c.log.Error("处理微信支付通知失败", "error", err)
		ctx.Data(200, "application/xml", wxService.BuildFailXMLResponse(err.Error()))
		return
	}

	// 返回成功响应
	ctx.Data(200, "application/xml", wxService.BuildSuccessXMLResponse())
}

// GetJSAPIConfig 获取微信 JS-SDK 配置
// @Summary 获取微信 JS-SDK 配置
// @Description 返回用于 wx.config() 的配置参数（appId、timestamp、nonceStr、signature）
// @Tags 微信支付
// @Accept json
// @Produce json
// @Success 200 {object} baseRes.Response{data=response.WXJSAPIConfigResponse,msg=string} "获取配置成功"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/wx/pay/jsapi-config [get]
func (c *PaymentController) GetJSAPIConfig(ctx *gin.Context) {
	config, err := c.service.GetJSAPIConfig()
	if err != nil {
		c.log.Error("获取微信 JS-SDK 配置失败", "error", err)
		baseRes.FailWithMessage("获取微信配置失败", ctx)
		return
	}

	resp := response.WXJSAPIConfigResponse{
		AppID:     config.AppID,
		Timestamp: config.Timestamp,
		NonceStr:  config.NonceStr,
		Signature: config.Signature,
	}

	baseRes.OkWithDetailed(resp, "获取微信 JS-SDK 配置成功", ctx)
}