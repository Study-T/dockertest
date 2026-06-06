package main

import (
	"ns-tracking-go/domain/tracking/queue/handler"

	"github.com/hibiken/asynq"
)

func RegisterAllHandlers(mux *asynq.ServeMux) {
	handler.RegisterRetryHandler(mux)
}
