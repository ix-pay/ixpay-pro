package persistence

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// wxAuthSessionModel 微信授权会话数据库模型
type wxAuthSessionModel struct {
	database.SnowflakeBaseModel
	WXUserID     int64     `gorm:"not null;index"`
	AccessToken  string    `gorm:"size:500;not null"`
	RefreshToken string    `gorm:"size:500;not null"`
	ExpiresIn    int64     `gorm:"not null;default:0"`
	Scope        string    `gorm:"size:100"`
	IsActive     bool      `gorm:"not null;default:true"`
	ExpiresAt    time.Time `gorm:"not null;index"`
}

// TableName 指定表名
func (wxAuthSessionModel) TableName() string {
	return "wx_auth_sessions"
}

// toDomain 将数据库模型转换为领域实体
func (m *wxAuthSessionModel) toDomain() *entity.WXAuthSession {
	if m == nil {
		return nil
	}
	return &entity.WXAuthSession{
		ID:           m.ID,
		WXUserID:     m.WXUserID,
		AccessToken:  m.AccessToken,
		RefreshToken: m.RefreshToken,
		ExpiresIn:    m.ExpiresIn,
		Scope:        m.Scope,
		IsActive:     m.IsActive,
		ExpiresAt:    m.ExpiresAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// fromDomainWXAuthSession 将领域实体转换为数据库模型
func fromDomainWXAuthSession(session *entity.WXAuthSession) *wxAuthSessionModel {
	return &wxAuthSessionModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        session.ID,
			CreatedBy: 0,
			UpdatedBy: 0,
		},
		WXUserID:     session.WXUserID,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    session.ExpiresIn,
		Scope:        session.Scope,
		IsActive:     session.IsActive,
		ExpiresAt:    session.ExpiresAt,
	}
}

// wxAuthSessionRepository Repository 实现
type wxAuthSessionRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.WXAuthSessionRepository = (*wxAuthSessionRepository)(nil)

// NewWXAuthSessionRepository 创建微信授权会话仓库实现
func NewWXAuthSessionRepository(db *database.PostgresDB) repo.WXAuthSessionRepository {
	return &wxAuthSessionRepository{db: db}
}

// GetByID 根据 ID 查询会话
func (r *wxAuthSessionRepository) GetByID(id int64) (*entity.WXAuthSession, error) {
	var dbModel wxAuthSessionModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// GetActiveSessionByWXUserID 根据微信用户 ID 查询活跃会话
func (r *wxAuthSessionRepository) GetActiveSessionByWXUserID(wxUserID int64) (*entity.WXAuthSession, error) {
	var dbModel wxAuthSessionModel
	result := r.db.Where("wx_user_id = ? AND is_active = ? AND expires_at > ?", wxUserID, true, time.Now()).
		Order("id DESC").First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// Create 创建会话
func (r *wxAuthSessionRepository) Create(session *entity.WXAuthSession) error {
	dbModel := fromDomainWXAuthSession(session)
	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}
	session.ID = dbModel.ID
	session.CreatedAt = dbModel.CreatedAt
	session.UpdatedAt = dbModel.UpdatedAt
	return nil
}

// Update 更新会话
func (r *wxAuthSessionRepository) Update(session *entity.WXAuthSession) error {
	dbModel := fromDomainWXAuthSession(session)
	return r.db.Save(dbModel).Error
}

// InvalidateSession 使会话失效
func (r *wxAuthSessionRepository) InvalidateSession(id int64) error {
	return r.db.Model(&wxAuthSessionModel{}).Where("id = ?", id).
		Update("is_active", false).Error
}

// InvalidateAllSessionsByWXUserID 使指定用户的所有会话失效
func (r *wxAuthSessionRepository) InvalidateAllSessionsByWXUserID(wxUserID int64) error {
	return r.db.Model(&wxAuthSessionModel{}).Where("wx_user_id = ?", wxUserID).
		Update("is_active", false).Error
}