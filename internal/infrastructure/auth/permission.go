package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/redis"
)

// PermissionManager 权限管理器
type PermissionManager struct {
	redisClient *redis.RedisClient
	log         logger.Logger
}

// Permission 权限信息
type Permission struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Roles       []string `json:"roles"`
	WechatGrant bool     `json:"wechat_grant"` // 是否允许微信授权登录访问
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager(redisClient *redis.RedisClient, log logger.Logger) *PermissionManager {
	return &PermissionManager{
		redisClient: redisClient,
		log:         log,
	}
}

// CachePermissions 缓存权限数据
func (p *PermissionManager) CachePermissions(permissions []Permission) error {
	// 缓存所有权限信息
	permissionsData, err := json.Marshal(permissions)
	if err != nil {
		p.log.Error("Failed to marshal permissions", "error", err)
		return err
	}

	// 缓存权限列表，设置过期时间为24小时
	if err := p.redisClient.Set("permissions:list", permissionsData, 24*time.Hour); err != nil {
		p.log.Error("Failed to cache permissions list", "error", err)
		return err
	}

	// 为每个路径和方法创建索引
	for _, perm := range permissions {
		key := fmt.Sprintf("permission:%s:%s", perm.Method, perm.Path)
		permData, err := json.Marshal(perm)
		if err != nil {
			p.log.Error("Failed to marshal permission", "error", err)
			continue
		}

		if err := p.redisClient.Set(key, permData, 24*time.Hour); err != nil {
			p.log.Error("Failed to cache permission", "error", err)
			continue
		}
	}

	return nil
}

// GetPermission 获取特定路径和方法的权限信息
func (p *PermissionManager) GetPermission(method, path string) (*Permission, error) {
	key := fmt.Sprintf("permission:%s:%s", method, path)

	// 从缓存获取权限信息
	permData, err := p.redisClient.Get(key)
	if err != nil {
		p.log.Error("Failed to get permission from cache", "error", err)
		return nil, err
	}

	var permission Permission
	if err := json.Unmarshal([]byte(permData), &permission); err != nil {
		p.log.Error("Failed to unmarshal permission", "error", err)
		return nil, err
	}

	return &permission, nil
}

// CheckPermission 检查用户是否有权限访问指定路径
func (p *PermissionManager) CheckPermission(ctx context.Context, method, path string) bool {
	// 从上下文中获取用户信息
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		p.log.Warn("No claims found in context")
		return false
	}

	// 获取权限信息
	permission, err := p.GetPermission(method, path)
	if err != nil {
		p.log.Error("Failed to get permission", "error", err)
		return false
	}

	// 检查微信登录的特殊权限
	if claims.LoginType == "wechat" && permission.WechatGrant {
		return true
	}

	// 检查角色权限
	for _, role := range permission.Roles {
		if role == claims.Role {
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
		p.log.Error("Failed to get permission", "error", err)
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

// RefreshPermissions 刷新权限缓存
func (p *PermissionManager) RefreshPermissions(permissions []Permission) error {
	// 先清除所有权限缓存
	if err := p.clearPermissionCache(); err != nil {
		p.log.Error("Failed to clear permission cache", "error", err)
	}

	// 重新缓存权限
	return p.CachePermissions(permissions)
}

// clearPermissionCache 清除权限缓存
func (p *PermissionManager) clearPermissionCache() error {
	// 获取所有权限键
	// 在实际生产环境中，应该使用SCAN命令而不是KEYS命令
	// 这里为了简化，假设使用Redis的KEYS命令

	// 首先删除权限列表
	if err := p.redisClient.Del("permissions:list"); err != nil {
		return err
	}

	// 由于我们没有直接方法获取所有权限键，这里省略了删除单个权限键的逻辑
	// 在实际应用中，应该使用SCAN命令遍历并删除所有权限相关的键

	return nil
}
