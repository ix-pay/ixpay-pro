package seed

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// DictSeed 字典种子数据
type DictSeed struct {
	dictRepo     repo.DictRepository
	dictItemRepo repo.DictItemRepository
}

// NewDictSeed 创建字典种子数据实例
func NewDictSeed(dictRepo repo.DictRepository, dictItemRepo repo.DictItemRepository) Seed {
	return &DictSeed{
		dictRepo:     dictRepo,
		dictItemRepo: dictItemRepo,
	}
}

// Version 返回种子数据版本
func (ds *DictSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (ds *DictSeed) Name() string {
	return "dict_seed"
}

// Order 返回初始化顺序（第六个执行）
func (ds *DictSeed) Order() int {
	return 6
}

// Init 初始化字典种子数据
func (ds *DictSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化字典种子数据")

	// 定义字典种子数据
	dicts := ds.getDictionaries()

	// 批量插入或更新字典（增量写入）
	for _, dict := range dicts {
		// 检查是否已存在
		existing, err := ds.dictRepo.GetByCode(dict.DictCode)
		if err != nil {
			// 不存在则创建
			if err := ds.dictRepo.Create(dict); err != nil {
				logger.Error("创建字典失败", "dict_code", dict.DictCode, "error", err)
				return err
			}
			logger.Info("创建字典成功", "dict_code", dict.DictCode)
		} else {
			// 存在则更新
			existing.DictName = dict.DictName
			existing.Description = dict.Description
			existing.Status = dict.Status
			if err := ds.dictRepo.Update(existing); err != nil {
				logger.Error("更新字典失败", "dict_code", dict.DictCode, "error", err)
				return err
			}
			logger.Info("更新字典成功", "id", existing.ID, "dict_code", dict.DictCode)
		}

		// 处理字典项的增量保存
		if err := ds.syncDictItems(dict, logger); err != nil {
			logger.Error("同步字典项失败", "dict_code", dict.DictCode, "error", err)
			return err
		}
	}

	logger.Info("字典种子数据初始化完成", "total", len(dicts))
	return nil
}

// syncDictItems 同步字典项（增量写入）
func (ds *DictSeed) syncDictItems(dict *entity.Dict, logger logger.Logger) error {
	// 获取已存在的字典项
	existingItems, err := ds.dictItemRepo.GetByDictID(dict.ID)
	if err != nil {
		// 如果查询失败但不是不存在错误，返回错误
		logger.Warn("查询字典项失败", "dict_id", dict.ID, "error", err)
		existingItems = nil
	}

	// 构建已存在字典项的映射（按 ItemKey）
	existingMap := make(map[string]*entity.DictItem)
	for _, item := range existingItems {
		existingMap[item.ItemKey] = item
	}

	// 遍历种子数据中的字典项
	for _, newItem := range dict.DictItems {
		existingItem, exists := existingMap[newItem.ItemKey]
		if !exists {
			// 不存在则创建
			if err := ds.dictItemRepo.Create(&newItem); err != nil {
				logger.Error("创建字典项失败", "dict_id", dict.ID, "item_key", newItem.ItemKey, "error", err)
				return err
			}
			logger.Info("创建字典项成功", "dict_id", dict.ID, "item_key", newItem.ItemKey)
		} else {
			// 存在则更新
			existingItem.ItemValue = newItem.ItemValue
			existingItem.Sort = newItem.Sort
			existingItem.Description = newItem.Description
			existingItem.Status = newItem.Status
			if err := ds.dictItemRepo.Update(existingItem); err != nil {
				logger.Error("更新字典项失败", "dict_id", dict.ID, "item_key", newItem.ItemKey, "error", err)
				return err
			}
			logger.Info("更新字典项成功", "dict_id", dict.ID, "item_key", newItem.ItemKey)
		}
	}

	return nil
}

// getDictionaries 获取所有字典定义
func (ds *DictSeed) getDictionaries() []*entity.Dict {
	now := time.Now()
	var systemUserID int64 = 0 // 系统用户 ID 使用 0

	return []*entity.Dict{
		// ==================== 用户类型 ====================
		{
			ID:          1000, // 字典 ID 使用固定值
			DictName:    "用户类型",
			DictCode:    "user_type",
			Description: "用户类型分类",
			Status:      1,
			CreatedBy:   systemUserID,
			CreatedAt:   now,
			UpdatedBy:   systemUserID,
			UpdatedAt:   now,
			DictItems: []entity.DictItem{
				{
					ID:          1001,
					DictID:      1000,
					ItemKey:     "admin",
					ItemValue:   "管理员",
					Sort:        1,
					Description: "系统管理员",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          1002,
					DictID:      1000,
					ItemKey:     "normal",
					ItemValue:   "普通用户",
					Sort:        2,
					Description: "普通用户",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          1003,
					DictID:      1000,
					ItemKey:     "wechat",
					ItemValue:   "微信",
					Sort:        3,
					Description: "微信用户",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          1004,
					DictID:      1000,
					ItemKey:     "alipay",
					ItemValue:   "支付宝",
					Sort:        4,
					Description: "支付宝用户",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
			},
		},

		// ==================== 性别 ====================
		{
			ID:          2000,
			DictName:    "性别",
			DictCode:    "gender",
			Description: "性别分类",
			Status:      1,
			CreatedBy:   systemUserID,
			CreatedAt:   now,
			UpdatedBy:   systemUserID,
			UpdatedAt:   now,
			DictItems: []entity.DictItem{
				{
					ID:          2001,
					DictID:      2000,
					ItemKey:     "unknown",
					ItemValue:   "未知",
					Sort:        0,
					Description: "未知性别",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          2002,
					DictID:      2000,
					ItemKey:     "male",
					ItemValue:   "男",
					Sort:        1,
					Description: "男性",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          2003,
					DictID:      2000,
					ItemKey:     "female",
					ItemValue:   "女",
					Sort:        2,
					Description: "女性",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
			},
		},

		// ==================== 公告类型 ====================
		{
			ID:          3000,
			DictName:    "公告类型",
			DictCode:    "notice_type",
			Description: "公告类型分类",
			Status:      1,
			CreatedBy:   systemUserID,
			CreatedAt:   now,
			UpdatedBy:   systemUserID,
			UpdatedAt:   now,
			DictItems: []entity.DictItem{
				{
					ID:          3001,
					DictID:      3000,
					ItemKey:     "system",
					ItemValue:   "系统公告",
					Sort:        1,
					Description: "系统维护、升级等通知",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          3002,
					DictID:      3000,
					ItemKey:     "activity",
					ItemValue:   "活动公告",
					Sort:        2,
					Description: "活动推广、优惠等通知",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          3003,
					DictID:      3000,
					ItemKey:     "notice",
					ItemValue:   "通知",
					Sort:        3,
					Description: "一般通知",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
			},
		},

		// ==================== 任务类型 ====================
		{
			ID:          4000,
			DictName:    "任务类型",
			DictCode:    "task_type",
			Description: "任务类型分类",
			Status:      1,
			CreatedBy:   systemUserID,
			CreatedAt:   now,
			UpdatedBy:   systemUserID,
			UpdatedAt:   now,
			DictItems: []entity.DictItem{
				{
					ID:          4001,
					DictID:      4000,
					ItemKey:     "http",
					ItemValue:   "HTTP请求",
					Sort:        1,
					Description: "HTTP请求任务",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          4002,
					DictID:      4000,
					ItemKey:     "database",
					ItemValue:   "数据库任务",
					Sort:        2,
					Description: "数据库SQL任务",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          4003,
					DictID:      4000,
					ItemKey:     "cache",
					ItemValue:   "缓存任务",
					Sort:        3,
					Description: "缓存操作任务",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          4004,
					DictID:      4000,
					ItemKey:     "script",
					ItemValue:   "脚本任务",
					Sort:        4,
					Description: "脚本执行任务",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
			},
		},

		// ==================== 配置类型 ====================
		{
			ID:          5000,
			DictName:    "配置类型",
			DictCode:    "config_type",
			Description: "系统配置类型分类",
			Status:      1,
			CreatedBy:   systemUserID,
			CreatedAt:   now,
			UpdatedBy:   systemUserID,
			UpdatedAt:   now,
			DictItems: []entity.DictItem{
				{
					ID:          5001,
					DictID:      5000,
					ItemKey:     "system",
					ItemValue:   "系统配置",
					Sort:        1,
					Description: "系统基础配置",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          5002,
					DictID:      5000,
					ItemKey:     "wechat",
					ItemValue:   "微信配置",
					Sort:        2,
					Description: "微信公众号/支付配置",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          5003,
					DictID:      5000,
					ItemKey:     "alipay",
					ItemValue:   "支付宝配置",
					Sort:        3,
					Description: "支付宝/支付配置",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          5004,
					DictID:      5000,
					ItemKey:     "business",
					ItemValue:   "业务配置",
					Sort:        4,
					Description: "业务相关配置",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
			},
		},
	}
}
