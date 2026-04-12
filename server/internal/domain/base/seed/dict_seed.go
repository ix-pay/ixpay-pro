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
	dictRepo repo.DictRepository
}

// NewDictSeed 创建字典种子数据实例
func NewDictSeed(dictRepo repo.DictRepository) Seed {
	return &DictSeed{
		dictRepo: dictRepo,
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
	}

	logger.Info("字典种子数据初始化完成", "total", len(dicts))
	return nil
}

// getDictionaries 获取所有字典定义
func (ds *DictSeed) getDictionaries() []*entity.Dict {
	now := time.Now()
	systemUserID := "system"

	return []*entity.Dict{
		// ==================== 用户类型 ====================
		{
			ID:          "dict_user_type",
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
					ID:          "dict_user_type_admin",
					DictID:      "dict_user_type",
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
					ID:          "dict_user_type_normal",
					DictID:      "dict_user_type",
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
					ID:          "dict_user_type_guest",
					DictID:      "dict_user_type",
					ItemKey:     "guest",
					ItemValue:   "访客",
					Sort:        3,
					Description: "访客用户",
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
			ID:          "dict_gender",
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
					ID:          "dict_gender_unknown",
					DictID:      "dict_gender",
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
					ID:          "dict_gender_male",
					DictID:      "dict_gender",
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
					ID:          "dict_gender_female",
					DictID:      "dict_gender",
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
			ID:          "dict_notice_type",
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
					ID:          "dict_notice_type_system",
					DictID:      "dict_notice_type",
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
					ID:          "dict_notice_type_activity",
					DictID:      "dict_notice_type",
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
					ID:          "dict_notice_type_notice",
					DictID:      "dict_notice_type",
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
			ID:          "dict_task_type",
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
					ID:          "dict_task_type_data_sync",
					DictID:      "dict_task_type",
					ItemKey:     "data_sync",
					ItemValue:   "数据同步",
					Sort:        1,
					Description: "数据同步任务",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          "dict_task_type_report",
					DictID:      "dict_task_type",
					ItemKey:     "report",
					ItemValue:   "报表生成",
					Sort:        2,
					Description: "报表生成任务",
					Status:      1,
					CreatedBy:   systemUserID,
					CreatedAt:   now,
					UpdatedBy:   systemUserID,
					UpdatedAt:   now,
				},
				{
					ID:          "dict_task_type_clean",
					DictID:      "dict_task_type",
					ItemKey:     "clean",
					ItemValue:   "清理任务",
					Sort:        3,
					Description: "数据清理任务",
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
			ID:          "dict_config_type",
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
					ID:          "dict_config_type_system",
					DictID:      "dict_config_type",
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
					ID:          "dict_config_type_wechat",
					DictID:      "dict_config_type",
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
					ID:          "dict_config_type_business",
					DictID:      "dict_config_type",
					ItemKey:     "business",
					ItemValue:   "业务配置",
					Sort:        3,
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
