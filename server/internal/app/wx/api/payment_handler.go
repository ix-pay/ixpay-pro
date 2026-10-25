package wxapi

import (
	"io"
	"net/http"
	"strconv"
	"time"

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

// CreatePayment 创建支付（待支付记录，创建后进入"待支付列表"）
// @Summary 创建支付记录
// @Description 创建一笔待支付订单，order_id 不传时由服务端自动生成；创建成功后可在待支付列表查看
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment body request.CreatePaymentRequest true "创建支付请求参数"
// @Success 201 {object} baseRes.Response{data=response.PaymentResponse} "支付创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/wx/payment [post]
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
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	paymentResponse := response.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        req.Amount,
		Currency:      payment.Currency,
		PaymentMethod: payment.Method,
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     payment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	baseRes.OkWithDetailed(paymentResponse, "创建支付成功", ctx)
}

// CreateUnifiedOrder 创建微信统一下单（Prepay）
// @Summary 微信统一下单
// @Description 对待支付订单发起微信 V3 JSAPI 下单，返回调起支付参数（公众号 H5 chooseWXPay 与小程序 wx.requestPayment 通用）；订单不存在时自动创建；已支付订单拒绝重复下单
// @Tags 微信支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CreateUnifiedOrderRequest true "统一下单请求参数"
// @Success 200 {object} baseRes.Response{data=response.UnifiedOrderResponse,msg=string} "统一下单成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/wx/payment/unified-order [post]
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
	if !exists || openID == nil || openID.(string) == "" {
		c.log.Error("缺少微信 openID")
		baseRes.FailWithMessage("缺少微信 openID", ctx)
		return
	}
	openIDStr := openID.(string)

	// 应用实例标志（登录态携带，多渠道路由）与可选商户号
	appKey, _ := ctx.Get("appKey")
	appKeyStr, _ := appKey.(string)

	// 金额：已存在的待支付订单以订单金额为准；新订单必传（元转分）
	var amount int64
	if req.Amount != nil {
		amount = int64(*req.Amount * 100)
	}

	// 调用服务层创建微信支付（V3 JSAPI 下单，按 app_key+mch_id 路由渠道；ctx 携带租户上下文路由租户库）
	payment, jsapiParams, err := c.service.CreateWechatPayment(
		ctx.Request.Context(),
		userIDInt,
		openIDStr,
		appKeyStr,
		req.MchID,
		req.OrderID,
		amount,
		req.Description,
	)
	if err != nil {
		c.log.Error("创建微信支付失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
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

// NativePrepay Native扫码下单
// @Summary Native扫码下单
// @Description 创建/复用待支付订单并发起 Native 下单，返回二维码链接 code_url(有效期2小时)，扫码拉起收银台
// @Tags 微信支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.NativePrepayRequest true "Native下单参数(金额单位：元)"
// @Success 200 {object} baseRes.Response{data=response.NativePrepayResponse} "二维码链接"
// @Router /api/wx/payment/native-prepay [post]
func (c *PaymentController) NativePrepay(ctx *gin.Context) {
	var req request.NativePrepayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}
	userIDInt, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	var amount int64
	if req.Amount != nil {
		amount = int64(*req.Amount * 100)
	}
	appKey, _ := ctx.Get("appKey")
	appKeyStr, _ := appKey.(string)
	payment, codeURL, err := c.service.NativePrepay(ctx.Request.Context(), userIDInt, appKeyStr, req.MchID, req.OrderID, amount, req.Description)
	if err != nil {
		c.log.Error("Native下单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(response.NativePrepayResponse{
		CodeURL:  codeURL,
		ExpireAt: payment.UpdatedAt.Add(2 * time.Hour).Format("2006-01-02 15:04:05"),
	}, "Native下单成功", ctx)
}

// WxQuery 微信侧订单查询
// @Summary 查询微信支付订单
// @Description 支持商户订单号(out_trade_no)与微信支付订单号(transaction_id)双方式查询微信侧订单状态，并同步本地支付单
// @Tags 微信支付
// @Produce json
// @Security BearerAuth
// @Param outTradeNo query string false "商户订单号"
// @Param transactionId query string false "微信支付订单号"
// @Success 200 {object} baseRes.Response{data=object} "微信侧订单状态"
// @Failure 400 {object} map[string]string "参数错误或查询失败"
// @Router /api/wx/payment/query [get]
func (c *PaymentController) WxQuery(ctx *gin.Context) {
	outTradeNo := ctx.Query("outTradeNo")
	transactionID := ctx.Query("transactionId")
	if outTradeNo == "" && transactionID == "" {
		baseRes.FailWithMessage("outTradeNo与transactionId至少传一个", ctx)
		return
	}
	result, err := c.service.WxQuery(outTradeNo, transactionID)
	if err != nil {
		c.log.Error("查询微信订单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(result, "查询微信订单成功", ctx)
}

// CombinePrepay JSAPI 合单支付下单
// @Summary 合单支付下单
// @Description 将 2-10 笔待支付订单合并为一笔合单订单统一下单(官方文档 4012556926，仅普通商户/直连模式支持)；返回合单订单号与调起支付参数，微信按子单逐笔回调
// @Tags 微信支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CombinePrepayRequest true "合单下单参数(待支付订单ID列表)"
// @Success 200 {object} baseRes.Response{data=response.CombinePrepayResponse} "合单订单号与调起支付参数"
// @Failure 400 {object} map[string]string "参数错误或下单失败"
// @Router /api/wx/payment/combine-prepay [post]
func (c *PaymentController) CombinePrepay(ctx *gin.Context) {
	var req request.CombinePrepayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误：payment_ids 须为2-10个支付订单ID", ctx)
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}
	userIDInt, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}
	openID, _ := ctx.Get("openID")
	openIDStr, _ := openID.(string)
	appKey, _ := ctx.Get("appKey")
	appKeyStr, _ := appKey.(string)

	ids := make([]int64, 0, len(req.PaymentIDs))
	for _, raw := range req.PaymentIDs {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			baseRes.FailWithMessage("支付订单ID格式错误: "+raw, ctx)
			return
		}
		ids = append(ids, id)
	}

	combineOrder, jsapiParams, err := c.service.CombinePrepay(ctx.Request.Context(), userIDInt, openIDStr, appKeyStr, req.MchID, ids, ctx.ClientIP())
	if err != nil {
		c.log.Error("合单下单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(response.CombinePrepayResponse{
		CombineOrderID: combineOrder.OrderID,
		TotalAmount:    float64(combineOrder.Amount) / 100.0,
		JSAPIParams: response.JSAPIParamsResponse{
			AppID:     jsapiParams.AppID,
			TimeStamp: jsapiParams.TimeStamp,
			NonceStr:  jsapiParams.NonceStr,
			Package:   jsapiParams.Package,
			SignType:  jsapiParams.SignType,
			PaySign:   jsapiParams.PaySign,
		},
	}, "合单下单成功", ctx)
}

// GetPayment 查询支付
// @Summary 查询支付详情
// @Description 根据 ID 查询支付详情；传 sync=true 时先向微信侧查单同步最新支付状态（用于支付后确认结果）
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "支付 ID"
// @Param sync query bool false "是否同步微信侧订单状态（默认 false）"
// @Success 200 {object} baseRes.Response{data=response.PaymentResponse,msg=string} "支付详情"
// @Failure 400 {object} map[string]string "无效的支付 ID"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/wx/payment/{id} [get]
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

	// sync=true 时向微信侧查单同步最新状态（用于支付完成后确认）
	if ctx.Query("sync") == "true" && payment.OrderID != "" {
		if synced, err := c.service.SyncPaymentStatus(payment.OrderID); err == nil && synced != nil {
			payment = synced
		} else if err != nil {
			c.log.Warn("同步微信侧订单状态失败", "orderID", payment.OrderID, "error", err)
		}
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
// @Description 获取当前登录用户的支付记录（按用户ID过滤）；传 status=pending 查询待支付列表
// @Tags 支付管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码（默认1）"
// @Param pageSize query int false "每页条数（默认10）"
// @Param status query string false "状态过滤（pending/success/failed/refunded/cancelled，默认全部）"
// @Success 200 {object} baseRes.Response{data=baseRes.PageResult,msg=string} "支付列表及分页信息"
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

	// 状态过滤（待支付列表传 pending）
	status := ctx.Query("status")

	// 调用服务层按用户ID查询支付记录（ctx 携带租户上下文路由租户库）
	payments, total, err := c.service.GetUserPayments(ctx.Request.Context(), userIDInt, status, page, pageSize)
	if err != nil {
		c.log.Error("获取用户支付列表失败", "error", err)
		baseRes.FailWithMessage("获取支付列表失败", ctx)
		return
	}

	// 构建响应
	paymentResponses := make([]response.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResponses = append(paymentResponses, response.PaymentResponse{
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

// HandleWechatPayNotify 处理微信支付通知（V3）
// @Summary 微信支付回调通知
// @Description 微信支付 APIv3 事件回调（TRANSACTION.SUCCESS 等），验签（微信支付公钥）+ AES-GCM 解密后更新本地支付状态；由微信服务端调用，业务方无需手动请求
// @Tags 支付管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "应答微信成功（code=SUCCESS）"
// @Failure 401 {object} map[string]string "验签失败（code=FAIL）"
// @Failure 500 {object} map[string]string "处理失败（code=FAIL）"
// @Router /api/wx/pay/notify/wechat [post]
func (c *PaymentController) HandleWechatPayNotify(ctx *gin.Context) {
	// 读取 JSON 请求体（V3 为 JSON 报文，非 V2 XML）
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		c.log.Error("读取通知请求体失败", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "读取报文失败"})
		return
	}

	if len(body) == 0 {
		c.log.Error("通知请求体为空")
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "请求体为空"})
		return
	}

	// 调用服务层处理通知（验签 + 解密 + 状态同步）
	_, err = c.service.HandleWechatPayNotify(ctx.Request.Context(), ctx.Request.Header, body)
	if err != nil {
		c.log.Error("处理微信支付通知失败", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "处理失败"})
		return
	}

	// 返回成功应答（V3 格式）
	ctx.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

// GetJSAPIConfig 获取微信 JS-SDK 配置
// @Summary 获取微信 JS-SDK 配置
// @Description 返回用于 wx.config() 的配置参数（appId、timestamp、nonceStr、signature），签名基于 jsapi_ticket + 当前页面 URL
// @Tags 微信支付
// @Accept json
// @Produce json
// @Param url query string true "当前网页 URL（不含 # 及其后部分）"
// @Success 200 {object} baseRes.Response{data=response.WXJSAPIConfigResponse,msg=string} "获取配置成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/wx/pay/jsapi-config [get]
func (c *PaymentController) GetJSAPIConfig(ctx *gin.Context) {
	pageURL := ctx.Query("url")
	if pageURL == "" {
		baseRes.FailWithMessage("url 参数不能为空", ctx)
		return
	}

	config, err := c.service.GetJSAPIConfig(pageURL)
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
