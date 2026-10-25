package wxapi

import (
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	medService "github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

	"github.com/gin-gonic/gin"
)

// MedicalController 医保支付控制器（对齐 Gpay handler/medical.go）
// @Summary 医保支付相关 API
// @Description 医保免密授权、医保自费混合下单、查单、收款回调
// @Tags 医保支付
type MedicalController struct {
	service *medService.MedicalService
	log     logger.Logger
}

// NewMedicalController 创建医保支付控制器
func NewMedicalController(service *medService.MedicalService, log logger.Logger) *MedicalController {
	return &MedicalController{service: service, log: log}
}

// GetMedAuthURL 医保免密授权链接
// @Summary 医保免密授权链接
// @Description 生成医保支付免密授权链接(本人/代亲人)；用户微信内打开完成授权后回跳 redirect?retCode=0&authCode=xxx
// @Tags 医保支付
// @Produce json
// @Security BearerAuth
// @Param pay_for_relatives query bool false "是否代亲人支付授权(默认 false)"
// @Param rel_name query string false "亲属姓名(代亲人时必填)"
// @Param rel_card_no query string false "亲属身份证号(代亲人时必填)"
// @Param redirect query string false "授权后回跳地址(空则用前端地址)"
// @Success 200 {object} baseRes.Response{data=map[string]string} "授权链接"
// @Failure 412 {object} map[string]string "医保授权未配置"
// @Router /api/wx/med/auth/url [get]
func (c *MedicalController) GetMedAuthURL(ctx *gin.Context) {
	var req request.MedAuthURLRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	url, err := c.service.MedAuthURL(req.PayForRelatives, req.RelName, req.RelCardNo, req.Redirect)
	if err != nil {
		c.log.Error("生成医保授权链接失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(gin.H{"url": url}, "医保授权链接生成成功", ctx)
}

// MedAuthQuery 医保授权查询（authCode 换 payAuthNo）
// @Summary 医保授权查询
// @Description 用授权回跳的 authCode 换取 payAuthNo 与用户经纬度(openid 取自登录态)
// @Tags 医保支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.MedAuthQueryRequest true "授权查询请求参数"
// @Success 200 {object} baseRes.Response{data=response.MedAuthQueryResponse} "payAuthNo 与用户信息"
// @Router /api/wx/med/auth/query [post]
func (c *MedicalController) MedAuthQuery(ctx *gin.Context) {
	var req request.MedAuthQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	openID, _ := ctx.Get("openID")
	openIDStr, _ := openID.(string)

	resp, err := c.service.MedQueryAuth(req.AuthCode, openIDStr)
	if err != nil {
		c.log.Error("医保授权查询失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	out := response.MedAuthQueryResponse{
		UserName:  resp.UserName,
		CityID:    resp.CityID,
		PayAuthNo: resp.PayAuthNo,
	}
	if resp.GeoLocation != nil {
		out.Longitude = resp.GeoLocation.Longitude
		out.Latitude = resp.GeoLocation.Latitude
	}
	baseRes.OkWithDetailed(out, "医保授权查询成功", ctx)
}

// CreateMedOrder 创建医保订单
// @Summary 创建医保订单
// @Description 创建医保自费混合收款订单（本地待支付，order_type=medical），随后调用 /med/prepay 混合下单
// @Tags 医保支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CreateMedOrderRequest true "医保订单创建参数(金额单位：分)"
// @Success 200 {object} baseRes.Response{data=response.PaymentResponse} "订单创建成功"
// @Router /api/wx/med/orders [post]
func (c *MedicalController) CreateMedOrder(ctx *gin.Context) {
	var req request.CreateMedOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	userID, exists := ctx.Get("userID")
	if !exists {
		baseRes.NoAuth("未授权", ctx)
		return
	}
	userIDInt, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	med := &entity.MedOrderInfo{
		OrderType:                 req.MedInfo.OrderType,
		SerialNo:                  req.MedInfo.SerialNo,
		PayOrderID:                req.MedInfo.PayOrderID,
		PayAuthNo:                 req.MedInfo.PayAuthNo,
		GeoLocation:               req.MedInfo.GeoLocation,
		PassthroughRequestContent: req.MedInfo.PassthroughRequestContent,
		PayerName:                 req.MedInfo.PayerName,
		PayerIDCard:               req.MedInfo.PayerIDCard,
		PayForRelatives:           req.MedInfo.PayForRelatives,
		RelativeName:              req.MedInfo.RelativeName,
		RelativeIDCard:            req.MedInfo.RelativeIDCard,
		MedInsGovFee:              req.MedInfo.MedInsGovFee,
		MedInsSelfFee:             req.MedInfo.MedInsSelfFee,
		MedInsOtherFee:            req.MedInfo.MedInsOtherFee,
		MedInsCashFee:             req.MedInfo.MedInsCashFee,
		WechatPayCashFee:          req.MedInfo.WechatPayCashFee,
		ChannelNo:                 req.MedInfo.ChannelNo,
	}
	for _, a := range req.MedInfo.CashAddDetail {
		med.CashAddDetail = append(med.CashAddDetail, entity.MedCashAddItem{CashAddFee: a.CashAddFee, CashAddType: a.CashAddType})
	}
	for _, r := range req.MedInfo.CashReduceDetail {
		med.CashReduceDetail = append(med.CashReduceDetail, entity.MedCashReduceItem{CashReduceFee: r.CashReduceFee, CashReduceType: r.CashReduceType})
	}

	payment, err := c.service.MedCreateOrder(userIDInt, req.Description, med)
	if err != nil {
		c.log.Error("创建医保订单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(paymentToResponse(payment), "医保订单创建成功", ctx)
}

// MedPrepay 医保自费混合下单
// @Summary 医保混合下单
// @Description 含自费时先 JSAPI 下单再医保混合下单(同一 out_trade_no)，返回调起医保支付参数(wx.requestMedicalInsurancePay)
// @Tags 医保支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.MedPrepayRequest true "医保下单参数"
// @Success 200 {object} baseRes.Response{data=response.MedInsPayParamsResponse} "调起医保支付参数"
// @Router /api/wx/med/prepay [post]
func (c *MedicalController) MedPrepay(ctx *gin.Context) {
	var req request.MedPrepayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	userID, exists := ctx.Get("userID")
	if !exists {
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

	params, err := c.service.MedPrepay(userIDInt, openIDStr, req.OutTradeNo)
	if err != nil {
		c.log.Error("医保混合下单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(response.MedInsPayParamsResponse{
		MixTradeNo: params.MixTradeNo,
		AppID:      params.AppID,
		TimeStamp:  params.TimeStamp,
		NonceStr:   params.NonceStr,
		Package:    params.Package,
		SignType:   params.SignType,
		PaySign:    params.PaySign,
	}, "医保混合下单成功", ctx)
}

// MedQuery 医保订单查询（out_trade_no 与 mix_trade_no 二选一，同步本地状态）
// @Summary 医保订单查询
// @Description 按商户订单号或医保混合订单号查询医保订单，并同步本地订单状态
// @Tags 医保支付
// @Produce json
// @Security BearerAuth
// @Param outTradeNo query string false "商户订单号"
// @Param mixTradeNo query string false "医保混合订单号"
// @Success 200 {object} baseRes.Response{data=object} "医保订单信息"
// @Router /api/wx/med/query [get]
func (c *MedicalController) MedQuery(ctx *gin.Context) {
	outTradeNo := ctx.Query("outTradeNo")
	mixTradeNo := ctx.Query("mixTradeNo")
	if outTradeNo == "" && mixTradeNo == "" {
		baseRes.FailWithMessage("outTradeNo与mixTradeNo至少传一个", ctx)
		return
	}
	var (
		resp any
		err  error
	)
	if mixTradeNo != "" {
		resp, err = c.service.MedQueryByMixTradeNo(mixTradeNo)
	} else {
		resp, err = c.service.MedQueryByOutTradeNo(outTradeNo)
	}
	if err != nil {
		c.log.Error("查询医保订单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(resp, "查询医保订单成功", ctx)
}

// HandleMedNotify 医保收款回调
// @Summary 医保收款回调通知
// @Description 医保混合收款 MEDICAL_INSURANCE.SUCCESS 事件回调，验签+解密后同步本地订单状态；由微信服务端调用
// @Tags 医保支付
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "应答微信成功"
// @Failure 500 {object} map[string]string "处理失败"
// @Router /api/wx/med/notify [post]
func (c *MedicalController) HandleMedNotify(ctx *gin.Context) {
	body, err := readBody(ctx)
	if err != nil {
		ctx.JSON(200, gin.H{"code": "FAIL", "message": "读取报文失败"})
		return
	}
	if _, err = c.service.HandleNotify(ctx.Request.Context(), ctx.Request.Header, body); err != nil {
		c.log.Error("处理医保回调失败", "error", err)
		ctx.JSON(500, gin.H{"code": "FAIL", "message": "处理失败"})
		return
	}
	ctx.JSON(200, gin.H{"code": "SUCCESS", "message": "成功"})
}

// paymentToResponse 支付实体转响应 DTO
func paymentToResponse(p *entity.Payment) response.PaymentResponse {
	resp := response.PaymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		UserID:        p.UserID,
		Amount:        float64(p.Amount) / 100.0,
		Currency:      p.Currency,
		PaymentMethod: p.Method,
		Status:        string(p.Status),
		TransactionID: p.TransactionID,
		Description:   p.Description,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if !p.CreatedAt.IsZero() {
		resp.CreatedAt = p.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}
