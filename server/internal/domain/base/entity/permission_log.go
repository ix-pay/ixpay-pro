package entity

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// PermissionLog 权限日志实体
// 对应数据库表 sys_permission_logs
// 表结构：id, operator_id, operator_name, action_type, target_type, target_id,
//
//	before_data, after_data, ip_address, user_agent, created_at
type PermissionLog struct {
	database.SnowflakeBaseModelWithoutDeleted
	OperatorID   int64  `gorm:"column:operator_id;not null" json:"operatorId"`
	OperatorName string `gorm:"column:operator_name;size:100" json:"operatorName"`
	ActionType   string `gorm:"column:action_type;size:50;not null" json:"actionType"`
	TargetType   string `gorm:"column:target_type;size:50" json:"targetType"`
	TargetID     int64  `gorm:"column:target_id" json:"targetId"`
	BeforeData   string `gorm:"column:before_data;type:jsonb" json:"beforeData"`
	AfterData    string `gorm:"column:after_data;type:jsonb" json:"afterData"`
	IPAddress    string `gorm:"column:ip_address;size:50" json:"ipAddress"`
	UserAgent    string `gorm:"column:user_agent;size:500" json:"userAgent"`
}

// TableName 指定表名
func (PermissionLog) TableName() string {
	return "sys_permission_logs"
}