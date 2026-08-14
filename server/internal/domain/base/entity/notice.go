package entity

import "time"

// NoticeType 公告类型
type NoticeType string

const (
	NoticeTypeSystem    NoticeType = "system"    // 系统公告
	NoticeTypeActivity  NoticeType = "activity"  // 活动公告
	NoticeTypeNotice    NoticeType = "notice"    // 普通通知
	NoticeTypeEmergency NoticeType = "emergency" // 紧急通知
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
	ID          int64        // 公告 ID
	Title       string       // 公告标题
	Content     string       // 公告内容
	Type        NoticeType   // 公告类型
	Status      NoticeStatus // 公告状态
	PublisherID int64        // 发布人 ID
	PublishTime *time.Time   // 发布时间
	ViewCount   int64        // 浏览次数
	IsTop       bool         // 是否置顶
	Sort        int          // 排序
	Description string       // 公告描述/摘要
	CreatedBy   int64        // 创建人 ID
	CreatedAt   time.Time    // 创建时间
	UpdatedBy   int64        // 更新人 ID
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

// ParseNoticeType 将字符串类型值转换为 NoticeType 常量
// 支持数字字符串（"1"="system", "2"="activity", "3"="notice", "4"="emergency"）和字符串常量值
func ParseNoticeType(typeStr string) (NoticeType, bool) {
	switch typeStr {
	case "1", "system":
		return NoticeTypeSystem, true
	case "2", "activity":
		return NoticeTypeActivity, true
	case "3", "notice":
		return NoticeTypeNotice, true
	case "4", "emergency":
		return NoticeTypeEmergency, true
	default:
		return "", false
	}
}

// NoticeReadRecord 公告阅读记录领域实体
// 纯业务模型，无 GORM 标签
type NoticeReadRecord struct {
	ID       int64     // 阅读记录 ID
	NoticeID int64     // 公告 ID
	UserID   int64     // 用户 ID
	ReadTime time.Time // 阅读时间
}

// NoticeStatistics 公告统计信息
type NoticeStatistics struct {
	TotalCount     int64 // 公告总数
	PublishedCount int64 // 已发布数量
	DraftCount     int64 // 草稿数量
	ArchivedCount  int64 // 已归档数量
}
