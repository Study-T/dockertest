package event

import "ns-tracking-go/domain/tracking/event/define"

func RegisterListeners(d define.Dispatcher) {
	define.Log().Infof("Listeners registered")
}
