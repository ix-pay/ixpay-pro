package response

// 微信多渠道配置响应（管理平台端）

// WxAppConfigResponse 微信应用配置响应（敏感字段脱敏）
// 所有 ID 字段使用 int64 格式，通过 json:",string" 标签自动序列化为字符串
// 默认商户字段（只读）来源于商户关联表 is_default 标记，服务端组装返回
type WxAppConfigResponse struct {
	ID                    int64  `json:"id,string"`
	AppKey                string `json:"appKey"` // 前端实例标志（约定命名，豁免 camelCase 例外规则中的保留项）
	AppType               string `json:"appType"`
	AppID                 string `json:"appId"`
	AppSecret             string `json:"appSecret"` // 脱敏：非空返回 ******
	DefaultMerchantID     int64  `json:"defaultMerchantId,string"`       // 默认商户配置ID（0=未设置）
	DefaultMerchantMchID  string `json:"defaultMerchantMchId,omitempty"` // 默认商户号（展示用）
	DefaultMerchantName   string `json:"defaultMerchantName,omitempty"`  // 默认商户名称（展示用）
	AuthScope             string `json:"authScope"`
	OauthCallbackURL      string `json:"oauthCallbackUrl"`
	FrontendURL           string `json:"frontendUrl"`
	Description           string `json:"description"`
	Status                int    `json:"status"`
	AuthURL               string `json:"authUrl,omitempty"` // 服务号授权登录入口链接
	CreatedAt             string `json:"createdAt,omitempty"`
}

// WxMerchantConfigResponse 微信商户配置响应（敏感字段脱敏）
type WxMerchantConfigResponse struct {
	ID              int64  `json:"id,string"`
	MchID           string `json:"mchId"`
	MchName         string `json:"mchName"`
	MchSerialNo     string `json:"mchSerialNo"`
	Apiv3Key        string `json:"apiv3Key"`   // 脱敏
	PrivateKey      string `json:"privateKey"` // 脱敏
	PublicKeyID     string `json:"publicKeyId"`
	PublicKey       string `json:"publicKey"`
	PayNotifyURL    string `json:"payNotifyUrl"`
	RefundNotifyURL string `json:"refundNotifyUrl"`
	Status          int    `json:"status"`
	CreatedAt       string `json:"createdAt,omitempty"`
}

// WxAppMerchantResponse 应用-商户关联响应
type WxAppMerchantResponse struct {
	ID               int64  `json:"id,string"`
	AppConfigID      int64  `json:"appConfigId,string"`
	MerchantConfigID int64  `json:"merchantConfigId,string"`
	MchID            string `json:"mchId,omitempty"`
	MchName          string `json:"mchName,omitempty"`
	IsDefault        bool   `json:"isDefault"`
}

// WxPublicAppResponse 公开应用配置响应（前端空壳启动时按 appKey 加载，无敏感信息）
type WxPublicAppResponse struct {
	AppKey      string                 `json:"appKey"`
	AppType     string                 `json:"appType"` // mp | miniapp
	AppID       string                 `json:"appId"`   // 前端做 wx.config 等需要
	Description string                 `json:"description"`
	AuthURL     string                 `json:"authUrl,omitempty"` // 服务号授权登录入口（后端生成）
	UI          map[string]interface{} `json:"ui,omitempty"`      // 租户个性化 UI 配置（菜单/图片/主题）
	Status      int                    `json:"status"`
}

// WxReconSummaryResponse 对账汇总响应
type WxReconSummaryResponse struct {
	TotalCount    int64   `json:"totalCount"`    // 支付单总数
	SuccessCount  int64   `json:"successCount"`  // 支付成功笔数
	SuccessAmount float64 `json:"successAmount"` // 支付成功金额(元)
	RefundCount   int64   `json:"refundCount"`   // 已退款笔数
	RefundAmount  float64 `json:"refundAmount"`  // 累计退款金额(元)
	PendingCount  int64   `json:"pendingCount"`  // 待支付笔数
}
