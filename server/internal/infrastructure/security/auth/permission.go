package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
)

// PermissionManager 权限管理器
type PermissionManager struct {
	cache cache.Cache
	log   logger.Logger
}

// Permission 权限信息
type Permission struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Roles       []string `json:"roles"`
	Buttons     []string `json:"buttons"`      // 所需按钮权限编码列表
	WechatGrant bool     `json:"wechat_grant"` // 是否允许微信授权登录访问
}

// NewPermissionManager 创建权限管理器
func SetupPermissionManager(cache cache.Cache, log logger.Logger) *PermissionManager {
	return &PermissionManager{
		cache: cache,
		log:   log,
	}
}

// CachePermissions 缓存权限数据
func (p *PermissionManager) CachePermissions(permissions []Permission) error {
	// 缓存所有权限信息
	permissionsData, err := json.Marshal(permissions)
	if err != nil {
		p.log.Error("权限数据序列化失败", "error", err)
		return err
	}

	// 缓存权限列表，设置过期时间为 24 小时
	if err := p.cache.Set("permissions:list", string(permissionsData), 24*time.Hour); err != nil {
		p.log.Error("缓存权限列表失败", "error", err)
		return err
	}

	// 为每个路径和方法创建索引
	for _, perm := range permissions {
		key := fmt.Sprintf("permission:%s:%s", perm.Method, perm.Path)
		permData, err := json.Marshal(perm)
		if err != nil {
			p.log.Error("权限序列化失败", "error", err)
			continue
		}

		if err := p.cache.Set(key, string(permData), 24*time.Hour); err != nil {
			p.log.Error("缓存权限失败", "error", err)
			continue
		}
	}

	return nil
}

// GetPermission 获取特定路径和方法的权限信息
func (p *PermissionManager) GetPermission(method, path string) (*Permission, error) {
	key := fmt.Sprintf("permission:%s:%s", method, path)

	// 从缓存获取权限信息
	permData, err := p.cache.Get(key)
	if err != nil {
		p.log.Warn("从缓存获取权限失败，返回默认权限", "error", err, "method", method, "path", path)

		// 返回一个默认权限，允许 user 和 admin 角色访问
		return &Permission{
			Path:    path,
			Method:  method,
			Roles:   []string{"user", "admin"},
			Buttons: []string{},
		}, nil
	}

	var permission Permission
	if err := json.Unmarshal([]byte(permData), &permission); err != nil {
		p.log.Error("权限反序列化失败", "error", err)
		return nil, err
	}

	return &permission, nil
}

// CheckPermission 检查用户是否有权限访问指定路径
func (p *PermissionManager) CheckPermission(ctx context.Context, method, path string) bool {
	// 从上下文中获取角色信息
	roleValue := ctx.Value("role")
	if roleValue == nil {
		p.log.Warn("上下文中未找到角色信息")
		return false
	}

	role, ok := roleValue.(string)
	if !ok {
		p.log.Warn("角色信息类型错误")
		return false
	}

	// 超管admin角色拥有所有权限
	if role == "admin" {
		p.log.Debug("管理员角色拥有完整访问权限", "method", method, "path", path)
		return true
	}

	// 获取权限信息
	permission, err := p.GetPermission(method, path)
	if err != nil {
		p.log.Error("获取权限失败", "error", err)
		return false
	}

	// 检查角色权限
	for _, r := range permission.Roles {
		if r == role {
			return true
		}
	}

	return false
}

// CheckRolePermission 检查用户角色是否有权限
func (p *PermissionManager) CheckRolePermission(role string, method, path string) bool {
	// 获取权限信息
	permission, err := p.GetPermission(method, path)
	if err != nil {
		p.log.Error("获取权限失败", "error", err)
		return false
	}

	// 检查角色权限
	for _, r := range permission.Roles {
		if r == role {
			return true
		}
	}

	return false
}

// CheckButtonPermission 检查用户是否有按钮权限
// buttonCode: 按钮权限编码
func (p *PermissionManager) CheckButtonPermission(ctx context.Context, buttonCode string) bool {
	// 从上下文中获取用户拥有的按钮权限
	userButtons, exists := ctx.Value("userButtons").([]string)
	if !exists {
		p.log.Warn("上下文中未找到用户按钮权限")
		return false
	}

	// 检查用户是否拥有该按钮权限
	for _, btn := range userButtons {
		if btn == buttonCode {
			return true
		}
	}

	return false
}

// CheckAPIPermissionWithButton 检查用户是否有权限访问带按钮权限的API
// 先检查基础的角色权限，如果有则进一步检查按钮权限
func (p *PermissionManager) CheckAPIPermissionWithButton(ctx context.Context, method, path string) bool {
	// 从上下文中获取角色信息
	roleValue := ctx.Value("role")
	var role string
	var roleExists bool

	if roleValue != nil {
		if r, ok := roleValue.(string); ok {
			role = r
			roleExists = true
		}
	}

	// 超管admin角色拥有所有权限，包括按钮权限
	if roleExists && role == "admin" {
		p.log.Debug("管理员角色拥有带按钮权限的完整访问权限", "method", method, "path", path)
		return true
	}

	// 获取权限信息
	permission, err := p.GetPermission(method, path)
	if err != nil {
		p.log.Error("获取权限失败", "error", err)
		return false
	}

	// 检查角色权限
	if roleExists {
		// 检查用户角色是否在权限的角色列表中
		hasRolePermission := false
		for _, r := range permission.Roles {
			if r == role {
				hasRolePermission = true
				break
			}
		}

		if !hasRolePermission {
			return false
		}
	}

	// 如果API不需要按钮权限，则直接返回true
	if len(permission.Buttons) == 0 {
		return true
	}

	// 检查用户是否拥有至少一个所需的按钮权限
	userButtons, exists := ctx.Value("userButtons").([]string)
	if !exists {
		p.log.Warn("上下文中未找到用户按钮权限")
		return false
	}

	// 检查用户是否拥有至少一个所需的按钮权限
	for _, requiredBtn := range permission.Buttons {
		for _, userBtn := range userButtons {
			if userBtn == requiredBtn {
				return true
			}
		}
	}

	return false
}

// RefreshPermissions 刷新权限缓存
func (p *PermissionManager) RefreshPermissions(permissions []Permission) error {
	// 先清除所有权限缓存
	if err := p.clearPermissionCache(); err != nil {
		p.log.Error("清除权限缓存失败", "error", err)
	}

	// 重新缓存权限
	return p.CachePermissions(permissions)
}

// clearPermissionCache 清除权限缓存
func (p *PermissionManager) clearPermissionCache() error {
	// 删除权限列表
	if err := p.cache.Delete("permissions:list"); err != nil {
		p.log.Error("删除权限列表失败", "error", err)
	}

	// 由于缓存接口不支持 SCAN 操作，我们只清除已知的权限键
	// 如果需要清除所有权限键，需要在调用此方法前收集所有需要清除的键
	// 这里简化处理，只清除权限列表，单个权限键会在下次访问时自动更新
	return nil
}
