package eventbus

import (
	"time"
)

// EventStatus 事件状态
type EventStatus int

const (
	// EventStatusPending 待处理
	EventStatusPending EventStatus = iota
	// EventStatusProcessing 处理中
	EventStatusProcessing
	// EventStatusSuccess 成功
	EventStatusSuccess
	// EventStatusFailed 失败
	EventStatusFailed
	// EventStatusDeadLetter 死信
	EventStatusDeadLetter
)

func (s EventStatus) String() string {
	switch s {
	case EventStatusPending:
		return "待处理"
	case EventStatusProcessing:
		return "处理中"
	case EventStatusSuccess:
		return "成功"
	case EventStatusFailed:
		return "失败"
	case EventStatusDeadLetter:
		return "死信"
	default:
		return "未知"
	}
}

// Event 事件定义
type Event struct {
	ID              int64           `json:"id,string" gorm:"primaryKey;autoIncrement"`
	Name            string          `json:"name" gorm:"size:255;not null;index"`
	Payload         string          `json:"payload" gorm:"type:text"`
	Status          EventStatus     `json:"status" gorm:"size:32;default:0"`
	Priority        int             `json:"priority" gorm:"default:0"`
	MaxRetries      int             `json:"maxRetries" gorm:"default:3"`
	RetryCount      int             `json:"retryCount" gorm:"default:0"`
	DelaySeconds    int             `json:"delaySeconds" gorm:"default:0"`
	ScheduledAt     *time.Time      `json:"scheduledAt,omitempty"`
	ProcessedAt     *time.Time      `json:"processedAt,omitempty"`
	NextRetryAt     *time.Time      `json:"nextRetryAt,omitempty"`
	ErrorMessage    string          `json:"errorMessage,omitempty" gorm:"type:text"`
	SubscriberCount int             `json:"subscriberCount" gorm:"default:0"`
	Metadata        string          `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	Deliveries      []EventDelivery `json:"deliveries,omitempty" gorm:"foreignKey:EventID"`
}

// TableName 指定表名
func (Event) TableName() string {
	return "base_events"
}

// EventDelivery 事件投递记录
type EventDelivery struct {
	ID           int64       `json:"id,string" gorm:"primaryKey;autoIncrement"`
	EventID      int64       `json:"eventId,string" gorm:"index;not null"`
	SubscriberID int64       `json:"subscriberId,string" gorm:"index;not null"`
	Status       EventStatus `json:"status" gorm:"size:32;default:0"`
	Attempts     int         `json:"attempts" gorm:"default:0"`
	MaxAttempts  int         `json:"maxAttempts" gorm:"default:3"`
	ErrorMessage string      `json:"errorMessage,omitempty" gorm:"type:text"`
	DeliveredAt  *time.Time  `json:"deliveredAt,omitempty"`
	NextRetryAt  *time.Time  `json:"nextRetryAt,omitempty"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

// TableName 指定表名
func (EventDelivery) TableName() string {
	return "base_event_deliveries"
}

// Subscriber 订阅者定义
type Subscriber struct {
	ID              int64                    `json:"id,string" gorm:"primaryKey;autoIncrement"`
	Name            string                   `json:"name" gorm:"size:255;not null;uniqueIndex"`
	EventName       string                   `json:"eventName" gorm:"size:255;not null;index"`
	Endpoint        string                   `json:"endpoint" gorm:"size:2048"`
	Handler         func(event *Event) error `json:"-" gorm:"-"`
	IsActive        bool                     `json:"isActive" gorm:"default:true"`
	MaxRetries      int                      `json:"maxRetries" gorm:"default:3"`
	Timeout         int                      `json:"timeout" gorm:"default:30"` // 秒
	Priority        int                      `json:"priority" gorm:"default:0"`
	LastDeliveredAt *time.Time               `json:"lastDeliveredAt,omitempty"`
	FailureCount    int                      `json:"failureCount" gorm:"default:0"`
	CreatedAt       time.Time                `json:"createdAt"`
	UpdatedAt       time.Time                `json:"updatedAt"`
}

// TableName 指定表名
func (Subscriber) TableName() string {
	return "base_subscribers"
}

// NewEvent 创建新事件
func NewEvent(name string, payload string, opts ...EventOption) *Event {
	event := &Event{
		Name:       name,
		Payload:    payload,
		Status:     EventStatusPending,
		MaxRetries: 3,
		Priority:   0,
	}

	for _, opt := range opts {
		opt(event)
	}

	return event
}

// EventOption 事件配置选项
type EventOption func(*Event)

// WithPriority 设置优先级
func WithPriority(priority int) EventOption {
	return func(e *Event) {
		e.Priority = priority
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) EventOption {
	return func(e *Event) {
		e.MaxRetries = maxRetries
	}
}

// WithDelaySeconds 设置延迟秒数
func WithDelaySeconds(delaySeconds int) EventOption {
	return func(e *Event) {
		e.DelaySeconds = delaySeconds
		now := time.Now()
		scheduledAt := now.Add(time.Duration(delaySeconds) * time.Second)
		e.ScheduledAt = &scheduledAt
	}
}

// WithScheduledAt 设置计划执行时间
func WithScheduledAt(scheduledAt time.Time) EventOption {
	return func(e *Event) {
		e.ScheduledAt = &scheduledAt
	}
}

// WithMetadata 设置元数据
func WithMetadata(metadata string) EventOption {
	return func(e *Event) {
		e.Metadata = metadata
	}
}
