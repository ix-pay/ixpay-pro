package wxapi

import (
	"io"
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	payService "github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

	"github.com/gin-gonic/gin"
)

// RefundController 退款控制器（对齐 Gpay handler/refund.go：普通退款 + 医保退款）
// @Summary 退款相关 API
// @Description 退款申请/查询/异常退款/退款回调/医保退款
// @Tags 退款管理
type RefundController struct {
	refundService *payService.RefundService
	medRefundSvc  *payService.MedRefundService
	log           logger.Logger
}

// NewRefundController 创建退款控制器
func NewRefundController(refundService *payService.RefundService, medRefundSvc *payService.MedRefundService, log logger.Logger) *RefundController {
	return &RefundController{refundService: refundService, medRefundSvc: medRefundSvc, log: log}
}

// Create 发起退款
// @Summary 退款申请
// @Description 对已支付订单发起退款（支持多次部分退款，累计退款不超订单金额；金额单位：分）
// @Tags 退款管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.RefundCreateRequest true "退款申请参数"
// @Success 200 {object} baseRes.Response{data=response.RefundResponse} "退款单(受理结果，最终以回调/查询为准)"
// @Router /api/wx/refund/create [post]
func (c *RefundController) Create(ctx *gin.Context) {
	var req request.RefundCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	record, err := c.refundService.Create(req.OutTradeNo, req.RefundFee, req.Reason)
	if err != nil {
		c.log.Error("退款申请失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(refundToResp(record), "退款申请已受理", ctx)
}

// Query 查询退款
// @Summary 查询单笔退款
// @Description 按商户退款单号查询微信侧退款状态并同步本地退款单与订单状态
// @Tags 退款管理
// @Produce json
// @Security BearerAuth
// @Param outRefundNo query string true "商户退款单号"
// @Success 200 {object} baseRes.Response{data=response.RefundResponse} "退款单"
// @Router /api/wx/refund/query [get]
func (c *RefundController) Query(ctx *gin.Context) {
	outRefundNo := ctx.Query("outRefundNo")
	if outRefundNo == "" {
		baseRes.FailWithMessage("outRefundNo不能为空", ctx)
		return
	}
	record, err := c.refundService.Query(outRefundNo)
	if err != nil {
		c.log.Error("查询退款失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(refundToResp(record), "查询退款成功", ctx)
}

// List 退款单列表
// @Summary 退款单列表
// @Description 分页查询本地退款单(按创建时间倒序)
// @Tags 退款管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码(默认1)"
// @Param pageSize query int false "每页条数(默认10)"
// @Success 200 {object} baseRes.Response{data=baseRes.PageResult} "退款单列表"
// @Router /api/wx/refund/list [get]
func (c *RefundController) List(ctx *gin.Context) {
	page, pageSize := parsePage(ctx)
	records, total, err := c.refundService.List(page, pageSize)
	if err != nil {
		c.log.Error("查询退款单列表失败", "error", err)
		baseRes.FailWithMessage("查询退款单列表失败", ctx)
		return
	}
	list := make([]response.RefundResponse, 0, len(records))
	for _, r := range records {
		list = append(list, refundToResp(r))
	}
	baseRes.OkWithDetailed(baseRes.PageResult{List: list, Total: int64(total), Page: page, PageSize: pageSize}, "查询退款单列表成功", ctx)
}

// ApplyAbnormal 发起异常退款
// @Summary 发起异常退款
// @Description 退款单状态为 ABNORMAL 时发起；退用户银行卡时卡号/姓名由服务端 OAEP 加密后出网
// @Tags 退款管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.RefundAbnormalRequest true "异常退款参数"
// @Success 200 {object} baseRes.Response{data=response.RefundResponse} "退款单(置为处理中)"
// @Router /api/wx/refund/abnormal [post]
func (c *RefundController) ApplyAbnormal(ctx *gin.Context) {
	var req request.RefundAbnormalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	record, err := c.refundService.ApplyAbnormal(req.OutRefundNo, req.Type, req.BankType, req.BankAccount, req.RealName)
	if err != nil {
		c.log.Error("发起异常退款失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(refundToResp(record), "异常退款已受理", ctx)
}

// HandleRefundNotify 退款结果回调
// @Summary 退款结果回调通知
// @Description REFUND.SUCCESS/ABNORMAL/CLOSED 事件回调，验签+解密后同步退款单与订单状态；由微信服务端调用
// @Tags 退款管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "应答微信成功"
// @Failure 500 {object} map[string]string "处理失败"
// @Router /api/wx/refund/notify [post]
func (c *RefundController) HandleRefundNotify(ctx *gin.Context) {
	body, err := readBody(ctx)
	if err != nil {
		ctx.JSON(200, gin.H{"code": "FAIL", "message": "读取报文失败"})
		return
	}
	if _, err = c.refundService.HandleNotify(ctx.Request.Context(), ctx.Request.Header, body); err != nil {
		c.log.Error("处理退款回调失败", "error", err)
		ctx.JSON(500, gin.H{"code": "FAIL", "message": "处理失败"})
		return
	}
	ctx.JSON(200, gin.H{"code": "SUCCESS", "message": "成功"})
}

// MedCreate 医保退款
// @Summary 医保退款
// @Description 医保部分(统筹/个账/其他)走 V3 医保退款通知，现金部分走普通退款接口；按退款类型自动拆分金额
// @Tags 医保支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.MedRefundCreateRequest true "医保退款参数(金额单位：分)"
// @Success 200 {object} baseRes.Response{data=response.MedRefundResponse} "医保退款单(受理结果)"
// @Router /api/wx/refund/med [post]
func (c *RefundController) MedCreate(ctx *gin.Context) {
	var req request.MedRefundCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	record, err := c.medRefundSvc.Create(&payService.MedRefundInput{
		OutTradeNo:      req.OutTradeNo,
		HospOutRefundNo: req.HospOutRefundNo,
		PartRefundType:  req.PartRefundType,
		CashRefundFee:   req.CashRefundFee,
		InsRefundFee:    req.InsRefundFee,
		RefReason:       req.RefReason,
	})
	if err != nil {
		c.log.Error("医保退款失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(medRefundToResp(record), "医保退款已受理", ctx)
}

// MedQuery 医保退款查询
// @Summary 医保退款查询
// @Description 现金部分以 V3 退款查询为准；医保部分以医保订单查询(mix_pay_status=MIX_PAY_REFUND)为准，同步退款单与订单状态
// @Tags 医保支付
// @Produce json
// @Security BearerAuth
// @Param hospOutRefundNo query string false "医院退款订单号"
// @Param outTradeNo query string false "本地订单号(二选一)"
// @Success 200 {object} baseRes.Response{data=response.MedRefundQueryResponse} "退款状态"
// @Router /api/wx/refund/med/query [get]
func (c *RefundController) MedQuery(ctx *gin.Context) {
	hospOutRefundNo := ctx.Query("hospOutRefundNo")
	outTradeNo := ctx.Query("outTradeNo")
	result, err := c.medRefundSvc.Query(hospOutRefundNo, outTradeNo)
	if err != nil {
		c.log.Error("查询医保退款失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(response.MedRefundQueryResponse{
		HospOutRefundNo:       result.HospOutRefundNo,
		HospOutTradeNo:        result.HospOutTradeNo,
		MedRefundID:           result.MedRefundID,
		CashRefundID:          result.CashRefundID,
		InsuranceRefundID:     result.InsuranceRefundID,
		CashRefundFee:         result.CashRefundFee,
		InsuranceRefundFee:    result.InsuranceRefundFee,
		CashRefundStatus:      result.CashRefundStatus,
		InsuranceRefundStatus: result.InsuranceRefundStatus,
		MedRefundState:        result.MedRefundState,
		RefundEndTime:         result.RefundEndTime,
		ResponseContent:       result.ResponseContent,
	}, "查询医保退款成功", ctx)
}

// MedList 医保退款单列表
// @Summary 医保退款单列表
// @Description 分页查询本地医保退款单(按创建时间倒序)
// @Tags 医保支付
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码(默认1)"
// @Param pageSize query int false "每页条数(默认10)"
// @Success 200 {object} baseRes.Response{data=baseRes.PageResult} "医保退款单列表"
// @Router /api/wx/refund/med/list [get]
func (c *RefundController) MedList(ctx *gin.Context) {
	page, pageSize := parsePage(ctx)
	records, total, err := c.medRefundSvc.List(page, pageSize)
	if err != nil {
		c.log.Error("查询医保退款单列表失败", "error", err)
		baseRes.FailWithMessage("查询医保退款单列表失败", ctx)
		return
	}
	list := make([]response.MedRefundResponse, 0, len(records))
	for _, r := range records {
		list = append(list, medRefundToResp(r))
	}
	baseRes.OkWithDetailed(baseRes.PageResult{List: list, Total: int64(total), Page: page, PageSize: pageSize}, "查询医保退款单列表成功", ctx)
}

// ---------- 内部转换 ----------

func refundToResp(r *entity.RefundRecord) response.RefundResponse {
	resp := response.RefundResponse{
		OutRefundNo:         r.OutRefundNo,
		OutTradeNo:          r.OutTradeNo,
		RefundID:            r.RefundID,
		Reason:              r.Reason,
		RefundFee:           r.RefundFee,
		TotalFee:            r.TotalFee,
		Status:              r.Status,
		UserReceivedAccount: r.UserReceivedAccount,
		SuccessTime:         r.SuccessTime,
	}
	if !r.CreatedAt.IsZero() {
		resp.CreatedAt = r.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}

func medRefundToResp(r *entity.MedRefundRecord) response.MedRefundResponse {
	resp := response.MedRefundResponse{
		HospOutRefundNo:       r.HospOutRefundNo,
		OutTradeNo:            r.OutTradeNo,
		PartRefundType:        r.PartRefundType,
		CashRefundFee:         r.CashRefundFee,
		InsRefundFee:          r.InsRefundFee,
		MedRefundState:        r.MedRefundState,
		CashRefundStatus:      r.CashRefundStatus,
		InsuranceRefundStatus: r.InsuranceRefundStatus,
		Reason:                r.Reason,
	}
	if !r.CreatedAt.IsZero() {
		resp.CreatedAt = r.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}

// readBody 读取请求体
func readBody(ctx *gin.Context) ([]byte, error) {
	return io.ReadAll(ctx.Request.Body)
}

// parsePage 解析分页参数
func parsePage(ctx *gin.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}
