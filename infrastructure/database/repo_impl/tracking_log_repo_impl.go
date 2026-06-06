package repo_impl

import (
	"context"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"
	"ns-tracking-go/infrastructure/database"

	"gorm.io/gorm"
)

type trackingLogModel struct {
	ID                   int64      `gorm:"column:id;primaryKey"`
	TrackingNumber       string     `gorm:"column:tracking_number"`
	SourceTrackingNumber string     `gorm:"column:source_tracking_number"`
	ChannelAlias         string     `gorm:"column:channel_alias"`
	ShippingAgent        string     `gorm:"column:shipping_agent"`
	ShippingChannel      string     `gorm:"column:shipping_channel"`
	CountryCode          string     `gorm:"column:country_code"`
	TrackStatus          int        `gorm:"column:track_status"`
	SyncedAt             *time.Time `gorm:"column:synced_at"`
	ReceivedAt           *time.Time `gorm:"column:received_at"`
	DeliveredAt          *time.Time `gorm:"column:delivered_at"`
	TrackedAt            *time.Time `gorm:"column:tracked_at"`
	FulfillAt            *time.Time `gorm:"column:fulfill_at"`
}

func (trackingLogModel) TableName() string { return "tracking_logs" }

type TrackingLogRepoImpl struct {
	db *database.DB
}

func NewTrackingLogRepoImpl(db *database.DB) *TrackingLogRepoImpl {
	return &TrackingLogRepoImpl{db: db}
}

var _ repo.TrackingLogRepo = (*TrackingLogRepoImpl)(nil)

func (r *TrackingLogRepoImpl) FindBySourceTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingLog, error) {
	var m trackingLogModel
	err := r.db.WithContext(ctx).
		Where("source_tracking_number = ?", trackingNumber).
		Order("id DESC").First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil when not found, not error
		}
		return nil, fmt.Errorf("find tracking log: %w", err)
	}
	return toTrackingLogEntity(&m), nil
}

func (r *TrackingLogRepoImpl) Create(ctx context.Context, log *entity.TrackingLog) error {
	m := &trackingLogModel{
		TrackingNumber:       log.TrackingNumber,
		SourceTrackingNumber: log.SourceTrackingNumber,
		ChannelAlias:         log.ChannelAlias,
		ShippingAgent:        log.ShippingAgent,
		ShippingChannel:      log.ShippingChannel,
		CountryCode:          log.CountryCode,
		TrackStatus:          log.TrackStatus,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create tracking log: %w", err)
	}
	log.ID = m.ID
	return nil
}

func (r *TrackingLogRepoImpl) UpdateTrackingFields(ctx context.Context, id int64, trackStatus int, syncedAt time.Time, receivedAt, deliveredAt, trackedAt *time.Time) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE tracking_logs
		SET track_status = ?, synced_at = ?, received_at = ?, delivered_at = ?, tracked_at = ?
		WHERE id = ?
	`, trackStatus, syncedAt, receivedAt, deliveredAt, trackedAt, id).Error
}

func (r *TrackingLogRepoImpl) UpdateTrackingFieldsBySourceNumber(ctx context.Context, sourceTrackingNumber string, trackStatus int, syncedAt time.Time, receivedAt, deliveredAt, trackedAt *time.Time) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE tracking_logs
		SET track_status = ?, synced_at = ?, received_at = ?, delivered_at = ?, tracked_at = ?
		WHERE source_tracking_number = ?
	`, trackStatus, syncedAt, receivedAt, deliveredAt, trackedAt, sourceTrackingNumber).Error
}

func (r *TrackingLogRepoImpl) FindSyncCandidates(ctx context.Context, limit int, cacheHours int) ([]*entity.TrackingLog, error) {
	var models []trackingLogModel
	err := r.db.WithContext(ctx).Raw(`
		SELECT tl.* FROM tracking_logs tl
		LEFT JOIN tracking_details td ON td.tracking_number = tl.source_tracking_number
		WHERE tl.source_tracking_number IS NOT NULL AND tl.source_tracking_number != ''
		  AND tl.fulfill_at IS NOT NULL
		  AND tl.track_status != ?
		  AND (td.id IS NULL OR td.synced_at < NOW() - INTERVAL '1 hour' * ?)
		ORDER BY tl.updated_at ASC
		LIMIT ?
	`, entity.TrackStatusDelivered, cacheHours, limit).Scan(&models).Error
	if err != nil {
		return nil, err
	}
	return toTrackingLogEntities(models), nil
}

func toTrackingLogEntities(models []trackingLogModel) []*entity.TrackingLog {
	result := make([]*entity.TrackingLog, len(models))
	for i := range models {
		result[i] = toTrackingLogEntity(&models[i])
	}
	return result
}

func toTrackingLogEntity(m *trackingLogModel) *entity.TrackingLog {
	return &entity.TrackingLog{
		ID: m.ID, TrackingNumber: m.TrackingNumber, SourceTrackingNumber: m.SourceTrackingNumber,
		ChannelAlias: m.ChannelAlias, ShippingAgent: m.ShippingAgent, ShippingChannel: m.ShippingChannel,
		CountryCode: m.CountryCode, TrackStatus: m.TrackStatus, SyncedAt: m.SyncedAt,
		ReceivedAt: m.ReceivedAt, DeliveredAt: m.DeliveredAt, TrackedAt: m.TrackedAt, FulfillAt: m.FulfillAt,
	}
}
