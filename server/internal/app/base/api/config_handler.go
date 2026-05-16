package baseapi

import (
	"fmt"
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

// ConfigController 配置控制器
//
//	@Summary		系统配置相关API
//	@Description	提供系统配置的增删改查功能
//	@Tags			系统配置管理
//	@Router			/api/admin/config [get]
type ConfigController struct {
	service *service.ConfigService
	log     logger.Logger
}

// NewConfigController 创建配置控制器
func NewConfigController(service *service.ConfigService, log logger.Logger) *ConfigController {
	return &ConfigController{
		service: service,
		log:     log,
	}
}

// convertToConfigResponse 将 entity.Config 转换为 response.ConfigResponse
func convertToConfigResponse(config *entity.Config) response.ConfigResponse {
	return response.ConfigResponse{
		ID:          config.ID,
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		ConfigType:  fmt.Sprintf("%d", config.ConfigType),
		Description: config.Description,
		Status:      config.Status,
		CreatedAt:   config.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   config.UpdatedAt.Format(time.RFC3339),
	}
}

// GetConfigByKey 根据配置键获取配置
//
//	@Summary		根据配置键获取配置
//	@Description	根据配置键获取配置详情
//	@Tags			系统配置管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_key	query		string														true	"配置键"
//	@Success		200			{object}	baseRes.Response{data=response.ConfigResponse,msg=string}	"配置详情"
//	@Failure		400			{object}	map[string]string											"请求参数错误"
//	@Failure		401			{object}	map[string]string											"未授权"
//	@Failure		500			{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin/config/key [get]
func (c *ConfigController) GetConfigByKey(ctx *gin.Context) {
	configKey := ctx.Query("config_key")
	if configKey == "" {
		c.log.Error("请求参数错误", "config_key", configKey)
		baseRes.FailWithMessage("配置键不能为空", ctx)
		return
	}

	config, err := c.service.GetConfigByKey(configKey)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	configResponse := convertToConfigResponse(config)

	baseRes.OkWithDetailed(configResponse, "获取配置成功", ctx)
}

// GetConfigByID 根据ID获取配置
//
//	@Summary		根据ID获取配置
//	@Description	根据ID获取配置详情
//	@Tags			系统配置管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int															true	"配置ID"
//	@Success		200	{object}	baseRes.Response{data=response.ConfigResponse,msg=string}	"配置详情"
//	@Failure		400	{object}	map[string]string											"请求参数错误"
//	@Failure		401	{object}	map[string]string											"未授权"
//	@Failure		500	{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin/config/:id [get]
func (c *ConfigController) GetConfigByID(ctx *gin.Context) {
	idStr := ctx.Param("id")

	// 将字符串 ID 转换为 int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", idStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	config, err := c.service.GetConfigByID(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	configResponse := convertToConfigResponse(config)

	baseRes.OkWithDetailed(configResponse, "获取配置成功", ctx)
}

// CreateConfig 创建配置
//
//	@Summary		创建配置
//	@Description	创建新的系统配置
//	@Tags			系统配置管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config	body		request.CreateConfigRequest									true	"创建配置请求参数"
//	@Success		201		{object}	baseRes.Response{data=response.ConfigResponse,msg=string}	"创建成功"
//	@Failure		400		{object}	map[string]string											"请求参数错误"
//	@Failure		401		{object}	map[string]string											"未授权"
//	@Failure		500		{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin/config [post]
func (c *ConfigController) CreateConfig(ctx *gin.Context) {
	var req request.CreateConfigRequest
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

	// 将 ConfigType 从 string 转换为 int
	configType, err := strconv.Atoi(req.ConfigType)
	if err != nil {
		c.log.Error("配置类型格式错误", "error", err)
		baseRes.FailWithMessage("配置类型格式错误", ctx)
		return
	}

	config, err := c.service.CreateConfig(
		req.ConfigKey,
		req.ConfigValue,
		configType,
		req.Description,
		req.Status,
		createdByInt,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	configResponse := convertToConfigResponse(config)

	baseRes.OkWithDetailed(configResponse, "创建配置成功", ctx)
}

// UpdateConfig 更新配置
//
//	@Summary		更新配置
//	@Description	更新系统配置
//	@Tags			系统配置管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config	body		request.UpdateConfigRequest		true	"更新配置请求参数"
//	@Success		200		{object}	baseRes.Response{msg=string}	"更新成功"
//	@Failure		400		{object}	map[string]string				"请求参数错误"
//	@Failure		401		{object}	map[string]string				"未授权"
//	@Failure		500		{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/config [put]
func (c *ConfigController) UpdateConfig(ctx *gin.Context) {
	var req request.UpdateConfigRequest
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

	// 将 ConfigType 从 string 转换为 int
	configType, err := strconv.Atoi(req.ConfigType)
	if err != nil {
		c.log.Error("配置类型格式错误", "error", err)
		baseRes.FailWithMessage("配置类型格式错误", ctx)
		return
	}

	err = c.service.UpdateConfig(
		req.ID,
		req.ConfigKey,
		req.ConfigValue,
		configType,
		req.Description,
		req.Status,
		updatedByInt,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("更新配置成功", ctx)
}

// DeleteConfig 删除配置
//
//	@Summary		删除配置
//	@Description	删除系统配置
//	@Tags			系统配置管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int								true	"配置ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/config/:id [delete]
func (c *ConfigController) DeleteConfig(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		c.log.Error("配置 ID 不能为空")
		baseRes.FailWithMessage("配置 ID 不能为空", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", idStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	err = c.service.DeleteConfig(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除配置成功", ctx)
}

// GetConfigList 获取配置列表
//
//	@Summary		获取配置列表
//	@Description	获取系统配置列表
//	@Tags			系统配置管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int																false	"页码"
//	@Param			pageSize	query		int																false	"每页条数"
//	@Param			config_key	query		string															false	"配置键"
//	@Param			status		query		int																false	"状态：1-启用 0-禁用"
//	@Success		200			{object}	baseRes.Response{data=response.ConfigListResponse,msg=string}	"配置列表"
//	@Failure		401			{object}	map[string]string												"未授权"
//	@Failure		500			{object}	map[string]string												"服务器内部错误"
//	@Router			/api/admin/config [get]
func (c *ConfigController) GetConfigList(ctx *gin.Context) {
	// 获取分页参数
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 构建过滤条件
	filters := make(map[string]interface{})
	configKey := ctx.Query("config_key")
	if configKey != "" {
		filters["config_key LIKE ?"] = "%" + configKey + "%"
	}
	statusStr := ctx.Query("status")
	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		filters["status"] = status
	}

	configs, total, err := c.service.GetConfigList(page, pageSize, filters)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	configResponses := make([]response.ConfigResponse, 0, len(configs))
	for _, config := range configs {
		configResponses = append(configResponses, convertToConfigResponse(config))
	}

	configListResponse := response.ConfigListResponse{
		PageResult: baseRes.PageResult{
			List:     configResponses,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		List: configResponses,
	}

	baseRes.OkWithDetailed(configListResponse, "获取配置列表成功", ctx)
}

// GetAllActiveConfigs 获取所有启用的配置
//
//	@Summary		获取所有启用的配置
//	@Description	获取所有启用的系统配置
//	@Tags			系统配置管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	baseRes.Response{data=[]response.ConfigResponse,msg=string}	"配置列表"
//	@Failure		500	{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin/config/active [get]
func (c *ConfigController) GetAllActiveConfigs(ctx *gin.Context) {
	configs, err := c.service.GetAllActiveConfigs()
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应
	configResponses := make([]response.ConfigResponse, 0, len(configs))
	for _, config := range configs {
		configResponses = append(configResponses, convertToConfigResponse(config))
	}

	baseRes.OkWithDetailed(configResponses, "获取启用的配置列表成功", ctx)
}
