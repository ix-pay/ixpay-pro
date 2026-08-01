package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// DepartmentSeed 部门种子数据
type DepartmentSeed struct {
	deptRepo repo.DepartmentRepository
}

// NewDepartmentSeed 创建部门种子数据实例
func NewDepartmentSeed(deptRepo repo.DepartmentRepository) Seed {
	return &DepartmentSeed{
		deptRepo: deptRepo,
	}
}

// Version 返回种子数据版本
func (ds *DepartmentSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (ds *DepartmentSeed) Name() string {
	return "department_seed"
}

// Order 返回初始化顺序
func (ds *DepartmentSeed) Order() int {
	return 3
}

// Init 初始化部门种子数据
func (ds *DepartmentSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化部门种子数据")

	// 初始化顶级部门
	topDept, err := ds.findOrCreate("ixpay", 0, 0, "公司顶级部门", logger)
	if err != nil {
		return err
	}
	logger.Info("顶级部门初始化完成", "id", topDept.ID)

	// 初始化运营中心
	opsCenter, err := ds.findOrCreate("运营中心", topDept.ID, 1, "负责业务运营和市场推广", logger)
	if err != nil {
		return err
	}
	logger.Info("运营中心初始化完成", "id", opsCenter.ID)

	// 初始化技术中心
	techCenter, err := ds.findOrCreate("技术中心", topDept.ID, 2, "负责产品研发和技术实现", logger)
	if err != nil {
		return err
	}
	logger.Info("技术中心初始化完成", "id", techCenter.ID)

	return nil
}

// findOrCreate 查找或创建部门（通用方法）
func (ds *DepartmentSeed) findOrCreate(name string, parentID int64, sort int, description string, logger logger.Logger) (*entity.Department, error) {
	allDepts, err := ds.deptRepo.GetAll()
	if err == nil {
		for _, dept := range allDepts {
			if dept.Name == name && dept.ParentID == parentID {
				logger.Info("部门已存在，跳过创建", "id", dept.ID, "name", name)
				return dept, nil
			}
		}
	}

	newDept := &entity.Department{
		Name:        name,
		ParentID:    parentID,
		LeaderID:    0,
		Sort:        sort,
		Status:      1,
		Description: description,
	}

	if err := ds.deptRepo.Create(newDept); err != nil {
		logger.Error("创建部门失败", "name", name, "error", err)
		return nil, err
	}

	logger.Info("创建部门成功", "id", newDept.ID, "name", name)
	return newDept, nil
}
