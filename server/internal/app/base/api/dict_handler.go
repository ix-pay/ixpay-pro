package baseapi

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// DictController 字典控制器
//
//	@Summary		字典管理相关API
//	@Description	提供字典和字典明细项的增删改查功能
//	@Tags			字典管理
//	@Router			/api/admin/dict [get]
type DictController struct {
	dictService     *service.DictService
	dictItemService *service.DictItemService
	log             logger.Logger
}

// NewDictController 创建字典控制器
func NewDictController(dictService *service.DictService, dictItemService *service.DictItemService, log logger.Logger) *DictController {
	return &DictController{
		dictService:     dictService,
		dictItemService: dictItemService,
		log:             log,
	}
}

// convertToDictItemResponse 将 entity.DictItem 转换为 response.DictItemResponse
func convertToDictItemResponse(item *entity.DictItem) response.DictItemResponse {
	return response.DictItemResponse{
		ID:          item.ID,
		DictID:      item.DictID,
		ItemKey:     item.ItemKey,
		ItemValue:   item.ItemValue,
		Sort:        item.Sort,
		Description: item.Description,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
}

// convertToDictResponse 将 entity.Dict 转换为 response.DictResponse
func convertToDictResponse(dict *entity.Dict) response.DictResponse {
	dictItems := make([]response.DictItemResponse, 0, len(dict.DictItems))
	for _, item := range dict.DictItems {
		dictItems = append(dictItems, convertToDictItemResponse(&item))
	}

	return response.DictResponse{
		ID:          dict.ID,
		DictName:    dict.DictName,
		DictCode:    dict.DictCode,
		Description: dict.Description,
		Status:      dict.Status,
		CreatedAt:   dict.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dict.UpdatedAt.Format(time.RFC3339),
		DictItems:   dictItems,
	}
}

// 字典表相关接口

// GetDictByID 根据ID获取字典
//
//	@Summary		根据ID获取字典
//	@Description	根据ID获取字典详情
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int														true	"字典ID"
//	@Success		200	{object}	baseRes.Response{data=response.DictResponse,msg=string}	"字典详情"
//	@Failure		400	{object}	map[string]string										"请求参数错误"
//	@Failure		401	{object}	map[string]string										"未授权"
//	@Failure		500	{object}	map[string]string										"服务器内部错误"
//	@Router			/api/admin/dict/:id [get]
func (c *DictController) GetDictByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		c.log.Error("字典 ID 不能为空")
		baseRes.FailWithMessage("字典 ID 不能为空", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", idStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	dict, err := c.dictService.GetDictByID(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	dictResponse := convertToDictResponse(dict)

	baseRes.OkWithDetailed(dictResponse, "获取字典成功", ctx)
}

// GetDictByCode 根据编码获取字典
//
//	@Summary		根据编码获取字典
//	@Description	根据编码获取字典详情
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dict_code	query		string													true	"字典编码"
//	@Success		200			{object}	baseRes.Response{data=response.DictResponse,msg=string}	"字典详情"
//	@Failure		400			{object}	map[string]string										"请求参数错误"
//	@Failure		401			{object}	map[string]string										"未授权"
//	@Failure		500			{object}	map[string]string										"服务器内部错误"
//	@Router			/api/admin/dict/code [get]
func (c *DictController) GetDictByCode(ctx *gin.Context) {
	dictCode := ctx.Query("dict_code")
	if dictCode == "" {
		c.log.Error("请求参数错误", "dict_code", dictCode)
		baseRes.FailWithMessage("字典编码不能为空", ctx)
		return
	}

	dict, err := c.dictService.GetDictByCode(dictCode)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 获取字典项
	dictItems, err := c.dictItemService.GetDictItemsByDictID(dict.ID)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	var dictItemResponses []response.DictItemResponse
	for _, item := range dictItems {
		dictItemResponses = append(dictItemResponses, convertToDictItemResponse(item))
	}

	dictResponse := response.DictResponse{
		ID:          dict.ID,
		DictCode:    dict.DictCode,
		DictName:    dict.DictName,
		Description: dict.Description,
		Status:      dict.Status,
		CreatedAt:   dict.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dict.UpdatedAt.Format(time.RFC3339),
		DictItems:   dictItemResponses,
	}

	baseRes.OkWithDetailed(dictResponse, "获取字典成功", ctx)
}

// CreateDict 创建字典
//
//	@Summary		创建字典
//	@Description	创建新的字典
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dict	body		request.CreateDictRequest								true	"创建字典请求参数"
//	@Success		201		{object}	baseRes.Response{data=response.DictResponse,msg=string}	"创建成功"
//	@Failure		400		{object}	map[string]string										"请求参数错误"
//	@Failure		401		{object}	map[string]string										"未授权"
//	@Failure		500		{object}	map[string]string										"服务器内部错误"
//	@Router			/api/admin/dict [post]
func (c *DictController) CreateDict(ctx *gin.Context) {
	var req request.CreateDictRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户 ID
	createdBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	// 将 createdBy 转换为 int64
	var createdByInt int64
	var err error
	switch v := createdBy.(type) {
	case string:
		createdByInt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.log.Error("用户 ID 格式错误", "error", err)
			baseRes.FailWithMessage("用户 ID 格式错误", ctx)
			return
		}
	case int64:
		createdByInt = v
	case int:
		createdByInt = int64(v)
	default:
		c.log.Error("用户 ID 类型错误", "actual_type", v)
		baseRes.FailWithMessage("用户 ID 类型错误", ctx)
		return
	}

	dict, err := c.dictService.CreateDict(
		req.DictCode,
		req.DictName,
		req.Description,
		req.Status,
		createdByInt,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	dictResponse := response.DictResponse{
		ID:          dict.ID,
		DictCode:    dict.DictCode,
		DictName:    dict.DictName,
		Description: dict.Description,
		Status:      dict.Status,
		CreatedAt:   dict.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dict.UpdatedAt.Format(time.RFC3339),
		DictItems:   []response.DictItemResponse{},
	}

	baseRes.OkWithDetailed(dictResponse, "创建字典成功", ctx)
}

// UpdateDict 更新字典
//
//	@Summary		更新字典
//	@Description	更新字典信息
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dict	body		request.UpdateDictRequest		true	"更新字典请求参数"
//	@Success		200		{object}	baseRes.Response{msg=string}	"更新成功"
//	@Failure		400		{object}	map[string]string				"请求参数错误"
//	@Failure		401		{object}	map[string]string				"未授权"
//	@Failure		500		{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/dict [put]
func (c *DictController) UpdateDict(ctx *gin.Context) {
	var req request.UpdateDictRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户 ID
	updatedBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	// 将 updatedBy 转换为 int64
	var updatedByInt int64
	var err error
	switch v := updatedBy.(type) {
	case string:
		updatedByInt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.log.Error("用户 ID 格式错误", "error", err)
			baseRes.FailWithMessage("用户 ID 格式错误", ctx)
			return
		}
	case int64:
		updatedByInt = v
	case int:
		updatedByInt = int64(v)
	default:
		c.log.Error("用户 ID 类型错误", "actual_type", v)
		baseRes.FailWithMessage("用户 ID 类型错误", ctx)
		return
	}

	err = c.dictService.UpdateDict(
		req.ID,
		req.DictCode,
		req.DictName,
		req.Description,
		req.Status,
		updatedByInt,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("更新字典成功", ctx)
}

// DeleteDict 删除字典
//
//	@Summary		删除字典
//	@Description	删除字典（会同时删除字典下的所有字典项）
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int								true	"字典ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/dict/:id [delete]
func (c *DictController) DeleteDict(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		c.log.Error("字典 ID 不能为空")
		baseRes.FailWithMessage("字典 ID 不能为空", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", idStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	err = c.dictService.DeleteDict(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除字典成功", ctx)
}

// GetDictList 获取字典列表
//
//	@Summary		获取字典列表
//	@Description	获取字典列表
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int															false	"页码"
//	@Param			pageSize	query		int															false	"每页条数"
//	@Param			dict_code	query		string														false	"字典编码"
//	@Param			dict_name	query		string														false	"字典名称"
//	@Param			status		query		int															false	"状态：1-启用 0-禁用"
//	@Success		200			{object}	baseRes.Response{data=response.DictListResponse,msg=string}	"字典列表"
//	@Failure		401			{object}	map[string]string											"未授权"
//	@Failure		500			{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin/dict/list [get]
func (c *DictController) GetDictList(ctx *gin.Context) {
	// 获取分页参数
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 构建过滤条件
	filters := make(map[string]interface{})
	dictCode := ctx.Query("dict_code")
	if dictCode != "" {
		filters["dict_code LIKE ?"] = "%" + dictCode + "%"
	}
	dictName := ctx.Query("dict_name")
	if dictName != "" {
		filters["dict_name LIKE ?"] = "%" + dictName + "%"
	}
	statusStr := ctx.Query("status")
	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		filters["status"] = status
	}

	dicts, total, err := c.dictService.GetDictList(int64(page), int64(pageSize), filters)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	var dictResponses []response.DictResponse
	for _, dict := range dicts {
		dictResponses = append(dictResponses, response.DictResponse{
			ID:          dict.ID,
			DictCode:    dict.DictCode,
			DictName:    dict.DictName,
			Description: dict.Description,
			Status:      dict.Status,
			CreatedAt:   dict.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   dict.UpdatedAt.Format(time.RFC3339),
			DictItems:   []response.DictItemResponse{},
		})
	}

	dictListResponse := response.DictListResponse{
		PageResult: baseRes.PageResult{
			List:     dictResponses,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		List: dictResponses,
	}

	baseRes.OkWithDetailed(dictListResponse, "获取字典列表成功", ctx)
}

// 字典项相关接口

// GetDictItemByID 根据ID获取字典项
//
//	@Summary		根据ID获取字典项
//	@Description	根据ID获取字典项详情
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int															true	"字典项ID"
//	@Success		200	{object}	baseRes.Response{data=response.DictItemResponse,msg=string}	"字典项详情"
//	@Failure		400	{object}	map[string]string											"请求参数错误"
//	@Failure		401	{object}	map[string]string											"未授权"
//	@Failure		500	{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin/dict/item/:id [get]
func (c *DictController) GetDictItemByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		c.log.Error("字典项 ID 不能为空")
		baseRes.FailWithMessage("字典项 ID 不能为空", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", idStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	dictItem, err := c.dictItemService.GetDictItemByID(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	dictItemResponse := convertToDictItemResponse(dictItem)

	baseRes.OkWithDetailed(dictItemResponse, "获取字典项成功", ctx)
}

// GetDictItemsByDictID 根据字典ID获取字典项列表
//
//	@Summary		根据字典ID获取字典项列表
//	@Description	根据字典ID获取该字典下的所有字典项
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dict_id	query		int																true	"字典ID"
//	@Success		200		{object}	baseRes.Response{data=response.DictItemListResponse,msg=string}	"字典项列表"
//	@Failure		400		{object}	map[string]string												"请求参数错误"
//	@Failure		401		{object}	map[string]string												"未授权"
//	@Failure		500		{object}	map[string]string												"服务器内部错误"
//	@Router			/api/admin/dict/items [get]
func (c *DictController) GetDictItemsByDictID(ctx *gin.Context) {
	dictIDStr := ctx.Query("dict_id")
	if dictIDStr == "" {
		c.log.Error("字典 ID 不能为空")
		baseRes.FailWithMessage("字典 ID 不能为空", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	dictID, err := strconv.ParseInt(dictIDStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "dict_id", dictIDStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	dictItems, err := c.dictItemService.GetDictItemsByDictID(dictID)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	var dictItemResponses []response.DictItemResponse
	for _, item := range dictItems {
		dictItemResponses = append(dictItemResponses, convertToDictItemResponse(item))
	}

	dictItemListResponse := response.DictItemListResponse{
		PageResult: baseRes.PageResult{
			List:     dictItemResponses,
			Total:    int64(len(dictItemResponses)),
			Page:     len(dictItemResponses),
			PageSize: len(dictItemResponses),
		},
		List: dictItemResponses,
	}

	baseRes.OkWithDetailed(dictItemListResponse, "获取字典项列表成功", ctx)
}

// CreateDictItem 创建字典项
//
//	@Summary		创建字典项
//	@Description	创建新的字典项
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dict_item	body		request.CreateDictItemRequest								true	"创建字典项请求参数"
//	@Success		201			{object}	baseRes.Response{data=response.DictItemResponse,msg=string}	"创建成功"
//	@Failure		400			{object}	map[string]string											"请求参数错误"
//	@Failure		401			{object}	map[string]string											"未授权"
//	@Failure		500			{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin/dict/item [post]
func (c *DictController) CreateDictItem(ctx *gin.Context) {
	var req request.CreateDictItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户 ID
	createdBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	// 将 createdBy 转换为 int64
	var createdByInt int64
	var err error
	switch v := createdBy.(type) {
	case string:
		createdByInt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.log.Error("用户 ID 格式错误", "error", err)
			baseRes.FailWithMessage("用户 ID 格式错误", ctx)
			return
		}
	case int64:
		createdByInt = v
	case int:
		createdByInt = int64(v)
	default:
		c.log.Error("用户 ID 类型错误", "actual_type", v)
		baseRes.FailWithMessage("用户 ID 类型错误", ctx)
		return
	}

	dictItem, err := c.dictItemService.CreateDictItem(
		req.DictID,
		req.ItemKey,
		req.ItemValue,
		req.Description,
		req.Sort,
		req.Status,
		createdByInt,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	dictItemResponse := convertToDictItemResponse(dictItem)

	baseRes.OkWithDetailed(dictItemResponse, "创建字典项成功", ctx)
}

// UpdateDictItem 更新字典项
//
//	@Summary		更新字典项
//	@Description	更新字典项信息
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dict_item	body		request.UpdateDictItemRequest	true	"更新字典项请求参数"
//	@Success		200			{object}	baseRes.Response{msg=string}	"更新成功"
//	@Failure		400			{object}	map[string]string				"请求参数错误"
//	@Failure		401			{object}	map[string]string				"未授权"
//	@Failure		500			{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/dict/item [put]
func (c *DictController) UpdateDictItem(ctx *gin.Context) {
	var req request.UpdateDictItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户 ID
	updatedBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未授权")
		baseRes.NoAuth("未授权", ctx)
		return
	}

	// 将 updatedBy 转换为 int64
	var updatedByInt int64
	var err error
	switch v := updatedBy.(type) {
	case string:
		updatedByInt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.log.Error("用户 ID 格式错误", "error", err)
			baseRes.FailWithMessage("用户 ID 格式错误", ctx)
			return
		}
	case int64:
		updatedByInt = v
	case int:
		updatedByInt = int64(v)
	default:
		c.log.Error("用户 ID 类型错误", "actual_type", v)
		baseRes.FailWithMessage("用户 ID 类型错误", ctx)
		return
	}

	err = c.dictItemService.UpdateDictItem(
		req.ID,
		req.DictID,
		req.ItemKey,
		req.ItemValue,
		req.Description,
		req.Sort,
		req.Status,
		updatedByInt,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("更新字典项成功", ctx)
}

// DeleteDictItem 删除字典项
//
//	@Summary		删除字典项
//	@Description	删除字典项
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int								true	"字典项ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/dict/item/:id [delete]
func (c *DictController) DeleteDictItem(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		c.log.Error("字典项 ID 不能为空")
		baseRes.FailWithMessage("字典项 ID 不能为空", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", idStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	err = c.dictItemService.DeleteDictItem(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除字典项成功", ctx)
}
