package eventbus

import (
	"context"
	"testing"

	eb "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	"github.com/stretchr/testify/assert"
)

type mockEventRepo struct {
	subscribers        map[int64]*eb.Subscriber
	events             map[int64]*eb.Event
	deliveries         []*eb.EventDelivery
	subscribersByEvent map[string][]*eb.Subscriber
	nextID             int64
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		subscribers:        make(map[int64]*eb.Subscriber),
		events:             make(map[int64]*eb.Event),
		deliveries:         make([]*eb.EventDelivery, 0),
		subscribersByEvent: make(map[string][]*eb.Subscriber),
		nextID:             1,
	}
}

func (m *mockEventRepo) CreateEvent(ctx context.Context, event *eb.Event) error {
	event.ID = m.nextID
	m.nextID++
	m.events[event.ID] = event
	return nil
}

func (m *mockEventRepo) GetEventByID(ctx context.Context, id int64) (*eb.Event, error) {
	if event, ok := m.events[id]; ok {
		return event, nil
	}
	return nil, nil
}

func (m *mockEventRepo) UpdateEvent(ctx context.Context, event *eb.Event) error {
	m.events[event.ID] = event
	return nil
}

func (m *mockEventRepo) UpdateEventStatus(ctx context.Context, id int64, status eb.EventStatus, errorMessage string) error {
	if event, ok := m.events[id]; ok {
		event.Status = status
		event.ErrorMessage = errorMessage
	}
	return nil
}

func (m *mockEventRepo) ListEvents(ctx context.Context, filter eb.EventFilter) ([]*eb.Event, int64, error) {
	events := make([]*eb.Event, 0, len(m.events))
	for _, e := range m.events {
		events = append(events, e)
	}
	return events, int64(len(events)), nil
}

func (m *mockEventRepo) CreateDelivery(ctx context.Context, delivery *eb.EventDelivery) error {
	m.deliveries = append(m.deliveries, delivery)
	return nil
}

func (m *mockEventRepo) UpdateDelivery(ctx context.Context, delivery *eb.EventDelivery) error {
	for i, d := range m.deliveries {
		if d.EventID == delivery.EventID && d.SubscriberID == delivery.SubscriberID {
			m.deliveries[i] = delivery
			return nil
		}
	}
	return nil
}

func (m *mockEventRepo) ListDeliveriesByEventID(ctx context.Context, eventID int64) ([]*eb.EventDelivery, error) {
	result := make([]*eb.EventDelivery, 0)
	for _, d := range m.deliveries {
		if d.EventID == eventID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockEventRepo) ListDeliveriesBySubscriberID(ctx context.Context, subscriberID int64) ([]*eb.EventDelivery, error) {
	return nil, nil
}

func (m *mockEventRepo) ListFailedDeliveries(ctx context.Context, limit int) ([]*eb.EventDelivery, error) {
	return nil, nil
}

func (m *mockEventRepo) CreateSubscriber(ctx context.Context, subscriber *eb.Subscriber) error {
	subscriber.ID = m.nextID
	m.nextID++
	m.subscribers[subscriber.ID] = subscriber
	m.subscribersByEvent[subscriber.EventName] = append(m.subscribersByEvent[subscriber.EventName], subscriber)
	return nil
}

func (m *mockEventRepo) GetSubscriberByID(ctx context.Context, id int64) (*eb.Subscriber, error) {
	if sub, ok := m.subscribers[id]; ok {
		return sub, nil
	}
	return nil, nil
}

func (m *mockEventRepo) GetSubscriberByName(ctx context.Context, name string) (*eb.Subscriber, error) {
	for _, sub := range m.subscribers {
		if sub.Name == name {
			return sub, nil
		}
	}
	return nil, nil
}

func (m *mockEventRepo) UpdateSubscriber(ctx context.Context, subscriber *eb.Subscriber) error {
	m.subscribers[subscriber.ID] = subscriber
	return nil
}

func (m *mockEventRepo) ListSubscribers(ctx context.Context, filter eb.SubscriberFilter) ([]*eb.Subscriber, int64, error) {
	if filter.IsActive != nil {
		result := make([]*eb.Subscriber, 0)
		for _, sub := range m.subscribers {
			if sub.IsActive == *filter.IsActive {
				result = append(result, sub)
			}
		}
		return result, int64(len(result)), nil
	}
	result := make([]*eb.Subscriber, 0, len(m.subscribers))
	for _, sub := range m.subscribers {
		result = append(result, sub)
	}
	return result, int64(len(result)), nil
}

func (m *mockEventRepo) DeleteSubscriber(ctx context.Context, id int64) error {
	if sub, ok := m.subscribers[id]; ok {
		delete(m.subscribers, id)
		subs := m.subscribersByEvent[sub.EventName]
		for i, s := range subs {
			if s.ID == id {
				m.subscribersByEvent[sub.EventName] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (m *mockEventRepo) ListDeadLetters(ctx context.Context, filter eb.EventFilter) ([]*eb.Event, int64, error) {
	return nil, 0, nil
}

func (m *mockEventRepo) MoveToDeadLetter(ctx context.Context, eventID int64, reason string) error {
	if event, ok := m.events[eventID]; ok {
		event.Status = eb.EventStatusDeadLetter
		event.ErrorMessage = reason
	}
	return nil
}

func (m *mockEventRepo) GetEventCount(ctx context.Context, status eb.EventStatus) (int64, error) {
	count := int64(0)
	for _, e := range m.events {
		if status < 0 || e.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *mockEventRepo) GetSubscriberCount(ctx context.Context, isActive bool) (int64, error) {
	count := int64(0)
	for _, sub := range m.subscribers {
		if sub.IsActive == isActive {
			count++
		}
	}
	return count, nil
}

func TestNewSubscriberManager(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)

	assert.NotNil(t, sm)
}

func TestAddSubscriber(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)
	ctx := context.Background()

	sub := &eb.Subscriber{
		Name:       "test.subscriber",
		EventName:  "test.event",
		IsActive:   true,
		MaxRetries: 3,
		Timeout:    30,
		Priority:   1,
	}

	err := sm.AddSubscriber(ctx, sub)
	assert.NoError(t, err)
	assert.Greater(t, sub.ID, int64(0))

	subs := sm.GetSubscribers("test.event")
	assert.Len(t, subs, 1)
	assert.Equal(t, "test.subscriber", subs[0].Name)
}

func TestRemoveSubscriber(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)
	ctx := context.Background()

	sub := &eb.Subscriber{
		Name:      "removable.sub",
		EventName: "test.event",
		IsActive:  true,
	}

	sm.AddSubscriber(ctx, sub)
	subID := sub.ID

	err := sm.RemoveSubscriber(ctx, subID)
	assert.NoError(t, err)

	subs := sm.GetSubscribers("test.event")
	assert.Len(t, subs, 0)
}

func TestGetSubscribers_SortedByPriority(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)
	ctx := context.Background()

	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "low.priority",
		EventName: "test.event",
		Priority:  10,
	})
	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "high.priority",
		EventName: "test.event",
		Priority:  1,
	})
	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "medium.priority",
		EventName: "test.event",
		Priority:  5,
	})

	subs := sm.GetSubscribers("test.event")
	assert.Len(t, subs, 3)
	assert.Equal(t, "high.priority", subs[0].Name)
	assert.Equal(t, "medium.priority", subs[1].Name)
	assert.Equal(t, "low.priority", subs[2].Name)
}

func TestUpdateSubscriber(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)
	ctx := context.Background()

	sub := &eb.Subscriber{
		Name:      "update.sub",
		EventName: "test.event",
		IsActive:  true,
		Priority:  1,
	}

	sm.AddSubscriber(ctx, sub)
	sub.Priority = 5
	sub.IsActive = false

	err := sm.UpdateSubscriber(ctx, sub)
	assert.NoError(t, err)

	subs := sm.GetSubscribers("test.event")
	assert.Len(t, subs, 1)
	assert.False(t, subs[0].IsActive)
	assert.Equal(t, 5, subs[0].Priority)
}

func TestGetSubscribers_Empty(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)

	subs := sm.GetSubscribers("non.existent.event")
	assert.Len(t, subs, 0)
}

func TestRefresh(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)
	ctx := context.Background()

	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "refresh.sub",
		EventName: "test.event",
		IsActive:  true,
	})

	err := sm.Refresh(ctx)
	assert.NoError(t, err)

	subs := sm.GetSubscribers("test.event")
	assert.Len(t, subs, 1)
}

func TestGetSubscriberCount(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)
	ctx := context.Background()

	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "active.sub",
		EventName: "test.event",
		IsActive:  true,
	})
	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "inactive.sub",
		EventName: "test.event",
		IsActive:  false,
	})

	activeCount, err := sm.GetSubscriberCount(ctx, true)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), activeCount)

	inactiveCount, err := sm.GetSubscriberCount(ctx, false)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), inactiveCount)
}

func TestMultipleEvents(t *testing.T) {
	repo := newMockEventRepo()
	sm := eb.NewSubscriberManager(repo)
	ctx := context.Background()

	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "sub1",
		EventName: "event1",
	})
	sm.AddSubscriber(ctx, &eb.Subscriber{
		Name:      "sub2",
		EventName: "event2",
	})

	subs1 := sm.GetSubscribers("event1")
	assert.Len(t, subs1, 1)
	assert.Equal(t, "sub1", subs1[0].Name)

	subs2 := sm.GetSubscribers("event2")
	assert.Len(t, subs2, 1)
	assert.Equal(t, "sub2", subs2[0].Name)
}
