package database

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/snowflake"
	"gorm.io/gorm"
)

var snowflakeInstance *snowflake.Snowflake

// SetSnowflakeInstance 设置全局雪花算法实例
func SetSnowflakeInstance(instance *snowflake.Snowflake) {
	snowflakeInstance = instance
}

// SnowflakeBaseModel 定义使用雪花算法 ID 的数据库表基础字段
// 包含所有实体共有的字段，使用 int64 类型的 ID
// 适用于需要使用雪花算法生成主键的表

type SnowflakeBaseModel struct {
	ID        int64          `gorm:"primaryKey;type:bigint;autoIncrement:false" json:"-"`
	CreatedBy int64          `gorm:"default:0;type:bigint" json:"-"` // 创建者 ID
	UpdatedBy int64          `gorm:"default:0;type:bigint" json:"-"` // 修改者 ID
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除时间，GORM 标准类型，不在 JSON 中暴露
}

// BeforeCreate GORM钩子函数，在创建记录前自动生成雪花算法ID
func (m *SnowflakeBaseModel) BeforeCreate(tx *gorm.DB) error {
	// 只有当ID为0时才生成新ID，允许手动指定ID
	if m.ID == 0 && snowflakeInstance != nil {
		m.ID = snowflakeInstance.Generate()
	}
	return nil
}

// SnowflakeBaseModelWithoutDeleted 定义不包含软删除字段的基础模型，使用雪花算法 ID
// 适用于不需要软删除功能但需要雪花算法 ID 的表

type SnowflakeBaseModelWithoutDeleted struct {
	ID        int64     `gorm:"primaryKey;type:bigint;autoIncrement:false" json:"id,string"`
	CreatedBy int64     `gorm:"default:0;type:bigint" json:"created_by,string"` // 创建者 ID
	UpdatedBy int64     `gorm:"default:0;type:bigint" json:"updated_by,string"` // 修改者 ID
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate GORM钩子函数，在创建记录前自动生成雪花算法ID
func (m *SnowflakeBaseModelWithoutDeleted) BeforeCreate(tx *gorm.DB) error {
	// 只有当ID为0时才生成新ID，允许手动指定ID
	if m.ID == 0 && snowflakeInstance != nil {
		m.ID = snowflakeInstance.Generate()
	}
	return nil
}
