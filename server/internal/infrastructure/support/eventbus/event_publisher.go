package eventbus

import (
	"context"
	"fmt"
)

// EventPublisher 事件发布器
type EventPublisher struct {
	eventBus *EventBus
}

// NewEventPublisher 创建事件发布器
func NewEventPublisher(eventBus *EventBus) *EventPublisher {
	return &EventPublisher{
		eventBus: eventBus,
	}
}

// PublishSync 同步发布事件
func (p *EventPublisher) PublishSync(ctx context.Context, event *Event) error {
	return p.eventBus.Publish(ctx, event)
}

// PublishAsync 异步发布事件
func (p *EventPublisher) PublishAsync(ctx context.Context, event *Event) error {
	return p.eventBus.PublishAsync(ctx, event)
}

// PublishBatchSync 同步批量发布事件
func (p *EventPublisher) PublishBatchSync(ctx context.Context, events []*Event) error {
	for _, event := range events {
		if err := p.PublishSync(ctx, event); err != nil {
			return fmt.Errorf("批量发布事件失败，事件: %s, 错误: %w", event.Name, err)
		}
	}
	return nil
}

// PublishBatchAsync 异步批量发布事件
func (p *EventPublisher) PublishBatchAsync(ctx context.Context, events []*Event) error {
	for _, event := range events {
		if err := p.PublishAsync(ctx, event); err != nil {
			return fmt.Errorf("批量异步发布事件失败，事件: %s, 错误: %w", event.Name, err)
		}
	}
	return nil
}

// PublishDelayed 发布延迟事件
func (p *EventPublisher) PublishDelayed(ctx context.Context, name string, payload string, delaySeconds int) error {
	event := NewEvent(name, payload, WithDelaySeconds(delaySeconds))
	return p.PublishAsync(ctx, event)
}

// PublishWithPriority 发布带优先级的事件
func (p *EventPublisher) PublishWithPriority(ctx context.Context, name string, payload string, priority int) error {
	event := NewEvent(name, payload, WithPriority(priority))
	return p.PublishSync(ctx, event)
}
