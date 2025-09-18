package repository

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
)

// WXAuthSessionRepository 实现微信授权会话仓库接口
type WXAuthSessionRepository struct {
	db *database.PostgresDB
}

// NewWXAuthSessionRepository 创建微信授权会话仓库实例
func NewWXAuthSessionRepository(db *database.PostgresDB) model.WXAuthSessionRepository {
	return &WXAuthSessionRepository{
		db: db,
	}
}

// GetByID 根据ID获取微信授权会话
func (r *WXAuthSessionRepository) GetByID(id uint) (*model.WXAuthSession, error) {
	var session model.WXAuthSession
	result := r.db.First(&session, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}

// GetActiveSessionByWXUserID 获取指定微信用户的有效会话
func (r *WXAuthSessionRepository) GetActiveSessionByWXUserID(wxUserID uint) (*model.WXAuthSession, error) {
	var session model.WXAuthSession
	result := r.db.Where("wx_user_id = ? AND is_active = ? AND expires_at > ?", wxUserID, true, time.Now()).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}

// Create 创建微信授权会话
func (r *WXAuthSessionRepository) Create(session *model.WXAuthSession) error {
	return r.db.Create(session).Error
}

// Update 更新微信授权会话
func (r *WXAuthSessionRepository) Update(session *model.WXAuthSession) error {
	return r.db.Save(session).Error
}

// InvalidateSession 使指定会话失效
func (r *WXAuthSessionRepository) InvalidateSession(id uint) error {
	return r.db.Model(&model.WXAuthSession{}).Where("id = ?", id).Update("is_active", false).Error
}

// InvalidateAllSessionsByWXUserID 使指定微信用户的所有会话失效
func (r *WXAuthSessionRepository) InvalidateAllSessionsByWXUserID(wxUserID uint) error {
	return r.db.Model(&model.WXAuthSession{}).Where("wx_user_id = ?", wxUserID).Update("is_active", false).Error
}
