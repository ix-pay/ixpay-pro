package repository

import (
	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
)

// WXUserRepository 实现微信用户仓库接口
type WXUserRepository struct {
	db *database.PostgresDB
}

// NewWXUserRepository 创建微信用户仓库实例
func NewWXUserRepository(db *database.PostgresDB) model.WXUserRepository {
	return &WXUserRepository{
		db: db,
	}
}

// GetByID 根据ID获取微信用户信息
func (r *WXUserRepository) GetByID(id uint) (*model.WXUser, error) {
	var user model.WXUser
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetByOpenID 根据OpenID获取微信用户信息
func (r *WXUserRepository) GetByOpenID(openID string) (*model.WXUser, error) {
	var user model.WXUser
	result := r.db.Where("open_id = ?", openID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetByUnionID 根据UnionID获取微信用户信息
func (r *WXUserRepository) GetByUnionID(unionID string) (*model.WXUser, error) {
	var user model.WXUser
	result := r.db.Where("union_id = ?", unionID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetByUserID 根据系统用户ID获取微信用户信息
func (r *WXUserRepository) GetByUserID(userID uint) (*model.WXUser, error) {
	var user model.WXUser
	result := r.db.Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create 创建微信用户
func (r *WXUserRepository) Create(user *model.WXUser) error {
	return r.db.Create(user).Error
}

// Update 更新微信用户信息
func (r *WXUserRepository) Update(user *model.WXUser) error {
	return r.db.Save(user).Error
}

// Delete 删除微信用户
func (r *WXUserRepository) Delete(id uint) error {
	return r.db.Delete(&model.WXUser{}, id).Error
}

// List 获取微信用户列表
func (r *WXUserRepository) List(page, pageSize int) ([]*model.WXUser, int64, error) {
	var users []*model.WXUser
	var total int64

	query := r.db.Model(&model.WXUser{})

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
