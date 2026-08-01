// request 包定义通知公告管理相关的请求模型
// 用于接收和验证 HTTP 请求参数
package request

// CreateNoticeRequest 创建公告请求
type CreateNoticeRequest struct {
	Title       string `json:"title" binding:"required"`   // 公告标题
	Content     string `json:"content" binding:"required"` // 公告内容
	Type        string `json:"type" binding:"required"`    // 公告类型：system-系统公告，activity-活动公告，notice-普通通知，emergency-紧急通知
	Status      int    `json:"status"`                     // 公告状态：0-草稿，1-已发布，2-已归档
	PublishTime string `json:"publishTime"`                // 发布时间（格式：2006-01-02 15:04:05）
	IsTop       bool   `json:"isTop"`                      // 是否置顶
	Sort        int    `json:"sort"`                       // 排序
	Description string `json:"description"`                // 公告描述/摘要
}

// UpdateNoticeRequest 更新公告请求
type UpdateNoticeRequest struct {
	ID          int64  `json:"id,string" binding:"required"` // 公告 ID
	Title       string `json:"title" binding:"required"`     // 公告标题
	Content     string `json:"content" binding:"required"`   // 公告内容
	Type        string `json:"type" binding:"required"`      // 公告类型：system-系统公告，activity-活动公告，notice-普通通知，emergency-紧急通知
	IsTop       bool   `json:"isTop"`                        // 是否置顶
	Sort        int    `json:"sort"`                         // 排序
	Description string `json:"description"`                  // 公告描述/摘要
}

// PublishNoticeRequest 发布公告请求
type PublishNoticeRequest struct {
	ID int64 `json:"id,string" binding:"required"` // 公告 ID
}

// GetNoticeListRequest 获取公告列表请求
type GetNoticeListRequest struct {
	Page     int    `form:"page" binding:"required"`     // 页码
	PageSize int    `form:"pageSize" binding:"required"` // 每页数量
	Title    string `form:"title"`                       // 公告标题（模糊查询）
	Type     *int   `form:"type"`                        // 公告类型（可选筛选条件）
	Status   *int   `form:"status"`                      // 公告状态（可选筛选条件）
	IsTop    *bool  `form:"isTop"`                       // 是否置顶（可选筛选条件）
}

// GetNoticeByIDRequest 获取公告详情请求
type GetNoticeByIDRequest struct {
	ID int64 `form:"id" binding:"required"` // 公告 ID
}

// DeleteNoticeRequest 删除公告请求
type DeleteNoticeRequest struct {
	ID int64 `json:"id,string" binding:"required"` // 公告 ID
}

// MarkNoticeAsReadRequest 标记公告已读请求
type MarkNoticeAsReadRequest struct {
	ID int64 `uri:"id" binding:"required"` // 公告 ID
}
