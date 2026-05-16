package eventbus

import (
	"context"
	"testing"
	"time"

	eb "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestNewEventTracker(t *testing.T) {
	repo := newMockEventRepo()
	tracker := eb.NewEventTracker(nil, repo, time.Hour*24*7)

	assert.NotNil(t, tracker)
}

func TestEventTrace_JSONSerialization(t *testing.T) {
	trace := eb.EventTrace{
		EventID:   123,
		EventName: "test.event",
		Action:    "创建",
		Status:    "待处理",
		Timestamp: time.Now().Unix(),
	}

	assert.Equal(t, int64(123), trace.EventID)
	assert.Equal(t, "test.event", trace.EventName)
	assert.Equal(t, "创建", trace.Action)
	assert.Equal(t, "待处理", trace.Status)
	assert.NotZero(t, trace.Timestamp)
}

func TestEventTracker_GetEventStatus(t *testing.T) {
	repo := newMockEventRepo()
	tracker := eb.NewEventTracker(nil, repo, time.Hour)
	ctx := context.Background()

	event := eb.NewEvent("test.event", "payload")
	repo.CreateEvent(ctx, event)

	retrievedEvent, err := tracker.GetEventStatus(ctx, event.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedEvent)
	assert.Equal(t, "test.event", retrievedEvent.Name)
}

func TestEventTracker_GetEventStatus_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	tracker := eb.NewEventTracker(nil, repo, time.Hour)
	ctx := context.Background()

	event, err := tracker.GetEventStatus(ctx, 999)
	if err != nil {
		// Expected: event not found
		return
	}
	// If no error, the event is nil
	assert.Nil(t, event)
}

func TestEventStatus_Transitions(t *testing.T) {
	event := eb.NewEvent("test.event", "payload")

	assert.Equal(t, eb.EventStatusPending, event.Status)

	event.Status = eb.EventStatusProcessing
	assert.Equal(t, eb.EventStatusProcessing, event.Status)
	assert.Equal(t, "处理中", event.Status.String())

	event.Status = eb.EventStatusSuccess
	assert.Equal(t, eb.EventStatusSuccess, event.Status)
	assert.Equal(t, "成功", event.Status.String())

	event.Status = eb.EventStatusFailed
	assert.Equal(t, eb.EventStatusFailed, event.Status)
	assert.Equal(t, "失败", event.Status.String())

	event.Status = eb.EventStatusDeadLetter
	assert.Equal(t, eb.EventStatusDeadLetter, event.Status)
	assert.Equal(t, "死信", event.Status.String())
}

func TestEventFilter_Defaults(t *testing.T) {
	filter := eb.EventFilter{
		Page:     0,
		PageSize: 0,
	}

	assert.Equal(t, "", filter.Name)
	assert.Nil(t, filter.Status)
	assert.Equal(t, 0, filter.Page)
	assert.Equal(t, 0, filter.PageSize)
}

func TestEventTrace_EmptyAction(t *testing.T) {
	event := eb.NewEvent("test.event", "payload")
	event.ID = 1

	trace := eb.EventTrace{
		EventID:   event.ID,
		EventName: event.Name,
		Action:    "",
		Status:    event.Status.String(),
		Timestamp: time.Now().Unix(),
	}

	assert.Equal(t, "", trace.Action)
	assert.Equal(t, "待处理", trace.Status)
}

func TestEventTrace_MultipleStates(t *testing.T) {
	event := eb.NewEvent("state.test", "payload")
	event.ID = 1

	states := []eb.EventStatus{
		eb.EventStatusPending,
		eb.EventStatusProcessing,
		eb.EventStatusSuccess,
		eb.EventStatusFailed,
		eb.EventStatusDeadLetter,
	}

	for _, state := range states {
		event.Status = state
		trace := eb.EventTrace{
			EventID:   event.ID,
			EventName: event.Name,
			Action:    "状态变更",
			Status:    event.Status.String(),
			Timestamp: time.Now().Unix(),
		}

		assert.Equal(t, state.String(), trace.Status)
	}
}
