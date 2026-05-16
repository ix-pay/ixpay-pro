package baseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// NodeController 节点管理控制器
type NodeController struct {
	nodeService *service.NodeService
	log         logger.Logger
}

// NewNodeController 创建节点控制器
func NewNodeController(
	nodeService *service.NodeService,
	log logger.Logger,
) *NodeController {
	return &NodeController{
		nodeService: nodeService,
		log:         log,
	}
}

// GetNodeList 获取节点列表
func (c *NodeController) GetNodeList(ctx *gin.Context) {
	nodes, err := c.nodeService.GetNodeList(ctx.Request.Context())
	if err != nil {
		c.log.Error("获取节点列表失败", "error", err)
		baseRes.FailWithMessage("获取节点列表失败", ctx)
		return
	}

	response := make([]NodeResponse, 0, len(nodes))
	for _, node := range nodes {
		response = append(response, c.toNodeResponse(node))
	}

	baseRes.OkWithDetailed(response, "获取节点列表成功", ctx)
}

// GetNodeById 获取节点详情
func (c *NodeController) GetNodeById(ctx *gin.Context) {
	nodeID := ctx.Param("id")
	if nodeID == "" {
		baseRes.FailWithMessage("节点ID不能为空", ctx)
		return
	}

	node, err := c.nodeService.GetNodeById(ctx.Request.Context(), nodeID)
	if err != nil {
		baseRes.FailWithMessage("获取节点信息失败", ctx)
		return
	}

	baseRes.OkWithDetailed(c.toNodeResponse(node), "获取节点详情成功", ctx)
}

// OfflineNode 下线节点
func (c *NodeController) OfflineNode(ctx *gin.Context) {
	nodeID := ctx.Param("id")
	if nodeID == "" {
		baseRes.FailWithMessage("节点ID不能为空", ctx)
		return
	}

	if err := c.nodeService.OfflineNode(ctx.Request.Context(), nodeID); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("节点下线成功", ctx)
}

// GetNodeStatistics 获取节点统计
func (c *NodeController) GetNodeStatistics(ctx *gin.Context) {
	stats, err := c.nodeService.GetNodeStatistics(ctx.Request.Context())
	if err != nil {
		c.log.Error("获取节点统计失败", "error", err)
		baseRes.FailWithMessage("获取节点统计失败", ctx)
		return
	}

	baseRes.OkWithDetailed(stats, "获取节点统计成功", ctx)
}

// NodeResponse 节点响应
type NodeResponse struct {
	NodeID        string `json:"nodeId"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	IPAddress     string `json:"ipAddress"`
	Port          int    `json:"port"`
	RunningTasks  int    `json:"runningTasks"`
	MaxConcurrent int    `json:"maxConcurrent"`
	LastHeartbeat string `json:"lastHeartbeat"`
	RegisteredAt  string `json:"registeredAt"`
	StartedAt     string `json:"startedAt"`
}

// toNodeResponse 将领域实体转换为响应DTO
func (c *NodeController) toNodeResponse(node *entity.Node) NodeResponse {
	return NodeResponse{
		NodeID:        node.NodeID,
		Role:          node.Role,
		Status:        node.Status,
		IPAddress:     node.IPAddress,
		Port:          node.Port,
		RunningTasks:  node.RunningTasks,
		MaxConcurrent: node.MaxConcurrent,
		LastHeartbeat: node.LastHeartbeat.Format("2006-01-02 15:04:05"),
		RegisteredAt:  node.RegisteredAt.Format("2006-01-02 15:04:05"),
		StartedAt:     node.StartedAt.Format("2006-01-02 15:04:05"),
	}
}
