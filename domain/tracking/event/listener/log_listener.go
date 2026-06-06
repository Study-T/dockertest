package listener

import "ns-tracking-go/domain/tracking/event/define"

type LogListener struct{}

func NewLogListener() *LogListener {
	return &LogListener{}
}

func (l *LogListener) Handle(event define.Event) {
	define.Log().Infof("Event: %s | EventID: %s", event.EventType(), event.EventID())
}
