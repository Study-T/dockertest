package handler

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type RetryHandler struct{}

func NewRetryHandler() *RetryHandler {
	return &RetryHandler{}
}

func (h *RetryHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	fmt.Printf("RetryHandler processing task: %s\n", task.Type())
	return nil
}

func RegisterRetryHandler(mux *asynq.ServeMux) {
	mux.HandleFunc("task:retry", NewRetryHandler().ProcessTask)
}
