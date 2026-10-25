package wxpayv3

import (
	"fmt"
)

// 医保自费混合收款协议层（与 Gpay wechatpay/medical.go 对齐）

// MedInsPrepay 医保自费混合收款下单: POST /v3/med-ins/orders
// 敏感字段(payer/relative的姓名与身份证摘要)需先经 BuildMedInsPayer 加密
func (c *Client) MedInsPrepay(req *MedInsPrepayRequest) (*MedInsOrderResponse, error) {
	if c.cfg.PublicKey == nil {
		return nil, fmt.Errorf("医保下单需配置微信支付公钥(wechat.public_key_id / wechat.public_key_path)")
	}
	var resp MedInsOrderResponse
	// Wechatpay-Serial: 敏感字段使用微信支付公钥加密，须传微信支付公钥ID
	headers := map[string]string{HeaderSerial: c.cfg.PublicKeyID}
	if err := c.doRequest("POST", "/v3/med-ins/orders", req, &resp, headers); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MedInsQueryByMixTradeNo 按医保自费混合订单号查单: GET /v3/med-ins/orders/mix-trade-no/{mix_trade_no}
// 服务商模式须携带sub_mchid查询参数
func (c *Client) MedInsQueryByMixTradeNo(mixTradeNo string) (*MedInsOrderResponse, error) {
	path := fmt.Sprintf("/v3/med-ins/orders/mix-trade-no/%s", mixTradeNo)
	if c.IsPartner() {
		path += "?sub_mchid=" + c.cfg.SubMchID
	}
	var resp MedInsOrderResponse
	if err := c.doRequest("GET", path, nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MedInsQueryByOutTradeNo 按商户订单号查医保订单: GET /v3/med-ins/orders/out-trade-no/{out_trade_no}
// 服务商模式须携带sub_mchid查询参数
func (c *Client) MedInsQueryByOutTradeNo(outTradeNo string) (*MedInsOrderResponse, error) {
	path := fmt.Sprintf("/v3/med-ins/orders/out-trade-no/%s", outTradeNo)
	if c.IsPartner() {
		path += "?sub_mchid=" + c.cfg.SubMchID
	}
	var resp MedInsOrderResponse
	if err := c.doRequest("GET", path, nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MedInsRefundNotify 医保退款通知: POST /v3/med-ins/refunds/notify?mix_trade_no={mix_trade_no}
// 直连与服务商同构；仅退医保部分(统筹/个账/其他)，现金部分走 CreateRefund 普通退款接口
func (c *Client) MedInsRefundNotify(mixTradeNo string, req *MedInsRefundNotifyRequest) error {
	if mixTradeNo == "" {
		return fmt.Errorf("缺少医保混合订单号(mix_trade_no)，无法发起医保退款")
	}
	if c.IsPartner() && req.SubMchID == "" {
		req.SubMchID = c.cfg.SubMchID
	}
	path := fmt.Sprintf("/v3/med-ins/refunds/notify?mix_trade_no=%s", mixTradeNo)
	return c.doRequest("POST", path, req, nil, nil)
}

// BuildMedInsPayer 加密医保支付人身份信息
// 姓名: RSA-OAEP加密(官方工具类); 身份证: MD5摘要(字母大写、15转18位)后RSA-OAEP加密
func (c *Client) BuildMedInsPayer(name, idCard string) (*MedInsPayer, error) {
	if c.cfg.PublicKey == nil {
		return nil, fmt.Errorf("医保支付需配置微信支付公钥")
	}
	if name == "" || idCard == "" {
		return nil, fmt.Errorf("支付人姓名与身份证号不能为空")
	}

	encName, err := EncryptOAEP(name, c.cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("姓名加密失败: %w", err)
	}

	digest, err := IDCardDigest(idCard)
	if err != nil {
		return nil, fmt.Errorf("身份证摘要计算失败: %w", err)
	}
	encDigest, err := EncryptOAEP(digest, c.cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("身份证摘要加密失败: %w", err)
	}

	return &MedInsPayer{
		Name:     encName,
		IDDigest: encDigest,
		CardType: "ID_CARD",
	}, nil
}
