package wxpayv3

import (
	"fmt"
	"net/url"
)

// 账单协议层（与 Gpay wechatpay/bill.go 对齐）

// TradeBill 申请交易账单: GET /v3/bill/tradebill
// bill_date为三个月内且非当天的日期；应答返回下载链接(5分钟内有效)与SHA1摘要
// 服务商模式附加sp_mchid/sub_mchid(账单范围限定为该子商户)
func (c *Client) TradeBill(req *TradeBillRequest) (*BillResponse, error) {
	if req.BillType == "" {
		req.BillType = "ALL"
	}
	path := fmt.Sprintf("/v3/bill/tradebill?bill_date=%s&bill_type=%s",
		url.QueryEscape(req.BillDate), url.QueryEscape(req.BillType))
	if c.IsPartner() {
		path += fmt.Sprintf("&sp_mchid=%s&sub_mchid=%s", c.cfg.SpMchID, c.cfg.SubMchID)
	}
	var resp BillResponse
	if err := c.doRequest("GET", path, nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FundFlowBill 申请资金账单: GET /v3/bill/fundflowbill
// account_type: BASIC基本账户/OPERATION运营账户/FEES手续费账户
// 服务商模式资金账单主体为服务商账户，附加sp_mchid
func (c *Client) FundFlowBill(req *FundFlowBillRequest) (*BillResponse, error) {
	if req.AccountType == "" {
		req.AccountType = "BASIC"
	}
	path := fmt.Sprintf("/v3/bill/fundflowbill?bill_date=%s&account_type=%s",
		url.QueryEscape(req.BillDate), url.QueryEscape(req.AccountType))
	if c.IsPartner() {
		path += fmt.Sprintf("&sp_mchid=%s", c.cfg.SpMchID)
	}
	var resp BillResponse
	if err := c.doRequest("GET", path, nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DownloadBill 下载账单: GET download_url(申请账单应答返回)
// 需按APIv3标准签名请求，浏览器无法直接打开；应答为账单数据本身(超2MB为gzip压缩)
// 完整性校验由上层使用申请账单应答的hash_value完成
func (c *Client) DownloadBill(downloadURL string) ([]byte, error) {
	_, body, err := c.doHTTP("GET", downloadURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("下载账单失败: %w", err)
	}
	return body, nil
}
