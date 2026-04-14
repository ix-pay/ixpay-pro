package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// userSettingModel 用户设置数据库模型
type userSettingModel struct {
	database.SnowflakeBaseModel
	UserID           int64  `gorm:"uniqueIndex;not null"`
	ThemeColor       string `gorm:"size:20"`
	SidebarColor     string `gorm:"size:20"`
	NavbarColor      string `gorm:"size:20"`
	FontSize         int    `gorm:"default:14"`
	Language         string `gorm:"size:20;default:zh-CN"`
	AutoLogin        bool   `gorm:"default:false"`
	RememberPassword bool   `gorm:"default:false"`
}

// TableName 指定表名
func (userSettingModel) TableName() string {
	return "base_user_settings"
}

// toDomain 将数据库模型转换为领域实体
func (m *userSettingModel) toDomain() *entity.UserSetting {
	if m == nil {
		return nil
	}
	return &entity.UserSetting{
		ID:               m.ID,
		UserID:           m.UserID,
		ThemeColor:       m.ThemeColor,
		SidebarColor:     m.SidebarColor,
		NavbarColor:      m.NavbarColor,
		FontSize:         m.FontSize,
		Language:         m.Language,
		AutoLogin:        m.AutoLogin,
		RememberPassword: m.RememberPassword,
		CreatedBy:        m.CreatedBy,
		CreatedAt:        m.CreatedAt,
		UpdatedBy:        m.UpdatedBy,
		UpdatedAt:        m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainUserSetting(setting *entity.UserSetting) (*userSettingModel, error) {
	return &userSettingModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        setting.ID,
			CreatedBy: setting.CreatedBy,
			UpdatedBy: setting.UpdatedBy,
		},
		UserID:           setting.UserID,
		ThemeColor:       setting.ThemeColor,
		SidebarColor:     setting.SidebarColor,
		NavbarColor:      setting.NavbarColor,
		FontSize:         setting.FontSize,
		Language:         setting.Language,
		AutoLogin:        setting.AutoLogin,
		RememberPassword: setting.RememberPassword,
	}, nil
}

// userSettingRepository Repository 实现
type userSettingRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.UserSettingRepository = (*userSettingRepository)(nil)

// NewUserSettingRepository 创建用户设置仓库实现
func NewUserSettingRepository(db *database.PostgresDB) repo.UserSettingRepository {
	return &userSettingRepository{db: db}
}

// GetByUserID 根据用户 ID 查询用户设置
func (r *userSettingRepository) GetByUserID(userID int64) (*entity.UserSetting, error) {
	var dbModel userSettingModel
	result := r.db.Where("user_id = ?", userID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建用户设置
func (r *userSettingRepository) Create(setting *entity.UserSetting) error {
	dbModel, err := fromDomainUserSetting(setting)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	setting.ID = dbModel.ID
	return nil
}

// Update 更新用户设置
func (r *userSettingRepository) Update(setting *entity.UserSetting) error {
	dbModel, err := fromDomainUserSetting(setting)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除用户设置
func (r *userSettingRepository) Delete(userID int64) error {
	return r.db.Where("user_id = ?", userID).Delete(&userSettingModel{}).Error
}
