package eventbus

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// EventRepositoryImpl 事件仓储实现
type EventRepositoryImpl struct {
	DB *gorm.DB
}

// NewEventRepository 创建事件仓储
func NewEventRepository(db *gorm.DB) EventRepository {
	return &EventRepositoryImpl{DB: db}
}

// CreateEvent 创建事件
func (r *EventRepositoryImpl) CreateEvent(ctx context.Context, event *Event) error {
	return r.DB.WithContext(ctx).Create(event).Error
}

// GetEventByID 根据ID获取事件
func (r *EventRepositoryImpl) GetEventByID(ctx context.Context, id int64) (*Event, error) {
	var event Event
	err := r.DB.WithContext(ctx).Preload("Deliveries").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// UpdateEvent 更新事件
func (r *EventRepositoryImpl) UpdateEvent(ctx context.Context, event *Event) error {
	return r.DB.WithContext(ctx).Save(event).Error
}

// UpdateEventStatus 更新事件状态
func (r *EventRepositoryImpl) UpdateEventStatus(ctx context.Context, id int64, status EventStatus, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	if status == EventStatusSuccess {
		now := time.Now()
		updates["processed_at"] = &now
	}
	return r.DB.WithContext(ctx).Model(&Event{}).Where("id = ?", id).Updates(updates).Error
}

// ListEvents 列出事件
func (r *EventRepositoryImpl) ListEvents(ctx context.Context, filter EventFilter) ([]*Event, int64, error) {
	var events []*Event
	var total int64

	query := r.DB.WithContext(ctx).Model(&Event{})

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	// 排序
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}
	query = query.Order(filter.SortBy + " " + filter.SortOrder)

	err := query.Find(&events).Error
	return events, total, err
}

// CreateDelivery 创建投递记录
func (r *EventRepositoryImpl) CreateDelivery(ctx context.Context, delivery *EventDelivery) error {
	return r.DB.WithContext(ctx).Create(delivery).Error
}

// UpdateDelivery 更新投递记录
func (r *EventRepositoryImpl) UpdateDelivery(ctx context.Context, delivery *EventDelivery) error {
	return r.DB.WithContext(ctx).Save(delivery).Error
}

// ListDeliveriesByEventID 根据事件ID列出投递记录
func (r *EventRepositoryImpl) ListDeliveriesByEventID(ctx context.Context, eventID int64) ([]*EventDelivery, error) {
	var deliveries []*EventDelivery
	err := r.DB.WithContext(ctx).Where("event_id = ?", eventID).Find(&deliveries).Error
	return deliveries, err
}

// ListDeliveriesBySubscriberID 根据订阅者ID列出投递记录
func (r *EventRepositoryImpl) ListDeliveriesBySubscriberID(ctx context.Context, subscriberID int64) ([]*EventDelivery, error) {
	var deliveries []*EventDelivery
	err := r.DB.WithContext(ctx).Where("subscriber_id = ?", subscriberID).Find(&deliveries).Error
	return deliveries, err
}

// ListFailedDeliveries 列出失败的投递记录
func (r *EventRepositoryImpl) ListFailedDeliveries(ctx context.Context, limit int) ([]*EventDelivery, error) {
	var deliveries []*EventDelivery
	err := r.DB.WithContext(ctx).
		Where("status = ? AND attempts < max_attempts", EventStatusFailed).
		Order("created_at ASC").
		Limit(limit).
		Find(&deliveries).Error
	return deliveries, err
}

// CreateSubscriber 创建订阅者
func (r *EventRepositoryImpl) CreateSubscriber(ctx context.Context, subscriber *Subscriber) error {
	return r.DB.WithContext(ctx).Create(subscriber).Error
}

// GetSubscriberByID 根据ID获取订阅者
func (r *EventRepositoryImpl) GetSubscriberByID(ctx context.Context, id int64) (*Subscriber, error) {
	var subscriber Subscriber
	err := r.DB.WithContext(ctx).First(&subscriber, id).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// GetSubscriberByName 根据名称获取订阅者
func (r *EventRepositoryImpl) GetSubscriberByName(ctx context.Context, name string) (*Subscriber, error) {
	var subscriber Subscriber
	err := r.DB.WithContext(ctx).Where("name = ?", name).First(&subscriber).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// UpdateSubscriber 更新订阅者
func (r *EventRepositoryImpl) UpdateSubscriber(ctx context.Context, subscriber *Subscriber) error {
	return r.DB.WithContext(ctx).Save(subscriber).Error
}

// ListSubscribers 列出订阅者
func (r *EventRepositoryImpl) ListSubscribers(ctx context.Context, filter SubscriberFilter) ([]*Subscriber, int64, error) {
	var subscribers []*Subscriber
	var total int64

	query := r.DB.WithContext(ctx).Model(&Subscriber{})

	if filter.EventName != "" {
		query = query.Where("event_name LIKE ?", "%"+filter.EventName+"%")
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	err := query.Order("priority ASC, created_at DESC").Find(&subscribers).Error
	return subscribers, total, err
}

// DeleteSubscriber 删除订阅者
func (r *EventRepositoryImpl) DeleteSubscriber(ctx context.Context, id int64) error {
	return r.DB.WithContext(ctx).Delete(&Subscriber{}, id).Error
}

// ListDeadLetters 列出死信
func (r *EventRepositoryImpl) ListDeadLetters(ctx context.Context, filter EventFilter) ([]*Event, int64, error) {
	filter.Status = ptrEventStatus(EventStatusDeadLetter)
	return r.ListEvents(ctx, filter)
}

// MoveToDeadLetter 将事件移动到死信队列
func (r *EventRepositoryImpl) MoveToDeadLetter(ctx context.Context, eventID int64, reason string) error {
	return r.DB.WithContext(ctx).Model(&Event{}).Where("id = ?", eventID).Updates(map[string]interface{}{
		"status":        EventStatusDeadLetter,
		"error_message": reason,
	}).Error
}

// GetEventCount 获取事件数量
func (r *EventRepositoryImpl) GetEventCount(ctx context.Context, status EventStatus) (int64, error) {
	var count int64
	query := r.DB.WithContext(ctx).Model(&Event{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

// GetSubscriberCount 获取订阅者数量
func (r *EventRepositoryImpl) GetSubscriberCount(ctx context.Context, isActive bool) (int64, error) {
	var count int64
	query := r.DB.WithContext(ctx).Model(&Subscriber{})
	query = query.Where("is_active = ?", isActive)
	err := query.Count(&count).Error
	return count, err
}

// 辅助函数：获取 EventStatus 指针
func ptrEventStatus(s EventStatus) *EventStatus {
	return &s
}
