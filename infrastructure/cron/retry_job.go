package cron

import (
	"context"
	"time"

	"ns-tracking-go/domain/tracking/service"

	"github.com/zeromicro/go-zero/core/logx"
)

// RetryJob runs failed event retry with backoff.
type RetryJob struct {
	retryService *service.RetryService
}

func NewRetryJob(retryService *service.RetryService) *RetryJob {
	return &RetryJob{retryService: retryService}
}

// Run executes the retry job. Called by cron scheduler.
func (j *RetryJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	logx.Info("retry_job: starting retry run")
	start := time.Now()

	result, err := j.retryService.RunRetry(ctx)
	if err != nil {
		logx.Errorf("retry_job: run failed: %v", err)
		return
	}

	logx.Infof("retry_job: completed attempted=%d succeeded=%d failed=%d dead_lettered=%d elapsed=%v",
		result.Attempted, result.Succeeded, result.Failed, result.DeadLettered, time.Since(start))
}
