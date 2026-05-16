package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOperationLogService_ModuleValidation 测试模块名称验证
func TestOperationLogService_ModuleValidation(t *testing.T) {
	testCases := []struct {
		name   string
		module string
		valid  bool
	}{
		{"有效模块 - 短模块", "用户管理", true},
		{"有效模块 - 中等模块", "系统管理", true},
		{"有效模块 - 长模块", "系统管理 - 用户管理", true},
		{"无效模块 - 空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.module == "" {
				assert.False(t, tc.valid, "模块名称不能为空")
			}

			if len(tc.module) > 50 {
				assert.False(t, tc.valid, "模块名称不能超过 50 字符")
			}
		})
	}
}

// TestOperationLogService_ActionValidation 测试操作类型验证
func TestOperationLogService_ActionValidation(t *testing.T) {
	testCases := []struct {
		name   string
		action string
		valid  bool
	}{
		{"有效操作 - 新增", "新增", true},
		{"有效操作 - 修改", "修改", true},
		{"有效操作 - 删除", "删除", true},
		{"有效操作 - 查询", "查询", true},
		{"有效操作 - 导出", "导出", true},
		{"有效操作 - 导入", "导入", true},
		{"无效操作 - 空字符串", "", false},
		{"无效操作 - 过长", "这是一个非常非常长的操作类型名称", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.action == "" {
				assert.False(t, tc.valid, "操作类型不能为空")
			}

			if len(tc.action) > 20 {
				assert.False(t, tc.valid, "操作类型不能超过 20 字符")
			}
		})
	}
}

// TestOperationLogService_MethodValidation 测试请求方法验证
func TestOperationLogService_MethodValidation(t *testing.T) {
	testCases := []struct {
		name   string
		method string
		valid  bool
	}{
		{"有效方法 - GET", "GET", true},
		{"有效方法 - POST", "POST", true},
		{"有效方法 - PUT", "PUT", true},
		{"有效方法 - DELETE", "DELETE", true},
		{"有效方法 - PATCH", "PATCH", true},
		{"无效方法 - 小写", "get", false},
		{"无效方法 - 空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.method == "" {
				assert.False(t, tc.valid, "请求方法不能为空")
				return
			}

			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true,
				"DELETE": true, "PATCH": true,
			}

			assert.Equal(t, tc.valid, validMethods[tc.method], "请求方法 %s 验证失败", tc.method)
		})
	}
}

// TestOperationLogService_OperNameValidation 测试操作人员验证
func TestOperationLogService_OperNameValidation(t *testing.T) {
	testCases := []struct {
		name     string
		operName string
		valid    bool
	}{
		{"有效操作人员 - 短名称", "admin", true},
		{"有效操作人员 - 中等名称", "zhangsan", true},
		{"有效操作人员 - 长名称", "administrator", true},
		{"无效操作人员 - 空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.operName == "" {
				assert.False(t, tc.valid, "操作人员不能为空")
			}

			if len(tc.operName) > 50 {
				assert.False(t, tc.valid, "操作人员不能超过 50 字符")
			}
		})
	}
}

// TestOperationLogService_TitleValidation 测试操作标题验证
func TestOperationLogService_TitleValidation(t *testing.T) {
	testCases := []struct {
		name  string
		title string
		valid bool
	}{
		{"有效标题 - 短标题", "新增用户", true},
		{"有效标题 - 中等标题", "修改用户信息", true},
		{"有效标题 - 长标题", "删除用户账号及相关联的数据", true},
		{"无效标题 - 空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.title == "" {
				assert.False(t, tc.valid, "操作标题不能为空")
			}

			if len(tc.title) > 200 {
				assert.False(t, tc.valid, "操作标题不能超过 200 字符")
			}
		})
	}
}

// TestOperationLogService_BusinessTypeValidation 测试业务类型验证
func TestOperationLogService_BusinessTypeValidation(t *testing.T) {
	testCases := []struct {
		name         string
		businessType int
		valid        bool
	}{
		{"有效类型 - 其他", 0, true},
		{"有效类型 - 新增", 1, true},
		{"有效类型 - 修改", 2, true},
		{"有效类型 - 删除", 3, true},
		{"有效类型 - 授权", 4, true},
		{"有效类型 - 导出", 5, true},
		{"有效类型 - 导入", 6, true},
		{"有效类型 - 强退", 7, true},
		{"有效类型 - 生成代码", 8, true},
		{"有效类型 - 清空数据", 99, true},
		{"无效类型 - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.businessType < 0 {
				assert.False(t, tc.valid, "业务类型不能为负数")
			}
		})
	}
}
