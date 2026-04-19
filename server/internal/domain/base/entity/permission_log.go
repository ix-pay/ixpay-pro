package entity

import "time"

// PermissionLog 权限日志实体
type PermissionLog struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"column:user_id;index;comment:用户 ID" json:"userId"`
	Username   string    `gorm:"column:user_name;size:64;comment:用户名" json:"userName"`
	Operation  string    `gorm:"column:operation;size:128;comment:操作类型" json:"operation"`
	Module     string    `gorm:"column:module;size:64;comment:模块" json:"module"`
	TargetType string    `gorm:"column:target_type;size:64;comment:目标类型" json:"targetType"`
	TargetID   int64     `gorm:"column:target_id;comment:目标 ID" json:"targetId"`
	OldValue   string    `gorm:"column:old_value;type:text;comment:旧值" json:"oldValue"`
	NewValue   string    `gorm:"column:new_value;type:text;comment:新值" json:"newValue"`
	IP         string    `gorm:"column:ip;size:64;comment:操作 IP" json:"ip"`
	UserAgent  string    `gorm:"column:user_agent;size:512;comment:User-Agent" json:"userAgent"`
	CreatedAt  time.Time `gorm:"column:created_at;index" json:"createdAt"`
}

// TableName 指定表名
func (PermissionLog) TableName() string {
	return "permission_logs"
}
