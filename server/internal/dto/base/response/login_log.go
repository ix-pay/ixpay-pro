package response

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// LoginLogListDTO 列表 DTO（精简版）
type LoginLogListDTO struct {
	ID        int64  `json:"id,string"`
	Username  string `json:"username"`
	IP        string `json:"ip"`
	Place     string `json:"place"`
	Result    int    `json:"result"`
	CreatedAt string `json:"createdAt"`
}

// LoginLogDetailDTO 详情 DTO（完整版）
type LoginLogDetailDTO struct {
	ID        int64     `json:"id,string"`
	UserID    int64     `json:"userId,string"`
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	Place     string    `json:"place"`
	Device    string    `json:"device"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	Result    int       `json:"result"`
	CreatedAt time.Time `json:"createdAt"`
}

// LoginLogStatisticsDTO 统计 DTO
type LoginLogStatisticsDTO struct {
	Date    string `json:"date"`
	Total   int    `json:"total"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
}

// AbnormalLoginInfoDTO 异常登录信息 DTO
type AbnormalLoginInfoDTO struct {
	IP              string   `json:"ip"`
	FailedCount     int64    `json:"failedCount"`
	LastFailedTime  string   `json:"lastFailedTime"`
	Usernames       []string `json:"usernames"`
	RiskLevel       string   `json:"riskLevel"`
	RiskDescription string   `json:"riskDescription"`
}

// LoginLogListResponse 登录日志列表响应
type LoginLogListResponse struct {
	PageResult baseRes.PageResult `json:"pageResult"`
	List       []LoginLogListDTO  `json:"list"`
}

// LoginLogDetailResponse 登录日志详情响应
type LoginLogDetailResponse struct {
	Data LoginLogDetailDTO `json:"data"`
}

// LoginLogStatisticsResponse 登录统计响应
type LoginLogStatisticsResponse struct {
	Data LoginLogStatisticsDTO `json:"data"`
}
