# Go语言网关开发提示词

## 项目概述
使用纯Go语言开发一个轻量级API网关，不依赖任何第三方框架，实现服务注册发现、负载均衡、健康检查和网关集群数据同步功能。

## 核心功能需求

### 1. 服务注册与发现
- 服务实例启动时自动向网关注册服务地址
- 支持同一服务的多个实例注册
- 网关维护服务注册列表

### 2. 负载均衡
- 采用轮询(Round Robin)算法分发请求
- 网关接收客户端请求并转发到后端服务

### 3. 健康检查
- 定时检查注册服务的健康状态
- 自动剔除不健康的服务实例

### 4. 网关集群
- 无主模式部署多个网关实例
- 网关间自动同步服务注册数据
- 故障网关自动从集群中剔除

## 技术架构设计

### 项目目录结构
```
gateway/
├── cmd/
│   └── gateway/
│       └── main.go           # 入口文件
├── internal/
│   ├── api/
│   │   ├── handler.go        # API请求处理器
│   │   └── router.go         # 请求路由
│   ├── cluster/
│   │   ├── node.go           # 网关节点管理
│   │   └── sync.go           # 数据同步
│   ├── discovery/
│   │   ├── registry.go       # 服务注册中心
│   │   └── health.go         # 健康检查
│   ├── loadbalance/
│   │   └── roundrobin.go     # 轮询负载均衡
│   └── proxy/
│       └── proxy.go          # 请求代理转发
├── pkg/
│   ├── config/
│   │   └── config.go         # 配置管理
│   └── utils/
│       ├── http.go           # HTTP工具
│       └── log.go            # 日志工具
└── go.mod
```

### 核心组件设计

#### 服务注册中心 (Registry)
- 数据结构：使用map存储服务名称到服务实例列表的映射，同时维护实例ID到实例的直接映射以支持O(1)查找
- 提供注册、注销、查询服务实例的接口
- 线程安全设计（使用sync.RWMutex保护共享数据）
- **性能优化**：
  - 添加instances map[string]*ServiceInstance实现O(1)时间复杂度的实例查找
  - 优化UpdateConnectionCount方法从O(n*m)提升到O(1)

```go
// internal/discovery/registry.go
type ServiceInstance struct {
    ID        string
    Name      string
    Address   string
    Port      int
    Metadata  map[string]string
    LastSeen  time.Time
    ActiveConnections int // 当前活跃连接数
}

type Registry struct {
    services map[string][]*ServiceInstance
    instances map[string]*ServiceInstance // instanceID -> ServiceInstance映射，支持O(1)查找
    mutex    sync.RWMutex
}

func (r *Registry) Register(instance *ServiceInstance)
func (r *Registry) Deregister(serviceName, instanceID string)
func (r *Registry) GetInstances(serviceName string) []*ServiceInstance
func (r *Registry) UpdateConnectionCount(instanceID string, delta int) // O(1)时间复杂度实现
```

#### 健康检查 (Health Checker)
- 定期发送HTTP请求检查服务健康
- 支持可配置的检查间隔和超时时间
- 自动移除不健康的服务实例
- **性能优化**：
  - 实现并发健康检查，支持可配置的并发限制
  - 使用复用的HTTP客户端减少资源消耗
  - 采用信号量控制并发，避免资源耗尽

```go
// internal/discovery/health.go
type HealthChecker struct {
    registry *Registry
    interval time.Duration
    timeout  time.Duration
    logger   *utils.Logger
    client   *http.Client // 复用的HTTP客户端
    concurrencyLimit int   // 并发检查限制
}

func (hc *HealthChecker) Start()
func (hc *HealthChecker) checkAllServices() // 并发检查所有服务实例
func (hc *HealthChecker) checkService(instance *ServiceInstance) bool
```

#### 负载均衡 (Round Robin)
- 实现轮询算法选择服务实例
- 考虑服务健康状态
- 线程安全实现
- **性能优化**：
  - 使用读写锁(RWMutex)减少锁竞争
  - 简化算法实现，降低时间复杂度
  - 减少不必要的内存分配和计算
  - 采用细粒度锁策略，读操作使用读锁，写操作使用写锁

```go
// internal/loadbalance/roundrobin.go
type RoundRobinBalancer struct {
    currentIndex map[string]int
    mutex        sync.RWMutex
    connectionThreshold int // 连接数阈值，超过此值的实例将被暂时排除在轮询之外
}

// NewRoundRobinBalancer 创建新的轮询负载均衡器
func NewRoundRobinBalancer(connectionThreshold int) *RoundRobinBalancer {
    return &RoundRobinBalancer{
        currentIndex: make(map[string]int),
        connectionThreshold: connectionThreshold,
    }
}

func (rb *RoundRobinBalancer) Select(serviceName string, instances []*ServiceInstance) *ServiceInstance
```

#### 网关集群同步
- 使用HTTP实现网关间通信
- 采用种子节点发现机制
- 节点心跳检测，自动剔除故障节点
- **性能优化**：
  - 使用复用的HTTP客户端减少资源消耗
  - 自定义makeHTTPRequest方法优化请求处理
  - 减少内存分配和网络连接开销
  - 实现高效的数据同步机制

```go
// internal/cluster/node.go
type ClusterNode struct {
    ID        string
    Address   string
    Port      int
    NodeDiscoveryPort int
    LastSeen  time.Time
    Status    string
}

// 集群发现消息结构
type ClusterDiscoveryMessage struct {
    Type      string `json:"type"`      // 消息类型：discovery, heartbeat
    NodeID    string `json:"node_id"`   // 节点ID
    NodeAddr  string `json:"node_addr"` // 节点地址
    NodePort  int    `json:"node_port"` // 节点端口
    NodeDiscoveryPort int `json:"node_discovery_port"` // 节点发现服务端口
    Timestamp int64  `json:"timestamp"` // 时间戳
}

// 数据同步请求结构
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

// 与新节点同步数据
func (cs *ClusterSync) syncWithNewNode(newNode *ClusterNode) {
    // 构建同步请求
    syncReq := &DataSyncRequest{
        Type:         "sync_request",
        NodeID:       cs.localNode.ID,
        Timestamp:    time.Now().Unix(),
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
    
    // 序列化请求数据
    data, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    // 发送HTTP请求
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
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
    cs.registry.mutex.Lock()
    defer cs.registry.mutex.Unlock()
    
    for serviceName, instances := range resp.Services {
        // 合并服务实例
        existingInstances, exists := cs.registry.services[serviceName]
        if !exists {
            cs.registry.services[serviceName] = instances
            continue
        }
        
        // 合并实例，去重
        mergedInstances := make([]*ServiceInstance, 0, len(existingInstances)+len(instances))
        instanceMap := make(map[string]bool)
        
        // 添加现有实例
        for _, inst := range existingInstances {
            mergedInstances = append(mergedInstances, inst)
            instanceMap[inst.ID] = true
        }
        
        // 添加新实例
        for _, inst := range instances {
            if !instanceMap[inst.ID] {
                mergedInstances = append(mergedInstances, inst)
                instanceMap[inst.ID] = true
            }
        }
        
        cs.registry.services[serviceName] = mergedInstances
    }
    
    cs.logger.Info("Successfully synced data from cluster")
}

// 启动集群同步服务
func (cs *ClusterSync) Start() {
    // 1. 初始化本地节点信息
    cs.localNode = &ClusterNode{
        ID:       generateNodeID(),
        Address:  getLocalIP(),
        Port:     cs.config.ListenPort,
        LastSeen: time.Now(),
        Status:   "online",
    }
    
    // 2. 启动自动发现服务
    cs.startAutoDiscovery()
    
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

// 启动心跳监控
func (cs *ClusterSync) startHeartbeatMonitor() {
    go func() {
        ticker := time.NewTicker(cs.config.HeartbeatInterval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
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
        }
    }()
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
            Services:     cs.registry.services,
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
        
        // 创建新节点
        newNode := &ClusterNode{
            ID:                registerMsg.NodeID,
            Address:           registerMsg.NodeAddr,
            Port:              registerMsg.NodePort,
            NodeDiscoveryPort: registerMsg.NodeDiscoveryPort,
            LastSeen:          time.Now(),
            Status:            "online",
        }
        
        // 处理节点发现
        cs.handleNodeDiscovery(&registerMsg, nil)
        
        // 返回成功响应
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })
}

func (cs *ClusterSync) broadcastServiceUpdate(instance *ServiceInstance, action string)
func (cs *ClusterSync) handleNodeHeartbeat(node *ClusterNode)
```

#### 请求代理转发
- 实现HTTP请求的接收和转发
- 处理请求头、请求体的转发
- 处理响应的返回
- **性能优化**：
  - 使用复用的HTTP客户端减少资源消耗
  - 优化HTTP连接池配置（MaxIdleConns、MaxIdleConnsPerHost等）
  - 减少内存分配和网络连接开销
  - 配置合理的超时参数

```go
// internal/proxy/proxy.go
type Proxy struct {
    registry *discovery.Registry
    balancer *loadbalance.RoundRobinBalancer
    logger   *utils.Logger
    client   *http.Client // 复用的HTTP客户端
}

func NewProxy(registry *discovery.Registry, balancer *loadbalance.RoundRobinBalancer, logger *utils.Logger) *Proxy
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (p *Proxy) forwardRequest(instance *discovery.ServiceInstance, w http.ResponseWriter, r *http.Request)
```

#### HTTP工具函数
- 提供通用的HTTP请求发送和响应解析功能
- 支持JSON格式的请求和响应处理
- **性能优化**：
  - 使用全局复用的HTTP客户端替代每次创建新客户端
  - 优化连接池配置，减少资源消耗
  - 减少内存分配和网络连接开销
  - 配置合理的超时参数，避免长时间阻塞

```go
// pkg/utils/http.go
// 全局复用的HTTP客户端，避免每次请求创建新客户端
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 20,
        IdleConnTimeout:     90 * time.Second,
    },
}

func MakeHTTPRequest(method, url string, headers map[string]string, body interface{}) (*http.Response, error)
func ParseHTTPResponse(resp *http.Response, v interface{}) error
```

## 核心功能实现要点

### 1. 服务注册机制
- 提供HTTP接口供服务实例注册
- **添加认证机制**：使用密码或密钥验证服务合法性
- 服务实例定期发送心跳保持注册状态
- 支持服务元数据注册

```go
// 服务注册HTTP接口
func handleServiceRegister(w http.ResponseWriter, r *http.Request) {
    // 1. 验证注册凭证
    authHeader := r.Header.Get("Authorization")
    if !validateRegistrationAuth(authHeader) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // 2. 解析服务注册请求
    // 3. 注册服务实例
    // 4. 返回注册结果
}

// 验证注册凭证
func validateRegistrationAuth(authHeader string) bool {
    // 从配置中获取预设的注册密码/密钥
    // 验证请求中的凭证是否匹配
    // 返回验证结果
}
```

### 2. 负载均衡算法
- 轮询算法实现，确保请求均匀分发
- 跳过不健康的服务实例
- 处理服务实例列表为空的边界情况

```go
// 轮询选择实现 - 结合最少连接优先策略
func (rb *RoundRobinBalancer) Select(serviceName string, instances []*ServiceInstance) *ServiceInstance {
    rb.mutex.Lock()
    defer rb.mutex.Unlock()
    
    // 过滤健康实例
    var healthyInstances []*ServiceInstance
    for _, inst := range instances {
        // 这里假设健康检查已经完成，只需要筛选出健康的实例
        healthyInstances = append(healthyInstances, inst)
    }
    
    if len(healthyInstances) == 0 {
        return nil
    }
    
    // 计算实例之间的连接数差异
    if len(healthyInstances) > 1 {
        // 找出连接数最少和最多的实例
        minConn := healthyInstances[0].ActiveConnections
        maxConn := healthyInstances[0].ActiveConnections
        for _, inst := range healthyInstances {
            if inst.ActiveConnections < minConn {
                minConn = inst.ActiveConnections
            }
            if inst.ActiveConnections > maxConn {
                maxConn = inst.ActiveConnections
            }
        }
        
        // 判断连接数是否相差一个数量级别（如100 vs 1000，相差一个数量级）
        // 使用对数来判断数量级别差异
        if minConn > 0 {
            minLevel := int(math.Log10(float64(minConn)))
            maxLevel := int(math.Log10(float64(maxConn)))
            
            if maxLevel - minLevel >= 1 {
                // 存在数量级别差异，只选择低连接数的实例
                var lowConnInstances []*ServiceInstance
                for _, inst := range healthyInstances {
                    if int(math.Log10(float64(inst.ActiveConnections))) == minLevel {
                        lowConnInstances = append(lowConnInstances, inst)
                    }
                }
                
                if len(lowConnInstances) > 0 {
                    healthyInstances = lowConnInstances
                }
            }
        } else {
            // 如果有实例连接数为0，优先选择这些实例
            var zeroConnInstances []*ServiceInstance
            for _, inst := range healthyInstances {
                if inst.ActiveConnections == 0 {
                    zeroConnInstances = append(zeroConnInstances, inst)
                }
            }
            
            if len(zeroConnInstances) > 0 {
                healthyInstances = zeroConnInstances
            }
        }
    }
    
    // 轮询选择实例
    index, exists := rb.currentIndex[serviceName]
    if !exists {
        index = 0
    }
    
    // 计算下一个索引
    selectedInstance := healthyInstances[index]
    rb.currentIndex[serviceName] = (index + 1) % len(healthyInstances)
    
    return selectedInstance
}
```

### 3. 健康检查实现
- 使用goroutine定期执行检查
- 支持可配置的检查路径（如/health）
- 检查失败达到阈值后标记服务不健康

```go
// 健康检查实现
func (hc *HealthChecker) checkService(instance *ServiceInstance) bool {
    // 构建健康检查URL
    // 发送HTTP请求
    // 检查响应状态码
    // 返回健康状态
}
```

### 4. 网关集群同步
- 节点发现机制（通过配置或自动发现）
- 使用UDP广播服务注册信息
- 节点心跳检测，超时剔除

#### 自动发现实现（基于种子节点）
```go
// 自动发现实现
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
        
        for {
            select {
            case <-ticker.C:
                cs.scanClusterNodes()
            }
        }
    }()
    
    
}

// 向种子节点注册
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
    
    // 构建注册请求URL
    url := fmt.Sprintf("http://%s:%d/node/register", seedNode.Address, seedNode.NodeDiscoveryPort)
    
    // 发送注册请求
    data, _ := json.Marshal(registerMsg)
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
    if err != nil {
        cs.logger.Error("Failed to register to seed node %s: %v", seedNode.Address, err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        cs.logger.Error("Failed to register to seed node %s: %s", seedNode.Address, resp.Status)
        return
    }
    
    // 注册成功后，获取种子节点的节点列表并同步数据
    cs.getNodeListFromSeed(seedNode)
}

// 从种子节点获取节点列表
func (cs *ClusterSync) getNodeListFromSeed(seedNode *ClusterNode) {
    // 构建请求URL
    url := fmt.Sprintf("http://%s:%d/node/list", seedNode.Address, seedNode.NodeDiscoveryPort)
    
    // 发送请求
    resp, err := http.Get(url)
    if err != nil {
        cs.logger.Error("Failed to get node list from seed node %s: %v", seedNode.Address, err)
        return
    }
    defer resp.Body.Close()
    
    // 解析响应
    var nodes map[string]*ClusterNode
    if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
        cs.logger.Error("Failed to parse node list from seed node %s: %v", seedNode.Address, err)
        return
    }
    
    // 处理获取到的节点列表
    cs.mutex.Lock()
    for nodeID, node := range nodes {
        // 忽略自己和已经存在的节点
        if nodeID == cs.localNode.ID || cs.nodes[nodeID] != nil {
            continue
        }
        
        // 添加新节点
        cs.nodes[nodeID] = node
        cs.logger.Info("Discovered new cluster node from seed: %s (%s:%d)", 
            nodeID, node.Address, node.Port)
        
        // 与新节点同步数据
        go cs.syncWithNewNode(node)
    }
    cs.mutex.Unlock()
}

// 扫描集群节点
func (cs *ClusterSync) scanClusterNodes() {
    cs.mutex.RLock()
    nodes := make(map[string]*ClusterNode)
    for id, node := range cs.nodes {
        nodes[id] = node
    }
    cs.mutex.RUnlock()
    
    // 检查每个节点的状态
    for _, node := range nodes {
        // 发送心跳检测
        cs.sendHeartbeatToNode(node)
    }
}

// 向节点发送心跳
func (cs *ClusterSync) sendHeartbeatToNode(node *ClusterNode) {
    // 构建心跳消息
    heartbeatMsg := &ClusterDiscoveryMessage{
        Type:              "heartbeat",
        NodeID:            cs.localNode.ID,
        NodeAddr:          cs.localNode.Address,
        NodePort:          cs.localNode.Port,
        NodeDiscoveryPort: cs.localNode.NodeDiscoveryPort,
        Timestamp:         time.Now().Unix(),
    }
    
    // 构建请求URL
    url := fmt.Sprintf("http://%s:%d/node/info", node.Address, node.NodeDiscoveryPort)
    
    // 发送请求
    resp, err := http.Get(url)
    if err != nil {
        cs.logger.Error("Failed to send heartbeat to node %s: %v", node.ID, err)
        
        // 标记节点为离线
        cs.mutex.Lock()
        delete(cs.nodes, node.ID)
        cs.mutex.Unlock()
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        cs.logger.Error("Heartbeat to node %s failed with status: %s", node.ID, resp.Status)
        return
    }
    
    // 更新节点最后活跃时间
    cs.mutex.Lock()
    if n, exists := cs.nodes[node.ID]; exists {
        n.LastSeen = time.Now()
    }
    cs.mutex.Unlock()
}



// 处理节点发现
func (cs *ClusterSync) handleNodeDiscovery(msg *ClusterDiscoveryMessage, _ *net.UDPAddr) {
    cs.mutex.Lock()
    defer cs.mutex.Unlock()
    
    // 检查节点是否已存在
    if _, exists := cs.nodes[msg.NodeID]; exists {
        // 更新节点信息
        cs.nodes[msg.NodeID].LastSeen = time.Now()
        cs.nodes[msg.NodeID].NodeDiscoveryPort = msg.NodeDiscoveryPort
        return
    }
    
    // 添加新节点
    newNode := &ClusterNode{
        ID:                msg.NodeID,
        Address:           msg.NodeAddr,
        Port:              msg.NodePort,
        NodeDiscoveryPort: msg.NodeDiscoveryPort,
        LastSeen:          time.Now(),
        Status:            "online",
    }
    
    cs.nodes[msg.NodeID] = newNode
    cs.logger.Info("Discovered new cluster node: %s (%s:%d) [DiscoveryPort: %d]", 
        msg.NodeID, msg.NodeAddr, msg.NodePort, msg.NodeDiscoveryPort)
    
    // 与新节点建立连接并同步数据
    go cs.syncWithNewNode(newNode)
}
```

## 网络通信设计

### 服务注册协议
```
POST /register
Content-Type: application/json
Authorization: Bearer your-registration-secret-key

{
    "id": "service-instance-1",
    "name": "user-service",
    "address": "192.168.1.100",
    "port": 8080,
    "metadata": {
        "version": "1.0.0",
        "health_path": "/health"
    }
}
```

**认证说明**：
- 使用`Authorization`头部传递注册凭证
- 凭证格式：`Bearer <your-registration-secret-key>`
- 密钥需与网关配置中的`RegisterAuthKey`一致
- 未提供或错误的凭证将返回401 Unauthorized响应

### 集群同步协议
- 使用HTTP协议进行节点间通信和数据同步
- 支持跨网络环境的节点发现和通信
- 基于种子节点的节点发现机制
- 提供以下HTTP接口：
  - `POST /node/register`：注册新节点
  - `GET /node/info`：获取节点信息
  - `GET /node/list`：获取集群节点列表
  - `POST /sync`：集群数据同步
- 消息格式：JSON
- 支持的消息类型：节点发现、节点心跳、数据同步请求、数据同步响应

#### 节点注册接口 (`POST /node/register`)
```
POST /node/register
Content-Type: application/json

{
    "type": "discovery",
    "node_id": "gateway-node-1",
    "node_addr": "192.168.1.100",
    "node_port": 8080,
    "node_discovery_port": 8384,
    "timestamp": 1620000000
}
```

**响应格式**：
```json
{
    "status": "ok"
}
```

#### 获取节点信息接口 (`GET /node/info`)
```
GET /node/info
```

**响应格式**：
```json
{
    "id": "gateway-node-1",
    "address": "192.168.1.100",
    "port": 8080,
    "node_discovery_port": 8384,
    "last_seen": "2023-01-01T12:00:00Z",
    "status": "online"
}
```

#### 获取节点列表接口 (`GET /node/list`)
```
GET /node/list
```

**响应格式**：
```json
{
    "gateway-node-1": {
        "id": "gateway-node-1",
        "address": "192.168.1.100",
        "port": 8080,
        "node_discovery_port": 8384,
        "last_seen": "2023-01-01T12:00:00Z",
        "status": "online"
    },
    "gateway-node-2": {
        "id": "gateway-node-2",
        "address": "192.168.1.101",
        "port": 8081,
        "node_discovery_port": 8385,
        "last_seen": "2023-01-01T12:00:00Z",
        "status": "online"
    }
}
```

#### 集群数据同步接口 (`POST /sync`)
```
POST /sync
Content-Type: application/json

{
    "type": "sync_request",
    "node_id": "gateway-node-1",
    "timestamp": 1620000000
}
```

**响应格式**：
```json
{
    "type": "sync_response",
    "node_id": "gateway-node-2",
    "services": {
        "user-service": [
            {
                "id": "service-instance-1",
                "name": "user-service",
                "address": "192.168.1.200",
                "port": 8000,
                "metadata": {
                    "version": "1.0.0"
                },
                "last_seen": "2023-01-01T12:00:00Z"
            }
        ]
    },
    "nodes": {
        "gateway-node-2": {
            "id": "gateway-node-2",
            "address": "192.168.1.101",
            "port": 8081,
            "node_discovery_port": 8385,
            "last_seen": "2023-01-01T12:00:00Z",
            "status": "online"
        }
    },
    "timestamp": 1620000000
}
```

## 配置管理

### 配置项
- 网关监听地址和端口
- 服务注册过期时间
- 健康检查间隔和超时
- 集群节点列表
- 心跳间隔

```go
// pkg/config/config.go
type Config struct {
    ListenAddr      string        `json:"listen_addr"`
    ListenPort      int           `json:"listen_port"`       // 监听端口
    RegisterTTL     time.Duration `json:"register_ttl"`
    HealthCheckInterval time.Duration `json:"health_check_interval"`
    HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
    ClusterNodes    []string      `json:"cluster_nodes"`     // 手动配置的集群节点
    SeedNodes       []string      `json:"seed_nodes"`        // 种子节点列表（用于跨网络自动发现）
    HeartbeatInterval time.Duration `json:"heartbeat_interval"`
    RegisterAuthKey string        `json:"register_auth_key"` // 服务注册认证密钥
    EnableAutoDiscovery bool       `json:"enable_auto_discovery"` // 是否启用自动发现
    NodeDiscoveryPort int          `json:"node_discovery_port"` // 节点发现服务端口
}
```

## 启动流程

1. 加载配置
2. 初始化服务注册中心
3. 启动健康检查
4. 初始化集群同步
5. 启动HTTP服务器处理请求
6. 启动管理API（服务注册接口）

```go
// cmd/gateway/main.go
func main() {
    // 加载配置
    // 初始化注册中心
    // 初始化负载均衡器
    // 初始化健康检查
    // 初始化集群同步
    // 启动HTTP服务器
    // 启动管理API
    // 等待中断信号
}
```

## 错误处理与日志

### 错误处理
- 统一的错误响应格式
- 详细的错误日志记录
- 适当的重试机制

### 日志设计
- 不同级别的日志（DEBUG, INFO, WARN, ERROR）
- 包含时间戳、日志级别、模块、消息内容
- 支持日志输出到文件

```go
// pkg/utils/log.go
type Logger struct {
    // 日志配置
}

func (l *Logger) Debug(format string, args ...interface{})
func (l *Logger) Info(format string, args ...interface{})
func (l *Logger) Warn(format string, args ...interface{})
func (l *Logger) Error(format string, args ...interface{})
```

## 测试策略

### 单元测试
- 测试服务注册中心的各种操作
- 测试负载均衡算法的正确性
- 测试健康检查逻辑
- 测试集群同步机制

### 集成测试
- 启动多个服务实例进行注册
- 测试请求转发功能
- 测试负载均衡效果
- 测试健康检查和故障转移
- 测试网关集群同步

### 性能测试
- 测试网关的请求处理能力
- 测试在高并发下的稳定性
- 测试服务实例增减时的性能影响

## 部署与运维

### 部署方式
- 编译为单个可执行文件
- 支持Docker容器化部署
- 配置文件外部化

### 集群自动发现配置与使用

#### 配置说明
在网关配置文件中启用并配置自动发现功能：

```json
{
  "listen_addr": "0.0.0.0",
  "listen_port": 8080,
  "enable_auto_discovery": true,
  "multicast_addr": "224.0.0.100",
  "multicast_port": 8384,
  "register_auth_key": "your-secret-key",
  "heartbeat_interval": "10s"
}
```

#### 使用说明
1. **启用自动发现**：将`enable_auto_discovery`设置为`true`
2. **配置多播地址**：使用默认的`224.0.0.100`或自定义多播地址（范围：224.0.0.0-239.255.255.255）
3. **配置多播端口**：默认使用8384端口
4. **启动多个网关实例**：在同一网络中启动多个网关实例，它们会自动发现彼此并形成集群

#### 自动发现工作原理
1. 网关启动时加入多播组
2. 定期发送多播发现消息
3. 接收其他节点的发现消息并建立连接
4. 自动同步服务注册数据
5. 定期发送心跳保持节点活跃状态
6. 自动剔除超时无响应的节点

### 多网关统一入口解决方案

#### Nginx负载均衡
- **方案描述**：使用Nginx作为前端负载均衡器，将请求分发到多个网关实例
- **配置示例**：
  ```nginx
  http {
    upstream gateway_cluster {
      server gateway1:8080;
      server gateway2:8080;
      server gateway3:8080;
    }
    
    server {
      listen 80;
      
      location / {
        proxy_pass http://gateway_cluster;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
      }
    }
  }
  ```
- **优势**：
  - 成熟稳定的负载均衡方案
  - 支持健康检查和故障转移
  - 可配置SSL终止、请求限流等高级功能
  - 性能优异，资源消耗低

### 监控与维护
- 提供管理API查看服务注册状态
- 提供管理API查看集群状态
- 支持动态调整配置

## 扩展功能建议

1. **支持更多负载均衡算法**：如权重轮询、最少连接等
2. **支持服务限流**：基于令牌桶或漏桶算法
3. **支持请求路由规则**：基于路径、请求头的路由
4. **支持认证授权**：集成JWT等认证机制
5. **支持请求/响应转换**：修改请求头、响应头
6. **支持日志记录**：请求日志、访问统计

## 技术要点总结

1. **纯Go实现**：不依赖第三方框架，充分利用Go标准库
2. **并发安全**：使用sync包确保线程安全
3. **高性能**：利用Go的并发特性处理高并发请求
4. **可靠性**：完善的错误处理和故障转移机制
5. **可扩展性**：模块化设计，便于功能扩展
6. **易维护**：清晰的代码结构和详细的文档

通过以上设计和实现要点，可以开发出一个功能完整、性能优良的Go语言API网关，满足服务注册发现、负载均衡、健康检查和网关集群同步的需求。