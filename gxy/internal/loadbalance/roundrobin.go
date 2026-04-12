package loadbalance

import (
	"sync"

	"github.com/ixpay-pro/gxy/internal/discovery"
)

// ServiceInstance 类型别名，便于使用
type ServiceInstance = discovery.ServiceInstance

type RoundRobinBalancer struct {
	currentIndex        map[string]int
	mutex               sync.RWMutex
	connectionThreshold int
}

func NewRoundRobinBalancer(connectionThreshold int) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		currentIndex:        make(map[string]int),
		connectionThreshold: connectionThreshold,
	}
}

func (rb *RoundRobinBalancer) Select(serviceName string, instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	// 过滤健康实例（目前假设所有实例都是健康的，后续可扩展健康检查逻辑）
	healthyCount := len(instances)
	if healthyCount == 0 {
		return nil
	}

	// 使用读写锁分离读和写操作，减少锁竞争
	rb.mutex.RLock()
	index, exists := rb.currentIndex[serviceName]
	rb.mutex.RUnlock()

	// 如果服务名不存在，初始化为0
	if !exists {
		rb.mutex.Lock()
		// 双重检查，避免竞争条件
		if index, exists = rb.currentIndex[serviceName]; !exists {
			index = 0
		}
		rb.mutex.Unlock()
	}

	// 轮询选择实例
	selectedIndex := index % healthyCount
	selectedInstance := instances[selectedIndex]

	// 更新索引，使用写锁保护
	rb.mutex.Lock()
	rb.currentIndex[serviceName] = selectedIndex + 1
	rb.mutex.Unlock()

	return selectedInstance
}
