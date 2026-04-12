package entity

import "time"

// OnlineUser 在线用户领域实体
// 存储在 Redis 中，用于实时跟踪在线用户状态
// 纯业务模型，无 GORM 标签
type OnlineUser struct {
	UserID       string    // 用户 ID
	Username     string    // 用户名
	Nickname     string    // 用户昵称
	SessionID    string    // 会话 ID
	LoginIP      string    // 登录 IP
	LoginPlace   string    // 登录地点
	LoginTime    time.Time // 登录时间
	LastActiveAt time.Time // 最后活跃时间
	Device       string    // 设备信息
	Browser      string    // 浏览器信息
	OS           string    // 操作系统
	UserAgent    string    // 原始 User-Agent
}

// IsActive 检查用户是否活跃（5 分钟内）
func (o *OnlineUser) IsActive() bool {
	return time.Since(o.LastActiveAt) < 5*time.Minute
}

// GetSessionID 获取会话 ID
func (o *OnlineUser) GetSessionID() string {
	return o.SessionID
}

// UpdateActiveTime 更新活跃时间
func (o *OnlineUser) UpdateActiveTime() {
	o.LastActiveAt = time.Now()
}

// OnlineUserStatistics 在线用户统计信息
// 注意：HourlyStat 已在 types.go 中定义
type OnlineUserStatistics struct {
	TotalOnline int64        // 总在线人数
	PeakOnline  int64        // 峰值在线人数
	PeakTime    time.Time    // 峰值时间
	HourlyStats []HourlyStat // 每小时统计
}
