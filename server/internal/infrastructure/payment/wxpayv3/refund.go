package wxpayv3

import (
	"fmt"
)

// 退款协议层（与 Gpay wechatpay/refund.go 对齐）

// CreateRefund 退款申请: POST /v3/refund/domestic/refunds
// 直连与服务商URL相同，服务商模式自动附加sub_mchid
// 返回成功仅表示退款单已受理，最终结果以退款回调/查询单笔退款为准
func (c *Client) CreateRefund(req *RefundCreateRequest) (*RefundResponse, error) {
	if c.IsPartner() {
		req.SubMchID = c.cfg.SubMchID
	}
	var resp RefundResponse
	if err := c.doRequest("POST", "/v3/refund/domestic/refunds", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryRefund 查询单笔退款(商户退款单号): GET /v3/refund/domestic/refunds/{out_refund_no}
// 服务商模式附加query参数sub_mchid
func (c *Client) QueryRefund(outRefundNo string) (*RefundResponse, error) {
	path := fmt.Sprintf("/v3/refund/domestic/refunds/%s", outRefundNo)
	if c.IsPartner() {
		path += fmt.Sprintf("?sub_mchid=%s", c.cfg.SubMchID)
	}
	var resp RefundResponse
	if err := c.doRequest("GET", path, nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApplyAbnormalRefund 发起异常退款: POST /v3/refund/domestic/refunds/{refund_id}/apply-abnormal-refund
// 退款状态为ABNORMAL时调用；退用户银行卡时bank_account/real_name须先经EncryptOAEP加密
func (c *Client) ApplyAbnormalRefund(refundID string, req *AbnormalRefundRequest) error {
	headers := map[string]string{}
	// 含加密敏感字段时需携带微信支付公钥ID
	if req.BankAccount != "" || req.RealName != "" {
		if c.cfg.PublicKey == nil {
			return fmt.Errorf("退款至用户银行卡需配置微信支付公钥以加密敏感字段")
		}
		headers[HeaderSerial] = c.cfg.PublicKeyID
	}
	if c.IsPartner() {
		req.SubMchID = c.cfg.SubMchID
	}
	return c.doRequest("POST", fmt.Sprintf("/v3/refund/domestic/refunds/%s/apply-abnormal-refund", refundID), req, nil, headers)
}
