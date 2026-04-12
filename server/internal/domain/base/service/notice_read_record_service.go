package service

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// NoticeReadRecordService 公告阅读记录服务实现
type NoticeReadRecordService struct {
	repo repo.NoticeReadRecordRepository
	log  logger.Logger
}

// NewNoticeReadRecordService 创建公告阅读记录服务实例
func NewNoticeReadRecordService(repo repo.NoticeReadRecordRepository, log logger.Logger) *NoticeReadRecordService {
	return &NoticeReadRecordService{
		repo: repo,
		log:  log,
	}
}

// CreateReadRecord 创建阅读记录
func (s *NoticeReadRecordService) CreateReadRecord(noticeID string, userID string) error {
	if err := s.repo.CreateOrUpdate(noticeID, userID); err != nil {
		s.log.Error("创建阅读记录失败", "error", err, "notice_id", noticeID, "user_id", userID)
		return err
	}

	s.log.Info("创建阅读记录成功", "notice_id", noticeID, "user_id", userID)
	return nil
}

// GetReadUserCount 获取公告阅读人数
func (s *NoticeReadRecordService) GetReadUserCount(noticeID string) (int64, error) {
	count, err := s.repo.GetReadUserCount(noticeID)
	if err != nil {
		s.log.Error("获取阅读人数失败", "error", err, "notice_id", noticeID)
		return 0, err
	}

	return count, nil
}
