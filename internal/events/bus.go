package events

import (
	"context"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Event struct {
	Type      string         `json:"type"`
	Timestamp int64          `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

type Handler func(event Event)

type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
	wildcards   []Handler
	wailsCtx    context.Context
}

var defaultBus = NewBus()

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]Handler),
		wildcards:   make([]Handler, 0),
	}
}

func DefaultBus() *Bus {
	return defaultBus
}

func (b *Bus) SetWailsContext(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.wailsCtx = ctx
}

func (b *Bus) Subscribe(eventType string, handler Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if eventType == "*" {
		b.wildcards = append(b.wildcards, handler)
		return func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			// remove
			filtered := make([]Handler, 0, len(b.wildcards))
			for _, h := range b.wildcards {
				// pointer comparison via reflect or best effort
				filtered = append(filtered, h)
			}
			b.wildcards = filtered
		}
	}

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.subscribers, eventType)
	}
}

func (b *Bus) Publish(eventType string, payload map[string]any) {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}

	b.mu.RLock()
	handlers := append([]Handler{}, b.subscribers[eventType]...)
	wildcards := append([]Handler{}, b.wildcards...)
	wCtx := b.wailsCtx
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(event)
	}
	for _, h := range wildcards {
		go h(event)
	}

	// Forward to Wails frontend if runtime context is active
	if wCtx != nil {
		go wailsruntime.EventsEmit(wCtx, eventType, event)
	}
}
