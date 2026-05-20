package eventbus

import (
	"testing"
	"time"

	eb "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestEventStatus_String(t *testing.T) {
	tests := []struct {
		status   eb.EventStatus
		expected string
	}{
		{eb.EventStatusPending, "待处理"},
		{eb.EventStatusProcessing, "处理中"},
		{eb.EventStatusSuccess, "成功"},
		{eb.EventStatusFailed, "失败"},
		{eb.EventStatusDeadLetter, "死信"},
		{eb.EventStatus(99), "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestNewEvent_Defaults(t *testing.T) {
	event := eb.NewEvent("test.event", `{"key": "value"}`)

	assert.Equal(t, "test.event", event.Name)
	assert.Equal(t, `{"key": "value"}`, event.Payload)
	assert.Equal(t, eb.EventStatusPending, event.Status)
	assert.Equal(t, 3, event.MaxRetries)
	assert.Equal(t, 0, event.Priority)
	assert.Equal(t, 0, event.DelaySeconds)
	assert.Nil(t, event.ScheduledAt)
}

func TestNewEvent_WithOptions(t *testing.T) {
	scheduledAt := time.Now().Add(time.Hour)

	event := eb.NewEvent("test.event", `{"key": "value"}`,
		eb.WithPriority(10),
		eb.WithMaxRetries(5),
		eb.WithDelaySeconds(60),
		eb.WithMetadata(`{"source": "test"}`),
		eb.WithScheduledAt(scheduledAt),
	)

	assert.Equal(t, 10, event.Priority)
	assert.Equal(t, 5, event.MaxRetries)
	assert.Equal(t, 60, event.DelaySeconds)
	assert.Equal(t, `{"source": "test"}`, event.Metadata)
	assert.Equal(t, scheduledAt.Unix(), event.ScheduledAt.Unix())
}

func TestWithDelaySeconds(t *testing.T) {
	event := eb.NewEvent("delay.test", "payload", eb.WithDelaySeconds(30))

	assert.Equal(t, 30, event.DelaySeconds)
	assert.NotNil(t, event.ScheduledAt)

	expectedTime := time.Now().Add(30 * time.Second)
	diff := expectedTime.Sub(*event.ScheduledAt)
	assert.True(t, diff.Abs() < time.Second, "ScheduledAt should be approximately 30 seconds from now")
}

func TestWithMaxRetries(t *testing.T) {
	event := eb.NewEvent("retry.test", "payload", eb.WithMaxRetries(10))
	assert.Equal(t, 10, event.MaxRetries)
}

func TestWithPriority(t *testing.T) {
	event := eb.NewEvent("priority.test", "payload", eb.WithPriority(5))
	assert.Equal(t, 5, event.Priority)
}

func TestWithMetadata(t *testing.T) {
	event := eb.NewEvent("meta.test", "payload", eb.WithMetadata(`{"key": "value"}`))
	assert.Equal(t, `{"key": "value"}`, event.Metadata)
}

func TestWithScheduledAt(t *testing.T) {
	scheduledAt := time.Now().Add(2 * time.Hour)
	event := eb.NewEvent("scheduled.test", "payload", eb.WithScheduledAt(scheduledAt))

	assert.NotNil(t, event.ScheduledAt)
	assert.Equal(t, scheduledAt.Unix(), event.ScheduledAt.Unix())
}

func TestEventDelivery_Defaults(t *testing.T) {
	delivery := &eb.EventDelivery{
		EventID:      1,
		SubscriberID: 2,
		Status:       eb.EventStatusPending,
		MaxAttempts:  3,
	}

	assert.Equal(t, int64(1), delivery.EventID)
	assert.Equal(t, int64(2), delivery.SubscriberID)
	assert.Equal(t, eb.EventStatusPending, delivery.Status)
	assert.Equal(t, 0, delivery.Attempts)
	assert.Equal(t, 3, delivery.MaxAttempts)
	assert.Nil(t, delivery.DeliveredAt)
	assert.Nil(t, delivery.NextRetryAt)
}

func TestNewEvent_MultipleOptions(t *testing.T) {
	event := eb.NewEvent("complex.test", `{"data": "test"}`,
		eb.WithPriority(8),
		eb.WithMaxRetries(7),
		eb.WithDelaySeconds(120),
	)

	assert.Equal(t, "complex.test", event.Name)
	assert.Equal(t, `{"data": "test"}`, event.Payload)
	assert.Equal(t, eb.EventStatusPending, event.Status)
	assert.Equal(t, 8, event.Priority)
	assert.Equal(t, 7, event.MaxRetries)
	assert.Equal(t, 120, event.DelaySeconds)
	assert.NotNil(t, event.ScheduledAt)
}

func TestSubscriber_Defaults(t *testing.T) {
	sub := &eb.Subscriber{
		Name:       "test.subscriber",
		EventName:  "test.event",
		IsActive:   true,
		MaxRetries: 3,
		Timeout:    30,
		Priority:   0,
	}

	assert.Equal(t, "test.subscriber", sub.Name)
	assert.Equal(t, "test.event", sub.EventName)
	assert.True(t, sub.IsActive)
	assert.Equal(t, 3, sub.MaxRetries)
	assert.Equal(t, 30, sub.Timeout)
	assert.Equal(t, 0, sub.Priority)
	assert.Equal(t, 0, sub.FailureCount)
	assert.Nil(t, sub.Handler)
}
