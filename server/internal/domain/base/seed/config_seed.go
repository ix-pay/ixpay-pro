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
			ConfigValue: "wx079ea8f5e3d25370",
			ConfigType:  1,
			Description: "微信公众号 AppID（服务号测试号）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_app_secret",
			ConfigValue: "a3eb4ff7a4162e5ffd905d04c73d5681",
			ConfigType:  1,
			Description: "微信公众号 AppSecret（服务号测试号）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_mch_id",
			ConfigValue: "your-wechat-mch-id",
			ConfigType:  1,
			Description: "微信支付商户号",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_mch_serial_no",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付商户 API 证书序列号（APIv3）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_apiv3_key",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付 APIv3 密钥（回调报文解密）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_private_key_path",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付商户私钥文件路径（apiclient_key.pem，与 wechat_private_key 二选一）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_private_key",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付商户私钥 PEM 内容（apiclient_key.pem 文本直存数据库，优先于文件路径）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_public_key_id",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付公钥 ID（PUB_KEY_ID_xxx，回调/应答验签）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_public_key_path",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付公钥文件路径（与 wechat_public_key 二选一）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_public_key",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付公钥 PEM 内容（文本直存数据库，优先于文件路径）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_miniapp_app_id",
			ConfigValue: "wxd0f3a22e75e60564",
			ConfigType:  1,
			Description: "微信小程序 AppID（小程序测试号，与 miniapp/project.config.json 一致）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_miniapp_app_secret",
			ConfigValue: "39db154b9e2d4907cfae8576ffefb11b",
			ConfigType:  1,
			Description: "微信小程序 AppSecret（小程序测试号）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_mode",
			ConfigValue: "direct",
			ConfigType:  1,
			Description: "微信支付模式：direct直连(默认) | partner服务商",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_sp_appid",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付服务商 APPID（服务商模式）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_sp_mchid",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付服务商商户号（服务商模式签名主体）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_sub_appid",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付子商户 APPID（服务商模式可选）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_sub_mchid",
			ConfigValue: "",
			ConfigType:  1,
			Description: "微信支付子商户号（服务商模式收款商户）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_refund_notify_url",
			ConfigValue: "https://ixpay.zcyxiaokeai.live/api/wx/refund/notify",
			ConfigType:  1,
			Description: "微信退款结果回调地址（V3，需公网 HTTPS）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_notify_url",
			ConfigValue: "https://ixpay.zcyxiaokeai.live/api/wx/med/notify",
			ConfigType:  1,
			Description: "微信医保收款回调地址（V3，需公网 HTTPS）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_city_id",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保支付城市ID（腾讯侧对接时提供）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_inst_name",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保支付医疗机构名称",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_inst_no",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保支付医疗机构编码（医保局提供）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_channel_no",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保支付渠道号（可选）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_test_env",
			ConfigValue: "false",
			ConfigType:  1,
			Description: "是否医保局测试环境（true/false）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_partner_id",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保授权查询合作方ID（wecity 平台分配）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_partner_secret",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保授权查询密钥（HmacSHA256 签名用）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_org_chnl_crtf_codg",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保机构渠道认证编码（国家局医保移动支付反馈单）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_med_org_app_id",
			ConfigValue: "",
			ConfigType:  1,
			Description: "医保定点医药机构应用ID（国家局医保移动支付反馈单）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_notify_url",
			ConfigValue: "https://ixpay.zcyxiaokeai.live/api/wx/pay/notify/wechat",
			ConfigType:  1,
			Description: "微信支付回调通知地址（V3，需公网 HTTPS）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_oauth_callback_url",
			ConfigValue: "https://ixpay.zcyxiaokeai.live/api/wx/auth/callback",
			ConfigType:  1,
			Description: "微信 OAuth 授权回调地址（后端接口，微信授权后回调到此地址）",
			Status:      1,
		},
		{
			ConfigKey:   "wechat_frontend_url",
			ConfigValue: "https://ixpay.zcyxiaokeai.live",
			ConfigType:  1,
			Description: "前端页面地址（微信 OAuth 回调后重定向的前端域名）",
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
