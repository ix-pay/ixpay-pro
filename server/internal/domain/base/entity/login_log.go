package entity

import "time"

// LoginResult 登录结果
type LoginResult int

const (
	LoginResultSuccess LoginResult = 1 // 登录成功
	LoginResultFailed  LoginResult = 0 // 登录失败
)

// LoginLog 登录日志领域实体
// 记录用户登录行为，用于安全审计和异常检测
// 纯业务模型，无 GORM 标签
type LoginLog struct {
	ID         string      // 日志 ID
	UserID     string      // 用户 ID
	Username   string      // 用户名
	LoginIP    string      // 登录 IP
	LoginTime  time.Time   // 登录时间
	LoginPlace string      // 登录地点
	Device     string      // 设备信息
	Browser    string      // 浏览器信息
	OS         string      // 操作系统
	Result     LoginResult // 登录结果
	ErrorMsg   string      // 错误信息
	UserAgent  string      // 原始 User-Agent
	CreatedBy  string      // 创建人 ID
	CreatedAt  time.Time   // 创建时间
	UpdatedBy  string      // 更新人 ID
	UpdatedAt  time.Time   // 更新时间
}

// IsSuccess 检查登录是否成功
func (l *LoginLog) IsSuccess() bool {
	return l.Result == LoginResultSuccess
}

// IsFailed 检查登录是否失败
func (l *LoginLog) IsFailed() bool {
	return l.Result == LoginResultFailed
}

// AbnormalLoginInfo 异常登录信息
type AbnormalLoginInfo struct {
	IP              string    // IP 地址
	FailedCount     int64     // 失败次数
	LastFailedTime  time.Time // 最后失败时间
	Usernames       []string  // 尝试的用户名列表
	RiskLevel       string    // 风险等级：low, medium, high
	RiskDescription string    // 风险描述
}
