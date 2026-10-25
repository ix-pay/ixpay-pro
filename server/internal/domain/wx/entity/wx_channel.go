package entity

import "time"

// 微信多渠道配置实体：多服务号/多小程序 × 多商户号（管理平台端配置，前端"空壳+appKey标志"动态获取）

// 微信应用类型
const (
	WxAppTypeMP      = "mp"      // 微信服务号（H5 网页授权 + JS-SDK 支付）
	WxAppTypeMiniapp = "miniapp" // 微信小程序（jscode2session 登录 + 小程序支付）
)

// WxAppConfig 微信应用配置（一个服务号或一个小程序实例）
// appKey 为前端实例标志：wxapp/miniapp 仅配置此标志，启动时按标志从后端获取全部配置
type WxAppConfig struct {
	ID               int64     // 应用配置 ID
	AppKey           string    // 应用实例标志（唯一，如 wxapp-main、mini-demo）
	AppType          string    // 应用类型：mp(服务号) | miniapp(小程序)
	AppID            string    // 微信 AppID
	AppSecret        string    // 微信 AppSecret（数据库加密存储）
	AuthScope        string    // 网页授权 scope（仅服务号生效：snsapi_base/snsapi_userinfo）
	OauthCallbackURL string    // 网页授权回调地址（空则用全局 wechat_oauth_callback_url）
	FrontendURL      string    // 授权后回跳的前端地址（空则用全局 wechat_frontend_url）
	Description      string    // 应用描述
	Status           int       // 状态：1-启用，0-禁用
	CreatedAt        time.Time // 创建时间
	UpdatedAt        time.Time // 更新时间
}

// WxMerchantConfig 微信支付商户配置
type WxMerchantConfig struct {
	ID              int64     // 商户配置 ID
	MchID           string    // 微信支付商户号（唯一）
	MchName         string    // 商户名称
	MchSerialNo     string    // 商户 API 证书序列号
	Apiv3Key        string    // APIv3 密钥（数据库加密存储）
	PrivateKey      string    // 商户私钥 PEM 内容（数据库加密存储）
	PublicKeyID     string    // 微信支付公钥 ID
	PublicKey       string    // 微信支付公钥 PEM 内容（公钥非敏感，明文存储）
	PayNotifyURL    string    // 支付回调地址（空则用全局 wechat_notify_url）
	RefundNotifyURL string    // 退款回调地址（空则用全局 wechat_refund_notify_url）
	Status          int       // 状态：1-启用，0-禁用
	CreatedAt       time.Time // 创建时间
	UpdatedAt       time.Time // 更新时间
}

// WxAppMerchant 应用-商户关联（一个应用可关联多个商户，一个商户可被多个应用使用）
// IsDefault 为「默认商户」唯一权威数据源：支付未指定商户号时按此标记路由（应用编辑表单不再单独存储默认商户）
type WxAppMerchant struct {
	ID               int64     // 关联 ID
	AppConfigID      int64     // 应用配置 ID
	MerchantConfigID int64     // 商户配置 ID
	IsDefault        bool      // 是否该应用的默认商户
	CreatedAt        time.Time // 创建时间
	UpdatedAt        time.Time // 更新时间
}

// IsEnabled 检查应用配置是否启用
func (a *WxAppConfig) IsEnabled() bool { return a.Status == 1 }

// IsEnabled 检查商户配置是否启用
func (m *WxMerchantConfig) IsEnabled() bool { return m.Status == 1 }
