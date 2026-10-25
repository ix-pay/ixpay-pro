package baseapi

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// TenantController 租户管理控制器（多租户平台管理，见设计文档 01 §4.1/§4.2）
// 响应遵循项目数据传输规范：JSON 标签 camelCase、雪花 ID 经 json:",string" 序列化为字符串
type TenantController struct {
	service *service.TenantService
	log     logger.Logger
}

// NewTenantController 创建租户管理控制器
func NewTenantController(service *service.TenantService, log logger.Logger) *TenantController {
	return &TenantController{
		service: service,
		log:     log,
	}
}

// timeLayout 列表时间展示格式
const timeLayout = "2006-01-02 15:04:05"

// tenantToResp 租户实体 → 响应 DTO（手动映射；不回显数据库密码密文）
func tenantToResp(t *entity.Tenant) response.TenantResponse {
	return response.TenantResponse{
		ID:          t.ID,
		TenantCode:  t.TenantCode,
		TenantName:  t.TenantName,
		DBName:      t.DBName,
		DBHost:      t.DBHost,
		DBPort:      t.DBPort,
		DBUser:      t.DBUser,
		HasDBPass:   t.DBPassword != "",
		Status:      t.Status,
		InitError:   t.InitError,
		Description: t.Description,
		CreatedAt:   t.CreatedAt.Format(timeLayout),
		UpdatedAt:   t.UpdatedAt.Format(timeLayout),
	}
}

// isAdmin 当前用户是否 admin 角色（租户管理写操作仅 admin；权限中间件对 admin 已直通，
// 此处为防御纵深：即使角色授权误配也不放行普通用户）
func isAdmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	return role == "admin"
}

// operatorID 当前操作人（管理平台登录用户）
func operatorID(c *gin.Context) int64 {
	uid, _ := c.Get("userID")
	if s, ok := uid.(string); ok {
		id, _ := strconv.ParseInt(s, 10, 64)
		return id
	}
	return 0
}

// Granted 当前用户可见租户（租户切换器数据源：admin 全量 / 普通用户按授权，见设计文档 01 §3.4）
func (t *TenantController) Granted(c *gin.Context) {
	tenants, err := t.service.ListVisibleTenants(operatorID(c), isAdmin(c))
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	list := make([]response.GrantedTenantResponse, 0, len(tenants))
	for _, item := range tenants {
		list = append(list, response.GrantedTenantResponse{
			ID:         item.ID,
			TenantCode: item.TenantCode,
			TenantName: item.TenantName,
			DBName:     item.DBName,
			Status:     item.Status,
		})
	}
	baseRes.OkWithDetailed(list, "获取可见租户成功", c)
}

// List 租户列表（分页 + 状态过滤）
func (t *TenantController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.Query("status")

	tenants, total, err := t.service.ListTenants(page, pageSize, status)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	list := make([]response.TenantResponse, 0, len(tenants))
	for _, item := range tenants {
		list = append(list, tenantToResp(item))
	}
	baseRes.OkWithDetailed(gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize}, "获取租户列表成功", c)
}

// Get 租户详情
func (t *TenantController) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	tenant, err := t.service.GetTenant(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithDetailed(tenantToResp(tenant), "获取租户成功", c)
}

// Create 创建租户（异步初始化：建库+模块迁移）
func (t *TenantController) Create(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	var req request.SaveTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), c)
		return
	}
	tenant, err := t.service.CreateTenant(req.TenantCode, req.TenantName, req.DBHost, req.DBPort, req.DBUser, req.DBPassword, req.Description, operatorID(c))
	if err != nil {
		t.log.Error("创建租户失败", "error", err)
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithDetailed(tenantToResp(tenant), "租户已创建，初始化任务已投递", c)
}

// Update 更新租户信息
func (t *TenantController) Update(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	var req request.SaveTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), c)
		return
	}
	tenant, err := t.service.UpdateTenant(id, req.TenantName, req.DBHost, req.DBPort, req.DBUser, req.DBPassword, req.Description, operatorID(c))
	if err != nil {
		t.log.Error("更新租户失败", "error", err)
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithDetailed(tenantToResp(tenant), "更新租户成功", c)
}

// Retry 重试/重跑租户库初始化（failed 重试；active 重跑即增量迁移）
func (t *TenantController) Retry(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	if err := t.service.RetryInit(id, operatorID(c)); err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithMessage("初始化任务已重新投递", c)
}

// Disable 禁用租户
func (t *TenantController) Disable(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	if err := t.service.DisableTenant(id, operatorID(c)); err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithMessage("租户已禁用", c)
}

// Enable 启用租户
func (t *TenantController) Enable(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	if err := t.service.EnableTenant(id, operatorID(c)); err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithMessage("租户已启用", c)
}

// Archive 归档租户
func (t *TenantController) Archive(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	if err := t.service.ArchiveTenant(id, operatorID(c)); err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithMessage("租户已归档", c)
}

// ---------- 应用路由 ----------

// ListRoutes 应用路由列表
func (t *TenantController) ListRoutes(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	routes, err := t.service.ListAppRoutes(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	list := make([]response.TenantRouteResponse, 0, len(routes))
	for _, r := range routes {
		list = append(list, response.TenantRouteResponse{
			ID:         r.ID,
			AppKey:     r.AppKey,
			TenantID:   r.TenantID,
			ModuleCode: r.ModuleCode,
			CreatedAt:  r.CreatedAt.Format(timeLayout),
		})
	}
	baseRes.OkWithDetailed(list, "获取应用路由成功", c)
}

// CreateRoute 创建应用路由
func (t *TenantController) CreateRoute(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	var req request.CreateTenantRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), c)
		return
	}
	route, err := t.service.CreateAppRoute(req.AppKey, id, req.ModuleCode, operatorID(c))
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithDetailed(response.TenantRouteResponse{
		ID:         route.ID,
		AppKey:     route.AppKey,
		TenantID:   route.TenantID,
		ModuleCode: route.ModuleCode,
		CreatedAt:  route.CreatedAt.Format(timeLayout),
	}, "绑定应用路由成功", c)
}

// DeleteRoute 删除应用路由
func (t *TenantController) DeleteRoute(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	routeID, err := strconv.ParseInt(c.Param("route_id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的路由ID", c)
		return
	}
	if err := t.service.DeleteAppRoute(routeID, operatorID(c)); err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithMessage("解绑应用路由成功", c)
}

// ---------- 用户授权 ----------

// ListGrants 授权用户列表
func (t *TenantController) ListGrants(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	grants, err := t.service.ListGrants(id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	list := make([]response.TenantGrantResponse, 0, len(grants))
	for _, g := range grants {
		list = append(list, response.TenantGrantResponse{
			ID:        g.ID,
			UserID:    g.UserID,
			TenantID:  g.TenantID,
			CreatedAt: g.CreatedAt.Format(timeLayout),
		})
	}
	baseRes.OkWithDetailed(list, "获取授权列表成功", c)
}

// CreateGrant 授权用户
func (t *TenantController) CreateGrant(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	var req request.CreateTenantGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), c)
		return
	}
	grant, err := t.service.CreateGrant(req.UserID, id, operatorID(c))
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithDetailed(response.TenantGrantResponse{
		ID:        grant.ID,
		UserID:    grant.UserID,
		TenantID:  grant.TenantID,
		CreatedAt: grant.CreatedAt.Format(timeLayout),
	}, "授权成功", c)
}

// DeleteGrant 取消授权
func (t *TenantController) DeleteGrant(c *gin.Context) {
	if !isAdmin(c) {
		baseRes.NoAuth("仅管理员可执行租户管理操作", c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的用户ID", c)
		return
	}
	if err := t.service.DeleteGrant(userID, id, operatorID(c)); err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	baseRes.OkWithMessage("取消授权成功", c)
}

// ---------- 运维审计 ----------

// ListOpLogs 租户操作审计日志
func (t *TenantController) ListOpLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的租户ID", c)
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	logs, err := t.service.ListOpLogs(id, limit)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), c)
		return
	}
	list := make([]response.TenantOpLogResponse, 0, len(logs))
	for _, l := range logs {
		list = append(list, response.TenantOpLogResponse{
			ID:         l.ID,
			TenantID:   l.TenantID,
			Action:     l.Action,
			Detail:     l.Detail,
			OperatorID: l.OperatorID,
			CreatedAt:  l.CreatedAt.Format(timeLayout),
		})
	}
	baseRes.OkWithDetailed(list, "获取操作日志成功", c)
}
