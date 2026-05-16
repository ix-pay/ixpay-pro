package seed

import (
	"sort"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// SeedManager 种子数据管理器
type SeedManager struct {
	seeds  []Seed
	logger logger.Logger
}

// NewSeedManager 创建种子数据管理器实例
func NewSeedManager(logger logger.Logger) *SeedManager {
	return &SeedManager{
		seeds:  []Seed{},
		logger: logger,
	}
}

// Register 注册种子数据
func (sm *SeedManager) Register(seed Seed) {
	if seed == nil {
		return
	}

	// 检查是否已经注册
	for _, s := range sm.seeds {
		if s.Name() == seed.Name() {
			sm.logger.Warn("种子数据已注册，跳过", "name", seed.Name())
			return
		}
	}

	sm.seeds = append(sm.seeds, seed)
	// 按Order排序
	sm.sortSeeds()
}

// RegisterAll 注册所有种子数据
func (sm *SeedManager) RegisterAll(seeds []Seed) {
	for _, seed := range seeds {
		sm.Register(seed)
	}
}

// Init 初始化所有种子数据
func (sm *SeedManager) Init(db *database.PostgresDB) error {
	sm.logger.Info("开始初始化种子数据")
	sm.logger.Info("已注册的种子数据数量", "count", len(sm.seeds))

	for _, seed := range sm.seeds {
		sm.logger.Info("初始化种子数据", "name", seed.Name(), "version", seed.Version(), "order", seed.Order())
		if err := seed.Init(db, sm.logger); err != nil {
			sm.logger.Error("初始化种子数据失败", "name", seed.Name(), "error", err)
			return err
		}
		sm.logger.Info("初始化种子数据成功", "name", seed.Name())
	}

	sm.logger.Info("所有种子数据初始化完成")
	return nil
}

// GetSeeds 返回所有注册的种子数据
func (sm *SeedManager) GetSeeds() []Seed {
	return sm.seeds
}

// sortSeeds 按Order排序种子数据
func (sm *SeedManager) sortSeeds() {
	sort.Slice(sm.seeds, func(i, j int) bool {
		return sm.seeds[i].Order() < sm.seeds[j].Order()
	})
}
