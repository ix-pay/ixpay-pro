package baseapi

import (
	"fmt"
	"strconv"
)

// convertStringSliceToInt64Slice 将字符串切片转换为 int64 切片
func convertStringSliceToInt64Slice(stringSlice []string) ([]int64, error) {
	if stringSlice == nil {
		return nil, nil
	}
	int64Slice := make([]int64, 0, len(stringSlice))
	for _, s := range stringSlice {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的 ID 格式：%s", s)
		}
		int64Slice = append(int64Slice, id)
	}
	return int64Slice, nil
}
