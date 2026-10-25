package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	wxpay_utility "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayutility"
	wxpayv3 "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayv3"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"

	"github.com/ix-pay/ixpay-pro/internal/config"
	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
)

// 微信支付 APIv3 平台地址（协议层已限定仅允许微信支付域名）
const wechatPlatformAPIBase = "https://api.weixin.qq.com"

// cstZone 统一东八区时间（skills 铁律：不依赖服务器本地时区）
var cstZone = time.FixedZone("CST", 8*3600)

// 数据库配置键（config_seed 中定义，数据库优先、YAML 降级）
const (
	cfgKeyAppID          = "wechat_app_id"
	cfgKeyAppSecret      = "wechat_app_secret"
	cfgKeyMchID          = "wechat_mch_id"
	cfgKeyMchSerialNo    = "wechat_mch_serial_no"
	cfgKeyAPIv3Key       = "wechat_apiv3_key"
	cfgKeyPrivateKeyPath = "wechat_private_key_path"
	cfgKeyPublicKeyID    = "wechat_public_key_id"
	cfgKeyPublicKeyPath  = "wechat_public_key_path"
	// 商户私钥/微信支付公钥 PEM 内容（数据库直存，优先于文件路径；官方工具类支持按文本加载）
	cfgKeyPrivateKeyPEM = "wechat_private_key"
	cfgKeyPublicKeyPEM  = "wechat_public_key"
	cfgKeyMiniAppID     = "wechat_miniapp_app_id"
	cfgKeyMiniAppSecret = "wechat_miniapp_app_secret"
	cfgKeyNotifyURL     = "wechat_notify_url"
	cfgKeyOAuthCallback = "wechat_oauth_callback_url"
	cfgKeyFrontendURL   = "wechat_frontend_url"
	// 支付模式与回调
	cfgKeyMode            = "wechat_mode"
	cfgKeySpAppID         = "wechat_sp_appid"
	cfgKeySpMchID         = "wechat_sp_mchid"
	cfgKeySubAppID        = "wechat_sub_appid"
	cfgKeySubMchID        = "wechat_sub_mchid"
	cfgKeyRefundNotifyURL = "wechat_refund_notify_url"
	cfgKeyMedNotifyURL    = "wechat_med_notify_url"
	// 医保支付
	cfgKeyMedCityID          = "wechat_med_city_id"
	cfgKeyMedInstName        = "wechat_med_inst_name"
	cfgKeyMedInstNo          = "wechat_med_inst_no"
	cfgKeyMedChannelNo       = "wechat_med_channel_no"
	cfgKeyMedTestEnv         = "wechat_med_test_env"
	cfgKeyMedPartnerID       = "wechat_med_partner_id"
	cfgKeyMedPartnerSecret   = "wechat_med_partner_secret"
	cfgKeyMedOrgChnlCrtfCodg = "wechat_med_org_chnl_crtf_codg"
	cfgKeyMedOrgAppID        = "wechat_med_org_app_id"
)

// WechatPayService 微信支付服务（微信支付 APIv3 直连模式 + 公众号平台接口）
// 参考项目 D:\UseCode\Gpay 的实现：
//   - 支付走 V3 JSON API（/v3/pay/transactions/jsapi），RSA 签名，V2 XML 统一下单已停更故弃用
//   - 网页授权走 api.weixin.qq.com/sns/*，接口调用凭证走 stable_token（服务号能力）
type WechatPayService struct {
	cfg         config.WechatConfig       // YAML 配置（历史兼容，已置空）
	cfgRepo     baseRepo.ConfigRepository // 数据库配置仓库
	channelRepo repo.WxChannelRepository  // 多渠道配置仓库（应用/商户）
	tenantRepo  baseRepo.TenantRepository // 商户号→租户映射（回调定位租户库商户，设计文档 01 §九）
	dbConfig    *config.WechatConfig      // 数据库加载的配置（覆盖 YAML）
	log         logger.Logger
	http        *http.Client
	cache       cache.Cache // token/ticket 缓存（可为 nil，退化为进程内缓存）

	mu              sync.RWMutex
	v3Client        *wxpayv3.Client           // 全局 V3 客户端（历史全局配置，向后兼容）
	v3Key           string                    // 构建 V3 客户端时的配置指纹
	merchantClients map[int64]*wxpayv3.Client // 多商户客户端池（key=商户配置ID）
	merchantKeys    map[int64]string          // 商户客户端配置指纹

	memMu        sync.Mutex
	memToken     string
	memTokenExp  time.Time
	memTicket    string
	memTicketExp time.Time
}

// NewWechatPayService 创建微信支付服务实例
func NewWechatPayService(cfg *config.Config, cfgRepo baseRepo.ConfigRepository, channelRepo repo.WxChannelRepository, tenantRepo baseRepo.TenantRepository, cacheClient cache.Cache, log logger.Logger) *WechatPayService {
	return &WechatPayService{
		cfg:             cfg.Wechat,
		cfgRepo:         cfgRepo,
		channelRepo:     channelRepo,
		tenantRepo:      tenantRepo,
		log:             log,
		http:            &http.Client{Timeout: 10 * time.Second},
		cache:           cacheClient,
		merchantClients: make(map[int64]*wxpayv3.Client),
		merchantKeys:    make(map[int64]string),
	}
}

// getConfig 返回当前有效的配置（数据库配置优先于 YAML 配置）
func (s *WechatPayService) getConfig() config.WechatConfig {
	s.mu.RLock()
	if s.dbConfig != nil {
		defer s.mu.RUnlock()
		return *s.dbConfig
	}
	s.mu.RUnlock()
	return s.cfg
}

// LoadConfigFromDB 从数据库加载微信配置，覆盖 YAML 配置
// 配置键在 config_seed.go 中定义：wechat_app_id/wechat_app_secret/wechat_mch_id/wechat_mch_serial_no/
// wechat_apiv3_key/wechat_private_key_path/wechat_public_key_id/wechat_public_key_path/
// wechat_miniapp_app_id/wechat_miniapp_app_secret/wechat_notify_url 等
func (s *WechatPayService) LoadConfigFromDB() error {
	keys := []string{
		cfgKeyAppID, cfgKeyAppSecret, cfgKeyMchID, cfgKeyMchSerialNo, cfgKeyAPIv3Key,
		cfgKeyPrivateKeyPath, cfgKeyPublicKeyID, cfgKeyPublicKeyPath, cfgKeyPrivateKeyPEM, cfgKeyPublicKeyPEM,
		cfgKeyMiniAppID, cfgKeyMiniAppSecret, cfgKeyNotifyURL, cfgKeyOAuthCallback, cfgKeyFrontendURL,
		// 支付模式与回调
		cfgKeyMode, cfgKeySpAppID, cfgKeySpMchID, cfgKeySubAppID, cfgKeySubMchID,
		cfgKeyRefundNotifyURL, cfgKeyMedNotifyURL,
		// 医保支付
		cfgKeyMedCityID, cfgKeyMedInstName, cfgKeyMedInstNo, cfgKeyMedChannelNo, cfgKeyMedTestEnv,
		cfgKeyMedPartnerID, cfgKeyMedPartnerSecret, cfgKeyMedOrgChnlCrtfCodg, cfgKeyMedOrgAppID,
	}
	configMap := make(map[string]string, len(keys))

	for _, key := range keys {
		cfg, err := s.cfgRepo.GetByKey(key)
		if err != nil {
			s.log.Warn("从数据库读取微信配置失败", "key", key, "error", err)
			continue
		}
		configMap[key] = cfg.ConfigValue
	}

	if len(configMap) == 0 {
		s.log.Warn("数据库微信配置为空，使用 YAML 配置降级")
		return fmt.Errorf("数据库微信配置为空")
	}

	yamlCfg := s.cfg
	dbCfg := &config.WechatConfig{
		AppID:            getConfigValue(configMap, cfgKeyAppID, yamlCfg.AppID),
		AppSecret:        getConfigValue(configMap, cfgKeyAppSecret, yamlCfg.AppSecret),
		MCHID:            getConfigValue(configMap, cfgKeyMchID, yamlCfg.MCHID),
		MchSerialNo:      getConfigValue(configMap, cfgKeyMchSerialNo, yamlCfg.MchSerialNo),
		APIv3Key:         getConfigValue(configMap, cfgKeyAPIv3Key, yamlCfg.APIv3Key),
		PrivateKeyPath:   getConfigValue(configMap, cfgKeyPrivateKeyPath, yamlCfg.PrivateKeyPath),
		PublicKeyID:      getConfigValue(configMap, cfgKeyPublicKeyID, yamlCfg.PublicKeyID),
		PublicKeyPath:    getConfigValue(configMap, cfgKeyPublicKeyPath, yamlCfg.PublicKeyPath),
		PrivateKeyPEM:    strings.TrimSpace(configMap[cfgKeyPrivateKeyPEM]),
		PublicKeyPEM:     strings.TrimSpace(configMap[cfgKeyPublicKeyPEM]),
		MiniAppID:        getConfigValue(configMap, cfgKeyMiniAppID, yamlCfg.MiniAppID),
		MiniAppSecret:    getConfigValue(configMap, cfgKeyMiniAppSecret, yamlCfg.MiniAppSecret),
		NotifyURL:        getConfigValue(configMap, cfgKeyNotifyURL, yamlCfg.NotifyURL),
		RedirectURL:      yamlCfg.RedirectURL,
		OAuthCallbackURL: getConfigValue(configMap, cfgKeyOAuthCallback, yamlCfg.OAuthCallbackURL),
		FrontendURL:      getConfigValue(configMap, cfgKeyFrontendURL, yamlCfg.FrontendURL),
		// 支付模式与回调
		Mode:            getConfigValue(configMap, cfgKeyMode, yamlCfg.Mode),
		SpAppID:         getConfigValue(configMap, cfgKeySpAppID, yamlCfg.SpAppID),
		SpMchID:         getConfigValue(configMap, cfgKeySpMchID, yamlCfg.SpMchID),
		SubAppID:        getConfigValue(configMap, cfgKeySubAppID, yamlCfg.SubAppID),
		SubMchID:        getConfigValue(configMap, cfgKeySubMchID, yamlCfg.SubMchID),
		RefundNotifyURL: getConfigValue(configMap, cfgKeyRefundNotifyURL, yamlCfg.RefundNotifyURL),
		MedNotifyURL:    getConfigValue(configMap, cfgKeyMedNotifyURL, yamlCfg.MedNotifyURL),
		// 医保支付
		MedCityID:          getConfigValue(configMap, cfgKeyMedCityID, yamlCfg.MedCityID),
		MedInstName:        getConfigValue(configMap, cfgKeyMedInstName, yamlCfg.MedInstName),
		MedInstNo:          getConfigValue(configMap, cfgKeyMedInstNo, yamlCfg.MedInstNo),
		MedChannelNo:       getConfigValue(configMap, cfgKeyMedChannelNo, yamlCfg.MedChannelNo),
		MedTestEnv:         strings.EqualFold(getConfigValue(configMap, cfgKeyMedTestEnv, fmt.Sprintf("%t", yamlCfg.MedTestEnv)), "true"),
		MedPartnerID:       getConfigValue(configMap, cfgKeyMedPartnerID, yamlCfg.MedPartnerID),
		MedPartnerSecret:   getConfigValue(configMap, cfgKeyMedPartnerSecret, yamlCfg.MedPartnerSecret),
		MedOrgChnlCrtfCodg: getConfigValue(configMap, cfgKeyMedOrgChnlCrtfCodg, yamlCfg.MedOrgChnlCrtfCodg),
		MedOrgAppID:        getConfigValue(configMap, cfgKeyMedOrgAppID, yamlCfg.MedOrgAppID),
	}

	s.mu.Lock()
	s.dbConfig = dbCfg
	// 配置可能已变化，重建 V3 客户端
	s.v3Client = nil
	s.v3Key = ""
	s.mu.Unlock()

	s.log.Info("从数据库加载微信配置成功")
	return nil
}

// getConfigValue 从 map 中取值，如果为空或为占位符则使用默认值
func getConfigValue(m map[string]string, key, defaultVal string) string {
	v, ok := m[key]
	if !ok || v == "" {
		return defaultVal
	}
	// 过滤种子占位符
	switch v {
	case "your-wechat-app-id", "your-wechat-app-secret", "your-wechat-mch-id",
		"your-wechat-api-key", "http://your-server.com":
		return defaultVal
	}
	return v
}

// ---------- V3 客户端 ----------

// v3ConfigFingerprint 生成 V3 客户端配置指纹（变化时重建客户端；PEM 内容以长度+首尾片段参与指纹，避免超长明文）
func (s *WechatPayService) v3ConfigFingerprint(cfg config.WechatConfig) string {
	pemFp := func(pem string) string {
		if pem == "" {
			return ""
		}
		head := pem
		if len(head) > 16 {
			head = head[:16]
		}
		tail := pem
		if len(tail) > 16 {
			tail = tail[len(tail)-16:]
		}
		return fmt.Sprintf("%d:%s..%s", len(pem), head, tail)
	}
	return strings.Join([]string{
		cfg.MCHID, cfg.MchSerialNo, cfg.APIv3Key,
		pemFp(cfg.PrivateKeyPEM), cfg.PrivateKeyPath,
		cfg.PublicKeyID, pemFp(cfg.PublicKeyPEM), cfg.PublicKeyPath,
	}, "|")
}

// getClient 获取 V3 客户端（懒加载；支付参数以数据库为唯一有效来源：私钥/公钥 PEM 内容优先，文件路径兜底）
func (s *WechatPayService) getClient() (*wxpayv3.Client, error) {
	cfg := s.getConfig()
	fingerprint := s.v3ConfigFingerprint(cfg)

	s.mu.RLock()
	client, currentKey := s.v3Client, s.v3Key
	s.mu.RUnlock()
	if client != nil && currentKey == fingerprint {
		return client, nil
	}

	// 校验必要配置（私钥：PEM 内容或文件路径二选一）
	if cfg.MCHID == "" || cfg.MchSerialNo == "" {
		return nil, fmt.Errorf("微信支付 V3 配置不完整（需要 mch_id/mch_serial_no，请在数据库配置表中维护）")
	}
	if cfg.PrivateKeyPEM == "" && cfg.PrivateKeyPath == "" {
		return nil, fmt.Errorf("商户私钥未配置（数据库配置键 wechat_private_key 存 PEM 内容，或 wechat_private_key_path 存文件路径）")
	}

	// 加载商户私钥（官方 WXPayUtility 工具类；数据库 PEM 内容优先）
	privKeyPEM := cfg.PrivateKeyPEM
	if privKeyPEM == "" {
		pemBytes, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("读取商户私钥文件失败: %w", err)
		}
		privKeyPEM = string(pemBytes)
	}
	privKey, err := wxpay_utility.LoadPrivateKey(privKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("加载商户私钥失败: %w", err)
	}

	clientCfg := wxpayv3.ClientConfig{
		MchID:      cfg.MCHID,
		SerialNo:   cfg.MchSerialNo,
		APIv3Key:   cfg.APIv3Key,
		PrivateKey: privKey,
		// 服务商模式
		IsPartner: cfg.Mode == "partner",
		SpMchID:   cfg.SpMchID,
		SubMchID:  cfg.SubMchID,
		SubAppID:  cfg.SubAppID,
	}
	if clientCfg.IsPartner && clientCfg.MchID == "" {
		clientCfg.MchID = cfg.SpMchID // 服务商模式签名主体为 sp_mchid
	}

	// 加载微信支付公钥（用于应答/回调验签与医保敏感字段加密；数据库 PEM 内容优先，未配置时跳过应答验签）
	if cfg.PublicKeyID != "" && (cfg.PublicKeyPEM != "" || cfg.PublicKeyPath != "") {
		pubKeyPEM := cfg.PublicKeyPEM
		if pubKeyPEM == "" {
			pubBytes, err := os.ReadFile(cfg.PublicKeyPath)
			if err != nil {
				return nil, fmt.Errorf("读取微信支付公钥文件失败: %w", err)
			}
			pubKeyPEM = string(pubBytes)
		}
		pubKey, err := wxpay_utility.LoadPublicKey(pubKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("加载微信支付公钥失败: %w", err)
		}
		clientCfg.PublicKeyID = cfg.PublicKeyID
		clientCfg.PublicKey = pubKey
	} else {
		s.log.Warn("未配置微信支付公钥，将跳过应答验签（建议在数据库配置 wechat_public_key_id + wechat_public_key 内容）")
	}

	newClient := wxpayv3.NewClient(clientCfg)
	s.mu.Lock()
	s.v3Client = newClient
	s.v3Key = fingerprint
	s.mu.Unlock()
	return newClient, nil
}

// ---------- 多渠道（多服务号/多小程序 × 多商户号） ----------

// PayChannel 渠道解析结果（一次支付所用的 应用 + 商户 + 协议客户端）
type PayChannel struct {
	App          *entity.WxAppConfig      // 微信应用配置（服务号/小程序）
	Merchant     *entity.WxMerchantConfig // 商户配置
	Client       *wxpayv3.Client          // 该商户的 V3 协议客户端
	PayNotifyURL string                   // 支付回调地址（商户配置优先，全局兜底）
}

// ResolveChannel 解析支付渠道：appKey 定位应用（空则用全局配置默认应用），mchID 定位商户（空则用应用默认商户）
// 应用与商户必须存在启用状态的有效关联；ctx 携带租户上下文时渠道路由到对应租户库（业务请求 JWT 已注入）
func (s *WechatPayService) ResolveChannel(ctx context.Context, appKey, mchID string) (*PayChannel, error) {
	if err := s.LoadConfigFromDB(); err != nil {
		s.log.Warn("从数据库加载微信配置失败，使用 YAML 降级配置", "error", err)
	}

	var app *entity.WxAppConfig
	var err error
	if appKey != "" {
		app, err = s.channelRepo.GetAppByKey(ctx, appKey)
		if err != nil {
			return nil, fmt.Errorf("微信应用配置不存在(app_key=%s)，请在管理平台配置", appKey)
		}
		if !app.IsEnabled() {
			return nil, fmt.Errorf("微信应用 %s 已禁用", appKey)
		}
	} else {
		// 兼容：未传 appKey 时回退到全局配置（历史单应用模式）
		return s.resolveGlobalChannel(mchID)
	}

	// 定位商户：优先显式 mchID（须与应用关联），否则用商户关联中标记的默认商户
	// （默认商户唯一数据源：wx_app_merchants.is_default，管理端「商户关联」中设置）
	var merchant *entity.WxMerchantConfig
	if mchID != "" {
		merchant, err = s.channelRepo.GetMerchantByMchID(ctx, mchID)
		if err != nil {
			return nil, fmt.Errorf("商户配置不存在(mch_id=%s)，请在管理平台配置", mchID)
		}
		if _, err := s.channelRepo.GetAppMerchant(ctx, app.ID, merchant.ID); err != nil {
			return nil, fmt.Errorf("应用 %s 未关联商户 %s，请在管理平台配置关联", app.AppKey, mchID)
		}
	} else {
		link, lerr := s.channelRepo.GetDefaultAppMerchant(ctx, app.ID)
		if lerr != nil {
			return nil, fmt.Errorf("应用 %s 未设置默认商户，请在管理平台「商户关联」中绑定商户并设为默认", app.AppKey)
		}
		merchant, err = s.channelRepo.GetMerchantByID(ctx, link.MerchantConfigID)
		if err != nil {
			return nil, fmt.Errorf("应用默认商户不存在(merchant_id=%d)", link.MerchantConfigID)
		}
	}
	if !merchant.IsEnabled() {
		return nil, fmt.Errorf("商户 %s 已禁用", merchant.MchID)
	}

	client, err := s.GetClientForMerchant(merchant)
	if err != nil {
		return nil, err
	}

	// 支付回调地址：商户配置优先，全局兜底
	notifyURL := merchant.PayNotifyURL
	if notifyURL == "" {
		notifyURL = s.getConfig().NotifyURL
	}
	return &PayChannel{App: app, Merchant: merchant, Client: client, PayNotifyURL: notifyURL}, nil
}

// resolveGlobalChannel 兼容路径：未传 appKey 时按全局配置解析（单应用历史模式）
func (s *WechatPayService) resolveGlobalChannel(mchID string) (*PayChannel, error) {
	cfg := s.getConfig()
	if mchID != "" && cfg.MCHID != "" && mchID != cfg.MCHID {
		return nil, fmt.Errorf("全局模式仅支持商户 %s，无法使用商户 %s（请配置多渠道应用）", cfg.MCHID, mchID)
	}
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	app := &entity.WxAppConfig{AppKey: "", AppType: entity.WxAppTypeMP, AppID: cfg.AppID, FrontendURL: cfg.FrontendURL, OauthCallbackURL: cfg.OAuthCallbackURL}
	merchant := &entity.WxMerchantConfig{MchID: cfg.MCHID}
	return &PayChannel{App: app, Merchant: merchant, Client: client, PayNotifyURL: cfg.NotifyURL}, nil
}

// GetClientForMerchant 获取指定商户的 V3 客户端（带缓存池，配置变更自动重建）
func (s *WechatPayService) GetClientForMerchant(mch *entity.WxMerchantConfig) (*wxpayv3.Client, error) {
	if mch.MchID == "" || mch.MchSerialNo == "" {
		return nil, fmt.Errorf("商户 %s 配置不完整（需要 mch_id/mch_serial_no）", mch.MchID)
	}
	if mch.PrivateKey == "" {
		return nil, fmt.Errorf("商户 %s 未配置私钥（请在管理平台录入商户私钥 PEM 内容）", mch.MchID)
	}

	fingerprint := s.merchantFingerprint(mch)

	s.mu.RLock()
	client, cached := s.merchantClients[mch.ID]
	currentKey := s.merchantKeys[mch.ID]
	s.mu.RUnlock()
	if cached && currentKey == fingerprint {
		return client, nil
	}

	privKey, err := wxpay_utility.LoadPrivateKey(mch.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("商户 %s 私钥加载失败: %w", mch.MchID, err)
	}

	clientCfg := wxpayv3.ClientConfig{
		MchID:      mch.MchID,
		SerialNo:   mch.MchSerialNo,
		APIv3Key:   mch.Apiv3Key,
		PrivateKey: privKey,
		IsPartner:  s.getConfig().Mode == "partner",
	}
	// 微信支付公钥（应答/回调验签与医保敏感字段加密；公钥明文存储）
	if mch.PublicKeyID != "" && mch.PublicKey != "" {
		pubKey, err := wxpay_utility.LoadPublicKey(mch.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("商户 %s 微信支付公钥加载失败: %w", mch.MchID, err)
		}
		clientCfg.PublicKeyID = mch.PublicKeyID
		clientCfg.PublicKey = pubKey
	} else {
		s.log.Warn("商户未配置微信支付公钥，将跳过应答验签", "mch_id", mch.MchID)
	}

	newClient := wxpayv3.NewClient(clientCfg)
	s.mu.Lock()
	s.merchantClients[mch.ID] = newClient
	s.merchantKeys[mch.ID] = fingerprint
	s.mu.Unlock()
	return newClient, nil
}

// merchantFingerprint 商户客户端配置指纹
func (s *WechatPayService) merchantFingerprint(mch *entity.WxMerchantConfig) string {
	pemFp := func(pem string) string {
		if pem == "" {
			return ""
		}
		head, tail := pem, pem
		if len(head) > 16 {
			head = head[:16]
		}
		if len(tail) > 16 {
			tail = tail[len(tail)-16:]
		}
		return fmt.Sprintf("%d:%s..%s", len(pem), head, tail)
	}
	return strings.Join([]string{mch.MchID, mch.MchSerialNo, mch.Apiv3Key, pemFp(mch.PrivateKey), mch.PublicKeyID, pemFp(mch.PublicKey)}, "|")
}

// FindMerchantByPublicKeyID 按微信支付公钥ID定位商户（回调验签路由：回调头 Wechatpay-Serial 即公钥ID）
// ctx 携带租户上下文时在对应租户库内查找；无租户上下文时查主库（默认租户存量模式）
func (s *WechatPayService) FindMerchantByPublicKeyID(ctx context.Context, publicKeyID string) (*entity.WxMerchantConfig, error) {
	merchants, err := s.channelRepo.ListEnabledMerchants(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range merchants {
		if m.PublicKeyID != "" && m.PublicKeyID == publicKeyID {
			return m, nil
		}
	}
	return nil, fmt.Errorf("未找到公钥ID=%s 的商户配置", publicKeyID)
}

// ---------- 支付（V3 JSAPI） ----------

// JSAPIParams 前端调起支付参数（小程序 wx.requestPayment 与公众号 H5 通用）
type JSAPIParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// JSAPIPrepay JSAPI 下单（H5 公众号与小程序通用，openid 须与 appid 匹配）
// 返回 prepay_id
func (s *WechatPayService) JSAPIPrepay(appID, openID, orderID, description string, amount int64) (string, error) {
	cfg := s.getConfig()
	if cfg.NotifyURL == "" {
		return "", fmt.Errorf("微信支付回调地址未配置（wechat_notify_url）")
	}
	if amount <= 0 {
		return "", fmt.Errorf("订单金额必须大于0(单位:分)")
	}

	client, err := s.getClient()
	if err != nil {
		return "", err
	}

	prepayID, err := client.JSAPIPrepay(appID, openID, orderID, description, cfg.NotifyURL, amount)
	if err != nil {
		s.log.Error("JSAPI下单失败", "appid", appID, "out_trade_no", orderID, "error", err)
		return "", fmt.Errorf("JSAPI下单失败: %w", err)
	}

	s.log.Info("JSAPI下单成功", "appid", appID, "out_trade_no", orderID, "prepay_id", prepayID)
	return prepayID, nil
}

// BuildJSAPIPayParams 根据 prepay_id 生成前端调起支付参数（RSA 签名）
// 小程序 wx.requestPayment 与公众号 H5（chooseWXPay/WeixinJSBridge）通用
func (s *WechatPayService) BuildJSAPIPayParams(prepayID, appID string) (*JSAPIParams, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	params, err := client.BuildJSAPIPayParams(prepayID, appID)
	if err != nil {
		return nil, err
	}
	return &JSAPIParams{
		AppID:     params.AppID,
		TimeStamp: params.TimeStamp,
		NonceStr:  params.NonceStr,
		Package:   params.Package,
		SignType:  params.SignType,
		PaySign:   params.PaySign,
	}, nil
}

// QueryOrderByOutTradeNo 商户订单号查询微信侧订单
func (s *WechatPayService) QueryOrderByOutTradeNo(outTradeNo string) (*wxpayv3.QueryOrderResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	return client.QueryOrderByOutTradeNo(outTradeNo)
}

// CloseOrder 关闭微信侧订单
func (s *WechatPayService) CloseOrder(outTradeNo string) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	return client.CloseOrder(outTradeNo)
}

// ParseNotify 验签并解密 V3 支付回调通知，返回通知与解密后的业务 JSON
// 回调租户定位（设计文档 01 §九）：先按回调头 Wechatpay-Serial 查主库商户映射得租户，
// 再到租户库取商户客户端验签解密；未登记映射时回退全局客户端（兼容存量单商户模式）
func (s *WechatPayService) ParseNotify(ctx context.Context, header http.Header, body []byte) (*wxpay_utility.Notification, []byte, error) {
	if mch := s.findNotifyMerchant(header.Get("Wechatpay-Serial")); mch != nil {
		client, cerr := s.GetClientForMerchant(mch)
		if cerr == nil {
			return client.ParseNotification(header, body)
		}
		s.log.Warn("回调商户客户端构建失败，回退全局客户端验签", "mch_id", mch.MchID, "error", cerr)
	}
	client, err := s.getClient()
	if err != nil {
		return nil, nil, err
	}
	return client.ParseNotification(header, body)
}

// findNotifyMerchant 按回调头 Wechatpay-Serial 经主库映射定位商户归属租户，返回该租户库的商户配置
// 未登记映射返回 nil（调用方回退全局客户端）
func (s *WechatPayService) findNotifyMerchant(publicKeyID string) *entity.WxMerchantConfig {
	if s.tenantRepo == nil || publicKeyID == "" {
		return nil
	}
	tenantID, err := s.tenantRepo.GetTenantByPublicKeyID(publicKeyID)
	if err != nil || tenantID <= 0 {
		return nil
	}
	tctx := database.WithTenantID(context.Background(), tenantID)
	mch, err := s.FindMerchantByPublicKeyID(tctx, publicKeyID)
	if err != nil {
		s.log.Warn("回调租户库未找到匹配商户，回退全局客户端验签", "serial", publicKeyID, "tenant_id", tenantID, "error", err)
		return nil
	}
	return mch
}

// ---------- 公众号平台接口（服务号能力） ----------

// GetStableAccessToken 获取服务号 access_token（stable_token 普通模式，带缓存）
// 缓存优先（Redis→进程内），有效期前 5 分钟过期；forceRefresh 为 true 时强刷
func (s *WechatPayService) GetStableAccessToken(forceRefresh bool) (string, error) {
	cfg := s.getConfig()
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return "", fmt.Errorf("未配置公众号 AppID/AppSecret，无法获取 access_token")
	}

	cacheKey := "wx:mp:stable_token:" + cfg.AppID

	// 快路径：缓存有效直接返回
	if !forceRefresh {
		if v := s.getCachedString(cacheKey); v != "" {
			return v, nil
		}
		if s.memToken != "" && time.Now().Before(s.memTokenExp) {
			return s.memToken, nil
		}
	}

	reqBody := map[string]interface{}{
		"grant_type": "client_credential",
		"appid":      cfg.AppID,
		"secret":     cfg.AppSecret,
	}
	if forceRefresh {
		reqBody["force_refresh"] = true
	}

	// 微信 stable_token 接口应答结构（access_token/expires_in/errcode 为微信协议字段名，豁免 camelCase 规范）
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := s.postJSON("/cgi-bin/stable_token", reqBody, &out); err != nil {
		return "", fmt.Errorf("获取access_token失败: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("获取access_token失败: errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}

	// 有效期前 5 分钟过期，至少缓存 60 秒
	ttl := time.Duration(out.ExpiresIn-300) * time.Second
	if ttl < 60*time.Second {
		ttl = 60 * time.Second
	}
	s.setCachedString(cacheKey, out.AccessToken, ttl)

	s.memMu.Lock()
	s.memToken = out.AccessToken
	s.memTokenExp = time.Now().Add(ttl)
	s.memMu.Unlock()

	return out.AccessToken, nil
}

// GetJSAPITicket 获取 jsapi_ticket（带缓存，有效期前 5 分钟过期）
func (s *WechatPayService) GetJSAPITicket() (string, error) {
	cfg := s.getConfig()
	cacheKey := "wx:mp:jsapi_ticket:" + cfg.AppID

	if v := s.getCachedString(cacheKey); v != "" {
		return v, nil
	}
	if s.memTicket != "" && time.Now().Before(s.memTicketExp) {
		return s.memTicket, nil
	}

	token, err := s.GetStableAccessToken(false)
	if err != nil {
		return "", err
	}

	// 微信 getticket 接口应答结构（ticket/expires_in/errcode 为微信协议字段名，豁免 camelCase 规范）
	var out struct {
		Ticket    string `json:"ticket"`
		ExpiresIn int64  `json:"expires_in"`
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
	}
	apiURL := fmt.Sprintf("%s/cgi-bin/ticket/getticket?access_token=%s&type=jsapi", wechatPlatformAPIBase, url.QueryEscape(token))
	if err := s.getJSON(apiURL, &out); err != nil {
		return "", fmt.Errorf("获取jsapi_ticket失败: %w", err)
	}
	if out.Ticket == "" {
		return "", fmt.Errorf("获取jsapi_ticket失败: errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}

	ttl := time.Duration(out.ExpiresIn-300) * time.Second
	if ttl < 60*time.Second {
		ttl = 60 * time.Second
	}
	s.setCachedString(cacheKey, out.Ticket, ttl)

	s.memMu.Lock()
	s.memTicket = out.Ticket
	s.memTicketExp = time.Now().Add(ttl)
	s.memMu.Unlock()

	return out.Ticket, nil
}

// JSAPIConfig JS-SDK 配置（wx.config 用）
type JSAPIConfig struct {
	AppID     string `json:"appId"`
	Timestamp string `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Signature string `json:"signature"`
}

// GetJSAPIConfig 获取微信 JS-SDK 配置（wx.config）
// pageURL 为当前网页 URL（不含 # 及其后的部分，由前端传入）
func (s *WechatPayService) GetJSAPIConfig(pageURL string) (*JSAPIConfig, error) {
	cfg := s.getConfig()
	if cfg.AppID == "" {
		return nil, fmt.Errorf("微信 AppID 未配置")
	}
	if pageURL == "" {
		return nil, fmt.Errorf("网页 URL 不能为空")
	}
	ticket, err := s.GetJSAPITicket()
	if err != nil {
		return nil, err
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := generateNonceStr()

	// JS-SDK 签名算法：sha1("jsapi_ticket=...&noncestr=...&timestamp=...&url=...")
	signStr := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%s&url=%s", ticket, nonceStr, timestamp, pageURL)
	hash := sha1.Sum([]byte(signStr))

	return &JSAPIConfig{
		AppID:     cfg.AppID,
		Timestamp: timestamp,
		NonceStr:  nonceStr,
		Signature: hex.EncodeToString(hash[:]),
	}, nil
}

// ---------- 平台 HTTP 帮助方法 ----------

// postJSON 调用微信公众平台 POST JSON 接口
func (s *WechatPayService) postJSON(path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := s.http.Post(wechatPlatformAPIBase+path, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}
	return nil
}

// getJSON 调用微信公众平台 GET 接口
func (s *WechatPayService) getJSON(apiURL string, out any) error {
	resp, err := s.http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}
	return nil
}

// getCachedString 读取缓存（未配置缓存或未命中返回 ""）
func (s *WechatPayService) getCachedString(key string) string {
	if s.cache == nil {
		return ""
	}
	v, err := s.cache.Get(key)
	if err != nil || v == "" {
		return ""
	}
	return v
}

// setCachedString 写入缓存（失败不阻断）
func (s *WechatPayService) setCachedString(key, value string, ttl time.Duration) {
	if s.cache == nil {
		return
	}
	if err := s.cache.Set(key, value, ttl); err != nil {
		s.log.Warn("写入缓存失败", "key", key, "error", err)
	}
}

// ---------- 配置访问器 ----------

// GetMCHID 获取商户号
func (s *WechatPayService) GetMCHID() string {
	return s.getConfig().MCHID
}

// GetAppID 获取微信公众号 AppID
func (s *WechatPayService) GetAppID() string {
	return s.getConfig().AppID
}

// GetMiniAppID 获取微信小程序 AppID
func (s *WechatPayService) GetMiniAppID() string {
	return s.getConfig().MiniAppID
}

// GetConfig 返回当前有效配置（供认证服务读取 OAuth 回调/前端地址等）
func (s *WechatPayService) GetConfig() config.WechatConfig {
	return s.getConfig()
}

// GetV3Client 获取 V3 协议客户端（供支付/医保/退款/账单服务直接调用协议层，与 Gpay s.cfg.Client 等价）
func (s *WechatPayService) GetV3Client() (*wxpayv3.Client, error) {
	return s.getClient()
}

// IsPartner 是否服务商模式
func (s *WechatPayService) IsPartner() bool {
	return s.getConfig().Mode == "partner"
}

// GetRefundNotifyURL 退款结果回调地址
func (s *WechatPayService) GetRefundNotifyURL() string {
	return s.getConfig().RefundNotifyURL
}

// GetMedNotifyURL 医保收款回调地址
func (s *WechatPayService) GetMedNotifyURL() string {
	return s.getConfig().MedNotifyURL
}
