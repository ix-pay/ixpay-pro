package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// NoticeRepository 公告仓库接口
type NoticeRepository interface {
	GetByID(id int64) (*entity.Notice, error)
	Create(notice *entity.Notice) error
	Update(notice *entity.Notice) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Notice, int64, error)
	GetPublishedList(page, pageSize int, filters map[string]interface{}) ([]*entity.Notice, int64, error)
	IncrementViewCount(id int64) error
	GetStatistics() (*entity.NoticeStatistics, error)
}

// NoticeReadRecordRepository 公告阅读记录仓库接口
type NoticeReadRecordRepository interface {
	Create(record *entity.NoticeReadRecord) error
	CreateOrUpdate(noticeID, userID int64) error
	GetByNoticeIDAndUserID(noticeID, userID int64) (*entity.NoticeReadRecord, error)
	GetReadUserCount(noticeID int64) (int64, error)
}
