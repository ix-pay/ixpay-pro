package repo

import (
	"context"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
)

// NodeRepository 节点仓库接口
type NodeRepository interface {
	GetAllNodes(ctx context.Context) ([]*entity.Node, error)
	GetNodeById(ctx context.Context, nodeID string) (*entity.Node, error)
	SetNodeStatus(ctx context.Context, nodeID string, status string) error
	GetNodeStatistics(ctx context.Context) (*NodeStatistics, error)
	GetActiveTaskNodes(ctx context.Context) ([]*entity.Node, error)
}

// NodeStatistics 节点统计信息
type NodeStatistics struct {
	Total   int `json:"total"`
	Online  int `json:"online"`
	Offline int `json:"offline"`
}
