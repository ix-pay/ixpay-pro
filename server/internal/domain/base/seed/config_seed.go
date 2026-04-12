package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// ConfigSeed 系统配置种子数据
type ConfigSeed struct {
	configRepo repo.ConfigRepository
}

// NewConfigSeed 创建系统配置种子数据实例
func NewConfigSeed(configRepo repo.ConfigRepository) Seed {
	return &ConfigSeed{
		configRepo: configRepo,
	}
}

// Version 返回种子数据版本
func (cs *ConfigSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (cs *ConfigSeed) Name() string {
	return "config_seed"
}

// Order 返回初始化顺序（最先执行）
func (cs *ConfigSeed) Order() int {
	return 0
}

// Init 初始化系统配置种子数据
func (cs *ConfigSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化系统配置种子数据")

	// 定义系统配置种子数据
	// 微信支付相关配置（供 wx 模块使用）
	configs := []*entity.Config{
		{
			ConfigKey:   "wechat_app_id",
			ConfigValue: "your-wechat-app-id",
			ConfigType:  "wechat",
			Description: "微信公众号 AppID",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_app_secret",
			ConfigValue: "your-wechat-app-secret",
			ConfigType:  "wechat",
			Description: "微信公众号 AppSecret",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_mch_id",
			ConfigValue: "your-wechat-mch-id",
			ConfigType:  "wechat",
			Description: "微信支付商户号",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_api_key",
			ConfigValue: "your-wechat-api-key",
			ConfigType:  "wechat",
			Description: "微信支付 API 密钥",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_notify_url",
			ConfigValue: "http://your-server.com/api/wx//pay/notify",
			ConfigType:  "wechat",
			Description: "微信支付回调通知地址",
			Status:      1,
		},
	}

	// 批量插入或更新配置（增量写入）
	for _, config := range configs {
		// 检查是否已存在
		existing, err := cs.configRepo.GetByKey(config.ConfigKey)
		if err != nil {
			// 不存在则创建
			if err := cs.configRepo.Create(config); err != nil {
				logger.Error("创建配置失败", "config_key", config.ConfigKey, "error", err)
				return err
			}
			logger.Info("创建配置成功", "config_key", config.ConfigKey)
		} else {
			// 存在则更新
			existing.ConfigValue = config.ConfigValue
			existing.ConfigType = config.ConfigType
			existing.Description = config.Description
			existing.Status = config.Status
			if err := cs.configRepo.Update(existing); err != nil {
				logger.Error("更新配置失败", "config_key", config.ConfigKey, "error", err)
				return err
			}
			logger.Info("更新配置成功", "id", existing.ID, "config_key", config.ConfigKey)
		}
	}

	logger.Info("系统配置种子数据初始化完成")
	return nil
}
