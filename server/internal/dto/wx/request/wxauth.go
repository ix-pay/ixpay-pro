package request

// WechatLoginRequest 微信登录请求参数（服务号网页授权 code）
// AppKey 字段名沿用前端实例标志约定（wxapp/miniapp 以 app_key 传参），豁免 camelCase 规则
type WechatLoginRequest struct {
	Code   string `json:"code" binding:"required"`
	AppKey string `json:"app_key"` // 微信应用实例标志（多服务号渠道，空则用全局配置）
}

// WechatMiniappLoginRequest 微信小程序登录请求参数（wx.login 获取的 js_code）
type WechatMiniappLoginRequest struct {
	Code   string `json:"code" binding:"required"`
	AppKey string `json:"app_key"` // 微信应用实例标志（多小程序渠道，空则用全局配置）
}

// WechatRefreshTokenRequest 刷新令牌请求参数
type WechatRefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
