package baseapi

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// EventController 事件管理控制器
//
//	@Summary		事件管理 API
//	@Description	提供事件发布、查询、重试、死信队列管理等功能（管理员权限）
//	@Tags			事件管理
//	@Router			/api/admin/event [get]
type EventController struct {
	eventBus *eventbus.EventBus
	log      logger.Logger
}

// NewEventController 创建事件控制器
func NewEventController(eventBus *eventbus.EventBus, log logger.Logger) *EventController {
	return &EventController{
		eventBus: eventBus,
		log:      log,
	}
}

// CreateEvent 创建事件
//
//	@Summary		创建事件
//	@Description	发布一个新的事件（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			event	body		request.CreateEventRequest		true	"创建事件请求参数"
//	@Success		201		{object}	map[string]response.EventResponse	"事件创建成功"
//	@Failure		400		{object}	map[string]string					"请求参数错误"
//	@Failure		401		{object}	map[string]string					"未授权"
//	@Failure		403		{object}	map[string]string					"无权限"
//	@Failure		500		{object}	map[string]string					"服务器内部错误"
//	@Router			/api/admin/event [post]
func (c *EventController) CreateEvent(ctx *gin.Context) {
	// 检查权限
	if !c.checkAdminRole(ctx) {
		return
	}

	var req request.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误："+err.Error(), ctx)
		return
	}

	// 构建事件选项
	opts := []eventbus.EventOption{
		eventbus.WithPriority(req.Priority),
	}
	if req.MaxRetries > 0 {
		opts = append(opts, eventbus.WithMaxRetries(req.MaxRetries))
	}
	if req.Delay > 0 {
		opts = append(opts, eventbus.WithDelaySeconds(req.Delay))
	}
	if req.Metadata != "" {
		opts = append(opts, eventbus.WithMetadata(req.Metadata))
	}

	// 创建并发布事件
	event := eventbus.NewEvent(req.Name, req.Payload, opts...)

	if req.Delay > 0 {
		// 异步发布延迟事件
		if err := c.eventBus.PublishAsync(ctx, event); err != nil {
			c.log.Error("发布延迟事件失败", "error", err)
			baseRes.FailWithMessage("发布事件失败："+err.Error(), ctx)
			return
		}
	} else {
		// 同步发布事件
		if err := c.eventBus.Publish(ctx, event); err != nil {
			c.log.Error("发布事件失败", "error", err)
			baseRes.FailWithMessage("发布事件失败："+err.Error(), ctx)
			return
		}
	}

	// 构建响应
	resp := c.convertEventToResponse(event)
	baseRes.OkWithDetailed(resp, "事件发布成功", ctx)
}

// GetEvent 获取事件详情
//
//	@Summary		获取事件详情
//	@Description	根据 ID 获取事件详情（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"事件 ID"
//	@Success		200	{object}	map[string]response.EventResponse	"事件详情"
//	@Failure		401	{object}	map[string]string					"未授权"
//	@Failure		403	{object}	map[string]string					"无权限"
//	@Failure		404	{object}	map[string]string					"事件不存在"
//	@Router			/api/admin/event/:id [get]
func (c *EventController) GetEvent(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	event, err := c.eventBus.GetEvent(ctx, id)
	if err != nil {
		baseRes.FailWithMessage("事件不存在", ctx)
		return
	}

	resp := c.convertEventToResponse(event)
	baseRes.OkWithDetailed(resp, "获取事件详情成功", ctx)
}

// ListEvents 获取事件列表
//
//	@Summary		获取事件列表
//	@Description	获取事件列表（支持分页和筛选）（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name		query		string								false	"事件名称"
//	@Param			status		query		int									false	"事件状态"
//	@Param			priority	query		int									false	"优先级"
//	@Param			startDate	query		string								false	"开始日期"
//	@Param			endDate		query		string								false	"结束日期"
//	@Param			page		query		int									false	"页码"
//	@Param			pageSize	query		int									false	"每页数量"
//	@Success		200	{object}	map[string]response.EventListResponse	"事件列表"
//	@Failure		401	{object}	map[string]string							"未授权"
//	@Failure		403	{object}	map[string]string							"无权限"
//	@Router			/api/admin/event [get]
func (c *EventController) ListEvents(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	var req request.ListEventsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误："+err.Error(), ctx)
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filter := eventbus.EventFilter{
		Name:      req.Name,
		Page:      req.Page,
		PageSize:  req.PageSize,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	if req.Status != nil {
		status := eventbus.EventStatus(*req.Status)
		filter.Status = &status
	}
	if req.Priority != nil {
		filter.Priority = req.Priority
	}

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			filter.StartTime = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			filter.EndTime = &t
		}
	}

	events, total, err := c.eventBus.ListEvents(ctx, filter)
	if err != nil {
		baseRes.FailWithMessage("查询事件列表失败", ctx)
		return
	}

	respList := make([]response.EventResponse, 0, len(events))
	for _, event := range events {
		respList = append(respList, c.convertEventToResponse(event))
	}

	resp := response.EventListResponse{
		List:  respList,
		Total: total,
	}

	baseRes.OkWithDetailed(resp, "获取事件列表成功", ctx)
}

// RetryEvent 重试事件
//
//	@Summary		重试事件
//	@Description	重试一个失败的事件（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"事件 ID"
//	@Success		200	{object}	map[string]string	"事件重试成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		404	{object}	map[string]string	"事件不存在"
//	@Router			/api/admin/event/:id/retry [post]
func (c *EventController) RetryEvent(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	event, err := c.eventBus.GetEvent(ctx, id)
	if err != nil {
		baseRes.FailWithMessage("事件不存在", ctx)
		return
	}

	// 重置事件状态
	event.Status = eventbus.EventStatusPending
	event.RetryCount = 0
	event.ErrorMessage = ""
	event.NextRetryAt = nil

	if err := c.eventBus.Publish(ctx, event); err != nil {
		c.log.Error("重试事件失败", "error", err)
		baseRes.FailWithMessage("重试事件失败："+err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("事件重试成功", ctx)
}

// ListDeadLetters 获取死信列表
//
//	@Summary		获取死信列表
//	@Description	获取死信队列中的事件列表（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int										false	"页码"
//	@Param			pageSize	query		int										false	"每页数量"
//	@Success		200	{object}	map[string]response.DeadLetterListResponse	"死信列表"
//	@Failure		401	{object}	map[string]string								"未授权"
//	@Failure		403	{object}	map[string]string								"无权限"
//	@Router			/api/admin/event/dead-letters [get]
func (c *EventController) ListDeadLetters(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := eventbus.EventFilter{
		Page:     page,
		PageSize: pageSize,
	}

	events, total, err := c.eventBus.ListDeadLetters(ctx, filter)
	if err != nil {
		baseRes.FailWithMessage("查询死信列表失败", ctx)
		return
	}

	respList := make([]response.DeadLetterResponse, 0, len(events))
	for _, event := range events {
		respList = append(respList, response.DeadLetterResponse{
			ID:         event.ID,
			EventName:  event.Name,
			Reason:     event.ErrorMessage,
			RetryCount: event.RetryCount,
			MaxRetries: event.MaxRetries,
			CreatedAt:  event.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	resp := response.DeadLetterListResponse{
		List:  respList,
		Total: total,
	}

	baseRes.OkWithDetailed(resp, "获取死信列表成功", ctx)
}

// RetryDeadLetter 重试死信
//
//	@Summary		重试死信
//	@Description	重试死信队列中的事件（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"死信 ID"
//	@Success		200	{object}	map[string]string	"死信重试成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		404	{object}	map[string]string	"死信不存在"
//	@Router			/api/admin/event/dead-letters/:id/retry [post]
func (c *EventController) RetryDeadLetter(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	if err := c.eventBus.RetryDeadLetter(ctx, id); err != nil {
		c.log.Error("重试死信失败", "error", err)
		baseRes.FailWithMessage("重试死信失败："+err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("死信重试成功", ctx)
}

// CreateSubscriber 创建订阅者
//
//	@Summary		创建订阅者
//	@Description	创建一个事件订阅者（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			subscriber	body		request.CreateSubscriberRequest			true	"创建订阅者请求参数"
//	@Success		201		{object}	map[string]response.SubscriberResponse	"订阅者创建成功"
//	@Failure		400		{object}	map[string]string							"请求参数错误"
//	@Failure		401		{object}	map[string]string							"未授权"
//	@Failure		403		{object}	map[string]string							"无权限"
//	@Failure		500		{object}	map[string]string							"服务器内部错误"
//	@Router			/api/admin/subscriber [post]
func (c *EventController) CreateSubscriber(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	var req request.CreateSubscriberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误："+err.Error(), ctx)
		return
	}

	subscriber := &eventbus.Subscriber{
		Name:       req.Name,
		EventName:  req.EventName,
		Endpoint:   req.Endpoint,
		IsActive:   true,
		MaxRetries: req.MaxRetries,
		Timeout:    req.Timeout,
		Priority:   req.Priority,
	}

	if err := c.eventBus.Subscribe(ctx, subscriber); err != nil {
		c.log.Error("创建订阅者失败", "error", err)
		baseRes.FailWithMessage("创建订阅者失败："+err.Error(), ctx)
		return
	}

	resp := c.convertSubscriberToResponse(subscriber)
	baseRes.OkWithDetailed(resp, "订阅者创建成功", ctx)
}

// ListSubscribers 获取订阅者列表
//
//	@Summary		获取订阅者列表
//	@Description	获取订阅者列表（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			eventName	query		string										false	"事件名称"
//	@Param			isActive	query		bool										false	"是否活跃"
//	@Param			page		query		int											false	"页码"
//	@Param			pageSize	query		int											false	"每页数量"
//	@Success		200	{object}	map[string]response.SubscriberListResponse		"订阅者列表"
//	@Failure		401	{object}	map[string]string									"未授权"
//	@Failure		403	{object}	map[string]string									"无权限"
//	@Router			/api/admin/subscriber [get]
func (c *EventController) ListSubscribers(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	var req request.ListSubscribersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误："+err.Error(), ctx)
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filter := eventbus.SubscriberFilter{
		EventName: req.EventName,
		IsActive:  req.IsActive,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}

	subscribers, total, err := c.eventBus.GetEventRepository().ListSubscribers(ctx, filter)
	if err != nil {
		baseRes.FailWithMessage("查询订阅者列表失败", ctx)
		return
	}

	respList := make([]response.SubscriberResponse, 0, len(subscribers))
	for _, sub := range subscribers {
		respList = append(respList, c.convertSubscriberToResponse(sub))
	}

	resp := response.SubscriberListResponse{
		List:  respList,
		Total: total,
	}

	baseRes.OkWithDetailed(resp, "获取订阅者列表成功", ctx)
}

// DeleteSubscriber 删除订阅者
//
//	@Summary		删除订阅者
//	@Description	删除一个订阅者（管理员权限）
//	@Tags			事件管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"订阅者 ID"
//	@Success		200	{object}	map[string]string	"订阅者删除成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		404	{object}	map[string]string	"订阅者不存在"
//	@Router			/api/admin/subscriber/:id [delete]
func (c *EventController) DeleteSubscriber(ctx *gin.Context) {
	if !c.checkAdminRole(ctx) {
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	if err := c.eventBus.Unsubscribe(ctx, id); err != nil {
		c.log.Error("删除订阅者失败", "error", err)
		baseRes.FailWithMessage("删除订阅者失败："+err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("订阅者删除成功", ctx)
}

// checkAdminRole 检查管理员角色
func (c *EventController) checkAdminRole(ctx *gin.Context) bool {
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		baseRes.FailWithMessage("权限不足", ctx)
		return false
	}
	return true
}

// convertEventToResponse 转换事件为响应格式
func (c *EventController) convertEventToResponse(event *eventbus.Event) response.EventResponse {
	resp := response.EventResponse{
		ID:              event.ID,
		Name:            event.Name,
		Payload:         event.Payload,
		Status:          int(event.Status),
		StatusLabel:     event.Status.String(),
		Priority:        event.Priority,
		MaxRetries:      event.MaxRetries,
		RetryCount:      event.RetryCount,
		DelaySeconds:    event.DelaySeconds,
		ErrorMessage:    event.ErrorMessage,
		SubscriberCount: event.SubscriberCount,
		Metadata:        event.Metadata,
		CreatedAt:       event.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       event.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if event.ScheduledAt != nil {
		t := event.ScheduledAt.Format("2006-01-02 15:04:05")
		resp.ScheduledAt = &t
	}
	if event.ProcessedAt != nil {
		t := event.ProcessedAt.Format("2006-01-02 15:04:05")
		resp.ProcessedAt = &t
	}
	if event.NextRetryAt != nil {
		t := event.NextRetryAt.Format("2006-01-02 15:04:05")
		resp.NextRetryAt = &t
	}

	return resp
}

// convertSubscriberToResponse 转换订阅者为响应格式
func (c *EventController) convertSubscriberToResponse(sub *eventbus.Subscriber) response.SubscriberResponse {
	resp := response.SubscriberResponse{
		ID:           sub.ID,
		Name:         sub.Name,
		EventName:    sub.EventName,
		Endpoint:     sub.Endpoint,
		IsActive:     sub.IsActive,
		MaxRetries:   sub.MaxRetries,
		Timeout:      sub.Timeout,
		Priority:     sub.Priority,
		FailureCount: sub.FailureCount,
		CreatedAt:    sub.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    sub.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if sub.LastDeliveredAt != nil {
		t := sub.LastDeliveredAt.Format("2006-01-02 15:04:05")
		resp.LastDeliveredAt = &t
	}

	return resp
}

// GetEventStatusText 获取事件状态文本
func GetEventStatusText(status eventbus.EventStatus) string {
	return status.String()
}
