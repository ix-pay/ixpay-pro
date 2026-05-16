package request

// CreateEventRequest 创建事件请求
type CreateEventRequest struct {
	Name       string `json:"name" binding:"required"`
	Payload    string `json:"payload"`
	Priority   int    `json:"priority"`
	MaxRetries int    `json:"maxRetries"`
	Delay      int    `json:"delay"` // 延迟秒数
	Metadata   string `json:"metadata"`
}

// CreateSubscriberRequest 创建订阅者请求
type CreateSubscriberRequest struct {
	Name       string `json:"name" binding:"required"`
	EventName  string `json:"eventName" binding:"required"`
	Endpoint   string `json:"endpoint"`
	MaxRetries int    `json:"maxRetries"`
	Timeout    int    `json:"timeout"`
	Priority   int    `json:"priority"`
}

// UpdateSubscriberRequest 更新订阅者请求
type UpdateSubscriberRequest struct {
	Name       string `json:"name"`
	Endpoint   string `json:"endpoint"`
	MaxRetries int    `json:"maxRetries"`
	Timeout    int    `json:"timeout"`
	Priority   int    `json:"priority"`
	IsActive   *bool  `json:"isActive"`
}

// ListEventsRequest 列出事件请求
type ListEventsRequest struct {
	Name      string `form:"name"`
	Status    *int   `form:"status"`
	Priority  *int   `form:"priority"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	SortBy    string `form:"sortBy"`
	SortOrder string `form:"sortOrder"`
}

// ListSubscribersRequest 列出订阅者请求
type ListSubscribersRequest struct {
	EventName string `form:"eventName"`
	IsActive  *bool  `form:"isActive"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}
