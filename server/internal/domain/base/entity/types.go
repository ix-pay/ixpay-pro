package entity

import "time"

// HourlyStat 每小时统计（通用结构）
type HourlyStat struct {
	Hour  int   // 小时 (0-23)
	Count int64 // 统计数量
}

// DailyStat 每日统计（通用结构）
type DailyStat struct {
	Date         string // 日期
	TotalCount   int64  // 总次数
	SuccessCount int64  // 成功次数
	FailedCount  int64  // 失败次数
}

// TopUserStat 用户统计（通用结构）
type TopUserStat struct {
	UserID     string // 用户 ID
	Username   string // 用户名
	LoginCount int64  // 登录次数
}

// TopIPStat IP 统计（通用结构）
type TopIPStat struct {
	IP         string // IP 地址
	LoginCount int64  // 登录次数
}

// FailedIPStat 失败 IP 统计（通用结构）
type FailedIPStat struct {
	IP          string // IP 地址
	FailedCount int64  // 失败次数
}

// LoginStatistics 登录统计信息（通用结构）
type LoginStatistics struct {
	TotalCount   int64          // 登录总次数
	SuccessCount int64          // 成功登录次数
	FailedCount  int64          // 失败登录次数
	UniqueUsers  int64          // 独立用户数
	UniqueIPs    int64          // 独立 IP 数
	DailyStats   []DailyStat    // 每日统计
	HourlyStats  []HourlyStat   // 每小时统计
	TopUsers     []TopUserStat  // 登录最频繁用户
	TopIPs       []TopIPStat    // 登录最频繁 IP
	FailedIPs    []FailedIPStat // 失败最多的 IP
	StartTime    time.Time      // 统计开始时间
	EndTime      time.Time      // 统计结束时间
}
