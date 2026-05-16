package eventbus

import (
	"context"
	"fmt"
	"sync"
)

// SubscriberManager 订阅者管理器
type SubscriberManager struct {
	repo        EventRepository
	subscribers map[string][]*Subscriber // eventName -> subscribers
	mu          sync.RWMutex
}

// NewSubscriberManager 创建订阅者管理器
func NewSubscriberManager(repo EventRepository) *SubscriberManager {
	return &SubscriberManager{
		repo:        repo,
		subscribers: make(map[string][]*Subscriber),
	}
}

// LoadFromDB 从数据库加载所有订阅者
func (sm *SubscriberManager) LoadFromDB(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 清空现有缓存
	sm.subscribers = make(map[string][]*Subscriber)

	// 获取所有活跃订阅者
	filter := SubscriberFilter{IsActive: boolPtr(true)}
	subscribers, _, err := sm.repo.ListSubscribers(ctx, filter)
	if err != nil {
		return fmt.Errorf("加载订阅者失败: %w", err)
	}

	// 按事件名称分组
	for _, sub := range subscribers {
		sm.subscribers[sub.EventName] = append(sm.subscribers[sub.EventName], sub)
	}

	return nil
}

// GetSubscribers 获取指定事件的订阅者
func (sm *SubscriberManager) GetSubscribers(eventName string) []*Subscriber {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subs := sm.subscribers[eventName]
	// 按优先级排序
	sorted := make([]*Subscriber, len(subs))
	copy(sorted, subs)
	// 简单冒泡排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority > sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// AddSubscriber 添加订阅者
func (sm *SubscriberManager) AddSubscriber(ctx context.Context, sub *Subscriber) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 保存到数据库
	if err := sm.repo.CreateSubscriber(ctx, sub); err != nil {
		return fmt.Errorf("创建订阅者失败: %w", err)
	}

	// 添加到内存缓存
	sm.subscribers[sub.EventName] = append(sm.subscribers[sub.EventName], sub)

	return nil
}

// RemoveSubscriber 移除订阅者
func (sm *SubscriberManager) RemoveSubscriber(ctx context.Context, id int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 从数据库删除
	if err := sm.repo.DeleteSubscriber(ctx, id); err != nil {
		return fmt.Errorf("删除订阅者失败: %w", err)
	}

	// 从内存缓存中移除
	for eventName, subs := range sm.subscribers {
		for i, sub := range subs {
			if sub.ID == id {
				sm.subscribers[eventName] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}

	return nil
}

// UpdateSubscriber 更新订阅者
func (sm *SubscriberManager) UpdateSubscriber(ctx context.Context, sub *Subscriber) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 更新数据库
	if err := sm.repo.UpdateSubscriber(ctx, sub); err != nil {
		return fmt.Errorf("更新订阅者失败: %w", err)
	}

	// 更新内存缓存
	for i, existing := range sm.subscribers[sub.EventName] {
		if existing.ID == sub.ID {
			sm.subscribers[sub.EventName][i] = sub
			break
		}
	}

	return nil
}

// Refresh 刷新订阅者缓存
func (sm *SubscriberManager) Refresh(ctx context.Context) error {
	return sm.LoadFromDB(ctx)
}

// GetSubscriberCount 获取订阅者数量
func (sm *SubscriberManager) GetSubscriberCount(ctx context.Context, isActive bool) (int64, error) {
	return sm.repo.GetSubscriberCount(ctx, isActive)
}

// 辅助函数
func boolPtr(b bool) *bool {
	return &b
}
