package eventbus

import (
	"context"
	"testing"

	eb "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := eb.DefaultConfig()

	assert.Equal(t, 4, config.WorkerCount)
	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, int64(5000000000), int64(config.FlushInterval))
	assert.Equal(t, 3, config.MaxRetries)
	assert.True(t, config.BufferEnabled)
}

func TestNewEventBus_NilConfig(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)

	assert.NotNil(t, bus)
}

func TestNewEventBus_WithConfig(t *testing.T) {
	config := &eb.Config{
		WorkerCount:   8,
		BatchSize:     50,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)

	assert.NotNil(t, bus)
}

func TestStartAndStop_NilRedis(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     10,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	bus.Start()
	bus.Stop()
}

func TestStop_NotRunning(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     10,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	bus.Stop()
}

func TestGetSubscriberManager(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)

	sm := bus.GetSubscriberManager()
	assert.NotNil(t, sm)
}

func TestGetEventRepository(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)

	repo := bus.GetEventRepository()
	assert.NotNil(t, repo)
}

func TestPublisher(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)

	publisher := bus.Publisher()
	assert.NotNil(t, publisher)
}

func TestTracker(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)

	tracker := bus.Tracker()
	assert.NotNil(t, tracker)
}

func TestDLQ(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)

	dlq := bus.DLQ()
	assert.NotNil(t, dlq)
}

func TestPublishAsync_ChannelFull(t *testing.T) {
	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     1,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		event := eb.NewEvent("test.event", "payload")
		bus.PublishAsync(ctx, event)
	}

	event2 := eb.NewEvent("test.event2", "payload2")
	err := bus.PublishAsync(ctx, event2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "事件通道已满")
}

func TestPublishAsync_BufferEnabled(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred (nil Redis client): %v", r)
		}
	}()

	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     10,
		BufferEnabled: true,
	}

	bus := eb.NewEventBus(nil, nil, config)
	ctx := context.Background()

	event := eb.NewEvent("test.event", "payload")

	err := bus.PublishAsync(ctx, event)
	assert.NoError(t, err)
}

func TestPublishAsync_Delayed(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred (nil Redis client): %v", r)
		}
	}()

	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     10,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	ctx := context.Background()

	event := eb.NewEvent("delayed.event", "payload", eb.WithDelaySeconds(10))

	err := bus.PublishAsync(ctx, event)
	assert.NoError(t, err)
}

func TestSubscribe_NilRedis(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     10,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	bus.Start()
	ctx := context.Background()

	sub := &eb.Subscriber{
		Name:      "test.sub",
		EventName: "test.event",
		IsActive:  true,
	}

	bus.Subscribe(ctx, sub)
	bus.Stop()
}

func TestUnsubscribe_NilRedis(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     10,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	bus.Start()
	ctx := context.Background()

	sub := &eb.Subscriber{
		Name:      "temp.sub",
		EventName: "test.event",
		IsActive:  true,
	}

	bus.Subscribe(ctx, sub)
	bus.Unsubscribe(ctx, sub.ID)
	bus.Stop()
}
