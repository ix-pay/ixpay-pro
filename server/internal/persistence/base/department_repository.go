package persistence

import (
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// departmentModel 部门数据库模型
type departmentModel struct {
	database.SnowflakeBaseModel
	Name        string `gorm:"size:100;not null"`
	ParentID    int64  `gorm:"default:0;index"`
	LeaderID    int64  `gorm:"index"`
	Sort        int    `gorm:"default:0"`
	Status      int    `gorm:"default:1"`
	Description string `gorm:"size:255"`
	// GORM 关联关系
	Children []departmentModel `gorm:"foreignKey:parent_id;references:id"`
	Parent   *departmentModel  `gorm:"foreignKey:parent_id;references:id"`
	Leader   *userModel        `gorm:"foreignKey:leader_id;references:id"`
}

// TableName 指定表名
func (departmentModel) TableName() string {
	return "base_departments"
}

// toDomain 将数据库模型转换为领域实体
func (m *departmentModel) toDomain() *entity.Department {
	if m == nil {
		return nil
	}
	dept := &entity.Department{
		ID:          common.ToString(m.ID),
		Name:        m.Name,
		ParentID:    common.ToString(m.ParentID),
		LeaderID:    common.ToString(m.LeaderID),
		Sort:        m.Sort,
		Status:      m.Status,
		Description: m.Description,
		CreatedBy:   common.ToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   common.ToString(m.UpdatedBy),
		UpdatedAt:   m.UpdatedAt,
	}

	// ⭐ 处理关联数据 - 子部门
	if len(m.Children) > 0 {
		children := make([]*entity.Department, len(m.Children))
		for i, child := range m.Children {
			children[i] = child.toDomain()
		}
		dept.Children = children
	}

	// ⭐ 处理关联数据 - 父部门
	if m.Parent != nil {
		dept.Parent = m.Parent.toDomain()
	}

	// ⭐ 处理关联数据 - 负责人
	if m.Leader != nil {
		dept.Leader = m.Leader.toDomain()
	}

	return dept
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainDepartment(dept *entity.Department) (*departmentModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(dept.ID, dept.CreatedBy, dept.UpdatedBy)

	return &departmentModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		Name:        dept.Name,
		ParentID:    common.TryParseInt64(dept.ParentID),
		LeaderID:    common.TryParseInt64(dept.LeaderID),
		Sort:        dept.Sort,
		Status:      dept.Status,
		Description: dept.Description,
	}, nil
}

// departmentRepository Repository 实现
type departmentRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.DepartmentRepository = (*departmentRepository)(nil)

// NewDepartmentRepository 创建部门仓库实现
func NewDepartmentRepository(db *database.PostgresDB) repo.DepartmentRepository {
	return &departmentRepository{db: db}
}

// GetByID 根据 ID 查询部门并支持加载关联数据
func (r *departmentRepository) GetByID(id string, relations ...repo.DepartmentRelation) (*entity.Department, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var dbModel departmentModel
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

// Create 创建部门
func (r *departmentRepository) Create(department *entity.Department) error {
	dbModel, err := fromDomainDepartment(department)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	department.ID = common.ToString(dbModel.ID)
	return nil
}

// Update 更新部门
func (r *departmentRepository) Update(department *entity.Department) error {
	dbModel, err := fromDomainDepartment(department)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除部门
func (r *departmentRepository) Delete(id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	return r.db.Delete(&departmentModel{}, intID).Error
}

// List 分页查询部门列表
func (r *departmentRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Department, int64, error) {
	var total int64
	var dbModels []departmentModel

	query := r.db.Model(&departmentModel{})

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

	departments := make([]*entity.Department, len(dbModels))
	for i, model := range dbModels {
		departments[i] = model.toDomain()
	}

	return departments, total, nil
}

// GetAll 获取所有部门
func (r *departmentRepository) GetAll() ([]*entity.Department, error) {
	var dbModels []departmentModel
	result := r.db.Order("parent_id ASC, sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	departments := make([]*entity.Department, len(dbModels))
	for i, model := range dbModels {
		departments[i] = model.toDomain()
	}

	return departments, nil
}

// GetChildrenByParentID 根据父部门 ID 获取子部门
func (r *departmentRepository) GetChildrenByParentID(parentID string) ([]*entity.Department, error) {
	intParentID, _ := strconv.ParseInt(parentID, 10, 64)
	var dbModels []departmentModel
	result := r.db.Where("parent_id = ?", intParentID).Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	departments := make([]*entity.Department, len(dbModels))
	for i, model := range dbModels {
		departments[i] = model.toDomain()
	}

	return departments, nil
}

// GetDepartmentTree 获取部门树
func (r *departmentRepository) GetDepartmentTree() ([]*entity.Department, error) {
	return r.GetAll()
}

// GetDepartmentPath 获取部门路径
func (r *departmentRepository) GetDepartmentPath(id string) ([]*entity.Department, error) {
	// TODO: 实现部门路径查询
	return nil, nil
}
