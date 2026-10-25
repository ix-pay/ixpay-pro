package wxpayv3

import (
	"fmt"
)

// JSAPI 合单支付协议层
// 官方文档: https://pay.weixin.qq.com/doc/v3/merchant/4012556926
// 一笔合单支付订单支持 2-10 笔商品单交易（普通商户模式）

// CombineAmount 商品单金额信息
type CombineAmount struct {
	TotalAmount int64  `json:"total_amount"` // 标价金额(分)
	Currency    string `json:"currency"`     // 标价币种，境内固定 CNY
}

// SettleInfo 结算信息
type SettleInfo struct {
	ProfitSharing bool `json:"profit_sharing,omitempty"` // 分账标识
}

// CombineSubOrder 合单商品单信息
type CombineSubOrder struct {
	Mchid       string         `json:"mchid"`        // 商品单商户号(须与 combine_appid 绑定)
	Attach      string         `json:"attach"`       // 商户数据包(合单下单必填)
	Amount      *CombineAmount `json:"amount"`       // 商品单金额
	OutTradeNo  string         `json:"out_trade_no"` // 商品单商户订单号(同合单商户号下唯一)
	Description string         `json:"description"`  // 商品描述
	SettleInfo  *SettleInfo    `json:"settle_info,omitempty"`
	GoodsTag    string         `json:"goods_tag,omitempty"` // 订单优惠标记
}

// CombineSceneInfo 场景信息
type CombineSceneInfo struct {
	DeviceID      string `json:"device_id,omitempty"` // 商户端设备号(选填)
	PayerClientIP string `json:"payer_client_ip"`     // 用户终端IP(scene_info 传入时必填)
}

// CombinePayerInfo 合单支付者信息
type CombinePayerInfo struct {
	Openid string `json:"openid"` // 用户在 combine_appid 下的唯一标识
}

// CombinePrepayRequest JSAPI 合单下单请求 POST /v3/combine-transactions/jsapi
type CombinePrepayRequest struct {
	CombineAppid      string             `json:"combine_appid"`         // 合单发起方 APPID(公众号类型)
	CombineMchid      string             `json:"combine_mchid"`         // 合单发起方商户号
	CombineOutTradeNo string             `json:"combine_out_trade_no"`  // 合单商户订单号(6-32位，同合单商户号下唯一)
	CombinePayerInfo  *CombinePayerInfo  `json:"combine_payer_info"`    // 支付者信息
	SubOrders         []*CombineSubOrder `json:"sub_orders"`            // 商品单列表(2-10笔)
	TimeExpire        string             `json:"time_expire,omitempty"` // rfc3339，7天以内
	NotifyURL         string             `json:"notify_url"`            // 合单支付成功回调地址
	SceneInfo         *CombineSceneInfo  `json:"scene_info,omitempty"`
}

// combinePrepayResponse 合单下单应答
type combinePrepayResponse struct {
	PrepayID string `json:"prepay_id"`
}

// CombineJSAPIPrepay JSAPI 合单下单
// 返回 prepay_id（调起支付参数与普通 JSAPI 一致，经 BuildJSAPIPayParams 签名）
func (c *Client) CombineJSAPIPrepay(req *CombinePrepayRequest) (string, error) {
	if len(req.SubOrders) < 2 || len(req.SubOrders) > 10 {
		return "", fmt.Errorf("合单支付商品单数量须为2-10笔，当前: %d", len(req.SubOrders))
	}
	var resp combinePrepayResponse
	if err := c.doRequest("POST", "/v3/combine-transactions/jsapi", req, &resp, nil); err != nil {
		return "", err
	}
	if resp.PrepayID == "" {
		return "", fmt.Errorf("合单下单失败: 应答缺少prepay_id")
	}
	return resp.PrepayID, nil
}

// CombineSubOrderResult 合单回调/查单中的子单结果
type CombineSubOrderResult struct {
	Mchid         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeType     string `json:"trade_type"`
	TradeState    string `json:"trade_state"`
	Attach        string `json:"attach,omitempty"`
	SuccessTime   string `json:"success_time,omitempty"`
	Amount        *struct {
		TotalAmount int64  `json:"total_amount"`
		Currency    string `json:"currency"`
		PayerAmount int64  `json:"payer_amount,omitempty"`
	} `json:"amount,omitempty"`
}

// CombineNotifyResult 合单支付成功回调解密后业务数据
// 官方文档: https://pay.weixin.qq.com/doc/v3/merchant/4013421231
type CombineNotifyResult struct {
	CombineAppid      string                   `json:"combine_appid"`
	CombineMchid      string                   `json:"combine_mchid"`
	CombineOutTradeNo string                   `json:"combine_out_trade_no"`
	SubOrders         []*CombineSubOrderResult `json:"sub_orders"`
}
