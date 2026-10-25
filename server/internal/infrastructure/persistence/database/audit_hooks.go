// Package database 审计字段保护钩子
package database

import (
	"reflect"

	"gorm.io/gorm"
)

// RegisterAuditPreserveHooks 注册全局审计字段保护回调（在 gorm.Open 后调用一次）
// 背景：仓储层 Update 统一为 Save(fromDomain(entity)) 全字段更新，而 fromDomain 通常
// 不回填 CreatedAt，零值时间会把库中 created_at 覆盖为 '0001-01-01 00:00:00'
// （base_tenants 曾因此丢失创建时间）。本回调在 UPDATE 前检查模型 created_at：
// 为零值则自动跳过该列（等效 Omit），非零值正常回写数据库原值，无需逐仓储修改。
func RegisterAuditPreserveHooks(db *gorm.DB) {
	if db == nil {
		return
	}
	_ = db.Callback().Update().Before("gorm:before_update").
		Register("ixpay:preserve_created_at", preserveCreatedAt)
}

// preserveCreatedAt 更新回调：模型 created_at 为零值时将其从更新列中排除
func preserveCreatedAt(tx *gorm.DB) {
	stmt := tx.Statement
	if stmt == nil || stmt.Schema == nil || stmt.Dest == nil {
		// 单列 Update / Map 更新等无完整 Schema 场景，直接跳过
		return
	}
	field := stmt.Schema.LookUpField("created_at")
	if field == nil {
		// 模型无 created_at 字段（如纯关联表）
		return
	}
	// 解析目标值（单条 struct 更新；批量 slice 场景项目内不存在，跳过处理）
	rv := stmt.ReflectValue
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	if _, isZero := field.ValueOf(stmt.Context, rv); isZero {
		stmt.Omit("created_at")
	}
}
