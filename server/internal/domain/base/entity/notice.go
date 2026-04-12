package entity

import "time"

// NoticeType 公告类型
type NoticeType int

const (
	NoticeTypeSystem    NoticeType = 1 // 系统公告
	NoticeTypeActivity  NoticeType = 2 // 活动公告
	NoticeTypeNotice    NoticeType = 3 // 普通通知
	NoticeTypeEmergency NoticeType = 4 // 紧急通知
)

// NoticeStatus 公告状态
type NoticeStatus int

const (
	NoticeStatusDraft     NoticeStatus = 0 // 草稿
	NoticeStatusPublished NoticeStatus = 1 // 已发布
	NoticeStatusArchived  NoticeStatus = 2 // 已归档
)

// Notice 公告领域实体
// 纯业务模型，无 GORM 标签
type Notice struct {
	ID          string       // 公告 ID
	Title       string       // 公告标题
	Content     string       // 公告内容
	Type        NoticeType   // 公告类型
	Status      NoticeStatus // 公告状态
	PublisherID string       // 发布人 ID
	PublishTime *time.Time   // 发布时间
	ViewCount   int64        // 浏览次数
	IsTop       bool         // 是否置顶
	Sort        int          // 排序
	Description string       // 公告描述/摘要
	CreatedBy   string       // 创建人 ID
	CreatedAt   time.Time    // 创建时间
	UpdatedBy   string       // 更新人 ID
	UpdatedAt   time.Time    // 更新时间
}

// IsPublished 检查公告是否已发布
func (n *Notice) IsPublished() bool {
	return n.Status == NoticeStatusPublished
}

// IsDraft 检查公告是否是草稿
func (n *Notice) IsDraft() bool {
	return n.Status == NoticeStatusDraft
}

// IsArchived 检查公告是否已归档
func (n *Notice) IsArchived() bool {
	return n.Status == NoticeStatusArchived
}

// NoticeReadRecord 公告阅读记录领域实体
// 纯业务模型，无 GORM 标签
type NoticeReadRecord struct {
	ID       string    // 阅读记录 ID
	NoticeID string    // 公告 ID
	UserID   string    // 用户 ID
	ReadTime time.Time // 阅读时间
}

// NoticeStatistics 公告统计信息
type NoticeStatistics struct {
	TotalCount     int64 // 公告总数
	PublishedCount int64 // 已发布数量
	DraftCount     int64 // 草稿数量
	ArchivedCount  int64 // 已归档数量
}
