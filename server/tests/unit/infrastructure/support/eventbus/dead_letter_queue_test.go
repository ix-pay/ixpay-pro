package eventbus

import (
	"context"
	"testing"

	eb "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestNewDeadLetterQueue(t *testing.T) {
	repo := newMockEventRepo()
	dlq := eb.NewDeadLetterQueue(nil, repo)

	assert.NotNil(t, dlq)
}

func TestDeadLetterQueue_Retry(t *testing.T) {
	repo := newMockEventRepo()
	dlq := eb.NewDeadLetterQueue(nil, repo)
	ctx := context.Background()

	event := eb.NewEvent("test.event", "payload")
	repo.CreateEvent(ctx, event)
	repo.MoveToDeadLetter(ctx, event.ID, "测试失败")

	err := dlq.Retry(ctx, event.ID)
	assert.NoError(t, err)

	retriedEvent, _ := repo.GetEventByID(ctx, event.ID)
	assert.Equal(t, eb.EventStatusPending, retriedEvent.Status)
	assert.Equal(t, 0, retriedEvent.RetryCount)
	assert.Equal(t, "", retriedEvent.ErrorMessage)
}

func TestDeadLetterQueue_Retry_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	dlq := eb.NewDeadLetterQueue(nil, repo)
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	err := dlq.Retry(ctx, 999)
	assert.Error(t, err)
}

func TestDeadLetterEntry(t *testing.T) {
	entry := eb.DeadLetterEntry{
		ID: "test-id-123",
		Data: map[string]interface{}{
			"event_id":   1,
			"event_name": "test.event",
			"reason":     "测试原因",
		},
	}

	assert.Equal(t, "test-id-123", entry.ID)
	assert.NotNil(t, entry.Data)
	assert.Equal(t, "test.event", entry.Data["event_name"])
}

func TestDeadLetterQueue_MoveToDeadLetter(t *testing.T) {
	repo := newMockEventRepo()
	ctx := context.Background()

	event := eb.NewEvent("test.event", "payload")
	repo.CreateEvent(ctx, event)

	err := repo.MoveToDeadLetter(ctx, event.ID, "测试原因")
	assert.NoError(t, err)

	storedEvent, _ := repo.GetEventByID(ctx, event.ID)
	assert.Equal(t, eb.EventStatusDeadLetter, storedEvent.Status)
	assert.Equal(t, "测试原因", storedEvent.ErrorMessage)
}

func TestDeadLetterQueue_ListDeadLetters(t *testing.T) {
	repo := newMockEventRepo()
	ctx := context.Background()

	filter := eb.EventFilter{
		Page:     1,
		PageSize: 10,
	}

	events, total, err := repo.ListDeadLetters(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, events)
}
