package service

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	wxpay_utility "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayutility"
	wxpayv3 "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayv3"
)

// BillService 账单业务编排（对齐 Gpay service/bill.go：交易账单/资金账单/下载）
type BillService struct {
	wechatPayService *WechatPayService
	log              logger.Logger
	_                repo.PaymentRepository // 预留：账单与本地订单对账扩展
}

// NewBillService 创建账单服务
func NewBillService(wechatPayService *WechatPayService, log logger.Logger) *BillService {
	return &BillService{wechatPayService: wechatPayService, log: log}
}

// TradeBill 申请交易账单，返回下载链接与SHA1摘要
// billType: ALL/SUCCESS/REFUND/RECHARGE_REFUND，默认ALL
func (s *BillService) TradeBill(billDate, billType string) (*wxpayv3.BillResponse, error) {
	if err := validateBillDate(billDate); err != nil {
		return nil, err
	}
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	return client.TradeBill(&wxpayv3.TradeBillRequest{BillDate: billDate, BillType: billType})
}

// FundFlowBill 申请资金账单
// accountType: BASIC基本账户/OPERATION运营账户/FEES手续费账户，默认BASIC
func (s *BillService) FundFlowBill(billDate, accountType string) (*wxpayv3.BillResponse, error) {
	if err := validateBillDate(billDate); err != nil {
		return nil, err
	}
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	return client.FundFlowBill(&wxpayv3.FundFlowBillRequest{BillDate: billDate, AccountType: accountType})
}

// DownloadBill 下载账单内容并校验SHA1摘要
// download_url为申请账单应答返回的链接(5分钟内有效)；
// 超过2MB的账单微信侧以gzip压缩返回，此处自动解压
func (s *BillService) DownloadBill(downloadURL, hashValue string) ([]byte, error) {
	if downloadURL == "" {
		return nil, fmt.Errorf("download_url不能为空")
	}
	client, err := s.wechatPayService.GetV3Client()
	if err != nil {
		return nil, err
	}
	data, err := client.DownloadBill(downloadURL)
	if err != nil {
		return nil, err
	}

	// gzip账单自动解压(魔数 1f 8b)
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("账单gzip解压失败: %w", err)
		}
		defer zr.Close()
		if data, err = io.ReadAll(zr); err != nil {
			return nil, fmt.Errorf("账单gzip解压失败: %w", err)
		}
	}

	// SHA1摘要校验(hash_value为账单源文件摘要，委托官方 WXPayUtility 工具类)
	if hashValue != "" {
		sum, err := wxpay_utility.GenerateSHA1FromStream(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("账单SHA1摘要计算失败: %w", err)
		}
		if !strings.EqualFold(sum, hashValue) {
			return nil, fmt.Errorf("账单SHA1摘要校验失败")
		}
	}
	return data, nil
}

// validateBillDate 账单日期校验: yyyy-MM-dd、非当天、三个月内
func validateBillDate(billDate string) error {
	t, err := time.ParseInLocation("2006-01-02", billDate, time.Local)
	if err != nil {
		return fmt.Errorf("账单日期格式应为yyyy-MM-dd")
	}
	if billDate == time.Now().Format("2006-01-02") {
		return fmt.Errorf("账单日期不能为当天，最早为昨天")
	}
	if time.Since(t) > 93*24*time.Hour {
		return fmt.Errorf("账单日期超出三个月范围")
	}
	return nil
}
