package service

import (
	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// WechatPayInfoService 实现支付领域服务接口
type WechatPayInfoService struct {
	repo                 repo.WechatPayInfoRepository
	log                  logger.Logger
	baseConfigRepository baseRepo.ConfigRepository
}

// NewWechatPayInfoService 创建支付服务实例
func NewWechatPayInfoService(repo repo.WechatPayInfoRepository, baseConfigRepository baseRepo.ConfigRepository, log logger.Logger) *WechatPayInfoService {
	return &WechatPayInfoService{
		repo:                 repo,
		log:                  log,
		baseConfigRepository: baseConfigRepository,
	}
}
