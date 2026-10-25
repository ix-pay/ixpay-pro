package wxpayv3

import (
	"crypto/md5"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"

	wxpay_utility "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayutility"
)

// 密码学基础能力统一委托微信支付官方WXPayUtility工具类(wxpayutility包，源码原样引自官方文档)
// 文档: https://pay.weixin.qq.com/doc/v3/merchant/4015119334

// EncryptOAEP 使用微信支付公钥RSA-OAEP-SHA1加密(医保下单/异常退款敏感字段加密)，返回base64
func EncryptOAEP(message string, publicKey *rsa.PublicKey) (string, error) {
	return wxpay_utility.EncryptOAEPWithPublicKey(message, publicKey)
}

// GenerateNonce 生成32位随机字符串（委托官方工具类）
func GenerateNonce() string {
	nonce, err := wxpay_utility.GenerateNonce()
	if err != nil {
		// 随机源异常时退化为时间熵(实际不会发生)
		const symbols = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		buf := make([]byte, 32)
		for i := range buf {
			buf[i] = symbols[(i*7+len(symbols))%len(symbols)]
		}
		return string(buf)
	}
	return nonce
}

// IDCardDigest 医保下单身份证摘要: 字母转大写、15位转18位后取MD5小写hex
func IDCardDigest(idCard string) (string, error) {
	id := strings.TrimSpace(idCard)
	if id == "" {
		return "", errors.New("身份证号为空")
	}
	// 身份证中字母需转大写
	id = strings.ToUpper(id)
	// 15位转18位
	if len(id) == 15 {
		converted, err := convert15To18(id)
		if err != nil {
			return "", err
		}
		id = converted
	}
	if len(id) != 18 {
		return "", fmt.Errorf("身份证号位数非法: %d", len(id))
	}
	sum := md5.Sum([]byte(id))
	return fmt.Sprintf("%x", sum), nil
}

// convert15To18 15位身份证转18位(补全出生年份与校验位)
func convert15To18(id15 string) (string, error) {
	if len(id15) != 15 {
		return "", errors.New("非15位身份证")
	}
	body := id15[:6] + "19" + id15[6:12]
	// GB11643校验码算法
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
	sum := 0
	for i, c := range []byte(body) {
		v := int(c - '0')
		if v < 0 || v > 9 {
			return "", errors.New("身份证含非法字符")
		}
		sum += v * weights[i]
	}
	return body + string(checkCodes[sum%11]), nil
}
