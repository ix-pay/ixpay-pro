package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ixpay-pro/gxy/internal/discovery"
	"github.com/ixpay-pro/gxy/internal/loadbalance"
	"github.com/ixpay-pro/gxy/pkg/utils"
)

type Proxy struct {
	registry *discovery.Registry
	balancer *loadbalance.RoundRobinBalancer
	logger   *utils.Logger
	client   *http.Client // 复用的HTTP客户端
}

func NewProxy(registry *discovery.Registry, balancer *loadbalance.RoundRobinBalancer, logger *utils.Logger) *Proxy {
	// 创建配置优化的HTTP客户端
	client := &http.Client{
		// 配置超时时间
		Timeout: 30 * time.Second,
		// 配置连接池
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}

	return &Proxy{
		registry: registry,
		balancer: balancer,
		logger:   logger,
		client:   client,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 解析服务名称和路径
	// 假设请求格式为 /service-name/path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 || pathParts[1] == "" {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	serviceName := pathParts[1]
	backendPath := strings.Join(pathParts[2:], "/")

	// 获取服务实例列表
	instances := p.registry.GetInstances(serviceName)
	if len(instances) == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// 选择服务实例
	selectedInstance := p.balancer.Select(serviceName, instances)
	if selectedInstance == nil {
		http.Error(w, "No available service instances", http.StatusServiceUnavailable)
		return
	}

	// 更新活跃连接数
	p.registry.UpdateConnectionCount(selectedInstance.ID, 1)
	defer p.registry.UpdateConnectionCount(selectedInstance.ID, -1)

	// 构建后端URL
	backendURL := &url.URL{
		Scheme:   "http",
		Host:     selectedInstance.Address + ":" + strconv.Itoa(selectedInstance.Port),
		Path:     "/" + backendPath,
		RawQuery: r.URL.RawQuery,
	}

	// 转发请求
	p.forwardRequest(selectedInstance, backendURL, w, r)
}

func (p *Proxy) forwardRequest(instance *discovery.ServiceInstance, backendURL *url.URL, w http.ResponseWriter, r *http.Request) {
	// 创建新的请求
	proxyReq, err := http.NewRequest(r.Method, backendURL.String(), r.Body)
	if err != nil {
		p.logger.Error("Failed to create proxy request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	proxyReq.Header = make(http.Header)
	for k, v := range r.Header {
		proxyReq.Header[k] = v
	}

	// 设置X-Forwarded-For头
	clientIP := r.RemoteAddr
	if xff, ok := r.Header["X-Forwarded-For"]; ok && len(xff) > 0 {
		clientIP = xff[0] + ", " + clientIP
	}
	proxyReq.Header.Set("X-Forwarded-For", clientIP)

	// 发送请求到后端服务
	resp, err := p.client.Do(proxyReq)
	if err != nil {
		p.logger.Error("Failed to forward request to %s:%d: %v", instance.Address, instance.Port, err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	if _, err := io.Copy(w, resp.Body); err != nil {
		p.logger.Error("Failed to copy response body: %v", err)
	}
}
