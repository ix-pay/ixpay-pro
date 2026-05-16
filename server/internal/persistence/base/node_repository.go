package persistence

import (
	"context"
	"strconv"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
)

// nodeRepository 节点仓库实现
type nodeRepository struct {
	nodeRegistry *task.NodeRegistry
}

// NewNodeRepository 创建节点仓库
func NewNodeRepository(nodeRegistry *task.NodeRegistry) repo.NodeRepository {
	return &nodeRepository{
		nodeRegistry: nodeRegistry,
	}
}

// GetAllNodes 获取所有节点
func (r *nodeRepository) GetAllNodes(ctx context.Context) ([]*entity.Node, error) {
	nodesData, err := r.nodeRegistry.GetAllNodes(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*entity.Node, 0, len(nodesData))
	for _, data := range nodesData {
		node := r.toDomain(data)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetNodeById 根据ID获取节点
func (r *nodeRepository) GetNodeById(ctx context.Context, nodeID string) (*entity.Node, error) {
	data, err := r.nodeRegistry.GetNodeById(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	return r.toDomain(data), nil
}

// SetNodeStatus 设置节点状态
func (r *nodeRepository) SetNodeStatus(ctx context.Context, nodeID string, status string) error {
	return r.nodeRegistry.SetNodeStatus(ctx, nodeID, status)
}

// GetNodeStatistics 获取节点统计
func (r *nodeRepository) GetNodeStatistics(ctx context.Context) (*repo.NodeStatistics, error) {
	statsData, err := r.nodeRegistry.GetNodeStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return &repo.NodeStatistics{
		Total:   statsData["total"].(int),
		Online:  statsData["online"].(int),
		Offline: statsData["offline"].(int),
	}, nil
}

// GetActiveTaskNodes 获取活跃任务节点
func (r *nodeRepository) GetActiveTaskNodes(ctx context.Context) ([]*entity.Node, error) {
	nodesData, err := r.nodeRegistry.GetActiveTaskNodes(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*entity.Node, 0, len(nodesData))
	for _, data := range nodesData {
		node := r.toDomain(data)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// toDomain 将 Redis 数据转换为领域实体
func (r *nodeRepository) toDomain(data map[string]string) *entity.Node {
	node := &entity.Node{
		NodeID:    data["node_id"],
		Role:      data["role"],
		Status:    data["status"],
		IPAddress: data["ip_address"],
	}

	if port, err := strconv.Atoi(data["port"]); err == nil {
		node.Port = port
	}

	if runningTasks, err := strconv.Atoi(data["running_tasks"]); err == nil {
		node.RunningTasks = runningTasks
	}

	if maxConcurrent, err := strconv.Atoi(data["max_concurrent"]); err == nil {
		node.MaxConcurrent = maxConcurrent
	}

	if lastHeartbeat, err := time.Parse(time.RFC3339, data["last_heartbeat"]); err == nil {
		node.LastHeartbeat = lastHeartbeat
	}

	if registeredAt, err := time.Parse(time.RFC3339, data["registered_at"]); err == nil {
		node.RegisteredAt = registeredAt
	}

	if startedAt, err := time.Parse(time.RFC3339, data["started_at"]); err == nil {
		node.StartedAt = startedAt
	}

	return node
}

// ensure interface implementation
var _ repo.NodeRepository = (*nodeRepository)(nil)
