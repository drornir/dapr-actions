package orch

import (
	"context"
	"sync"
)

type InMemQueueBus struct {
	channel chan Event
	once    sync.Once
}

func (b *InMemQueueBus) init() {
	b.once.Do(func() {
		b.channel = make(chan Event, 1)
	})
}

func (b *InMemQueueBus) Incoming(ctx context.Context) <-chan Event {
	b.init()
	return b.channel
}

func (b *InMemQueueBus) Emit(ctx context.Context, event Event) {
	b.init()
	go func() {
		select {
		case <-ctx.Done():
			return
		case b.channel <- event:
			return
		}
	}()
}

func (b *InMemQueueBus) Close() {
	if b.channel == nil {
		return
	}
	close(b.channel)
}
