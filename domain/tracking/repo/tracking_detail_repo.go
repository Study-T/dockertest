package repo

import (
	"context"

	"ns-tracking-go/domain/tracking/entity"
)

// TrackingDetailRepo defines persistence for tracking_details.
type TrackingDetailRepo interface {
	Save(ctx context.Context, detail *entity.TrackingDetail) error
	Update(ctx context.Context, detail *entity.TrackingDetail) error
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingDetail, error)
}
