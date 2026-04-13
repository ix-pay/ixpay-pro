package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// userModel 数据库模型（带 GORM 标签）
type userModel struct {
	database.SnowflakeBaseModel
	Username      string `gorm:"size:50;not null;unique"`
	PasswordHash  string `gorm:"size:100;not null"`
	Nickname      string `gorm:"size:50"`
	Email         string `gorm:"size:100"`
	Phone         string `gorm:"size:20"`
	Avatar        string `gorm:"size:255"`
	Status        int    `gorm:"default:1"`
	Gender        int    `gorm:"default:0"`
	Birthday      string `gorm:"size:20"`
	Address       string `gorm:"size:255"`
	PositionID    int64  `gorm:"index"`
	DepartmentID  int64  `gorm:"index"`
	EntryDate     string `gorm:"size:20"`
	LastLoginIP   string `gorm:"size:50"`
	LastLoginTime string `gorm:"size:50"`
	WechatOpenID  string `gorm:"size:100;uniqueIndex;default:null"`

	// GORM 关联标签 - 多对一
	Department *departmentModel `gorm:"foreignKey:department_id;references:id"`
	Position   *positionModel   `gorm:"foreignKey:position_id;references:id"`

	// GORM 关联标签 - 多对多
	Roles []*roleModel `gorm:"many2many:base_role_users;joinForeignKey:user_id;joinReferences:role_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName 指定表名
func (userModel) TableName() string {
	return "base_users"
}

// toDomain 将数据库模型转换为领域实体
func (m *userModel) toDomain() *entity.User {
	user := &entity.User{
		ID:            common.ToString(m.ID),
		Username:      m.Username,
		PasswordHash:  m.PasswordHash,
		Nickname:      m.Nickname,
		Email:         m.Email,
		Phone:         m.Phone,
		Avatar:        m.Avatar,
		Status:        m.Status,
		Gender:        m.Gender,
		Birthday:      m.Birthday,
		Address:       m.Address,
		PositionID:    common.ToString(m.PositionID),
		DepartmentID:  common.ToString(m.DepartmentID),
		EntryDate:     m.EntryDate,
		LastLoginIP:   m.LastLoginIP,
		LastLoginTime: m.LastLoginTime,
		WechatOpenID:  m.WechatOpenID,
		CreatedBy:     common.ToString(m.CreatedBy),
		CreatedAt:     m.CreatedAt,
		UpdatedBy:     common.ToString(m.UpdatedBy),
		UpdatedAt:     m.UpdatedAt,
	}

	// 处理关联数据 - 部门
	if m.Department != nil {
		user.Department = m.Department.toDomain()
	}

	// 处理关联数据 - 岗位
	if m.Position != nil {
		user.Position = m.Position.toDomain()
	}

	// 处理关联数据 - 角色
	if len(m.Roles) > 0 {
		// 将 roleModel 转换为 *entity.Role，同时填充角色 ID 列表
		roles := make([]*entity.Role, len(m.Roles))
		roleIDs := make([]string, len(m.Roles))
		for i, role := range m.Roles {
			roles[i] = role.toDomain()
			roleIDs[i] = common.ToString(role.ID)
		}
		user.Roles = roles
		user.RoleIds = roleIDs
	}

	return user
}

// fromDomain 将领域实体转换为数据库模型
func fromDomain(user *entity.User) (*userModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(user.ID, user.CreatedBy, user.UpdatedBy)

	return &userModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		Username:      user.Username,
		PasswordHash:  user.PasswordHash,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Phone:         user.Phone,
		Avatar:        user.Avatar,
		Status:        user.Status,
		Gender:        user.Gender,
		Birthday:      user.Birthday,
		Address:       user.Address,
		PositionID:    common.TryParseInt64(user.PositionID),
		DepartmentID:  common.TryParseInt64(user.DepartmentID),
		EntryDate:     user.EntryDate,
		LastLoginIP:   user.LastLoginIP,
		LastLoginTime: user.LastLoginTime,
		WechatOpenID:  user.WechatOpenID,
	}, nil
}

// userRepository Repository 实现
type userRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.UserRepository = (*userRepository)(nil)

// NewUserRepository 创建用户仓库实现
func NewUserRepository(db *database.PostgresDB) repo.UserRepository {
	return &userRepository{db: db}
}

// GetByID 根据 ID 查询用户并支持加载关联数据（使用 Preload）
// relations 参数可以是："Department", "Position", "Roles" 等
// 使用示例：
//   - 只查用户：GetByID("123")
//   - 查用户 + 部门：GetByID("123", "Department")
//   - 查用户 + 部门 + 岗位 + 角色：GetByID("123", "Department", "Position", "Roles")
func (r *userRepository) GetByID(id string, relations ...repo.UserRelation) (*entity.User, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel userModel
	query := r.db.Where("id = ?", intID)

	// 根据指定的关联关系进行 Preload
	for _, relation := range relations {
		query = query.Preload(string(relation))
	}

	result := query.First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByUsername 根据用户名查询用户
func (r *userRepository) GetByUsername(username string) (*entity.User, error) {
	var dbModel userModel
	result := r.db.Where("username = ?", username).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByEmail 根据邮箱查询用户
func (r *userRepository) GetByEmail(email string) (*entity.User, error) {
	var dbModel userModel
	result := r.db.Where("email = ?", email).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByPhone 根据手机号查询用户
func (r *userRepository) GetByPhone(phone string) (*entity.User, error) {
	var dbModel userModel
	result := r.db.Where("phone = ?", phone).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByWechatOpenID 根据微信 OpenID 查询用户
func (r *userRepository) GetByWechatOpenID(openID string) (*entity.User, error) {
	var dbModel userModel
	result := r.db.Where("wechat_open_id = ?", openID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建用户
func (r *userRepository) Create(user *entity.User) error {
	dbModel, err := fromDomain(user)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	user.ID = common.ToString(dbModel.ID)
	return nil
}

// Update 更新用户
func (r *userRepository) Update(user *entity.User) error {
	dbModel, err := fromDomain(user)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除用户
func (r *userRepository) Delete(id string) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&userModel{}, intID).Error
}

// List 分页查询用户列表
func (r *userRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.User, int64, error) {
	var total int64
	var dbModels []userModel

	query := r.db.Model(&userModel{})

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*entity.User, len(dbModels))
	for i, model := range dbModels {
		users[i] = model.toDomain()
	}

	return users, total, nil
}

// UpdateFields 更新用户指定字段
func (r *userRepository) UpdateFields(id string, updates map[string]interface{}) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Model(&userModel{}).Where("id = ?", intID).Updates(updates).Error
}

// SetUserSpecialPermissions 设置用户特殊 API 权限
// TODO: 需要创建用户 -API 关联表，目前暂时不实现
func (r *userRepository) SetUserSpecialPermissions(userID string, apiIDs []string) error {
	// TODO: 实现用户特殊 API 权限设置
	// 需要创建中间表 base_user_apis
	return nil
}

// SetUserSpecialBtnPermissions 设置用户特殊按钮权限
// TODO: 需要创建用户 - 按钮权限关联表，目前暂时不实现
func (r *userRepository) SetUserSpecialBtnPermissions(userID string, btnPermIDs []string) error {
	// TODO: 实现用户特殊按钮权限设置
	// 需要创建中间表 base_user_btn_perms
	return nil
}

// GetUserSpecialPermissions 获取用户特殊 API 权限
// TODO: 需要创建用户-API 关联表，目前暂时不实现
func (r *userRepository) GetUserSpecialPermissions(userID string) ([]*entity.API, error) {
	// TODO: 实现用户特殊 API 权限获取
	return []*entity.API{}, nil
}

// GetUserSpecialBtnPermissions 获取用户特殊按钮权限
// TODO: 需要创建用户 - 按钮权限关联表，目前暂时不实现
func (r *userRepository) GetUserSpecialBtnPermissions(userID string) ([]*entity.BtnPerm, error) {
	// TODO: 实现用户特殊按钮权限获取
	return []*entity.BtnPerm{}, nil
}
