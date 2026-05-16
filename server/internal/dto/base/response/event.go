package response

// EventResponse 事件响应 DTO
type EventResponse struct {
	ID              int64   `json:"id,string"`
	Name            string  `json:"name"`
	Payload         string  `json:"payload"`
	Status          int     `json:"status"`
	StatusLabel     string  `json:"statusLabel"`
	Priority        int     `json:"priority"`
	MaxRetries      int     `json:"maxRetries"`
	RetryCount      int     `json:"retryCount"`
	DelaySeconds    int     `json:"delaySeconds"`
	ScheduledAt     *string `json:"scheduledAt,omitempty"`
	ProcessedAt     *string `json:"processedAt,omitempty"`
	NextRetryAt     *string `json:"nextRetryAt,omitempty"`
	ErrorMessage    string  `json:"errorMessage,omitempty"`
	SubscriberCount int     `json:"subscriberCount"`
	Metadata        string  `json:"metadata,omitempty"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// EventListResponse 事件列表响应 DTO
type EventListResponse struct {
	List  []EventResponse `json:"list"`
	Total int64           `json:"total"`
}

// EventDeliveryResponse 事件投递响应 DTO
type EventDeliveryResponse struct {
	ID           int64   `json:"id,string"`
	EventID      int64   `json:"eventId,string"`
	SubscriberID int64   `json:"subscriberId,string"`
	Status       int     `json:"status"`
	StatusLabel  string  `json:"statusLabel"`
	Attempts     int     `json:"attempts"`
	MaxAttempts  int     `json:"maxAttempts"`
	ErrorMessage string  `json:"errorMessage,omitempty"`
	DeliveredAt  *string `json:"deliveredAt,omitempty"`
	NextRetryAt  *string `json:"nextRetryAt,omitempty"`
	CreatedAt    string  `json:"createdAt"`
}

// SubscriberResponse 订阅者响应 DTO
type SubscriberResponse struct {
	ID              int64   `json:"id,string"`
	Name            string  `json:"name"`
	EventName       string  `json:"eventName"`
	Endpoint        string  `json:"endpoint"`
	IsActive        bool    `json:"isActive"`
	MaxRetries      int     `json:"maxRetries"`
	Timeout         int     `json:"timeout"`
	Priority        int     `json:"priority"`
	LastDeliveredAt *string `json:"lastDeliveredAt,omitempty"`
	FailureCount    int     `json:"failureCount"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// SubscriberListResponse 订阅者列表响应 DTO
type SubscriberListResponse struct {
	List  []SubscriberResponse `json:"list"`
	Total int64                `json:"total"`
}

// DeadLetterResponse 死信响应 DTO
type DeadLetterResponse struct {
	ID         int64  `json:"id,string"`
	EventName  string `json:"eventName"`
	Reason     string `json:"reason"`
	RetryCount int    `json:"retryCount"`
	MaxRetries int    `json:"maxRetries"`
	CreatedAt  string `json:"createdAt"`
}

// DeadLetterListResponse 死信列表响应 DTO
type DeadLetterListResponse struct {
	List  []DeadLetterResponse `json:"list"`
	Total int64                `json:"total"`
}
