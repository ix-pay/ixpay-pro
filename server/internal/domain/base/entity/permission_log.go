package entity

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// PermissionLog 权限日志实体
type PermissionLog struct {
	database.SnowflakeBaseModelWithoutDeleted
	UserID     int64  `json:"userId"`
	Username   string `json:"userName"`
	Operation  string `json:"operation"`
	Module     string `json:"module"`
	TargetType string `json:"targetType"`
	TargetID   int64  `json:"targetId"`
	OldValue   string `json:"oldValue"`
	NewValue   string `json:"newValue"`
	IP         string `json:"ip"`
	UserAgent  string `json:"userAgent"`
}

// TableName 指定表名
func (PermissionLog) TableName() string {
	return "sys_permission_logs"
}
