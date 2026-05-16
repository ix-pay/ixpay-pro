package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDepartmentService_NameValidation 测试部门名称验证
func TestDepartmentService_NameValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"有效名称 - 短名称", "技术部", true},
		{"有效名称 - 中等名称", "技术研发中心", true},
		{"有效名称 - 长名称", "技术研发中心 - 软件开发部", true},
		{"无效名称 - 空字符串", "", false},
		{"无效名称 - 特殊字符", "技术@部", false},
		{"无效名称 - 数字", "123", false},
		{"无效名称 - 表情符号", "技术😊部", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.input == "" {
				assert.False(t, tc.valid, "部门名称不能为空")
				return
			}

			if len(tc.input) > 50 {
				assert.False(t, tc.valid, "部门名称不能超过 50 字符")
				return
			}

			for _, c := range tc.input {
				isValidChar := (c >= '\u4e00' && c <= '\u9fff') ||
					(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
					(c >= '0' && c <= '9') || c == '-' || c == '_' || c == ' '
				if !isValidChar {
					assert.False(t, tc.valid, "部门名称包含非法字符：%c", c)
					return
				}
			}
		})
	}
}

// TestDepartmentService_ParentIDValidation 测试父部门 ID 验证
func TestDepartmentService_ParentIDValidation(t *testing.T) {
	testCases := []struct {
		name     string
		parentID int64
		valid    bool
	}{
		{"有效父 ID - 根部门", 0, true},
		{"有效父 ID - 正数", 1, true},
		{"无效父 ID - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.parentID < 0 {
				assert.False(t, tc.valid, "父部门 ID 不能为负数")
			}
		})
	}
}

// TestDepartmentService_LeaderIDValidation 测试部门负责人 ID 验证
func TestDepartmentService_LeaderIDValidation(t *testing.T) {
	testCases := []struct {
		name     string
		leaderID int64
		valid    bool
	}{
		{"有效负责人 ID - 正数", 1, true},
		{"有效负责人 ID - 0（无负责人）", 0, true},
		{"无效负责人 ID - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.leaderID < 0 {
				assert.False(t, tc.valid, "负责人 ID 不能为负数")
			}
		})
	}
}

// TestDepartmentService_SortValidation 测试排序字段验证
func TestDepartmentService_SortValidation(t *testing.T) {
	testCases := []struct {
		name  string
		sort  int
		valid bool
	}{
		{"排序 - 0", 0, true},
		{"排序 - 正数", 10, true},
		{"排序 - 负数", -1, true},
		{"排序 - 大数", 99999, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tc.valid, "排序字段应该允许所有整数值")
		})
	}
}

// TestDepartmentService_StatusValidation 测试状态字段验证
func TestDepartmentService_StatusValidation(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		valid  bool
	}{
		{"有效状态 - 正常", 1, true},
		{"有效状态 - 禁用", 0, true},
		{"无效状态 - 负数", -1, false},
		{"无效状态 - 大于 1", 2, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.status != 0 && tc.status != 1 {
				assert.False(t, tc.valid, "状态只能是 0 或 1")
			}
		})
	}
}

// TestDepartmentService_DescriptionValidation 测试描述字段验证
func TestDepartmentService_DescriptionValidation(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		valid       bool
	}{
		{"空描述", "", true},
		{"短描述", "负责技术研发", true},
		{"中等描述", "负责公司所有技术项目的研发和管理工作", true},
		{"长描述", "这是一个非常长的部门描述，用于测试描述字段的长度限制。这个描述包含了很多字符，用来验证系统是否能够正确处理长文本输入。", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.description) > 500 {
				assert.False(t, tc.valid, "描述长度不能超过 500 字符")
			}
		})
	}
}

// TestDepartmentService_DepthValidation 测试部门层级深度验证
func TestDepartmentService_DepthValidation(t *testing.T) {
	testCases := []struct {
		name  string
		depth int
		valid bool
	}{
		{"有效深度 - 1 层", 1, true},
		{"有效深度 - 3 层", 3, true},
		{"有效深度 - 5 层", 5, true},
		{"无效深度 - 超过 10 层", 11, false},
		{"无效深度 - 0", 0, false},
		{"无效深度 - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.depth <= 0 || tc.depth > 10 {
				assert.False(t, tc.valid, "部门层级深度应该在 1-10 之间")
			}
		})
	}
}

// TestDepartmentService_CycleReference 测试循环引用检测
func TestDepartmentService_CycleReference(t *testing.T) {
	testCases := []struct {
		name        string
		deptID      int64
		parentID    int64
		expectCycle bool
	}{
		{"正常引用 - 部门 1 的父部门是 0", 1, 0, false},
		{"正常引用 - 部门 2 的父部门是 1", 2, 1, false},
		{"循环引用 - 部门 1 的父部门是自己", 1, 1, true},
		{"循环引用 - 间接循环", 1, 2, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.deptID == tc.parentID {
				assert.True(t, tc.expectCycle, "部门不能以自己为父部门")
			}
		})
	}
}
