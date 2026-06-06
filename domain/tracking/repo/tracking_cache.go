package repo

import (
	"context"
	"time"
)

// TrackingCache defines caching interface for tracking data.
type TrackingCache interface {
	// GetTrackingDetail retrieves cached tracking detail.
	GetTrackingDetail(ctx context.Context, orderNumber string) (string, error)

	// SetTrackingDetail caches tracking detail with TTL.
	SetTrackingDetail(ctx context.Context, orderNumber, data string, ttl time.Duration) error

	// DeleteTrackingDetail removes cached tracking detail.
	DeleteTrackingDetail(ctx context.Context, orderNumber string) error
}
