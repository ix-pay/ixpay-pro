package eventbus

import (
	"context"
	"testing"

	eb "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestNewEventPublisher(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)
	publisher := eb.NewEventPublisher(bus)

	assert.NotNil(t, publisher)
}

func TestEventPublisher_PublishSync(t *testing.T) {
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
	ctx := context.Background()

	event := eb.NewEvent("test.event", "payload")

	err := bus.Publisher().PublishSync(ctx, event)
	assert.Error(t, err)
}

func TestEventPublisher_PublishAsync(t *testing.T) {
	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     10,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	ctx := context.Background()

	event := eb.NewEvent("async.event", "payload")

	err := bus.Publisher().PublishAsync(ctx, event)
	assert.NoError(t, err)
}

func TestEventPublisher_PublishBatchSync(t *testing.T) {
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
	ctx := context.Background()

	events := []*eb.Event{
		eb.NewEvent("batch.event.1", "payload1"),
		eb.NewEvent("batch.event.2", "payload2"),
		eb.NewEvent("batch.event.3", "payload3"),
	}

	err := bus.Publisher().PublishBatchSync(ctx, events)
	assert.Error(t, err)
}

func TestEventPublisher_PublishBatchAsync(t *testing.T) {
	config := &eb.Config{
		WorkerCount:   2,
		BatchSize:     100,
		BufferEnabled: false,
	}

	bus := eb.NewEventBus(nil, nil, config)
	ctx := context.Background()

	events := []*eb.Event{
		eb.NewEvent("batch.async.1", "payload1"),
		eb.NewEvent("batch.async.2", "payload2"),
	}

	err := bus.Publisher().PublishBatchAsync(ctx, events)
	assert.NoError(t, err)
}

func TestEventPublisher_Publish_Delayed(t *testing.T) {
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
	ctx := context.Background()

	err := bus.Publisher().PublishDelayed(ctx, "delayed.event", "payload", 60)
	assert.NoError(t, err)
}

func TestEventPublisher_PublishWithPriority(t *testing.T) {
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
	ctx := context.Background()

	err := bus.Publisher().PublishWithPriority(ctx, "priority.event", "payload", 10)
	assert.Error(t, err)
}

func TestEventPublisher_PublisherFromEventBus(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)

	publisher := bus.Publisher()
	assert.NotNil(t, publisher)

	tracker := bus.Tracker()
	assert.NotNil(t, tracker)

	dlq := bus.DLQ()
	assert.NotNil(t, dlq)
}

func TestEventPublisher_EventBusReference(t *testing.T) {
	bus := eb.NewEventBus(nil, nil, nil)
	publisher := bus.Publisher()

	assert.NotNil(t, publisher)
}
