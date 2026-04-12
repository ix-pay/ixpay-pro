package discovery

import (
	"sync"
	"time"
)

type ServiceInstance struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Address           string            `json:"address"`
	Port              int               `json:"port"`
	Metadata          map[string]string `json:"metadata"`
	LastSeen          time.Time         `json:"last_seen"`
	ActiveConnections int               `json:"active_connections"`
}

type Registry struct {
	services  map[string][]*ServiceInstance
	instances map[string]*ServiceInstance // instanceID -> ServiceInstance mapping for O(1) lookups
	mutex     sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		services:  make(map[string][]*ServiceInstance),
		instances: make(map[string]*ServiceInstance),
	}
}

func (r *Registry) Register(instance *ServiceInstance) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance.LastSeen = time.Now()

	serviceList, exists := r.services[instance.Name]
	if !exists {
		serviceList = make([]*ServiceInstance, 0)
	}

	// 检查是否已经存在相同ID的实例
	for i, existingInstance := range serviceList {
		if existingInstance.ID == instance.ID {
			// 更新现有实例
			serviceList[i] = instance
			r.services[instance.Name] = serviceList
			r.instances[instance.ID] = instance // 更新instances映射
			return
		}
	}

	// 添加新实例
	serviceList = append(serviceList, instance)
	r.services[instance.Name] = serviceList
	r.instances[instance.ID] = instance // 添加到instances映射
}

func (r *Registry) Deregister(serviceName, instanceID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	serviceList, exists := r.services[serviceName]
	if !exists {
		return
	}

	for i, instance := range serviceList {
		if instance.ID == instanceID {
			// 从列表中删除实例
			serviceList = append(serviceList[:i], serviceList[i+1:]...)
			// 从instances映射中删除
			delete(r.instances, instanceID)
			break
		}
	}

	if len(serviceList) == 0 {
		// 如果服务实例列表为空，删除该服务
		delete(r.services, serviceName)
	} else {
		r.services[serviceName] = serviceList
	}
}

func (r *Registry) GetInstances(serviceName string) []*ServiceInstance {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instances, exists := r.services[serviceName]
	if !exists {
		return []*ServiceInstance{}
	}

	// 返回副本，避免外部修改内部数据
	result := make([]*ServiceInstance, len(instances))
	copy(result, instances)
	return result
}

func (r *Registry) GetAllServices() map[string][]*ServiceInstance {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 返回副本，避免外部修改内部数据
	result := make(map[string][]*ServiceInstance)
	for serviceName, instances := range r.services {
		result[serviceName] = make([]*ServiceInstance, len(instances))
		copy(result[serviceName], instances)
	}

	return result
}

func (r *Registry) UpdateConnectionCount(instanceID string, delta int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 使用instances映射直接查找实例，O(1)时间复杂度
	if instance, exists := r.instances[instanceID]; exists {
		instance.ActiveConnections += delta
		if instance.ActiveConnections < 0 {
			instance.ActiveConnections = 0
		}
	}
}
