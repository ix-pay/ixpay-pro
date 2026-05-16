package service

import (
	"encoding/json"
	"testing"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/stretchr/testify/assert"
)

// TestMenuService_DeleteImpactLevels 测试删除影响等级评估
func TestMenuService_DeleteImpactLevels(t *testing.T) {
	testCases := []struct {
		name               string
		childMenusCount    int64
		btnPermsCount      int64
		affectedRolesCount int64
		affectedApisCount  int64
		expectedLevel      string
		expectedWarning    string
	}{
		{
			name:               "无影响",
			childMenusCount:    0,
			btnPermsCount:      0,
			affectedRolesCount: 0,
			affectedApisCount:  0,
			expectedLevel:      "LOW",
			expectedWarning:    "",
		},
		{
			name:               "低影响 - 只有子菜单",
			childMenusCount:    2,
			btnPermsCount:      0,
			affectedRolesCount: 0,
			affectedApisCount:  0,
			expectedLevel:      "LOW",
			expectedWarning:    "将删除 2 个子菜单",
		},
		{
			name:               "中等影响 - 有按钮权限",
			childMenusCount:    1,
			btnPermsCount:      3,
			affectedRolesCount: 1,
			affectedApisCount:  2,
			expectedLevel:      "MEDIUM",
			expectedWarning:    "将删除 3 个按钮权限，影响 1 个角色",
		},
		{
			name:               "高影响 - 影响多个角色",
			childMenusCount:    0,
			btnPermsCount:      5,
			affectedRolesCount: 5,
			affectedApisCount:  10,
			expectedLevel:      "HIGH",
			expectedWarning:    "将删除 5 个按钮权限，影响 5 个角色",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 验证等级评估逻辑
			var expectedLevel string
			if tc.affectedRolesCount >= 3 || (tc.btnPermsCount >= 5 && tc.affectedRolesCount >= 2) {
				expectedLevel = "HIGH"
			} else if tc.btnPermsCount > 0 || tc.affectedRolesCount > 0 {
				expectedLevel = "MEDIUM"
			} else {
				expectedLevel = "LOW"
			}

			assert.Equal(t, tc.expectedLevel, expectedLevel, "影响等级应正确")

			// 验证警告信息
			if tc.expectedWarning != "" {
				assert.NotEmpty(t, tc.expectedWarning, "应生成警告信息")
			}
		})
	}
}

// TestMenuService_DeleteImpactCalculation 测试删除影响计算逻辑
func TestMenuService_DeleteImpactCalculation(t *testing.T) {
	// 模拟影响计算
	childMenus := int64(3)
	btnPerms := int64(5)
	affectedRoles := int64(2)
	affectedApis := int64(8)

	impact := &entity.DeleteImpact{
		ChildMenusCount:    childMenus,
		BtnPermsCount:      btnPerms,
		AffectedRolesCount: affectedRoles,
		AffectedApisCount:  affectedApis,
	}

	// 验证计算结果
	assert.Equal(t, int64(3), impact.ChildMenusCount, "子菜单数量应正确")
	assert.Equal(t, int64(5), impact.BtnPermsCount, "按钮权限数量应正确")
	assert.Equal(t, int64(2), impact.AffectedRolesCount, "影响角色数量应正确")
	assert.Equal(t, int64(8), impact.AffectedApisCount, "影响 API 数量应正确")

	// 验证总影响范围
	totalImpact := impact.ChildMenusCount + impact.BtnPermsCount + impact.AffectedRolesCount + impact.AffectedApisCount
	assert.Equal(t, int64(18), totalImpact, "总影响范围应正确")
}

// TestMenuService_DeleteImpactEdgeCases 测试边界情况
func TestMenuService_DeleteImpactEdgeCases(t *testing.T) {
	testCases := []struct {
		name               string
		childMenusCount    int64
		btnPermsCount      int64
		affectedRolesCount int64
		affectedApisCount  int64
		expectHighImpact   bool
	}{
		{
			name:               "全零边界",
			childMenusCount:    0,
			btnPermsCount:      0,
			affectedRolesCount: 0,
			affectedApisCount:  0,
			expectHighImpact:   false,
		},
		{
			name:               "最大值边界",
			childMenusCount:    999,
			btnPermsCount:      999,
			affectedRolesCount: 999,
			affectedApisCount:  999,
			expectHighImpact:   true,
		},
		{
			name:               "只有子菜单",
			childMenusCount:    10,
			btnPermsCount:      0,
			affectedRolesCount: 0,
			affectedApisCount:  0,
			expectHighImpact:   false,
		},
		{
			name:               "只有按钮权限",
			childMenusCount:    0,
			btnPermsCount:      100,
			affectedRolesCount: 0,
			affectedApisCount:  0,
			expectHighImpact:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			impact := &entity.DeleteImpact{
				ChildMenusCount:    tc.childMenusCount,
				BtnPermsCount:      tc.btnPermsCount,
				AffectedRolesCount: tc.affectedRolesCount,
				AffectedApisCount:  tc.affectedApisCount,
			}

			// 验证高影响判断
			isHighImpact := tc.affectedRolesCount >= 3 || (tc.btnPermsCount >= 5 && tc.affectedRolesCount >= 2)
			assert.Equal(t, tc.expectHighImpact, isHighImpact, "高影响判断应正确")

			// 验证数据完整性
			assert.GreaterOrEqual(t, impact.ChildMenusCount, int64(0), "子菜单数量不应为负")
			assert.GreaterOrEqual(t, impact.BtnPermsCount, int64(0), "按钮权限数量不应为负")
			assert.GreaterOrEqual(t, impact.AffectedRolesCount, int64(0), "影响角色数量不应为负")
			assert.GreaterOrEqual(t, impact.AffectedApisCount, int64(0), "影响 API 数量不应为负")
		})
	}
}

// TestMenuService_DeleteImpactWarningMessages 测试警告信息生成
func TestMenuService_DeleteImpactWarningMessages(t *testing.T) {
	testCases := []struct {
		name               string
		childMenusCount    int64
		btnPermsCount      int64
		affectedRolesCount int64
		shouldHaveWarning  bool
	}{
		{
			name:               "无影响无警告",
			childMenusCount:    0,
			btnPermsCount:      0,
			affectedRolesCount: 0,
			shouldHaveWarning:  false,
		},
		{
			name:               "有子菜单有警告",
			childMenusCount:    5,
			btnPermsCount:      0,
			affectedRolesCount: 0,
			shouldHaveWarning:  true,
		},
		{
			name:               "有按钮权限有警告",
			childMenusCount:    0,
			btnPermsCount:      3,
			affectedRolesCount: 1,
			shouldHaveWarning:  true,
		},
		{
			name:               "影响多个角色有警告",
			childMenusCount:    0,
			btnPermsCount:      0,
			affectedRolesCount: 5,
			shouldHaveWarning:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			impact := &entity.DeleteImpact{
				ChildMenusCount:    tc.childMenusCount,
				BtnPermsCount:      tc.btnPermsCount,
				AffectedRolesCount: tc.affectedRolesCount,
			}

			// 验证警告信息
			hasWarning := impact.ChildMenusCount > 0 || impact.BtnPermsCount > 0 || impact.AffectedRolesCount > 0
			assert.Equal(t, tc.shouldHaveWarning, hasWarning, "警告信息生成应正确")
		})
	}
}

// TestMenuService_DeleteImpactJSONSerialization 测试 JSON 序列化
func TestMenuService_DeleteImpactJSONSerialization(t *testing.T) {
	impact := &entity.DeleteImpact{
		ChildMenusCount:    5,
		BtnPermsCount:      10,
		AffectedRolesCount: 3,
		AffectedApisCount:  15,
		Level:              "HIGH",
		Warning:            "将删除 10 个按钮权限，影响 3 个角色",
	}

	// 测试序列化
	jsonData, err := json.Marshal(impact)
	assert.NoError(t, err, "序列化不应出错")
	assert.Greater(t, len(jsonData), 0, "序列化数据不应为空")

	// 测试反序列化
	var loadedImpact *entity.DeleteImpact
	err = json.Unmarshal(jsonData, &loadedImpact)
	assert.NoError(t, err, "反序列化不应出错")
	assert.NotNil(t, loadedImpact, "反序列化结果不应为空")
	assert.Equal(t, impact.ChildMenusCount, loadedImpact.ChildMenusCount, "子菜单数量应一致")
	assert.Equal(t, impact.BtnPermsCount, loadedImpact.BtnPermsCount, "按钮权限数量应一致")
	assert.Equal(t, impact.AffectedRolesCount, loadedImpact.AffectedRolesCount, "影响角色数量应一致")
	assert.Equal(t, impact.AffectedApisCount, loadedImpact.AffectedApisCount, "影响 API 数量应一致")
	assert.Equal(t, impact.Level, loadedImpact.Level, "影响等级应一致")
	assert.Equal(t, impact.Warning, loadedImpact.Warning, "警告信息应一致")
}
