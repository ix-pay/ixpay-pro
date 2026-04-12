package captcha

import (
	"errors"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/redis"
	"github.com/mojocn/base64Captcha"
)

// captcha包提供验证码生成和验证功能
// 使用base64Captcha库生成验证码图片，并通过Redis存储验证码信息
// 支持配置验证码长度、过期时间和是否开启验证码功能

// SetupCaptcha 初始化验证码服务
// 根据配置创建验证码实例，设置验证码长度、过期时间和开关状态
// 参数:
// - cfg: 应用配置对象，包含验证码相关配置
// - redis: Redis客户端，用于存储验证码
// 返回:
// - Captcha指针: 验证码服务实例
// - error: 错误信息
func SetupCaptcha(cfg *config.Config, redis *redis.RedisClient) (*Captcha, error) {
	// 获取验证码长度，默认4位
	len := cfg.Server.CaptchaLen
	if len <= 0 {
		len = 4
	}

	// 获取验证码过期时间，默认60秒
	expiry := cfg.Server.CaptchaTimeOut
	if expiry <= 0 {
		expiry = 60
	}
	// 获取验证码开关状态
	open := cfg.Server.OpenCaptcha

	return new(len, expiry, open, redis)
}

// new 创建验证码实例（内部函数）
// 配置验证码驱动、存储和相关参数
// 参数:
// - len: 验证码长度
// - expiry: 验证码过期时间（秒）
// - open: 是否开启验证码功能
// - redis: Redis客户端，用于存储验证码
// 返回:
// - Captcha指针: 验证码服务实例
// - error: 错误信息
func new(len int, expiry int, open bool, redis *redis.RedisClient) (*Captcha, error) {
	// 创建数字验证码驱动配置
	driver := base64Captcha.NewDriverDigit(
		80,  // 验证码高度
		240, // 验证码宽度
		len, 0.7, 80)

	// 设置验证码在Redis中的有效期
	redis.Expiration = time.Duration(expiry) * time.Second
	// 设置Redis中的验证码键前缀
	redis.PreKey = "captcha:"

	// 使用驱动和存储创建验证码实例
	captcha := base64Captcha.NewCaptcha(driver, redis)

	// 创建并返回Captcha结构体实例
	return &Captcha{
			captcha:       captcha,
			captchaLen:    len,
			captchaExpiry: expiry,
			openCaptcha:   open,
		},
		nil
}

// Captcha 验证码服务结构体
// 封装验证码生成和验证功能，包含验证码相关配置
// 字段:
// - captcha: base64Captcha库的验证码实例
// - captchaLen: 验证码长度
// - captchaExpiry: 验证码过期时间（秒）
// - openCaptcha: 验证码功能开关状态
type Captcha struct {
	captcha       *base64Captcha.Captcha // base64Captcha验证码实例
	captchaLen    int                    // 验证码长度
	captchaExpiry int                    // 验证码过期时间（秒）
	openCaptcha   bool                   // 验证码功能开关状态
}

// NewCaptcha 使用第三方库生成验证码
// 入参: len - 验证码长度
// 返回: 验证码文本, base64编码的图片, 错误信息
func (c *Captcha) NewCaptcha() (string, string, error) {

	// 生成验证码
	id, b64s, _, err := c.captcha.Generate()
	if err != nil {
		return "", "", err
	}

	// 库已经返回了完整的base64图片数据，直接返回
	return id, b64s, nil
}

// VerifyCaptcha 校验验证码是否正确
// 入参: captchaId - 验证码ID, captcha - 用户输入的验证码
// 返回: 校验结果, 错误信息
func (c *Captcha) VerifyCaptcha(captchaId string, captcha string) (bool, error) {
	// 检查参数
	if captchaId == "" || captcha == "" {
		return false, errors.New("验证码ID或验证码为空")
	}

	return c.captcha.Verify(captchaId, captcha, true), nil
}

// GetCaptchaLen 获取验证码长度配置
// 返回: 验证码长度
func (c *Captcha) GetCaptchaLen() int {
	return c.captchaLen
}

// GetCaptchaExpiry 获取验证码过期时间配置
// 返回: 验证码过期时间（秒）
func (c *Captcha) GetCaptchaExpiry() int {
	return c.captchaExpiry
}

// IsOpenCaptcha 检查验证码功能是否开启
// 返回: 验证码功能开关状态
func (c *Captcha) IsOpenCaptcha() bool {
	return c.openCaptcha
}
