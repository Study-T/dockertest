package repo

import (
	"context"
	"time"

	"ns-tracking-go/domain/tracking/entity"
)

// TrackingLogRepo defines persistence for tracking_logs (read + limited writes).
type TrackingLogRepo interface {
	FindBySourceTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingLog, error)
	Create(ctx context.Context, log *entity.TrackingLog) error
	UpdateTrackingFields(ctx context.Context, id int64, trackStatus int, syncedAt time.Time, receivedAt, deliveredAt, trackedAt *time.Time) error
	UpdateTrackingFieldsBySourceNumber(ctx context.Context, sourceTrackingNumber string, trackStatus int, syncedAt time.Time, receivedAt, deliveredAt, trackedAt *time.Time) error
	FindSyncCandidates(ctx context.Context, limit int, cacheHours int) ([]*entity.TrackingLog, error)
}
