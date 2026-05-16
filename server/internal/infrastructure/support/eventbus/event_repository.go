package eventbus

import (
	"context"
	"time"
)

// EventRepository 事件仓储接口
type EventRepository interface {
	// 事件操作
	CreateEvent(ctx context.Context, event *Event) error
	GetEventByID(ctx context.Context, id int64) (*Event, error)
	UpdateEvent(ctx context.Context, event *Event) error
	UpdateEventStatus(ctx context.Context, id int64, status EventStatus, errorMessage string) error
	ListEvents(ctx context.Context, filter EventFilter) ([]*Event, int64, error)

	// 投递记录操作
	CreateDelivery(ctx context.Context, delivery *EventDelivery) error
	UpdateDelivery(ctx context.Context, delivery *EventDelivery) error
	ListDeliveriesByEventID(ctx context.Context, eventID int64) ([]*EventDelivery, error)
	ListDeliveriesBySubscriberID(ctx context.Context, subscriberID int64) ([]*EventDelivery, error)
	ListFailedDeliveries(ctx context.Context, limit int) ([]*EventDelivery, error)

	// 订阅者操作
	CreateSubscriber(ctx context.Context, subscriber *Subscriber) error
	GetSubscriberByID(ctx context.Context, id int64) (*Subscriber, error)
	GetSubscriberByName(ctx context.Context, name string) (*Subscriber, error)
	UpdateSubscriber(ctx context.Context, subscriber *Subscriber) error
	ListSubscribers(ctx context.Context, filter SubscriberFilter) ([]*Subscriber, int64, error)
	DeleteSubscriber(ctx context.Context, id int64) error

	// 死信队列
	ListDeadLetters(ctx context.Context, filter EventFilter) ([]*Event, int64, error)
	MoveToDeadLetter(ctx context.Context, eventID int64, reason string) error

	// 统计
	GetEventCount(ctx context.Context, status EventStatus) (int64, error)
	GetSubscriberCount(ctx context.Context, isActive bool) (int64, error)
}

// EventFilter 事件过滤条件
type EventFilter struct {
	Name      string       `json:"name,omitempty"`
	Status    *EventStatus `json:"status,omitempty"`
	Priority  *int         `json:"priority,omitempty"`
	StartTime *time.Time   `json:"startTime,omitempty"`
	EndTime   *time.Time   `json:"endTime,omitempty"`
	Page      int          `json:"page"`
	PageSize  int          `json:"pageSize"`
	SortBy    string       `json:"sortBy"`
	SortOrder string       `json:"sortOrder"`
}

// SubscriberFilter 订阅者过滤条件
type SubscriberFilter struct {
	EventName string `json:"eventName,omitempty"`
	IsActive  *bool  `json:"isActive,omitempty"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
}
