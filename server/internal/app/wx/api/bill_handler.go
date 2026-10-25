package wxapi

import (
	"net/http"

	payService "github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
	"github.com/skip2/go-qrcode"

	"github.com/gin-gonic/gin"
)

// BillController 账单控制器（对齐 Gpay handler/bill.go：交易账单/资金账单/下载 + Native二维码）
// @Summary 账单相关 API
// @Description 申请交易账单/资金账单/下载账单/Native二维码
// @Tags 账单管理
type BillController struct {
	service    *payService.BillService
	paymentSvc *payService.PaymentService
	log        logger.Logger
}

// NewBillController 创建账单控制器
func NewBillController(service *payService.BillService, paymentSvc *payService.PaymentService, log logger.Logger) *BillController {
	return &BillController{service: service, paymentSvc: paymentSvc, log: log}
}

// TradeBill 申请交易账单
// @Summary 申请交易账单
// @Description 返回账单下载链接(5分钟有效)与SHA1摘要；bill_date 为三个月内且非当天的日期
// @Tags 账单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.TradeBillRequest true "交易账单参数"
// @Success 200 {object} baseRes.Response{data=response.BillResponse} "下载链接与摘要"
// @Router /api/wx/bill/trade [post]
func (c *BillController) TradeBill(ctx *gin.Context) {
	var req request.TradeBillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	resp, err := c.service.TradeBill(req.BillDate, req.BillType)
	if err != nil {
		c.log.Error("申请交易账单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(response.BillResponse{
		HashType:    resp.HashType,
		HashValue:   resp.HashValue,
		DownloadURL: resp.DownloadURL,
	}, "申请交易账单成功", ctx)
}

// FundFlowBill 申请资金账单
// @Summary 申请资金账单
// @Description 返回账单下载链接(5分钟有效)与SHA1摘要；account_type 默认 BASIC
// @Tags 账单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.FundFlowBillRequest true "资金账单参数"
// @Success 200 {object} baseRes.Response{data=response.BillResponse} "下载链接与摘要"
// @Router /api/wx/bill/fundflow [post]
func (c *BillController) FundFlowBill(ctx *gin.Context) {
	var req request.FundFlowBillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	resp, err := c.service.FundFlowBill(req.BillDate, req.AccountType)
	if err != nil {
		c.log.Error("申请资金账单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(response.BillResponse{
		HashType:    resp.HashType,
		HashValue:   resp.HashValue,
		DownloadURL: resp.DownloadURL,
	}, "申请资金账单成功", ctx)
}

// DownloadBill 下载账单
// @Summary 下载账单
// @Description 按申请账单返回的下载链接拉取账单内容(gzip自动解压+SHA1校验)，以附件形式返回
// @Tags 账单管理
// @Accept json
// @Produce text/csv
// @Security BearerAuth
// @Param body body request.DownloadBillRequest true "下载参数"
// @Success 200 {file} file "账单文件"
// @Router /api/wx/bill/download [post]
func (c *BillController) DownloadBill(ctx *gin.Context) {
	var req request.DownloadBillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	data, err := c.service.DownloadBill(req.DownloadURL, req.HashValue)
	if err != nil {
		c.log.Error("下载账单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	ctx.Header("Content-Disposition", "attachment; filename=wxpay_bill.csv")
	ctx.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// NativeQRCode Native 支付二维码 PNG
// @Summary Native 支付二维码
// @Description 将已下单订单的 code_url 生成为二维码 PNG 图片，供扫码支付页展示
// @Tags 微信支付
// @Produce png
// @Security BearerAuth
// @Param outTradeNo query string true "商户订单号"
// @Success 200 {file} file "二维码PNG"
// @Failure 400 {object} map[string]string "订单不存在或未下单"
// @Router /api/wx/payment/native/qrcode [get]
func (c *BillController) NativeQRCode(ctx *gin.Context) {
	outTradeNo := ctx.Query("outTradeNo")
	if outTradeNo == "" {
		baseRes.FailWithMessage("outTradeNo不能为空", ctx)
		return
	}
	codeURL, err := c.paymentSvc.NativeCodeURL(outTradeNo)
	if err != nil {
		c.log.Error("获取Native二维码链接失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	png, err := qrcode.Encode(codeURL, qrcode.Medium, 320)
	if err != nil {
		c.log.Error("生成二维码失败", "error", err)
		baseRes.FailWithMessage("生成二维码失败", ctx)
		return
	}
	ctx.Data(http.StatusOK, "image/png", png)
}
