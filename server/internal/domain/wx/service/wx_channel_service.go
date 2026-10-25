package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	baseEntity "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// WxChannelService 微信多渠道配置服务（管理平台端：应用/商户配置、授权入口、对账）
// 查询/写入方法携带 ctx：管理端带 ?tenant_id= 时由 AdminTenantAuth 注入租户上下文，
// 仓储层据此路由到对应租户库（单租户切换模式，见设计文档 01 §3.4/D4/D5）
type WxChannelService struct {
	channelRepo repo.WxChannelRepository
	paymentRepo repo.PaymentRepository
	authService *WXAuthService
	tenantRepo  baseRepo.TenantRepository // base_tenant_apps 路由（公开配置接口按 app_key 定位租户）
	log         logger.Logger

	// 小程序 access_token 进程内缓存（按 appid 隔离，供小程序码生成使用）
	miniappTokenMu sync.Mutex
	miniappTokens  map[string]miniappTokenMem
}

// NewWxChannelService 创建微信多渠道配置服务
func NewWxChannelService(channelRepo repo.WxChannelRepository, paymentRepo repo.PaymentRepository, authService *WXAuthService, tenantRepo baseRepo.TenantRepository, log logger.Logger) *WxChannelService {
	return &WxChannelService{
		channelRepo:   channelRepo,
		paymentRepo:   paymentRepo,
		authService:   authService,
		tenantRepo:    tenantRepo,
		log:           log,
		miniappTokens: make(map[string]miniappTokenMem),
	}
}

// maskSecret 敏感字段脱敏展示（非空返回 ******）
func maskSecret(v string) string {
	if v == "" {
		return ""
	}
	return "******"
}

// keepSecret 更新时保留原敏感值（入参为空或 "******" 表示不修改）
func keepSecret(newVal, oldVal string) string {
	if newVal == "" || newVal == "******" {
		return oldVal
	}
	return newVal
}

// withAppTenant 按 app_key 查路由表并注入租户上下文（公开配置接口在主库完成路由，见设计文档 01 §3.1）
// 未绑定路由 = 默认租户（主库），原 ctx 原样返回
func (s *WxChannelService) withAppTenant(ctx context.Context, appKey string) context.Context {
	if ctx == nil || appKey == "" || s.tenantRepo == nil {
		return ctx
	}
	if route, err := s.tenantRepo.GetAppRoute(appKey); err == nil && route != nil {
		return database.WithTenantID(ctx, route.TenantID)
	}
	return ctx
}

// ---------- 应用配置 ----------

// ListApps 应用配置列表（敏感字段脱敏，附带授权入口链接）
func (s *WxChannelService) ListApps(ctx context.Context, page, pageSize int) ([]*entity.WxAppConfig, int, error) {
	return s.channelRepo.ListApps(ctx, page, pageSize)
}

// GetApp 应用配置详情
func (s *WxChannelService) GetApp(ctx context.Context, id int64) (*entity.WxAppConfig, error) {
	return s.channelRepo.GetAppByID(ctx, id)
}

// CreateApp 创建应用配置
func (s *WxChannelService) CreateApp(ctx context.Context, app *entity.WxAppConfig) error {
	if app.AppKey == "" || app.AppID == "" || app.AppSecret == "" {
		return errors.New("app_key/app_id/app_secret 不能为空")
	}
	if app.AppType != entity.WxAppTypeMP && app.AppType != entity.WxAppTypeMiniapp {
		return errors.New("app_type 仅支持 mp 或 miniapp")
	}
	app.Status = 1
	if err := s.channelRepo.CreateApp(ctx, app); err != nil {
		return fmt.Errorf("创建应用配置失败: %w", err)
	}
	// 自动注册主库路由（app_key → 当前租户）：简化接入流程，
	// 租户详情页"绑定路由"仅作手动修正/补绑入口
	s.bindAppRouteQuietly(ctx, app.AppKey)
	s.log.Info("创建微信应用配置", "app_key", app.AppKey, "app_id", app.AppID)
	return nil
}

// bindAppRouteQuietly 创建应用后自动注册主库路由（app_key → 当前租户）
// 未携带租户上下文视为默认租户（主库，tenant_id=1）；失败不阻断应用创建，
// 留日志并可在租户详情页手动绑定兜底
func (s *WxChannelService) bindAppRouteQuietly(ctx context.Context, appKey string) {
	if s.tenantRepo == nil || appKey == "" {
		return
	}
	tenantID := int64(1)
	if tid, ok := database.TenantIDFromContext(ctx); ok && tid > 0 {
		tenantID = tid
	}
	if route, err := s.tenantRepo.GetAppRoute(appKey); err == nil && route != nil {
		if route.TenantID != tenantID {
			s.log.Warn("app_key 已绑定其他租户，未自动改绑（如需迁移请先在原租户详情页解绑）",
				"app_key", appKey, "bound_tenant", route.TenantID, "current_tenant", tenantID)
		}
		return
	}
	if err := s.tenantRepo.CreateAppRoute(&baseEntity.TenantApp{AppKey: appKey, TenantID: tenantID, ModuleCode: "wx"}); err != nil {
		s.log.Warn("自动绑定应用路由失败，可在租户详情页手动绑定", "app_key", appKey, "tenant_id", tenantID, "error", err)
		return
	}
	s.log.Info("应用路由已自动绑定", "app_key", appKey, "tenant_id", tenantID)
}

// UpdateApp 更新应用配置（app_secret 传空或 ****** 表示保留原值）
func (s *WxChannelService) UpdateApp(ctx context.Context, id int64, app *entity.WxAppConfig) error {
	old, err := s.channelRepo.GetAppByID(ctx, id)
	if err != nil {
		return fmt.Errorf("应用配置不存在: %w", err)
	}
	old.AppType = app.AppType
	old.AppID = app.AppID
	old.AppSecret = keepSecret(app.AppSecret, old.AppSecret)
	old.AuthScope = app.AuthScope
	old.OauthCallbackURL = app.OauthCallbackURL
	old.FrontendURL = app.FrontendURL
	old.Description = app.Description
	if app.Status != 0 {
		old.Status = app.Status
	}
	return s.channelRepo.UpdateApp(ctx, old)
}

// DeleteApp 删除应用配置（同时清理关联）
func (s *WxChannelService) DeleteApp(ctx context.Context, id int64) error {
	app, err := s.channelRepo.GetAppByID(ctx, id)
	if err != nil {
		return err
	}
	links, err := s.channelRepo.ListAppMerchants(ctx, id)
	if err == nil {
		for _, l := range links {
			_ = s.channelRepo.DeleteAppMerchant(ctx, l.ID)
		}
	}
	// 级联解绑主库应用路由：app_key 全局唯一，应用删除后路由残留会永久占用该
	// app_key（他租户重建同名应用后绑定路由报"已绑定"），且公开配置会按残留路由
	// 定位到旧租户库查不到应用；路由表在主库，与租户上下文无关
	if s.tenantRepo != nil {
		if route, rerr := s.tenantRepo.GetAppRoute(app.AppKey); rerr == nil && route != nil {
			_ = s.tenantRepo.DeleteAppRoute(route.ID)
		}
	}
	return s.channelRepo.DeleteApp(ctx, id)
}

// ---------- 商户配置 ----------

// ListMerchants 商户配置列表（敏感字段脱敏）
func (s *WxChannelService) ListMerchants(ctx context.Context, page, pageSize int) ([]*entity.WxMerchantConfig, int, error) {
	return s.channelRepo.ListMerchants(ctx, page, pageSize)
}

// GetMerchant 商户配置详情
func (s *WxChannelService) GetMerchant(ctx context.Context, id int64) (*entity.WxMerchantConfig, error) {
	return s.channelRepo.GetMerchantByID(ctx, id)
}

// registerMchRef 登记商户号→租户映射（支付回调按 Wechatpay-Serial 定位租户库商户，见设计文档 01 §九）
// 未携带租户上下文视为默认租户（主库，tenant_id=1）
func (s *WxChannelService) registerMchRef(ctx context.Context, mchID, publicKeyID string) {
	if s.tenantRepo == nil || mchID == "" {
		return
	}
	tenantID := int64(1)
	if tid, ok := database.TenantIDFromContext(ctx); ok && tid > 0 {
		tenantID = tid
	}
	_ = s.tenantRepo.UpsertMchRef(mchID, publicKeyID, tenantID)
}

// CreateMerchant 创建商户配置
func (s *WxChannelService) CreateMerchant(ctx context.Context, mch *entity.WxMerchantConfig) error {
	if mch.MchID == "" || mch.MchSerialNo == "" || mch.Apiv3Key == "" || mch.PrivateKey == "" {
		return errors.New("mch_id/mch_serial_no/apiv3_key/private_key 不能为空")
	}
	mch.Status = 1
	if err := s.channelRepo.CreateMerchant(ctx, mch); err != nil {
		return fmt.Errorf("创建商户配置失败: %w", err)
	}
	s.registerMchRef(ctx, mch.MchID, mch.PublicKeyID)
	s.log.Info("创建微信商户配置", "mch_id", mch.MchID)
	return nil
}

// UpdateMerchant 更新商户配置（敏感字段传空或 ****** 表示保留原值）
func (s *WxChannelService) UpdateMerchant(ctx context.Context, id int64, mch *entity.WxMerchantConfig) error {
	old, err := s.channelRepo.GetMerchantByID(ctx, id)
	if err != nil {
		return fmt.Errorf("商户配置不存在: %w", err)
	}
	old.MchID = mch.MchID
	old.MchName = mch.MchName
	old.MchSerialNo = mch.MchSerialNo
	old.Apiv3Key = keepSecret(mch.Apiv3Key, old.Apiv3Key)
	old.PrivateKey = keepSecret(mch.PrivateKey, old.PrivateKey)
	old.PublicKeyID = keepSecret(mch.PublicKeyID, old.PublicKeyID)
	old.PublicKey = keepSecret(mch.PublicKey, old.PublicKey)
	old.PayNotifyURL = mch.PayNotifyURL
	old.RefundNotifyURL = mch.RefundNotifyURL
	if mch.Status != 0 {
		old.Status = mch.Status
	}
	if err := s.channelRepo.UpdateMerchant(ctx, old); err != nil {
		return fmt.Errorf("更新商户配置失败: %w", err)
	}
	// 商户号可能变更：清旧映射、登记新映射
	if old.MchID != mch.MchID && old.MchID != "" {
		_ = s.tenantRepo.DeleteMchRef(old.MchID)
	}
	s.registerMchRef(ctx, old.MchID, old.PublicKeyID)
	return nil
}

// DeleteMerchant 删除商户配置
func (s *WxChannelService) DeleteMerchant(ctx context.Context, id int64) error {
	mch, err := s.channelRepo.GetMerchantByID(ctx, id)
	if err != nil {
		return fmt.Errorf("商户配置不存在: %w", err)
	}
	if err := s.channelRepo.DeleteMerchant(ctx, id); err != nil {
		return err
	}
	_ = s.tenantRepo.DeleteMchRef(mch.MchID)
	return nil
}

// ---------- 应用-商户关联 ----------

// ListAppMerchants 应用已关联商户列表
func (s *WxChannelService) ListAppMerchants(ctx context.Context, appConfigID int64) ([]*entity.WxAppMerchant, []*entity.WxMerchantConfig, error) {
	links, err := s.channelRepo.ListAppMerchants(ctx, appConfigID)
	if err != nil {
		return nil, nil, err
	}
	merchants := make([]*entity.WxMerchantConfig, 0, len(links))
	for _, l := range links {
		if m, err := s.channelRepo.GetMerchantByID(ctx, l.MerchantConfigID); err == nil {
			merchants = append(merchants, m)
		}
	}
	return links, merchants, nil
}

// BindAppMerchant 应用关联商户（可设为默认）
// 默认商户唯一数据源为本关联表 is_default 标记：设为默认时自动清除该应用的其他默认标记；
// 首个关联且应用当前无默认时自动设为默认（保证应用始终有默认商户，支付路由不中断）
func (s *WxChannelService) BindAppMerchant(ctx context.Context, appConfigID, merchantID int64, isDefault bool) error {
	if _, err := s.channelRepo.GetAppByID(ctx, appConfigID); err != nil {
		return fmt.Errorf("应用配置不存在: %w", err)
	}
	mch, err := s.channelRepo.GetMerchantByID(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("商户配置不存在: %w", err)
	}
	if !mch.IsEnabled() {
		return fmt.Errorf("商户 %s 已禁用，无法关联", mch.MchID)
	}

	// 新建关联且应用当前无任何默认时，自动设为默认（兜底，避免应用失去默认商户）
	if !isDefault {
		if _, derr := s.channelRepo.GetDefaultAppMerchant(ctx, appConfigID); derr != nil {
			isDefault = true
		}
	}

	// 设为默认时先清除该应用的其他默认标记
	if isDefault {
		links, _ := s.channelRepo.ListAppMerchants(ctx, appConfigID)
		for _, l := range links {
			if l.IsDefault {
				l.IsDefault = false
				_ = s.channelRepo.UpdateAppMerchant(ctx, l)
			}
		}
	}

	link, err := s.channelRepo.GetAppMerchant(ctx, appConfigID, merchantID)
	if err != nil {
		link = &entity.WxAppMerchant{AppConfigID: appConfigID, MerchantConfigID: merchantID, IsDefault: isDefault}
		return s.channelRepo.CreateAppMerchant(ctx, link)
	}
	link.IsDefault = isDefault
	return s.channelRepo.UpdateAppMerchant(ctx, link)
}

// UnbindAppMerchant 解除应用-商户关联
func (s *WxChannelService) UnbindAppMerchant(ctx context.Context, linkID int64) error {
	return s.channelRepo.DeleteAppMerchant(ctx, linkID)
}

// reassignDefault 解除默认商户关联后的自动补位：将剩余关联中最早创建的商户接任默认
// （保证应用始终有默认商户，下单未指定商户号时支付路由不中断）
func (s *WxChannelService) reassignDefault(ctx context.Context, appConfigID int64) {
	links, err := s.channelRepo.ListAppMerchants(ctx, appConfigID)
	if err != nil || len(links) == 0 {
		return
	}
	// ListAppMerchants 按 id ASC 排序，links[0] 即最早关联（默认关联已被删除，剩余均为非默认）
	links[0].IsDefault = true
	_ = s.channelRepo.UpdateAppMerchant(ctx, links[0])
}

// UnbindAppMerchantByIDs 按应用ID+商户ID解除关联
func (s *WxChannelService) UnbindAppMerchantByIDs(ctx context.Context, appConfigID, merchantConfigID int64) error {
	link, err := s.channelRepo.GetAppMerchant(ctx, appConfigID, merchantConfigID)
	if err != nil {
		return fmt.Errorf("关联不存在: %w", err)
	}
	if err := s.channelRepo.DeleteAppMerchant(ctx, link.ID); err != nil {
		return err
	}
	// 解除的是默认商户时自动补位
	if link.IsDefault {
		s.reassignDefault(ctx, appConfigID)
	}
	return nil
}

// MapDefaultMerchants 批量查询应用的默认商户（应用ID → 商户配置，无默认的应用不在映射中）
// 管理端应用列表「默认商户」列展示用，批量查询避免逐应用 N+1
func (s *WxChannelService) MapDefaultMerchants(ctx context.Context, appConfigIDs []int64) map[int64]*entity.WxMerchantConfig {
	out := make(map[int64]*entity.WxMerchantConfig, len(appConfigIDs))
	if len(appConfigIDs) == 0 {
		return out
	}
	links, err := s.channelRepo.ListDefaultAppMerchants(ctx, appConfigIDs)
	if err != nil {
		return out
	}
	mchCache := make(map[int64]*entity.WxMerchantConfig)
	for _, l := range links {
		m, ok := mchCache[l.MerchantConfigID]
		if !ok {
			m, err = s.channelRepo.GetMerchantByID(ctx, l.MerchantConfigID)
			if err != nil {
				continue // 商户已被删除的悬挂关联跳过
			}
			mchCache[l.MerchantConfigID] = m
		}
		out[l.AppConfigID] = m
	}
	return out
}

// GetAuthURL 生成应用授权登录入口链接（服务号类型）
// ctx 透传租户上下文：admin 端（渠道配置页）在所选租户库查应用配置；
// 公开配置/登录链路 ctx 无租户上下文时由 authService 按路由表定位
func (s *WxChannelService) GetAuthURL(ctx context.Context, appKey, redirectURI string) (string, error) {
	return s.authService.GetOAuthURL(ctx, redirectURI, appKey)
}

// wxMiniappQrcodeAPI 小程序码获取接口（getwxacodeunlimit：数量不限，支持携带 scene 参数）
const wxMiniappQrcodeAPI = "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s"

// GetMiniappEntryQrcode 生成小程序入口码 PNG（小程序类型应用专属入口：
// 服务端按应用 appid/secret 获取 access_token 后调用微信 getwxacodeunlimit，
// scene 携带 app_key，小程序端 onLoad 解析后完成实例归属识别）
func (s *WxChannelService) GetMiniappEntryQrcode(ctx context.Context, appID int64) ([]byte, string, error) {
	_ = ctx
	app, err := s.channelRepo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, "", fmt.Errorf("应用配置不存在: %w", err)
	}
	if app.AppType != entity.WxAppTypeMiniapp {
		return nil, "", errors.New("仅小程序类型应用支持生成小程序码")
	}
	if app.AppID == "" || app.AppSecret == "" {
		return nil, "", errors.New("该应用未配置 AppID/AppSecret，无法生成小程序码")
	}

	// 获取小程序接口调用凭证 access_token（带缓存，按 appid 隔离）
	token, err := s.getMiniappAccessToken(app.AppID, app.AppSecret)
	if err != nil {
		return nil, "", err
	}

	// 调用微信 getwxacodeunlimit 生成小程序码（scene 携带 app_key，进入首页默认支付页）
	apiURL := fmt.Sprintf(wxMiniappQrcodeAPI, token)
	body := map[string]interface{}{
		"scene":       "app_key=" + app.AppKey,
		"page":        "pages/payment/index",
		"check_path":  false,
		"env_version": "release",
		"width":       430,
	}
	bodyJSON, _ := json.Marshal(body)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(apiURL, "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, "", fmt.Errorf("请求微信小程序码接口失败: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取小程序码响应失败: %w", err)
	}

	// 微信错误时返回 JSON（errcode/errmsg），成功时返回图片二进制
	// 注意：成功时 Content-Type 可能是 image/png 或 image/jpeg，不能只按 PNG 魔数(0x89)判断，
	// 这里按 Content-Type(image/*) 与魔数（PNG/JPEG/GIF）双通道识别，JSON 结构（{ 或 [开头）才走错误解析
	contentType := resp.Header.Get("Content-Type")
	isImage := strings.Contains(contentType, "image/") ||
		(len(raw) >= 4 && raw[0] == 0x89 && raw[1] == 0x50 && raw[2] == 0x4E && raw[3] == 0x47) || // PNG 魔数
		(len(raw) >= 3 && raw[0] == 0xFF && raw[1] == 0xD8 && raw[2] == 0xFF) || // JPEG 魔数
		(len(raw) >= 4 && raw[0] == 'G' && raw[1] == 'I' && raw[2] == 'F' && raw[3] == '8') // GIF 魔数
	isJSONBody := len(raw) > 0 && (raw[0] == '{' || raw[0] == '[')
	if !isImage || strings.Contains(contentType, "application/json") || isJSONBody {
		var wxErr struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if json.Unmarshal(raw, &wxErr) == nil && wxErr.ErrCode != 0 {
			if wxErr.ErrCode == 41030 {
				return nil, "", errors.New("小程序默认页面路径不存在（pages/payment/index），请确认小程序已发布该页面或检查 env_version")
			}
			return nil, "", fmt.Errorf("生成小程序码失败: errcode=%d errmsg=%s", wxErr.ErrCode, wxErr.ErrMsg)
		}
		// 非图片且非约定错误结构：附响应体前缀便于排障
		preview := string(raw)
		if len(preview) > 120 {
			preview = preview[:120]
		}
		return nil, "", fmt.Errorf("生成小程序码失败：微信返回非图片内容(content-type=%s, body=%q)", contentType, preview)
	}
	return raw, app.AppKey, nil
}

// miniappTokenMem 小程序 access_token 进程内缓存（按 appid 隔离）
type miniappTokenMem struct {
	token string
	exp   time.Time
}

// getMiniappAccessToken 获取小程序 access_token（stable_token 普通模式，进程内缓存按 appid 隔离）
func (s *WxChannelService) getMiniappAccessToken(appID, appSecret string) (string, error) {
	// 快路径：进程内缓存有效直接返回（多实例部署时依赖微信侧 token 并行安全，stable_token 普通模式允许多 token 共存）
	var mem miniappTokenMem
	s.miniappTokenMu.Lock()
	mem = s.miniappTokens[appID]
	s.miniappTokenMu.Unlock()
	if mem.token != "" && time.Now().Before(mem.exp) {
		return mem.token, nil
	}

	reqBody := map[string]interface{}{
		"grant_type": "client_credential",
		"appid":      appID,
		"secret":     appSecret,
	}
	reqJSON, _ := json.Marshal(reqBody)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("https://api.weixin.qq.com/cgi-bin/stable_token", "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		return "", fmt.Errorf("获取小程序 access_token 失败: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 access_token 响应失败: %w", err)
	}
	// 微信 stable_token 接口应答结构（access_token/expires_in/errcode 为微信协议字段名，豁免 camelCase 规范）
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("解析 access_token 响应失败: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("获取小程序 access_token 失败: errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}

	// 有效期前 5 分钟过期，至少缓存 60 秒；进程内缓存兜底（无 Redis 场景）
	ttl := time.Duration(out.ExpiresIn-300) * time.Second
	if ttl < 60*time.Second {
		ttl = 60 * time.Second
	}
	s.miniappTokenMu.Lock()
	s.miniappTokens[appID] = miniappTokenMem{token: out.AccessToken, exp: time.Now().Add(ttl)}
	s.miniappTokenMu.Unlock()
	return out.AccessToken, nil
}

// ListPayments 对账支付单列表（filters 原样透传给仓储；ctx 决定查询哪个租户库）
func (s *WxChannelService) ListPayments(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*entity.Payment, int, error) {
	return s.paymentRepo.List(ctx, page, pageSize, filters)
}

// GetPublicApp 公开应用配置（前端空壳启动加载，脱敏）
// 先查 base_tenant_apps 定位租户并注入 ctx，渠道配置随 ctx 路由（主库存路由，租户库存明细）
func (s *WxChannelService) GetPublicApp(ctx context.Context, appKey string) (*entity.WxAppConfig, error) {
	app, err := s.channelRepo.GetAppByKey(s.withAppTenant(ctx, appKey), appKey)
	if err != nil {
		return nil, fmt.Errorf("微信应用配置不存在(app_key=%s)", appKey)
	}
	if !app.IsEnabled() {
		return nil, fmt.Errorf("微信应用 %s 已禁用", appKey)
	}
	return app, nil
}

// GetPublicAppUI 公开应用 UI 配置（前端空壳启动加载，返回解析后的 JSON 对象；未配置返回 nil）
func (s *WxChannelService) GetPublicAppUI(ctx context.Context, app *entity.WxAppConfig) (map[string]interface{}, error) {
	cfg, err := s.channelRepo.GetUIByAppConfigID(ctx, app.ID)
	if err != nil {
		return nil, nil // 未配置 UI 返回 nil，前端使用默认渲染
	}
	if cfg.Status != 1 || cfg.UIJSON == "" {
		return nil, nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(cfg.UIJSON), &out); err != nil {
		s.log.Error("UI 配置 JSON 解析失败（配置非法，返回空）", "app_config_id", app.ID, "error", err)
		return nil, nil
	}
	return out, nil
}

// GetAppUIRaw 管理端读取 UI 配置原文（未配置返回空串）
func (s *WxChannelService) GetAppUIRaw(ctx context.Context, appConfigID int64) (string, error) {
	cfg, err := s.channelRepo.GetUIByAppConfigID(ctx, appConfigID)
	if err != nil {
		return "", nil
	}
	return cfg.UIJSON, nil
}

// SaveAppUI 管理端保存 UI 配置（校验 JSON 可解析）
func (s *WxChannelService) SaveAppUI(ctx context.Context, appConfigID int64, uiJSON string) error {
	if _, err := s.channelRepo.GetAppByID(ctx, appConfigID); err != nil {
		return fmt.Errorf("应用配置不存在: %w", err)
	}
	if uiJSON == "" {
		return errors.New("UI 配置内容不能为空")
	}
	var check map[string]interface{}
	if err := json.Unmarshal([]byte(uiJSON), &check); err != nil {
		return fmt.Errorf("UI 配置不是合法 JSON: %w", err)
	}
	return s.channelRepo.SaveUIConfig(ctx, &entity.TenantUIConfig{
		AppConfigID: appConfigID,
		UIJSON:      uiJSON,
		Status:      1,
	})
}

// ---------- 对账 ----------

// ReconSummary 对账汇总统计
// 领域层值对象：不带 json 标签，接口响应由 API Handler 映射到 dto/wx/response.WxReconSummaryResponse
type ReconSummary struct {
	TotalCount    int64
	SuccessCount  int64
	SuccessAmount float64
	RefundCount   int64
	RefundAmount  float64
	PendingCount  int64
}

// GetReconSummary 对账汇总（时间范围 + 可选状态；ctx 决定查询哪个租户库）
func (s *WxChannelService) GetReconSummary(ctx context.Context, start, end time.Time) (*ReconSummary, error) {
	summary, err := s.paymentRepo.GetReconSummary(ctx, start, end)
	if err != nil {
		return nil, err
	}
	return &ReconSummary{
		TotalCount:    summary.TotalCount,
		SuccessCount:  summary.SuccessCount,
		SuccessAmount: float64(summary.SuccessAmount) / 100.0,
		RefundCount:   summary.RefundCount,
		RefundAmount:  float64(summary.RefundAmount) / 100.0,
		PendingCount:  summary.PendingCount,
	}, nil
}

// ParseDateRange 解析对账时间范围（默认近 7 天）
func ParseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()
	if startDate != "" {
		t, err := time.ParseInLocation("2006-01-02", startDate, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("start_date 格式应为 yyyy-MM-dd")
		}
		start = t
	}
	if endDate != "" {
		t, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("end_date 格式应为 yyyy-MM-dd")
		}
		end = t.Add(24*time.Hour - time.Second)
	}
	return start, end, nil
}
