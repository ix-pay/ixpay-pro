package request

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username  string `json:"userName" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CaptchaId string `json:"captchaId" binding:"required_if=OpenCaptcha true"`
	Captcha   string `json:"captcha" binding:"required_if=OpenCaptcha true"`
}

// RefreshTokenRequest 刷新令牌请求参数
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// JsonInBlacklistRequest JWT 加入黑名单请求参数
type JsonInBlacklistRequest struct {
	Token string `json:"token" binding:"required"` // JWT token
}
