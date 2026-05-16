package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Address           string            `json:"address"`
	Port              int               `json:"port"`
	Metadata          map[string]string `json:"metadata"`
	LastSeen          time.Time         `json:"last_seen"`
	ActiveConnections int               `json:"active_connections"`
}

// GatewayClient 网关客户端
type GatewayClient struct {
	gatewayURL string
	authKey    string
	instance   *ServiceInstance
	httpClient *http.Client
	ticker     *time.Ticker
	stopChan   chan struct{}
	mux        sync.RWMutex
	running    bool
}

// NewGatewayClient 创建网关客户端
func NewGatewayClient(gatewayURL, authKey string) *GatewayClient {
	return &GatewayClient{
		gatewayURL: gatewayURL,
		authKey:    authKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		stopChan: make(chan struct{}),
	}
}

// Register 注册服务到网关
func (gc *GatewayClient) Register(ctx context.Context, instance *ServiceInstance) error {
	gc.mux.Lock()
	gc.instance = instance
	gc.mux.Unlock()

	url := gc.gatewayURL + "/api/register"
	payload, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("序列化服务实例失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", gc.authKey)

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送注册请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("注册失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// Deregister 从网关注销服务
func (gc *GatewayClient) Deregister(ctx context.Context) error {
	gc.mux.RLock()
	if gc.instance == nil {
		gc.mux.RUnlock()
		return fmt.Errorf("服务实例未注册")
	}
	instanceID := gc.instance.ID
	serviceName := gc.instance.Name
	gc.mux.RUnlock()

	url := gc.gatewayURL + "/api/deregister"
	payload := map[string]string{
		"service_name": serviceName,
		"instance_id":  instanceID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化注销请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", gc.authKey)

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送注销请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("注销失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// StartHeartbeat 启动心跳
func (gc *GatewayClient) StartHeartbeat(ctx context.Context, interval time.Duration) {
	gc.mux.Lock()
	if gc.running {
		gc.mux.Unlock()
		return
	}
	gc.running = true
	gc.ticker = time.NewTicker(interval)
	gc.mux.Unlock()

	go func() {
		for {
			select {
			case <-gc.ticker.C:
				gc.updateLastSeen(ctx)
			case <-ctx.Done():
				gc.Stop()
				return
			case <-gc.stopChan:
				return
			}
		}
	}()
}

// Stop 停止心跳
func (gc *GatewayClient) Stop() {
	gc.mux.Lock()
	defer gc.mux.Unlock()

	if !gc.running {
		return
	}

	gc.running = false
	if gc.ticker != nil {
		gc.ticker.Stop()
	}
	close(gc.stopChan)
}

// updateLastSeen 更新最后心跳时间
func (gc *GatewayClient) updateLastSeen(ctx context.Context) {
	gc.mux.RLock()
	if gc.instance == nil {
		gc.mux.RUnlock()
		return
	}
	instance := *gc.instance
	gc.mux.RUnlock()

	instance.LastSeen = time.Now()
	payload, err := json.Marshal(&instance)
	if err != nil {
		return
	}

	url := gc.gatewayURL + "/api/register"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", gc.authKey)

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// GetInstance 获取服务实例
func (gc *GatewayClient) GetInstance() *ServiceInstance {
	gc.mux.RLock()
	defer gc.mux.RUnlock()
	return gc.instance
}

// BuildServiceInstance 构建服务实例
func BuildServiceInstance(name, host string, port int, nodeRole, nodeID string, metadata map[string]string) *ServiceInstance {
	if nodeID == "" {
		hostname, _ := os.Hostname()
		nodeID = fmt.Sprintf("%s-%s-%d", name, hostname, port)
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["node_role"] = nodeRole
	metadata["health_check_path"] = "/health"
	metadata["started_at"] = time.Now().Format(time.RFC3339)

	return &ServiceInstance{
		ID:                nodeID,
		Name:              name,
		Address:           host,
		Port:              port,
		Metadata:          metadata,
		LastSeen:          time.Now(),
		ActiveConnections: 0,
	}
}

// ParsePort 从字符串解析端口
func ParsePort(portStr string, defaultPort int) int {
	if portStr == "" {
		return defaultPort
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return defaultPort
	}
	return port
}
