package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// NoticeResponse 公告响应 DTO
// 所有 ID 字段使用 int64 格式，通过 json:",string" 标签自动序列化为字符串
type NoticeResponse struct {
	ID          int64  `json:"id,string"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Type        int    `json:"type"`
	Status      int    `json:"status"`
	PublisherID int64  `json:"publisherId,string"`
	PublishTime string `json:"publishTime,omitempty"`
	ViewCount   int64  `json:"viewCount"`
	IsTop       bool   `json:"isTop"`
	Sort        int    `json:"sort"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// NoticeListResponse 公告列表响应 DTO
type NoticeListResponse struct {
	baseRes.PageResult
	List []NoticeResponse `json:"list"`
}

// NoticeStatisticsResponse 公告统计响应 DTO
type NoticeStatisticsResponse struct {
	TotalCount     int64 `json:"totalCount"`
	PublishedCount int64 `json:"publishedCount"`
	DraftCount     int64 `json:"draftCount"`
	ArchivedCount  int64 `json:"archivedCount"`
}
