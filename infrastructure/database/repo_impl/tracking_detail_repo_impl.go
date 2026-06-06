package repo_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"
	"ns-tracking-go/infrastructure/database"
)

type trackingDetailModel struct {
	ID              int64      `gorm:"column:id;primaryKey"`
	TrackingNumber  string     `gorm:"column:tracking_number"`
	ServiceClass    string     `gorm:"column:service_class"`
	Status          int        `gorm:"column:status"`
	SyncedAt        time.Time  `gorm:"column:synced_at"`
	ErrorMessage    string     `gorm:"column:error_message"`
	TrackingLogID   int64      `gorm:"column:tracking_log_id"`
	AutoDeliveredAt *time.Time `gorm:"column:auto_delivered_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (trackingDetailModel) TableName() string { return "tracking_details" }

type TrackingDetailRepoImpl struct {
	db *database.DB
}

func NewTrackingDetailRepoImpl(db *database.DB) *TrackingDetailRepoImpl {
	return &TrackingDetailRepoImpl{db: db}
}

var _ repo.TrackingDetailRepo = (*TrackingDetailRepoImpl)(nil)

func (r *TrackingDetailRepoImpl) Save(ctx context.Context, detail *entity.TrackingDetail) error {
	now := time.Now()
	detailJSON := toJSON(detail.Detail)
	lastDetailJSON := toJSON(detail.LastDetail)

	return r.db.WithContext(ctx).Exec(`
		INSERT INTO tracking_details (tracking_number, service_class, status, detail, last_detail,
			synced_at, error_message, tracking_log_id, created_at, updated_at)
		VALUES (?, ?, ?, ?::jsonb, ?::jsonb, ?, ?, ?, ?, ?)
		ON CONFLICT (tracking_number) DO UPDATE SET
			last_detail = tracking_details.detail,
			detail = EXCLUDED.detail,
			status = EXCLUDED.status,
			service_class = EXCLUDED.service_class,
			synced_at = EXCLUDED.synced_at,
			error_message = EXCLUDED.error_message,
			tracking_log_id = COALESCE(EXCLUDED.tracking_log_id, tracking_details.tracking_log_id),
			updated_at = EXCLUDED.updated_at
		WHERE tracking_details.auto_delivered_at IS NULL
	`, detail.TrackingNumber, detail.ServiceClass, detail.Status,
		detailJSON, lastDetailJSON, detail.SyncedAt,
		detail.ErrorMessage, detail.TrackingLogID, now, now).Error
}

func (r *TrackingDetailRepoImpl) Update(ctx context.Context, detail *entity.TrackingDetail) error {
	detailJSON := toJSON(detail.Detail)
	lastDetailJSON := toJSON(detail.LastDetail)

	return r.db.WithContext(ctx).Exec(`
		UPDATE tracking_details
		SET detail = ?::jsonb, last_detail = ?::jsonb, status = ?, synced_at = ?,
			error_message = ?, updated_at = NOW()
		WHERE tracking_number = ?
		  AND auto_delivered_at IS NULL
		  AND synced_at < ?
	`, detailJSON, lastDetailJSON, detail.Status,
		detail.SyncedAt, detail.ErrorMessage,
		detail.TrackingNumber, detail.SyncedAt).Error
}

func (r *TrackingDetailRepoImpl) FindByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingDetail, error) {
	var m trackingDetailModel
	err := r.db.WithContext(ctx).
		Where("tracking_number = ?", trackingNumber).
		Order("id DESC").First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("find tracking detail: %w", err)
	}

	// 查询 detail JSONB 字段（PostgreSQL 语法）
	var detailJSON, lastDetailJSON string
	err = r.db.WithContext(ctx).
		Table("tracking_details").
		Select("detail::text, last_detail::text").
		Where("tracking_number = ?", trackingNumber).
		Order("id DESC").
		Row().Scan(&detailJSON, &lastDetailJSON)
	if err != nil {
		return nil, fmt.Errorf("find tracking detail json: %w", err)
	}

	detail := parseJSON(detailJSON)
	lastDetail := parseJSON(lastDetailJSON)

	return &entity.TrackingDetail{
		ID: m.ID, TrackingNumber: m.TrackingNumber, ServiceClass: m.ServiceClass,
		Status: m.Status, SyncedAt: m.SyncedAt, ErrorMessage: m.ErrorMessage,
		TrackingLogID: m.TrackingLogID, AutoDeliveredAt: m.AutoDeliveredAt,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		Detail: detail, LastDetail: lastDetail,
	}, nil
}

func parseJSON(jsonStr string) map[string]interface{} {
	if jsonStr == "" {
		return nil
	}
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}

func toJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
