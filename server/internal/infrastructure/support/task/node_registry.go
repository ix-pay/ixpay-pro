package task

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// NodeRole 节点角色
type NodeRole string

const (
	NodeRoleAPI  NodeRole = "api"
	NodeRoleTask NodeRole = "task"
	NodeRoleAll  NodeRole = "all"
)

// NodeRegistry 节点注册表
type NodeRegistry struct {
	redis     *redis.Client
	nodeID    string
	role      NodeRole
	keyPrefix string
	stopChan  chan struct{}
}

// NewNodeRegistry 创建节点注册表
func NewNodeRegistry(redis *redis.Client, nodeID string, role NodeRole) *NodeRegistry {
	if nodeID == "" {
		nodeID = generateDefaultNodeID()
	}

	return &NodeRegistry{
		redis:     redis,
		nodeID:    nodeID,
		role:      role,
		keyPrefix: "task:nodes:",
		stopChan:  make(chan struct{}),
	}
}

// GetNodeID 获取节点ID
func (nr *NodeRegistry) GetNodeID() string {
	return nr.nodeID
}

// Register 注册节点
func (nr *NodeRegistry) Register(ctx context.Context) error {
	key := fmt.Sprintf("%s%s", nr.keyPrefix, nr.nodeID)

	ipAddress := getLocalIP()

	data := map[string]interface{}{
		"node_id":        nr.nodeID,
		"role":           string(nr.role),
		"status":         "online",
		"ip_address":     ipAddress,
		"port":           0,
		"running_tasks":  0,
		"max_concurrent": 10,
		"started_at":     time.Now().Format(time.RFC3339),
		"last_heartbeat": time.Now().Format(time.RFC3339),
		"registered_at":  time.Now().Format(time.RFC3339),
	}

	if err := nr.redis.HSet(ctx, key, data).Err(); err != nil {
		return err
	}

	return nr.redis.Expire(ctx, key, 30*time.Second).Err()
}

// StartHeartbeat 启动心跳
func (nr *NodeRegistry) StartHeartbeat(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				nr.heartbeat(ctx)
			case <-ctx.Done():
				nr.Unregister(ctx)
				return
			case <-nr.stopChan:
				nr.Unregister(ctx)
				return
			}
		}
	}()
}

func (nr *NodeRegistry) heartbeat(ctx context.Context) {
	key := fmt.Sprintf("%s%s", nr.keyPrefix, nr.nodeID)
	nr.redis.HSet(ctx, key, "last_heartbeat", time.Now().Format(time.RFC3339))
	nr.redis.Expire(ctx, key, 30*time.Second)
}

// Unregister 注销节点
func (nr *NodeRegistry) Unregister(ctx context.Context) {
	key := fmt.Sprintf("%s%s", nr.keyPrefix, nr.nodeID)
	nr.redis.HSet(ctx, key, "status", "offline")
	nr.redis.Expire(ctx, key, 60*time.Second)
}

// Stop 停止心跳
func (nr *NodeRegistry) Stop() {
	close(nr.stopChan)
}

// GetAllNodes 获取所有节点（包括在线和离线）
func (nr *NodeRegistry) GetAllNodes(ctx context.Context) ([]map[string]string, error) {
	pattern := fmt.Sprintf("%s*", nr.keyPrefix)
	keys, err := nr.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	var nodes []map[string]string
	for _, key := range keys {
		node, err := nr.redis.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		if len(node) > 0 {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// GetNodeById 根据ID获取单个节点
func (nr *NodeRegistry) GetNodeById(ctx context.Context, nodeID string) (map[string]string, error) {
	key := fmt.Sprintf("%s%s", nr.keyPrefix, nodeID)
	node, err := nr.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("获取节点信息失败: %w", err)
	}

	if len(node) == 0 {
		return nil, fmt.Errorf("节点不存在")
	}

	return node, nil
}

// SetNodeStatus 设置节点状态
func (nr *NodeRegistry) SetNodeStatus(ctx context.Context, nodeID string, status string) error {
	key := fmt.Sprintf("%s%s", nr.keyPrefix, nodeID)

	exists, err := nr.redis.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("检查节点状态失败: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("节点不存在")
	}

	if err := nr.redis.HSet(ctx, key, "status", status).Err(); err != nil {
		return fmt.Errorf("设置节点状态失败: %w", err)
	}

	return nil
}

// GetNodeStatistics 获取节点统计信息
func (nr *NodeRegistry) GetNodeStatistics(ctx context.Context) (map[string]interface{}, error) {
	nodes, err := nr.GetAllNodes(ctx)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":   len(nodes),
		"online":  0,
		"offline": 0,
	}

	for _, node := range nodes {
		if node["status"] == "online" {
			stats["online"] = stats["online"].(int) + 1
		} else {
			stats["offline"] = stats["offline"].(int) + 1
		}
	}

	return stats, nil
}

// GetActiveTaskNodes 获取活跃的任务节点
func (nr *NodeRegistry) GetActiveTaskNodes(ctx context.Context) ([]map[string]string, error) {
	pattern := fmt.Sprintf("%s*", nr.keyPrefix)
	keys, err := nr.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var nodes []map[string]string
	for _, key := range keys {
		node, err := nr.redis.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		if node["status"] == "online" && (node["role"] == "task" || node["role"] == "all") {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func generateDefaultNodeID() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}

// ParseIntFromString 从字符串解析整数
func ParseIntFromString(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}

// SetupNodeRegistry 创建节点注册表实例（用于依赖注入）
func SetupNodeRegistry(redis *redis.Client) *NodeRegistry {
	return NewNodeRegistry(redis, "", NodeRoleAll)
}
