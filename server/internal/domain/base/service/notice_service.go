package service

import (
	"errors"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// NoticeService 公告服务实现
type NoticeService struct {
	repo       repo.NoticeRepository
	recordRepo repo.NoticeReadRecordRepository
	log        logger.Logger
}

// NewNoticeService 创建公告服务实例
func NewNoticeService(repo repo.NoticeRepository, recordRepo repo.NoticeReadRecordRepository, log logger.Logger) *NoticeService {
	return &NoticeService{
		repo:       repo,
		recordRepo: recordRepo,
		log:        log,
	}
}

// CreateNotice 创建公告
func (s *NoticeService) CreateNotice(title, content, description string, noticeType entity.NoticeType, publisherID int64, isTop bool, sort int) (*entity.Notice, error) {
	// 验证公告类型
	if noticeType < entity.NoticeTypeSystem || noticeType > entity.NoticeTypeEmergency {
		s.log.Error("无效的公告类型", "type", noticeType)
		return nil, errors.New("无效的公告类型")
	}

	// 创建公告
	notice := &entity.Notice{
		Title:       title,
		Content:     content,
		Type:        noticeType,
		Status:      entity.NoticeStatusDraft, // 默认为草稿状态
		PublisherID: publisherID,
		IsTop:       isTop,
		Sort:        sort,
		Description: description,
		CreatedBy:   publisherID,
		UpdatedBy:   publisherID,
	}

	if err := s.repo.Create(notice); err != nil {
		s.log.Error("创建公告失败", "error", err, "title", title)
		return nil, err
	}

	s.log.Info("创建公告成功", "id", notice.ID, "title", title, "publisher_id", publisherID)
	return notice, nil
}

// UpdateNotice 更新公告
func (s *NoticeService) UpdateNotice(id int64, title, content, description string, noticeType entity.NoticeType, publisherID int64, isTop bool, sort int) (*entity.Notice, error) {
	// 获取公告
	notice, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取公告失败", "error", err, "id", id)
		return nil, errors.New("公告不存在")
	}

	// 已发布的公告不允许修改（如需修改应先下架）
	if notice.Status == entity.NoticeStatusPublished {
		s.log.Error("已发布的公告不允许直接修改", "id", id, "status", notice.Status)
		return nil, errors.New("已发布的公告不允许直接修改，请先归档")
	}

	// 验证公告类型
	if noticeType < entity.NoticeTypeSystem || noticeType > entity.NoticeTypeEmergency {
		s.log.Error("无效的公告类型", "type", noticeType)
		return nil, errors.New("无效的公告类型")
	}

	// 更新公告信息
	notice.Title = title
	notice.Content = content
	notice.Type = noticeType
	notice.IsTop = isTop
	notice.Sort = sort
	notice.Description = description
	notice.UpdatedBy = publisherID

	if err := s.repo.Update(notice); err != nil {
		s.log.Error("更新公告失败", "error", err, "id", id)
		return nil, err
	}

	s.log.Info("更新公告成功", "id", id, "title", title, "publisher_id", publisherID)
	return notice, nil
}

// PublishNotice 发布公告
func (s *NoticeService) PublishNotice(id int64, publisherID int64) error {
	// 获取公告
	notice, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取公告失败", "error", err, "id", id)
		return errors.New("公告不存在")
	}

	// 检查公告状态
	if notice.Status == entity.NoticeStatusPublished {
		s.log.Error("公告已发布", "id", id)
		return errors.New("公告已发布")
	}

	if notice.Status == entity.NoticeStatusArchived {
		s.log.Error("已归档的公告不能发布", "id", id)
		return errors.New("已归档的公告不能发布")
	}

	// 更新公告状态为已发布
	notice.Status = entity.NoticeStatusPublished
	publishTime := time.Now()
	notice.PublishTime = &publishTime
	notice.UpdatedBy = publisherID

	if err := s.repo.Update(notice); err != nil {
		s.log.Error("发布公告失败", "error", err, "id", id)
		return err
	}

	s.log.Info("发布公告成功", "id", id, "title", notice.Title, "publisher_id", publisherID)
	return nil
}

// DeleteNotice 删除公告
func (s *NoticeService) DeleteNotice(id int64) error {
	// 获取公告
	notice, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取公告失败", "error", err, "id", id)
		return errors.New("公告不存在")
	}

	// 删除公告
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除公告失败", "error", err, "id", id)
		return err
	}

	s.log.Info("删除公告成功", "id", id, "title", notice.Title)
	return nil
}

// GetNoticeByID 获取公告详情
func (s *NoticeService) GetNoticeByID(id int64) (*entity.Notice, error) {
	notice, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取公告失败", "error", err, "id", id)
		return nil, errors.New("公告不存在")
	}
	return notice, nil
}

// GetNoticeList 获取公告列表
func (s *NoticeService) GetNoticeList(page, pageSize int, filters map[string]interface{}) ([]*entity.Notice, int64, error) {
	notices, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取公告列表失败", "error", err)
		return nil, 0, err
	}
	return notices, total, nil
}

// GetPublishedNoticeList 获取已发布公告列表
func (s *NoticeService) GetPublishedNoticeList(page, pageSize int, filters map[string]interface{}) ([]*entity.Notice, int64, error) {
	notices, total, err := s.repo.GetPublishedList(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取已发布公告列表失败", "error", err)
		return nil, 0, err
	}
	return notices, total, nil
}

// MarkAsRead 标记公告为已读
func (s *NoticeService) MarkAsRead(noticeID int64, userID int64) error {
	// 检查公告是否存在
	_, err := s.repo.GetByID(noticeID)
	if err != nil {
		s.log.Error("公告不存在", "error", err, "notice_id", noticeID)
		return errors.New("公告不存在")
	}

	// 创建或更新阅读记录
	if err := s.recordRepo.CreateOrUpdate(noticeID, userID); err != nil {
		s.log.Error("创建阅读记录失败", "error", err, "notice_id", noticeID, "user_id", userID)
		return err
	}

	// 增加浏览次数
	if err := s.repo.IncrementViewCount(noticeID); err != nil {
		s.log.Error("增加浏览次数失败", "error", err, "notice_id", noticeID)
		// 浏览次数增加失败不影响标记已读，只记录日志
	}

	s.log.Info("标记公告已读", "notice_id", noticeID, "user_id", userID)
	return nil
}

// IsRead 检查用户是否已读公告
func (s *NoticeService) IsRead(noticeID int64, userID int64) (bool, error) {
	_, err := s.recordRepo.GetByNoticeIDAndUserID(noticeID, userID)
	if err != nil {
		// 如果没有找到记录，返回 false
		return false, nil
	}
	return true, nil
}

// GetStatistics 获取公告统计
func (s *NoticeService) GetStatistics() (*entity.NoticeStatistics, error) {
	stats, err := s.repo.GetStatistics()
	if err != nil {
		s.log.Error("获取公告统计失败", "error", err)
		return nil, err
	}
	return stats, nil
}
