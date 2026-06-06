package event

import (
	"sync"
	"sync/atomic"

	"ns-tracking-go/domain/tracking/event/define"
	"ns-tracking-go/domain/tracking/event/dispatcher"

	"github.com/zeromicro/go-zero/core/logx"
)

var droppedEvents uint64

func DroppedEvents() uint64 { return atomic.LoadUint64(&droppedEvents) }

type EventBus struct {
	inner     *dispatcher.InMemoryDispatcher
	eventCh   chan define.Event
	quit      chan struct{}
	closed    atomic.Bool
	closeOnce sync.Once
}

func NewEventBus() *EventBus {
	bus := &EventBus{
		inner:   dispatcher.NewInMemoryDispatcher(),
		eventCh: make(chan define.Event, 1024),
		quit:    make(chan struct{}),
	}
	logx.Info("EventBus started")
	go bus.run()
	return bus
}

func (b *EventBus) Register(eventType string, listener define.Listener) {
	b.inner.Register(eventType, listener)
}

func (b *EventBus) RegisterMultiple(eventType string, listeners []define.Listener) {
	b.inner.RegisterMultiple(eventType, listeners)
}

func (b *EventBus) Dispatch(event define.Event) {
	if b.closed.Load() {
		return
	}
	select {
	case b.eventCh <- event:
		logx.Infof("Event queued: %s", event.EventType())
	default:
		atomic.AddUint64(&droppedEvents, 1)
		logx.Errorf("Event channel full, dropping: %s (total dropped: %d)", event.EventType(), DroppedEvents())
	}
}

func (b *EventBus) Close() {
	b.closeOnce.Do(func() {
		b.closed.Store(true)
		close(b.quit)
		for len(b.eventCh) > 0 {
			event := <-b.eventCh
			b.inner.Dispatch(event)
		}
	})
}

func (b *EventBus) run() {
	for {
		select {
		case event := <-b.eventCh:
			b.inner.Dispatch(event)
		case <-b.quit:
			return
		}
	}
}
