package define

import "time"

type Event interface {
	EventID() string
	EventType() string
	EventTimestamp() int64
}

type Dispatcher interface {
	Dispatch(event Event)
	Register(eventType string, listener Listener)
	RegisterMultiple(eventType string, listeners []Listener)
}

type Listener interface {
	Handle(event Event)
}

type Logger interface {
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type defaultLogger struct{}

func (defaultLogger) Infof(format string, v ...interface{})  {}
func (defaultLogger) Errorf(format string, v ...interface{}) {}

var log Logger = defaultLogger{}

func SetLogger(l Logger) { log = l }
func Log() Logger        { return log }

func NewEvent(eventType, eventID string) *BaseEvent {
	return &BaseEvent{eventType: eventType, eventID: eventID, timestamp: time.Now().Unix()}
}

type BaseEvent struct {
	eventType string
	eventID   string
	timestamp int64
}

func (e *BaseEvent) EventID() string       { return e.eventID }
func (e *BaseEvent) EventType() string     { return e.eventType }
func (e *BaseEvent) EventTimestamp() int64 { return e.timestamp }
