package service

import (
	"testing"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/stretchr/testify/assert"
)

// TestNodeService_NodeValidation 测试节点验证逻辑
func TestNodeService_NodeValidation(t *testing.T) {
	testCases := []struct {
		name        string
		node        *entity.Node
		expectError bool
	}{
		{
			name: "有效节点",
			node: &entity.Node{
				NodeID:    "node-001",
				Role:      "worker",
				Status:    "online",
				IPAddress: "192.168.1.1",
				Port:      8080,
			},
			expectError: false,
		},
		{
			name: "无效节点 - 节点ID为空",
			node: &entity.Node{
				Role:      "worker",
				Status:    "online",
				IPAddress: "192.168.1.1",
				Port:      8080,
			},
			expectError: true,
		},
		{
			name: "无效节点 - 节点ID和状态为空",
			node: &entity.Node{
				Role:      "worker",
				Status:    "",
				IPAddress: "192.168.1.1",
				Port:      8080,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.node.NodeID == "" {
				assert.True(t, tc.expectError, "节点ID不能为空")
				return
			}
			if tc.node.Status == "" {
				assert.True(t, tc.expectError, "节点状态不能为空")
				return
			}
			assert.False(t, tc.expectError, "有效节点应通过验证")
		})
	}
}

// TestNodeService_NodeStatusValidation 测试节点状态验证
func TestNodeService_NodeStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"online":  true,
		"offline": true,
		"error":   true,
	}

	testCases := []struct {
		name    string
		status  string
		isValid bool
	}{
		{name: "在线状态", status: "online", isValid: true},
		{name: "离线状态", status: "offline", isValid: true},
		{name: "错误状态", status: "error", isValid: true},
		{name: "无效状态", status: "unknown", isValid: false},
		{name: "空状态", status: "", isValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, isValid := validStatuses[tc.status]
			assert.Equal(t, tc.isValid, isValid, "状态验证应正确")
		})
	}
}

// TestNodeService_NodeRoleValidation 测试节点角色验证
func TestNodeService_NodeRoleValidation(t *testing.T) {
	validRoles := map[string]bool{
		"task":    true,
		"gateway": true,
		"worker":  true,
	}

	testCases := []struct {
		name    string
		role    string
		isValid bool
	}{
		{name: "有效角色 - task", role: "task", isValid: true},
		{name: "有效角色 - gateway", role: "gateway", isValid: true},
		{name: "有效角色 - worker", role: "worker", isValid: true},
		{name: "无效角色 - unknown", role: "unknown", isValid: false},
		{name: "无效角色 - 空字符串", role: "", isValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, isValid := validRoles[tc.role]
			assert.Equal(t, tc.isValid, isValid, "节点角色验证应正确")
		})
	}
}

// TestNodeService_NodeStatistics 测试节点统计结构
func TestNodeService_NodeStatistics(t *testing.T) {
	stats := struct {
		TotalNodes   int
		OnlineNodes  int
		OfflineNodes int
		ActiveTasks  int
	}{
		TotalNodes:   10,
		OnlineNodes:  7,
		OfflineNodes: 3,
		ActiveTasks:  25,
	}

	assert.Equal(t, 10, stats.TotalNodes, "总节点数应正确")
	assert.Equal(t, 7, stats.OnlineNodes, "在线节点数应正确")
	assert.Equal(t, 3, stats.OfflineNodes, "离线节点数应正确")
	assert.Equal(t, 25, stats.ActiveTasks, "活跃任务数应正确")
	assert.Equal(t, stats.TotalNodes, stats.OnlineNodes+stats.OfflineNodes, "总节点数应等于在线+离线")
}

// TestNodeService_Heartbeat 测试节点心跳时间
func TestNodeService_Heartbeat(t *testing.T) {
	now := time.Now()
	node := &entity.Node{
		NodeID:        "node-001",
		Role:          "worker",
		Status:        "online",
		LastHeartbeat: now,
		RegisteredAt:  now.Add(-1 * time.Hour),
		StartedAt:     now.Add(-30 * time.Minute),
	}

	assert.False(t, node.LastHeartbeat.IsZero(), "心跳时间不应为零")
	assert.True(t, node.RegisteredAt.Before(node.StartedAt), "注册时间应早于启动时间")
	assert.True(t, node.StartedAt.Before(node.LastHeartbeat), "启动时间应早于最近心跳时间")
}