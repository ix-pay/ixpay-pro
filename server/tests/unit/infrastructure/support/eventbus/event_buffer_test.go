package eventbus

import (
	"context"
	"testing"
	"time"

	eb "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestEventBuffer_New(t *testing.T) {
	buffer := eb.NewEventBuffer(nil, 100, time.Second*5)

	assert.NotNil(t, buffer)
}

func TestEventBuffer_SetFlushFunc(t *testing.T) {
	buffer := eb.NewEventBuffer(nil, 100, time.Second*5)

	called := false
	buffer.SetFlushFunc(func(events []*eb.Event) error {
		called = true
		return nil
	})

	assert.False(t, called)
}

func TestEventBuffer_Stop_Empty(t *testing.T) {
	buffer := eb.NewEventBuffer(nil, 100, time.Second*5)

	err := buffer.Stop(context.Background())
	assert.NoError(t, err)
}

func TestEventTrace(t *testing.T) {
	event := eb.NewEvent("test.event", "payload")
	event.ID = 123

	trace := eb.EventTrace{
		EventID:   event.ID,
		EventName: event.Name,
		Action:    "测试操作",
		Status:    event.Status.String(),
		Timestamp: time.Now().Unix(),
	}

	assert.Equal(t, int64(123), trace.EventID)
	assert.Equal(t, "test.event", trace.EventName)
	assert.Equal(t, "测试操作", trace.Action)
	assert.Equal(t, "待处理", trace.Status)
}

func TestEventFilter(t *testing.T) {
	status := eb.EventStatusPending
	filter := eb.EventFilter{
		Name:      "test.event",
		Status:    &status,
		Page:      1,
		PageSize:  20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	assert.Equal(t, "test.event", filter.Name)
	assert.Equal(t, eb.EventStatusPending, *filter.Status)
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 20, filter.PageSize)
}

func TestSubscriberFilter(t *testing.T) {
	isActive := true
	filter := eb.SubscriberFilter{
		EventName: "test.event",
		IsActive:  &isActive,
		Page:      1,
		PageSize:  10,
	}

	assert.Equal(t, "test.event", filter.EventName)
	assert.True(t, *filter.IsActive)
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 10, filter.PageSize)
}

func TestEventStatus_Constants(t *testing.T) {
	assert.Equal(t, eb.EventStatus(0), eb.EventStatusPending)
	assert.Equal(t, eb.EventStatus(1), eb.EventStatusProcessing)
	assert.Equal(t, eb.EventStatus(2), eb.EventStatusSuccess)
	assert.Equal(t, eb.EventStatus(3), eb.EventStatusFailed)
	assert.Equal(t, eb.EventStatus(4), eb.EventStatusDeadLetter)
}

func TestConfig_Defaults(t *testing.T) {
	config := eb.DefaultConfig()

	assert.Equal(t, 4, config.WorkerCount)
	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, time.Second*5, config.FlushInterval)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5, config.DeadLetterAfter)
	assert.True(t, config.BufferEnabled)
}

func TestConfig_Custom(t *testing.T) {
	config := &eb.Config{
		WorkerCount:     8,
		BatchSize:       50,
		FlushInterval:   time.Second * 10,
		MaxRetries:      5,
		DeadLetterAfter: 10,
		BufferEnabled:   false,
	}

	assert.Equal(t, 8, config.WorkerCount)
	assert.Equal(t, 50, config.BatchSize)
	assert.Equal(t, time.Second*10, config.FlushInterval)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 10, config.DeadLetterAfter)
	assert.False(t, config.BufferEnabled)
}
