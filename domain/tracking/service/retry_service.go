package service

import (
	"context"
	"encoding/json"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"

	"github.com/zeromicro/go-zero/core/logx"
)

// Backoff schedule: 30s → 120s → 600s → 3600s
var backoffSchedule = []time.Duration{
	30 * time.Second,
	120 * time.Second,
	600 * time.Second,
	3600 * time.Second,
}

const retryLogPrefix = "retry_job"

// RetryService handles retrying failed raw events.
type RetryService struct {
	rawEventRepo repo.RawEventRepo
	detailSvc    *TrackingDetailService
}

func NewRetryService(rawEventRepo repo.RawEventRepo, detailSvc *TrackingDetailService) *RetryService {
	return &RetryService{rawEventRepo: rawEventRepo, detailSvc: detailSvc}
}

// RetryResult holds the outcome of a retry run.
type RetryResult struct {
	Attempted int
	Succeeded int
	Failed    int
	DeadLettered int
}

// RunRetry finds retryable events and processes them.
func (s *RetryService) RunRetry(ctx context.Context) (*RetryResult, error) {
	events, err := s.rawEventRepo.FindRetryable(ctx, 50)
	if err != nil {
		return nil, err
	}

	result := &RetryResult{}
	for _, evt := range events {
		if !s.shouldRetry(evt) {
			continue
		}
		result.Attempted++

		if err := s.retryEvent(ctx, evt); err != nil {
			result.Failed++
			logx.Errorf("%s: retry failed id=%d err=%v", retryLogPrefix, evt.ID, err)
			continue
		}

		result.Succeeded++
		if err := s.rawEventRepo.UpdateStatus(ctx, evt.ID, entity.RawEventProcessed, ""); err != nil {
			logx.Errorf("%s: mark processed failed id=%d err=%v", retryLogPrefix, evt.ID, err)
		}
	}

	logx.Infof("%s: attempted=%d succeeded=%d failed=%d dead_lettered=%d",
		retryLogPrefix, result.Attempted, result.Succeeded, result.Failed, result.DeadLettered)
	return result, nil
}

func (s *RetryService) shouldRetry(evt *entity.RawEvent) bool {
	if evt.RetryCount >= evt.MaxRetries {
		return false
	}
	backoffIdx := evt.RetryCount
	if backoffIdx >= len(backoffSchedule) {
		backoffIdx = len(backoffSchedule) - 1
	}
	return time.Since(evt.UpdatedAt) >= backoffSchedule[backoffIdx]
}

func (s *RetryService) retryEvent(ctx context.Context, evt *entity.RawEvent) error {
	if err := s.rawEventRepo.IncrementRetry(ctx, evt.ID); err != nil {
		return err
	}

	// Parse payload back to map and re-run normalization
	data := parseJSONToMap(evt.Payload)
	if err := s.detailSvc.SaveFromWebhook(ctx, data); err != nil {
		if evt.RetryCount+1 >= evt.MaxRetries {
			_ = s.rawEventRepo.MarkDeadLettered(ctx, evt.ID, err.Error())
		}
		return err
	}
	return nil
}

func parseJSONToMap(s string) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil
	}
	return m
}
