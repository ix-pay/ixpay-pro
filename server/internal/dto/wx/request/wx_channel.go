package request

// 微信多渠道配置管理请求参数（管理平台端）

// WxAppConfigRequest 微信应用配置（创建/更新）
// 所有 ID 字段使用 int64 格式，通过 json:",string" 标签与前端字符串 ID 交互
// 默认商户不在本表单设置：唯一入口为「商户关联」接口的 is_default 标记（单一数据源）
type WxAppConfigRequest struct {
	AppKey            string `json:"appKey" binding:"required,max=50"` // 前端实例标志
	AppType           string `json:"appType" binding:"required,oneof=mp miniapp"`
	AppID             string `json:"appId" binding:"required,max=50"`
	AppSecret         string `json:"appSecret"` // 更新时传 "******" 或空表示保留原值
	AuthScope         string `json:"authScope" binding:"omitempty,oneof=snsapi_base snsapi_userinfo"`
	OauthCallbackURL  string `json:"oauthCallbackUrl"`
	FrontendURL       string `json:"frontendUrl"`
	Description       string `json:"description" binding:"max=255"`
	Status            *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// WxMerchantConfigRequest 微信商户配置（创建/更新）
type WxMerchantConfigRequest struct {
	MchID           string `json:"mchId" binding:"required,max=32"`
	MchName         string `json:"mchName" binding:"max=100"`
	MchSerialNo     string `json:"mchSerialNo" binding:"max=64"`
	Apiv3Key        string `json:"apiv3Key"`   // 更新时传 "******" 或空表示保留原值
	PrivateKey      string `json:"privateKey"` // 商户私钥 PEM；更新时传 "******" 或空表示保留原值
	PublicKeyID     string `json:"publicKeyId" binding:"max=64"`
	PublicKey       string `json:"publicKey"` // 微信支付公钥 PEM
	PayNotifyURL    string `json:"payNotifyUrl"`
	RefundNotifyURL string `json:"refundNotifyUrl"`
	Status          *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// WxAppMerchantRequest 应用-商户关联请求
type WxAppMerchantRequest struct {
	MerchantID int64 `json:"merchantId,string" binding:"required"` // 商户配置 ID
	IsDefault  bool  `json:"isDefault"`
}

// WxReconQueryRequest 对账查询请求参数
type WxReconQueryRequest struct {
	StartDate string `json:"startDate" form:"startDate"` // yyyy-MM-dd
	EndDate   string `json:"endDate" form:"endDate"`     // yyyy-MM-dd
	Status    string `json:"status" form:"status"`       // pending/success/failed/refunded/cancelled
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
}

// SaveAppUIRequest 保存租户 UI 配置请求参数
type SaveAppUIRequest struct {
	UIJSON string `json:"uiJson" binding:"required"` // UI 配置 JSON 字符串（schema 见 specs/02）
}
