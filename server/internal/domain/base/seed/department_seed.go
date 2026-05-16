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
	topDept, err := ds.initTopDepartment(logger)
	if err != nil {
		return err
	}
	logger.Info("顶级部门初始化完成", "id", topDept.ID)

	// 初始化技术部
	techDept, err := ds.initTechDepartment(topDept.ID, logger)
	if err != nil {
		return err
	}
	logger.Info("技术部初始化完成", "id", techDept.ID)

	// 初始化产品部
	productDept, err := ds.initProductDepartment(topDept.ID, logger)
	if err != nil {
		return err
	}
	logger.Info("产品部初始化完成", "id", productDept.ID)

	// 初始化运营部
	opsDept, err := ds.initOpsDepartment(topDept.ID, logger)
	if err != nil {
		return err
	}
	logger.Info("运营部初始化完成", "id", opsDept.ID)

	return nil
}

// initTopDepartment 初始化顶级部门
func (ds *DepartmentSeed) initTopDepartment(logger logger.Logger) (*entity.Department, error) {
	allDepts, err := ds.deptRepo.GetAll()
	if err == nil {
		for _, dept := range allDepts {
			if dept.Name == "_ixpay" && dept.ParentID == 0 {
				logger.Info("顶级部门已存在，跳过创建", "id", dept.ID, "name", dept.Name)
				return dept, nil
			}
		}
	}

	newDept := &entity.Department{
		Name:        "ixpay",
		ParentID:    0,
		LeaderID:    0,
		Sort:        0,
		Status:      1,
		Description: "公司顶级部门",
	}

	if err := ds.deptRepo.Create(newDept); err != nil {
		logger.Error("创建顶级部门失败", "error", err)
		return nil, err
	}

	logger.Info("创建顶级部门成功", "id", newDept.ID)
	return newDept, nil
}

// initTechDepartment 初始化技术部
func (ds *DepartmentSeed) initTechDepartment(parentID int64, logger logger.Logger) (*entity.Department, error) {
	allDepts, err := ds.deptRepo.GetAll()
	if err == nil {
		for _, dept := range allDepts {
			if dept.Name == "技术部" && dept.ParentID == parentID {
				logger.Info("技术部已存在，跳过创建", "id", dept.ID, "name", dept.Name)
				return dept, nil
			}
		}
	}

	newDept := &entity.Department{
		Name:        "技术部",
		ParentID:    parentID,
		LeaderID:    0,
		Sort:        1,
		Status:      1,
		Description: "负责产品研发和技术实现",
	}

	if err := ds.deptRepo.Create(newDept); err != nil {
		logger.Error("创建技术部失败", "error", err)
		return nil, err
	}

	logger.Info("创建技术部成功", "id", newDept.ID)
	return newDept, nil
}

// initProductDepartment 初始化产品部
func (ds *DepartmentSeed) initProductDepartment(parentID int64, logger logger.Logger) (*entity.Department, error) {
	allDepts, err := ds.deptRepo.GetAll()
	if err == nil {
		for _, dept := range allDepts {
			if dept.Name == "产品部" && dept.ParentID == parentID {
				logger.Info("产品部已存在，跳过创建", "id", dept.ID, "name", dept.Name)
				return dept, nil
			}
		}
	}

	newDept := &entity.Department{
		Name:        "产品部",
		ParentID:    parentID,
		LeaderID:    0,
		Sort:        2,
		Status:      1,
		Description: "负责产品规划和设计",
	}

	if err := ds.deptRepo.Create(newDept); err != nil {
		logger.Error("创建产品部失败", "error", err)
		return nil, err
	}

	logger.Info("创建产品部成功", "id", newDept.ID)
	return newDept, nil
}

// initOpsDepartment 初始化运营部
func (ds *DepartmentSeed) initOpsDepartment(parentID int64, logger logger.Logger) (*entity.Department, error) {
	allDepts, err := ds.deptRepo.GetAll()
	if err == nil {
		for _, dept := range allDepts {
			if dept.Name == "运营部" && dept.ParentID == parentID {
				logger.Info("运营部已存在，跳过创建", "id", dept.ID, "name", dept.Name)
				return dept, nil
			}
		}
	}

	newDept := &entity.Department{
		Name:        "运营部",
		ParentID:    parentID,
		LeaderID:    0,
		Sort:        3,
		Status:      1,
		Description: "负责业务运营和市场推广",
	}

	if err := ds.deptRepo.Create(newDept); err != nil {
		logger.Error("创建运营部失败", "error", err)
		return nil, err
	}

	logger.Info("创建运营部成功", "id", newDept.ID)
	return newDept, nil
}
