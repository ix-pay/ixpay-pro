package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// OnlineUserResponse 在线用户响应 DTO
type OnlineUserResponse struct {
	UserID       string `json:"userId,string"`
	Username     string `json:"userName"`
	Nickname     string `json:"nickname"`
	SessionID    string `json:"sessionId"`
	LoginIP      string `json:"loginIp"`
	LoginPlace   string `json:"loginPlace"`
	LoginTime    string `json:"loginTime"`
	LastActiveAt string `json:"lastActiveAt"`
	Device       string `json:"device"`
	Browser      string `json:"browser"`
	OS           string `json:"os"`
	UserAgent    string `json:"userAgent"`
}

// OnlineUserListResponse 在线用户列表响应 DTO
type OnlineUserListResponse struct {
	baseRes.PageResult
	List []OnlineUserResponse `json:"list"`
}

// OnlineUserStatisticsResponse 在线用户统计响应 DTO
type OnlineUserStatisticsResponse struct {
	TotalOnline int64  `json:"totalOnline"`
	PeakOnline  int64  `json:"peakOnline"`
	PeakTime    string `json:"peakTime"`
}
