package dispatcher

import (
	"sync"

	"ns-tracking-go/domain/tracking/event/define"
)

type InMemoryDispatcher struct {
	listeners map[string][]define.Listener
	mu        sync.RWMutex
}

func NewInMemoryDispatcher() *InMemoryDispatcher {
	return &InMemoryDispatcher{
		listeners: make(map[string][]define.Listener),
	}
}

func (d *InMemoryDispatcher) Dispatch(event define.Event) {
	d.mu.RLock()
	listeners := d.listeners[event.EventType()]
	d.mu.RUnlock()

	if len(listeners) == 0 {
		return
	}

	for _, listener := range listeners {
		func() {
			defer func() {
				if r := recover(); r != nil {
					define.Log().Errorf("listener panic: event=%s panic=%v", event.EventType(), r)
				}
			}()
			listener.Handle(event)
		}()
	}
	define.Log().Infof("Event dispatched: %s, listeners=%d", event.EventType(), len(listeners))
}

func (d *InMemoryDispatcher) Register(eventType string, listener define.Listener) {
	d.mu.Lock()
	d.listeners[eventType] = append(d.listeners[eventType], listener)
	d.mu.Unlock()
	define.Log().Infof("Listener registered: event_type=%s", eventType)
}

func (d *InMemoryDispatcher) RegisterMultiple(eventType string, listeners []define.Listener) {
	d.mu.Lock()
	d.listeners[eventType] = append(d.listeners[eventType], listeners...)
	d.mu.Unlock()
}
