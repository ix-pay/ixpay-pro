package entity

import "time"

// WechatPayInfo 微信支付信息领域实体
// 存储微信支付的详细信息，包括预支付、签名、回调等
// 纯业务模型，无 GORM 标签
type WechatPayInfo struct {
	ID         string    // 微信支付信息 ID
	PaymentID  string    // 关联的支付 ID
	AppID      string    // 微信公众号/小程序 AppID
	MCHID      string    // 商户号
	NonceStr   string    // 随机字符串
	PrepayID   string    // 预支付交易会话标识
	CodeURL    string    // 二维码链接（Native 支付）
	Sign       string    // 签名
	Timestamp  string    // 时间戳
	Package    string    // 包信息
	PaySign    string    // 支付签名
	ReturnCode string    // 返回码
	ReturnMsg  string    // 返回信息
	ResultCode string    // 业务结果
	ErrCode    string    // 错误码
	ErrCodeDes string    // 错误描述
	NotifyData string    // 回调原始数据
	CreatedBy  string    // 创建人 ID
	CreatedAt  time.Time // 创建时间
	UpdatedBy  string    // 更新人 ID
	UpdatedAt  time.Time // 更新时间
}

// IsSuccess 检查微信支付是否成功
func (w *WechatPayInfo) IsSuccess() bool {
	return w.ReturnCode == "SUCCESS" && w.ResultCode == "SUCCESS"
}

// IsError 检查微信支付是否有错误
func (w *WechatPayInfo) IsError() bool {
	return w.ReturnCode != "SUCCESS" || w.ResultCode != "SUCCESS"
}

// HasPrepayID 检查是否有预支付 ID
func (w *WechatPayInfo) HasPrepayID() bool {
	return w.PrepayID != ""
}

// HasCodeURL 检查是否有二维码链接
func (w *WechatPayInfo) HasCodeURL() bool {
	return w.CodeURL != ""
}
