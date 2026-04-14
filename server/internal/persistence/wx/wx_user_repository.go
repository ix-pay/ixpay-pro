package persistence

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// wxUserModel 微信用户数据库模型
type wxUserModel struct {
	database.SnowflakeBaseModel
	OpenID        string     `gorm:"size:100;not null;uniqueIndex"`
	UnionID       string     `gorm:"size:100;uniqueIndex"`
	Nickname      string     `gorm:"size:100"`
	Avatar        string     `gorm:"size:255"`
	Gender        int        `gorm:"default:0"`
	Country       string     `gorm:"size:50"`
	Province      string     `gorm:"size:50"`
	City          string     `gorm:"size:50"`
	Language      string     `gorm:"size:20"`
	Subscribe     bool       `gorm:"default:false"`
	SubscribeTime *time.Time `gorm:"index"`
	Remark        string     `gorm:"size:255"`
	GroupID       int64      `gorm:"default:0"`
	UserID        int64      `gorm:"index"`
}

// TableName 指定表名
func (wxUserModel) TableName() string {
	return "wx_users"
}

// toDomain 将数据库模型转换为领域实体
func (m *wxUserModel) toDomain() *entity.WXUser {
	if m == nil {
		return nil
	}
	return &entity.WXUser{
		ID:            m.ID,
		OpenID:        m.OpenID,
		UnionID:       m.UnionID,
		Nickname:      m.Nickname,
		Avatar:        m.Avatar,
		Gender:        m.Gender,
		Country:       m.Country,
		Province:      m.Province,
		City:          m.City,
		Language:      m.Language,
		Subscribe:     m.Subscribe,
		SubscribeTime: m.SubscribeTime,
		Remark:        m.Remark,
		GroupID:       m.GroupID,
		UserID:        m.UserID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainWXUser(user *entity.WXUser) (*wxUserModel, error) {
	return &wxUserModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        user.ID,
			CreatedBy: 0,
			UpdatedBy: 0,
		},
		OpenID:        user.OpenID,
		UnionID:       user.UnionID,
		Nickname:      user.Nickname,
		Avatar:        user.Avatar,
		Gender:        user.Gender,
		Country:       user.Country,
		Province:      user.Province,
		City:          user.City,
		Language:      user.Language,
		Subscribe:     user.Subscribe,
		SubscribeTime: user.SubscribeTime,
		Remark:        user.Remark,
		GroupID:       user.GroupID,
		UserID:        user.UserID,
	}, nil
}

// wxUserRepository Repository 实现
type wxUserRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.WXUserRepository = (*wxUserRepository)(nil)

// NewWXUserRepository 创建微信用户仓库实现
func NewWXUserRepository(db *database.PostgresDB) repo.WXUserRepository {
	return &wxUserRepository{db: db}
}

// GetByID 根据 ID 查询微信用户
func (r *wxUserRepository) GetByID(id int64) (*entity.WXUser, error) {
	var dbModel wxUserModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByOpenID 根据 OpenID 查询微信用户
func (r *wxUserRepository) GetByOpenID(openID string) (*entity.WXUser, error) {
	var dbModel wxUserModel
	result := r.db.Where("open_id = ?", openID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByUnionID 根据 UnionID 查询微信用户
func (r *wxUserRepository) GetByUnionID(unionID string) (*entity.WXUser, error) {
	var dbModel wxUserModel
	result := r.db.Where("union_id = ?", unionID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByUserID 根据系统用户 ID 查询微信用户
func (r *wxUserRepository) GetByUserID(userID int64) (*entity.WXUser, error) {
	var dbModel wxUserModel
	result := r.db.Where("user_id = ?", userID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建微信用户
func (r *wxUserRepository) Create(user *entity.WXUser) error {
	dbModel, err := fromDomainWXUser(user)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	user.ID = dbModel.ID
	return nil
}

// Update 更新微信用户
func (r *wxUserRepository) Update(user *entity.WXUser) error {
	dbModel, err := fromDomainWXUser(user)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除微信用户
func (r *wxUserRepository) Delete(id int64) error {
	return r.db.Delete(&wxUserModel{}, id).Error
}

// List 分页查询微信用户列表
func (r *wxUserRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.WXUser, int, error) {
	var total64 int64
	var dbModels []wxUserModel

	query := r.db.Model(&wxUserModel{})

	// 应用过滤器
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total64).Error; err != nil {
		return nil, 0, err
	}
	total := int(total64)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*entity.WXUser, len(dbModels))
	for i, model := range dbModels {
		users[i] = model.toDomain()
	}

	return users, total, nil
}
