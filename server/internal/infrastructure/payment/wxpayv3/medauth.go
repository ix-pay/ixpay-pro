package wxpayv3

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	wxpay_utility "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayutility"
)

// 医保免密授权(页面授权)与授权查询协议层（与 Gpay wechatpay/auth.go 对齐）

// 医保免密授权域名与授权查询接口
const (
	medAuthHostSelfTest   = "https://mitest.wecity.qq.com" // 本人支付授权(测试)
	medAuthHostFamilyTest = "https://exp.wecity.qq.com"    // 代亲人支付授权(测试)
	medAuthHostProd       = "https://card.wecity.qq.com"   // 本人/代亲人支付授权(正式)

	medUserQueryHostTest = "https://test-receiver.wecity.qq.com" // 授权查询(测试)
	medUserQueryHostProd = "https://receiver.wecity.qq.com"      // 授权查询(正式)
)

// MedAuthLinkInput 医保支付授权链接拼接参数
type MedAuthLinkInput struct {
	AppID           string // 公众号appid
	CityCode        string // 城市编码
	Channel         string // 渠道号(微信医保平台分配)
	OrgChnlCrtfCodg string // 机构渠道认证编码(反馈单)
	OrgCodg         string // 定点医药机构编码(反馈单/机构编码)
	OrgAppID        string // 定点医药机构应用ID(反馈单)
	RedirectURL     string // 回跳地址(需外网可访问)
	FamilyID        string // 代亲人支付: MD5(亲属姓名+身份证后4位)，为空表示本人支付
	TestEnv         bool
}

// BuildMedAuthLink 拼接医保支付授权链接(免密授权页)
// 用户在微信内打开后完成授权，平台回跳: REDIRECT_URL/?retCode=RETCODE&authCode=AUTHCODE
// 代亲人支付时附加familyId参数并使用测试域名exp(正式域名与本人相同)
func BuildMedAuthLink(in *MedAuthLinkInput) string {
	host := medAuthHostProd
	if in.TestEnv {
		host = medAuthHostSelfTest
		if in.FamilyID != "" {
			host = medAuthHostFamilyTest
		}
	}
	q := url.Values{}
	q.Set("authType", "2")
	q.Set("isDepart", "2")
	q.Set("appid", in.AppID)
	q.Set("cityCode", in.CityCode)
	q.Set("channel", in.Channel)
	if in.FamilyID != "" {
		q.Set("familyId", in.FamilyID)
	}
	q.Set("orgChnlCrtfCodg", in.OrgChnlCrtfCodg)
	q.Set("orgCodg", in.OrgCodg)
	q.Set("bizType", "04107") // 医保移动支付
	q.Set("orgAppId", in.OrgAppID)
	q.Set("redirectUrl", in.RedirectURL)
	return host + "/oauth/code?" + q.Encode()
}

// FamilyIDDigest 亲情授权familyId: MD5(就诊亲属姓名+身份证后4位)
// 32位小写hex，utf-8编码(身份证末位x须大写)
func FamilyIDDigest(name, idCard string) string {
	last4 := ""
	if len(idCard) >= 4 {
		last4 = idCard[len(idCard)-4:]
	}
	sum := md5.Sum([]byte(name + strings.ToUpper(last4)))
	return hex.EncodeToString(sum[:])
}

// BuildMedExtends 亲情移动支付下单extends字段
// {"rel_user_name_md5":"亲属姓名MD5","rel_user_card_no_md5":"亲属身份证MD5"}
// MD5为32位小写、utf-8编码(身份证字母转大写后计算)
func BuildMedExtends(relName, relIDCard string) string {
	nameSum := md5.Sum([]byte(relName))
	cardSum := md5.Sum([]byte(strings.ToUpper(relIDCard)))
	b, _ := json.Marshal(map[string]string{
		"rel_user_name_md5":    hex.EncodeToString(nameSum[:]),
		"rel_user_card_no_md5": hex.EncodeToString(cardSum[:]),
	})
	return string(b)
}

// MedUserQuery 调用wecity授权查询接口，用授权码换取payAuthNo与经纬度
// POST {host}/api/mipuserquery/userQuery/{partnerId}
// 签名: HmacSHA256(partnerSecret, partnerId+timestamp) 请求头god-portal-*
func (c *Client) MedUserQuery(partnerID, partnerSecret string, testEnv bool, authCode, openid string) (*MedUserQueryResponse, error) {
	host := medUserQueryHostProd
	if testEnv {
		host = medUserQueryHostTest
	}
	apiURL := fmt.Sprintf("%s/api/mipuserquery/userQuery/%s", host, partnerID)

	body, _ := json.Marshal(map[string]string{"qrcode": authCode, "openid": openid})
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	mac := hmac.New(sha256.New, []byte(partnerSecret))
	mac.Write([]byte(partnerID + timestamp))
	signature := hex.EncodeToString(mac.Sum(nil))
	nonce, err := wxpay_utility.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("生成请求ID失败: %w", err)
	}
	requestID := strings.ReplaceAll(nonce, "-", "")

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("构造授权查询请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("god-portal-timestamp", timestamp)
	req.Header.Set("god-portal-signature", signature)
	req.Header.Set("god-portal-request-id", requestID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求授权查询接口失败: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取授权查询应答失败: %w", err)
	}

	var out MedUserQueryResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("解析授权查询应答失败: %w (body: %s)", err, string(respBody))
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("授权查询失败[%d]: %s", out.Code, out.Message)
	}
	return &out, nil
}
