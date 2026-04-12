package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ixpay-pro/gxy/internal/discovery"
	"github.com/ixpay-pro/gxy/pkg/config"
	"github.com/ixpay-pro/gxy/pkg/utils"
)

type ClusterNode struct {
	ID                string    `json:"id"`
	Address           string    `json:"address"`
	Port              int       `json:"port"`
	NodeDiscoveryPort int       `json:"node_discovery_port"`
	LastSeen          time.Time `json:"last_seen"`
	Status            string    `json:"status"`
}

type ClusterDiscoveryMessage struct {
	Type              string `json:"type"`                // 消息类型：discovery, heartbeat
	NodeID            string `json:"node_id"`             // 节点ID
	NodeAddr          string `json:"node_addr"`           // 节点地址
	NodePort          int    `json:"node_port"`           // 节点端口
	NodeDiscoveryPort int    `json:"node_discovery_port"` // 节点发现服务端口
	Timestamp         int64  `json:"timestamp"`           // 时间戳
}

type DataSyncRequest struct {
	Type         string                                  `json:"type"`      // sync_request, sync_response
	NodeID       string                                  `json:"node_id"`   // 节点ID
	Services     map[string][]*discovery.ServiceInstance `json:"services"`  // 服务注册信息
	ClusterNodes map[string]*ClusterNode                 `json:"nodes"`     // 集群节点信息
	Timestamp    int64                                   `json:"timestamp"` // 时间戳
}

type ClusterSync struct {
	nodes     map[string]*ClusterNode
	registry  *discovery.Registry
	config    *config.Config
	localNode *ClusterNode
	logger    *utils.Logger
	mutex     sync.RWMutex
	client    *http.Client // 复用的HTTP客户端
}

func NewClusterSync(registry *discovery.Registry, config *config.Config, logger *utils.Logger) *ClusterSync {
	// 创建复用的HTTP客户端
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &ClusterSync{
		nodes:    make(map[string]*ClusterNode),
		registry: registry,
		config:   config,
		logger:   logger,
		client:   client,
	}
}

func (cs *ClusterSync) Start() {
	// 1. 初始化本地节点信息
	cs.localNode = &ClusterNode{
		ID:                generateNodeID(),
		Address:           getLocalIP(),
		Port:              cs.config.ListenPort,
		NodeDiscoveryPort: cs.config.NodeDiscoveryPort,
		LastSeen:          time.Now(),
		Status:            "online",
	}

	// 2. 启动自动发现服务
	if cs.config.EnableAutoDiscovery {
		cs.startAutoDiscovery()
	}

	// 3. 启动心跳检测
	cs.startHeartbeatMonitor()

	// 4. 启动数据同步HTTP服务
	cs.startSyncHTTPServer()

	cs.logger.Info("Cluster sync service started. Local node: %s", cs.localNode.ID)
}

// 生成节点ID（不依赖第三方库）
func generateNodeID() string {
	// 使用时间戳 + 随机数生成唯一ID
	timestamp := time.Now().UnixNano()
	random := rand.Int63()
	id := fmt.Sprintf("gw-%x-%x", timestamp, random)
	// 取前16个字符作为节点ID
	if len(id) > 16 {
		id = id[:16]
	}
	return id
}

// 获取本地IP地址
func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addresses {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return "127.0.0.1"
}

// 启动自动发现实现（基于种子节点）
func (cs *ClusterSync) startAutoDiscovery() {
	// 1. 初始化本地节点发现端口
	cs.localNode.NodeDiscoveryPort = cs.config.NodeDiscoveryPort

	// 2. 连接到种子节点
	go func() {
		// 等待集群同步服务完全初始化
		time.Sleep(time.Second * 2)

		// 连接到所有种子节点
		for _, seedNode := range cs.config.SeedNodes {
			if seedNode == "" {
				continue
			}

			// 解析种子节点地址
			seedAddr, err := net.ResolveTCPAddr("tcp", seedNode)
			if err != nil {
				cs.logger.Error("Failed to resolve seed node address %s: %v", seedNode, err)
				continue
			}

			// 构建种子节点对象
			seedNodeObj := &ClusterNode{
				ID:                fmt.Sprintf("seed-%s", seedAddr.String()),
				Address:           seedAddr.IP.String(),
				Port:              seedAddr.Port,
				NodeDiscoveryPort: seedAddr.Port, // 假设种子节点的发现端口与服务端口相同
				LastSeen:          time.Now(),
				Status:            "online",
			}

			// 向种子节点注册自己
			cs.registerToSeedNode(seedNodeObj)
		}
	}()

	// 3. 启动定期扫描任务
	go func() {
		ticker := time.NewTicker(time.Second * 15) // 每15秒扫描一次
		defer ticker.Stop()

		for range ticker.C {
			cs.scanClusterNodes()
		}
	}()

	cs.logger.Info("Auto discovery started with seed nodes: %v", cs.config.SeedNodes)
}

// 发送HTTP请求（使用复用的客户端）
func (cs *ClusterSync) makeHTTPRequest(method, url string, body interface{}) (*http.Response, error) {
	var requestBody []byte
	var err error

	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	return cs.client.Do(req)
}

func (cs *ClusterSync) registerToSeedNode(seedNode *ClusterNode) {
	// 构建节点注册消息
	registerMsg := &ClusterDiscoveryMessage{
		Type:              "discovery",
		NodeID:            cs.localNode.ID,
		NodeAddr:          cs.localNode.Address,
		NodePort:          cs.localNode.Port,
		NodeDiscoveryPort: cs.localNode.NodeDiscoveryPort,
		Timestamp:         time.Now().Unix(),
	}

	// 发送注册请求到种子节点
	url := fmt.Sprintf("http://%s:%d/node/register", seedNode.Address, seedNode.Port)

	resp, err := cs.makeHTTPRequest("POST", url, registerMsg)
	if err != nil {
		cs.logger.Error("Failed to register to seed node %s: %v", seedNode.Address, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		cs.logger.Info("Successfully registered to seed node %s", seedNode.Address)
		// 同步数据
		cs.syncWithNewNode(seedNode)
	} else {
		cs.logger.Error("Failed to register to seed node %s, status code: %d", seedNode.Address, resp.StatusCode)
	}
}

// 与新节点同步数据
func (cs *ClusterSync) syncWithNewNode(newNode *ClusterNode) {
	// 构建同步请求
	syncReq := &DataSyncRequest{
		Type:      "sync_request",
		NodeID:    cs.localNode.ID,
		Timestamp: time.Now().Unix(),
	}

	// 发送同步请求到新节点
	resp, err := cs.sendSyncRequest(newNode, syncReq)
	if err != nil {
		cs.logger.Error("Failed to send sync request to node %s: %v", newNode.ID, err)
		return
	}

	// 处理同步响应
	if resp.Type == "sync_response" {
		cs.applySyncData(resp)
	}
}

// 发送同步请求
func (cs *ClusterSync) sendSyncRequest(node *ClusterNode, req *DataSyncRequest) (*DataSyncRequest, error) {
	// 构建同步请求URL
	url := fmt.Sprintf("http://%s:%d/sync", node.Address, node.Port)

	// 发送HTTP请求
	resp, err := cs.makeHTTPRequest("POST", url, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var syncResp DataSyncRequest
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, err
	}

	return &syncResp, nil
}

// 应用同步数据
func (cs *ClusterSync) applySyncData(resp *DataSyncRequest) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// 同步集群节点信息
	for nodeID, node := range resp.ClusterNodes {
		// 忽略自己
		if nodeID == cs.localNode.ID {
			continue
		}

		// 更新或添加节点
		cs.nodes[nodeID] = node
	}

	// 同步服务注册信息
	for _, instances := range resp.Services {
		for _, inst := range instances {
			cs.registry.Register(inst)
		}
	}

	cs.logger.Info("Successfully synced data from cluster")
}

// 启动心跳监控
func (cs *ClusterSync) startHeartbeatMonitor() {
	go func() {
		ticker := time.NewTicker(cs.config.HeartbeatInterval)
		defer ticker.Stop()

		for range ticker.C {
			cs.mutex.Lock()
			// 检查所有节点的心跳
			now := time.Now()
			for nodeID, node := range cs.nodes {
				// 如果节点超过3倍心跳间隔没有心跳，标记为离线
				if now.Sub(node.LastSeen) > cs.config.HeartbeatInterval*3 {
					node.Status = "offline"
					cs.logger.Info("Node %s marked as offline due to heartbeat timeout", nodeID)

					// 从节点列表中移除
					delete(cs.nodes, nodeID)
				}
			}
			cs.mutex.Unlock()
		}
	}()

	cs.logger.Info("Heartbeat monitor started with interval %v", cs.config.HeartbeatInterval)
}

// 启动同步HTTP服务
func (cs *ClusterSync) startSyncHTTPServer() {
	// 处理集群数据同步
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		// 解析同步请求
		var syncReq DataSyncRequest
		if err := json.NewDecoder(r.Body).Decode(&syncReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 构建同步响应
		syncResp := &DataSyncRequest{
			Type:         "sync_response",
			NodeID:       cs.localNode.ID,
			Services:     cs.registry.GetAllServices(),
			ClusterNodes: cs.nodes,
			Timestamp:    time.Now().Unix(),
		}

		// 序列化响应
		data, err := json.Marshal(syncResp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 发送响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)

		// 应用接收到的同步数据
		cs.applySyncData(&syncReq)
	})

	// 获取当前节点信息
	http.HandleFunc("/node/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cs.localNode)
	})

	// 获取集群节点列表
	http.HandleFunc("/node/list", func(w http.ResponseWriter, r *http.Request) {
		cs.mutex.RLock()
		defer cs.mutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cs.nodes)
	})

	// 注册新节点
	http.HandleFunc("/node/register", func(w http.ResponseWriter, r *http.Request) {
		// 解析节点注册请求
		var registerMsg ClusterDiscoveryMessage
		if err := json.NewDecoder(r.Body).Decode(&registerMsg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 处理节点发现
		cs.handleNodeDiscovery(&registerMsg, nil)

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}

func (cs *ClusterSync) handleNodeDiscovery(msg *ClusterDiscoveryMessage, _ net.Addr) {
	newNode := &ClusterNode{
		ID:                msg.NodeID,
		Address:           msg.NodeAddr,
		Port:              msg.NodePort,
		NodeDiscoveryPort: msg.NodeDiscoveryPort,
		LastSeen:          time.Now(),
		Status:            "online",
	}

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// 忽略自己
	if newNode.ID == cs.localNode.ID {
		return
	}

	// 检查节点是否已存在
	if _, exists := cs.nodes[newNode.ID]; !exists {
		cs.nodes[newNode.ID] = newNode
		cs.logger.Info("Discovered new cluster node: %s at %s:%d", newNode.ID, newNode.Address, newNode.Port)

		// 同步数据
		go cs.syncWithNewNode(newNode)
	} else {
		// 更新节点信息
		cs.nodes[newNode.ID].LastSeen = time.Now()
		cs.nodes[newNode.ID].Status = "online"
	}
}

func (cs *ClusterSync) scanClusterNodes() {
	// 这个函数会定期检查所有节点的状态
	// 目前的实现是通过种子节点发现，所以这里可以留空
	// 或者添加额外的节点发现逻辑
}

// 暂时保留，用于后续功能扩展
/*
func (cs *ClusterSync) broadcastServiceUpdate(instance *discovery.ServiceInstance, action string) {
	// 向所有集群节点广播服务更新
	cs.mutex.RLock()
	nodes := make(map[string]*ClusterNode, len(cs.nodes))
	for id, node := range cs.nodes {
		nodes[id] = node
	}
	cs.mutex.RUnlock()

	for _, node := range nodes {
		if node.ID == cs.localNode.ID || node.Status != "online" {
			continue
		}

		// 这里可以实现向其他节点广播服务更新的逻辑
		// 例如发送HTTP请求通知其他节点服务的变化
	}
}*/

// 暂时保留，用于后续功能扩展
/*
func (cs *ClusterSync) handleNodeHeartbeat(node *ClusterNode) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if existingNode, exists := cs.nodes[node.ID]; exists {
		existingNode.LastSeen = time.Now()
		existingNode.Status = "online"
	}
}*/
