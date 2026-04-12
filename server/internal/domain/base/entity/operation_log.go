package entity

import "time"

// OperationType 操作类型枚举
type OperationType int

const (
	OperationTypeCreate OperationType = 1 // 创建
	OperationTypeUpdate OperationType = 2 // 更新
	OperationTypeDelete OperationType = 3 // 删除
	OperationTypeQuery  OperationType = 4 // 查询
	OperationTypeLogin  OperationType = 5 // 登录
	OperationTypeLogout OperationType = 6 // 登出
	OperationTypeOther  OperationType = 9 // 其他
)

// OperationLog 用户操作日志领域实体
// 纯业务模型，无 GORM 标签
type OperationLog struct {
	ID            string        // 日志 ID
	UserID        string        // 用户 ID
	Username      string        // 用户名
	Nickname      string        // 用户昵称
	OperationType OperationType // 操作类型
	Module        string        // 操作模块
	Description   string        // 操作描述
	Method        string        // 请求方法
	Path          string        // 请求路径
	Params        string        // 请求参数
	ClientIP      string        // 客户端 IP
	UserAgent     string        // 用户代理
	StatusCode    int           // 响应状态码
	Result        string        // 操作结果
	Duration      int64         // 操作耗时 (毫秒)
	ErrorMessage  string        // 错误信息
	IsSuccess     bool          // 是否成功
	CreatedBy     string        // 创建人 ID
	CreatedAt     time.Time     // 创建时间
	UpdatedBy     string        // 更新人 ID
	UpdatedAt     time.Time     // 更新时间
}

// IsSuccessOperation 检查操作是否成功
func (l *OperationLog) IsSuccessOperation() bool {
	return l.IsSuccess
}

// IsFailedOperation 检查操作是否失败
func (l *OperationLog) IsFailedOperation() bool {
	return !l.IsSuccess
}

// IsCreateOperation 检查是否是创建操作
func (l *OperationLog) IsCreateOperation() bool {
	return l.OperationType == OperationTypeCreate
}

// IsUpdateOperation 检查是否是更新操作
func (l *OperationLog) IsUpdateOperation() bool {
	return l.OperationType == OperationTypeUpdate
}

// IsDeleteOperation 检查是否是删除操作
func (l *OperationLog) IsDeleteOperation() bool {
	return l.OperationType == OperationTypeDelete
}

// IsQueryOperation 检查是否是查询操作
func (l *OperationLog) IsQueryOperation() bool {
	return l.OperationType == OperationTypeQuery
}
