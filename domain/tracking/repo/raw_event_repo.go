package repo

import (
	"context"

	"ns-tracking-go/domain/tracking/entity"
)

// SaveResult indicates whether the insert was new or a duplicate.
type SaveResult struct {
	Event *entity.RawEvent
	IsNew bool
}

// RawEventRepo defines persistence operations for raw webhook events.
type RawEventRepo interface {
	Save(ctx context.Context, event *entity.RawEvent) (int64, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*entity.RawEvent, error)
	UpdateStatus(ctx context.Context, id int64, status string, errMsg string) error
	IncrementRetry(ctx context.Context, id int64) error
	MarkDeadLettered(ctx context.Context, id int64, errMsg string) error
	FindPending(ctx context.Context, limit int) ([]*entity.RawEvent, error)
	FindRetryable(ctx context.Context, limit int) ([]*entity.RawEvent, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
}
