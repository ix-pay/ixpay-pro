package service

import (
	"context"
	"fmt"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// NodeService 节点服务
type NodeService struct {
	repo repo.NodeRepository
	log  logger.Logger
}

// NewNodeService 创建节点服务实例
func NewNodeService(repo repo.NodeRepository, log logger.Logger) *NodeService {
	return &NodeService{
		repo: repo,
		log:  log,
	}
}

// GetNodeList 获取节点列表
func (s *NodeService) GetNodeList(ctx context.Context) ([]*entity.Node, error) {
	nodes, err := s.repo.GetAllNodes(ctx)
	if err != nil {
		s.log.Error("获取节点列表失败", "error", err)
		return nil, err
	}

	return nodes, nil
}

// GetNodeById 根据ID获取节点
func (s *NodeService) GetNodeById(ctx context.Context, nodeID string) (*entity.Node, error) {
	node, err := s.repo.GetNodeById(ctx, nodeID)
	if err != nil {
		s.log.Error("获取节点信息失败", "node_id", nodeID, "error", err)
		return nil, err
	}

	return node, nil
}

// OfflineNode 下线节点
func (s *NodeService) OfflineNode(ctx context.Context, nodeID string) error {
	node, err := s.repo.GetNodeById(ctx, nodeID)
	if err != nil {
		s.log.Error("节点不存在", "node_id", nodeID, "error", err)
		return fmt.Errorf("节点不存在")
	}

	if node.Status != "online" {
		return fmt.Errorf("节点当前不在线，无法下线")
	}

	if err := s.repo.SetNodeStatus(ctx, nodeID, "offline"); err != nil {
		s.log.Error("下线节点失败", "node_id", nodeID, "error", err)
		return fmt.Errorf("下线节点失败")
	}

	s.log.Info("节点下线成功", "node_id", nodeID)
	return nil
}

// GetNodeStatistics 获取节点统计信息
func (s *NodeService) GetNodeStatistics(ctx context.Context) (*repo.NodeStatistics, error) {
	stats, err := s.repo.GetNodeStatistics(ctx)
	if err != nil {
		s.log.Error("获取节点统计失败", "error", err)
		return nil, err
	}

	return stats, nil
}

// GetActiveTaskNodes 获取活跃任务节点
func (s *NodeService) GetActiveTaskNodes(ctx context.Context) ([]*entity.Node, error) {
	nodes, err := s.repo.GetActiveTaskNodes(ctx)
	if err != nil {
		s.log.Error("获取活跃任务节点失败", "error", err)
		return nil, err
	}

	return nodes, nil
}
