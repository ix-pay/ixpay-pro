package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"

// WXAuthSessionRepository 微信授权会话仓库接口
// 提供对微信授权会话数据的访问方法
type WXAuthSessionRepository interface {
	GetByID(id int64) (*entity.WXAuthSession, error)
	GetActiveSessionByWXUserID(wxUserID int64) (*entity.WXAuthSession, error)
	Create(session *entity.WXAuthSession) error
	Update(session *entity.WXAuthSession) error
	InvalidateSession(id int64) error
	InvalidateAllSessionsByWXUserID(wxUserID int64) error
}
