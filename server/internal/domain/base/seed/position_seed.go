package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// PositionSeed 岗位种子数据
type PositionSeed struct {
	positionRepo repo.PositionRepository
}

// NewPositionSeed 创建岗位种子数据实例
func NewPositionSeed(positionRepo repo.PositionRepository) Seed {
	return &PositionSeed{
		positionRepo: positionRepo,
	}
}

// Version 返回种子数据版本
func (ps *PositionSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (ps *PositionSeed) Name() string {
	return "position_seed"
}

// Order 返回初始化顺序（第三个执行）
func (ps *PositionSeed) Order() int {
	return 3
}

// Init 初始化岗位种子数据
func (ps *PositionSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化岗位种子数据")

	// 初始化技术岗位
	positions := []struct {
		Name        string
		Code        string
		Sort        int
		Description string
	}{
		{
			Name:        "CEO",
			Code:        "ceo",
			Sort:        0,
			Description: "首席执行官",
		},
		{
			Name:        "CTO",
			Code:        "cto",
			Sort:        1,
			Description: "首席技术官",
		},
		{
			Name:        "高级后端工程师",
			Code:        "senior_backend_engineer",
			Sort:        2,
			Description: "负责后端系统架构和核心功能开发",
		},
		{
			Name:        "高级前端工程师",
			Code:        "senior_frontend_engineer",
			Sort:        3,
			Description: "负责前端系统架构和核心功能开发",
		},
		{
			Name:        "产品经理",
			Code:        "product_manager",
			Sort:        4,
			Description: "负责产品规划和设计",
		},
		{
			Name:        "运营经理",
			Code:        "ops_manager",
			Sort:        5,
			Description: "负责业务运营和市场推广",
		},
		{
			Name:        "测试工程师",
			Code:        "qa_engineer",
			Sort:        6,
			Description: "负责系统测试和质量保障",
		},
		{
			Name:        "运维工程师",
			Code:        "devops_engineer",
			Sort:        7,
			Description: "负责系统运维和部署",
		},
	}

	for _, pos := range positions {
		if _, err := ps.initPosition(pos.Name, pos.Code, pos.Sort, pos.Description, logger); err != nil {
			logger.Error("初始化岗位失败", "name", pos.Name, "error", err)
			return err
		}
		logger.Info("岗位初始化完成", "name", pos.Name, "code", pos.Code)
	}

	return nil
}

// initPosition 初始化单个岗位
func (ps *PositionSeed) initPosition(name, code string, sort int, description string, logger logger.Logger) (*entity.Position, error) {
	// 1. 先检查 code 是否存在
	position, err := ps.positionRepo.GetByCode(code)
	if err == nil {
		logger.Info("岗位已存在，跳过创建", "id", position.ID, "name", position.Name, "code", position.Code)
		return position, nil
	}

	// 2. 检查 name 是否存在
	position, err = ps.positionRepo.GetByName(name)
	if err == nil {
		logger.Info("岗位名已存在，跳过创建", "id", position.ID, "name", position.Name, "code", position.Code)
		return position, nil
	}

	// 3. 只有当 code 和 name 都不存在时才创建
	newPosition := &entity.Position{
		Name:        name,
		Code:        code,
		Sort:        sort,
		Status:      1,
		Description: description,
	}

	if err := ps.positionRepo.Create(newPosition); err != nil {
		logger.Error("创建岗位失败", "name", name, "error", err)
		return nil, err
	}

	logger.Info("创建岗位成功", "id", newPosition.ID, "name", name, "code", code)
	return newPosition, nil
}
