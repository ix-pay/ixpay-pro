package discovery

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ixpay-pro/gxy/pkg/utils"
)

type HealthChecker struct {
	registry         *Registry
	interval         time.Duration
	timeout          time.Duration
	logger           *utils.Logger
	client           *http.Client // 复用的HTTP客户端
	concurrencyLimit int          // 并发检查限制
}

func NewHealthChecker(registry *Registry, interval, timeout time.Duration, logger *utils.Logger) *HealthChecker {
	// 创建复用的HTTP客户端
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &HealthChecker{
		registry:         registry,
		interval:         interval,
		timeout:          timeout,
		logger:           logger,
		client:           client,
		concurrencyLimit: 10, // 默认并发限制为10
	}
}

func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for range ticker.C {
			hc.checkAllServices()
		}
	}()

	hc.logger.Info("Health checker started with interval %v and timeout %v", hc.interval, hc.timeout)
}

func (hc *HealthChecker) checkAllServices() {
	services := hc.registry.GetAllServices()

	// 收集所有需要检查的实例
	var allInstances []struct {
		serviceName string
		instance    *ServiceInstance
	}

	for serviceName, instances := range services {
		for _, instance := range instances {
			allInstances = append(allInstances, struct {
				serviceName string
				instance    *ServiceInstance
			}{serviceName, instance})
		}
	}

	// 使用信号量控制并发
	semaphore := make(chan struct{}, hc.concurrencyLimit)

	for _, item := range allInstances {
		semaphore <- struct{}{} // 获取信号量

		go func(serviceName string, instance *ServiceInstance) {
			defer func() {
				<-semaphore // 释放信号量
			}()

			if !hc.checkService(instance) {
				hc.logger.Warn("Service instance %s (%s:%d) is unhealthy, deregistering", instance.ID, instance.Address, instance.Port)
				hc.registry.Deregister(serviceName, instance.ID)
			}
		}(item.serviceName, item.instance)
	}

	// 等待所有并发检查完成
	for i := 0; i < hc.concurrencyLimit; i++ {
		semaphore <- struct{}{}
	}
	close(semaphore)
}

func (hc *HealthChecker) checkService(instance *ServiceInstance) bool {
	// 检查最后一次心跳时间
	if time.Since(instance.LastSeen) > hc.interval*3 {
		return false
	}

	// 构建健康检查URL
	healthCheckPath := "/health"
	if path, ok := instance.Metadata["health_check_path"]; ok {
		healthCheckPath = path
	}

	url := "http://" + instance.Address + ":" + strconv.Itoa(instance.Port) + healthCheckPath

	// 发送健康检查请求，使用复用的HTTP客户端
	resp, err := hc.client.Get(url)
	if err != nil {
		hc.logger.Debug("Health check failed for %s: %v", url, err)
		return false
	}
	defer resp.Body.Close()

	// 检查响应状态码
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
