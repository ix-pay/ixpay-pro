package captcha

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/redis"
	"github.com/mojocn/base64Captcha"
)

// NewCaptcha 使用第三方库生成验证码
// 入参: len - 验证码长度
// 返回: 验证码文本, base64编码的图片, 错误信息
func NewCaptcha(len int, expiry int, redis *redis.RedisClient) (string, string, error) {
	// 创建配置对象
	driver := base64Captcha.NewDriverDigit(
		80,  // 验证码高度
		240, // 验证码宽度
		len, 0.7, 80)

	// 设置验证码有效期
	redis.Expiration = time.Duration(expiry) * time.Second
	// 设置验证码前缀
	redis.PreKey = "captcha:"

	// 创建验证码实例
	captcha := base64Captcha.NewCaptcha(driver, redis)

	// 生成验证码
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		return "", "", err
	}

	// 库已经返回了完整的base64图片数据，直接返回
	return id, b64s, nil
}
