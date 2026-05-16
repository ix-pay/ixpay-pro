package converter

// EntityConverter 泛型转换器
// 提供统一的 Entity 到 DTO 转换方法
type EntityConverter struct{}

// NewEntityConverter 创建转换器实例（用于 Wire 依赖注入）
func NewEntityConverter() *EntityConverter {
	return &EntityConverter{}
}

// ConvertWithFunc 使用自定义转换函数转换 Entity
// 支持一个 Entity 转换成多种 DTO
func ConvertWithFunc[E any, T any](entity *E, convertFunc func(*E) T) T {
	return convertFunc(entity)
}

// ConvertSliceWithFunc 批量使用自定义转换函数转换 Entity
func ConvertSliceWithFunc[E any, T any](entities []*E, convertFunc func(*E) T) []T {
	if len(entities) == 0 {
		return []T{}
	}

	result := make([]T, 0, len(entities))
	for _, entity := range entities {
		result = append(result, convertFunc(entity))
	}
	return result
}
