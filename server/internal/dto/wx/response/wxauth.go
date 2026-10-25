package response

// WXUserInfoResponse 用户信息响应
type WXUserInfoResponse struct {
	ID       int64  `json:"id,string"` // 用户 ID（int64 + ,string 避免前端精度丢失）
	Username string `json:"userName"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	Status   int    `json:"status"`
}

// WXLoginResponse 登录响应
type WXLoginResponse struct {
	User         WXUserInfoResponse `json:"user"`
	AccessToken  string             `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
	AccessExpire int64              `json:"accessExpire"`
}

// WechatAuthResult 微信授权结果
// 注意：本结构体用于解析微信官方 API 应答（openid/unionid/access_token 等为微信协议字段名），
// 豁免 camelCase 规范；不应直接作为接口响应返回
type WechatAuthResult struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// WechatUserInfo 微信用户信息（解析微信官方 API 应答，字段名为微信协议约定，豁免 camelCase 规范）
type WechatUserInfo struct {
	OpenID    string `json:"openid"`
	UnionID   string `json:"unionid"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"headimgurl"`
	Sex       int    `json:"sex"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Language  string `json:"language"`
}
