package common

import (
	"strconv"
	"time"
)

// Ptr 辅助函数
func IntPtr(v int) *int       { return &v }
func BoolPtr(v bool) *bool    { return &v }
func Int64Ptr(v int64) *int64 { return &v }

// TryParseInt64 可选的 string 到 int64 转换，空字符串返回 0
func TryParseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

// ParseInt64 将 string ID 转换为 int64
func ParseInt64(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

// ToString 将 int64 转换为 string ID
func ToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

// Int64ToString 将 int64 转换为 string（新增函数，与 ToString 功能相同）
func Int64ToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

// StringToInt64s 将 string ID 数组转换为 int64 数组
func StringToInt64s(ids []string) ([]int64, error) {
	result := make([]int64, len(ids))
	for i, id := range ids {
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = intID
	}
	return result, nil
}

// Int64ToStrings 将 int64 数组转换为 string ID 数组
func Int64ToStrings(ids []int64) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = strconv.FormatInt(id, 10)
	}
	return result
}

// BaseModelFields 基础模型字段集合
type BaseModelFields struct {
	ID        int64
	CreatedBy int64
	UpdatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ExtractBaseFields 从数据库模型中提取基础字段
func ExtractBaseFields(id, createdBy, updatedBy int64, createdAt, updatedAt time.Time) BaseModelFields {
	return BaseModelFields{
		ID:        id,
		CreatedBy: createdBy,
		UpdatedBy: updatedBy,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// SetBaseFields 设置基础字段到数据库模型
func SetBaseFields(id, createdBy, updatedBy string) (int64, int64, int64) {
	return TryParseInt64(id), TryParseInt64(createdBy), TryParseInt64(updatedBy)
}
