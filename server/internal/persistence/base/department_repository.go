package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// departmentModel 部门数据库模型
type departmentModel struct {
	database.SnowflakeBaseModel
	Name        string `gorm:"size:100;not null"`
	ParentID    *int64 `gorm:"not null;default:0;index"`
	LeaderID    *int64 `gorm:"not null;default:0;index"`
	Sort        *int   `gorm:"not null;default:0"`
	Status      *int   `gorm:"not null;default:1"`
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
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   m.UpdatedBy,
		UpdatedAt:   m.UpdatedAt,
	}

	// 安全解引用，提供默认值
	if m.ParentID != nil {
		dept.ParentID = *m.ParentID
	} else {
		dept.ParentID = 0
	}

	if m.LeaderID != nil {
		dept.LeaderID = *m.LeaderID
	} else {
		dept.LeaderID = 0
	}

	if m.Sort != nil {
		dept.Sort = *m.Sort
	} else {
		dept.Sort = 0
	}

	if m.Status != nil {
		dept.Status = *m.Status
	} else {
		dept.Status = 1
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
	return &departmentModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        dept.ID,
			CreatedBy: dept.CreatedBy,
			UpdatedBy: dept.UpdatedBy,
		},
		Name:        dept.Name,
		ParentID:    common.Int64Ptr(dept.ParentID),
		LeaderID:    common.Int64Ptr(dept.LeaderID),
		Sort:        common.IntPtr(dept.Sort),
		Status:      common.IntPtr(dept.Status),
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
func (r *departmentRepository) GetByID(id int64, relations ...repo.DepartmentRelation) (*entity.Department, error) {
	var dbModel departmentModel
	query := r.db.Where("id = ?", id)

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
	department.ID = dbModel.ID
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
func (r *departmentRepository) Delete(id int64) error {
	return r.db.Delete(&departmentModel{}, id).Error
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
func (r *departmentRepository) GetChildrenByParentID(parentID int64) ([]*entity.Department, error) {
	var dbModels []departmentModel
	result := r.db.Where("parent_id = ?", parentID).Order("sort ASC").Find(&dbModels)
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
func (r *departmentRepository) GetDepartmentPath(id int64) ([]*entity.Department, error) {
	// TODO: 实现部门路径查询
	return nil, nil
}
