package database

import "context"

// tenantIDKey 上下文键（租户ID路由）
type tenantIDKeyType struct{}

// WithTenantID 将租户ID注入上下文（此后该 ctx 的 GORM 查询路由到对应租户库）
func WithTenantID(ctx context.Context, tenantID int64) context.Context {
	return context.WithValue(ctx, tenantIDKeyType{}, tenantID)
}

// TenantIDFromContext 从上下文提取租户ID（不存在返回 false）
func TenantIDFromContext(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	id, ok := ctx.Value(tenantIDKeyType{}).(int64)
	if !ok || id == 0 {
		return 0, false
	}
	return id, true
}
