package repo_impl

import (
	"context"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"
	"ns-tracking-go/infrastructure/database"
)

type rawEventModel struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	IdempotencyKey string     `gorm:"column:idempotency_key"`
	ProviderCode   string     `gorm:"column:provider_code"`
	DataCode       string     `gorm:"column:data_code"`
	WaybillNumber  string     `gorm:"column:waybill_number"`
	TrackingNumber string     `gorm:"column:tracking_number"`
	CustomerCode   string     `gorm:"column:customer_code"`
	TrackNodeCode  string     `gorm:"column:track_node_code"`
	ProcessTime    string     `gorm:"column:process_time"`
	Payload        string     `gorm:"column:payload"`
	EnvelopeMeta   string     `gorm:"column:envelope_meta"`
	Status         string     `gorm:"column:status"`
	RetryCount     int        `gorm:"column:retry_count"`
	MaxRetries     int        `gorm:"column:max_retries"`
	LastError      string     `gorm:"column:last_error"`
	ProcessedAt    *time.Time `gorm:"column:processed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (rawEventModel) TableName() string { return "raw_events" }

type RawEventRepoImpl struct {
	db *database.DB
}

func NewRawEventRepoImpl(db *database.DB) *RawEventRepoImpl {
	return &RawEventRepoImpl{db: db}
}

var _ repo.RawEventRepo = (*RawEventRepoImpl)(nil)

func (r *RawEventRepoImpl) Save(ctx context.Context, event *entity.RawEvent) (int64, error) {
	m := toModel(event)
	tx := r.db.WithContext(ctx).Exec(`
		INSERT INTO raw_events (idempotency_key, provider_code, data_code, waybill_number,
			tracking_number, customer_code, track_node_code, process_time,
			payload, envelope_meta, status, retry_count, max_retries,
			last_error, processed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, m.IdempotencyKey, m.ProviderCode, m.DataCode, m.WaybillNumber,
		m.TrackingNumber, m.CustomerCode, m.TrackNodeCode, m.ProcessTime,
		m.Payload, m.EnvelopeMeta, m.Status, m.RetryCount, m.MaxRetries,
		m.LastError, m.ProcessedAt, m.CreatedAt, m.UpdatedAt)
	if tx.Error != nil {
		return 0, fmt.Errorf("save raw event: %w", tx.Error)
	}
	return tx.RowsAffected, nil
}

func (r *RawEventRepoImpl) FindByIdempotencyKey(ctx context.Context, key string) (*entity.RawEvent, error) {
	var m rawEventModel
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("find by idempotency key: %w", err)
	}
	return toEntity(&m), nil
}

func (r *RawEventRepoImpl) UpdateStatus(ctx context.Context, id int64, status, errMsg string) error {
	return r.db.WithContext(ctx).Table("raw_events").Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status, "last_error": errMsg, "updated_at": time.Now(),
		}).Error
}

func (r *RawEventRepoImpl) IncrementRetry(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE raw_events SET retry_count = retry_count + 1, updated_at = NOW() WHERE id = ?
	`, id).Error
}

func (r *RawEventRepoImpl) MarkDeadLettered(ctx context.Context, id int64, errMsg string) error {
	return r.db.WithContext(ctx).Table("raw_events").Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": entity.RawEventDeadLettered, "last_error": errMsg, "updated_at": time.Now(),
		}).Error
}

func (r *RawEventRepoImpl) FindPending(ctx context.Context, limit int) ([]*entity.RawEvent, error) {
	return r.findByStatus(ctx, entity.RawEventPending, limit)
}

func (r *RawEventRepoImpl) FindRetryable(ctx context.Context, limit int) ([]*entity.RawEvent, error) {
	var models []rawEventModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND retry_count < max_retries", entity.RawEventFailed).
		Order("updated_at ASC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("find retryable: %w", err)
	}
	return toEntities(models), nil
}

func (r *RawEventRepoImpl) CountByStatus(ctx context.Context) (map[string]int64, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("raw_events").
		Select("status, count(*) as count").Group("status").Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("count by status: %w", err)
	}
	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.Status] = r.Count
	}
	return result, nil
}

func (r *RawEventRepoImpl) findByStatus(ctx context.Context, status string, limit int) ([]*entity.RawEvent, error) {
	var models []rawEventModel
	err := r.db.WithContext(ctx).Where("status = ?", status).
		Order("created_at ASC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("find %s: %w", status, err)
	}
	return toEntities(models), nil
}

func toEntities(models []rawEventModel) []*entity.RawEvent {
	result := make([]*entity.RawEvent, len(models))
	for i := range models {
		result[i] = toEntity(&models[i])
	}
	return result
}

func toModel(e *entity.RawEvent) *rawEventModel {
	return &rawEventModel{
		ID: e.ID, IdempotencyKey: e.IdempotencyKey, ProviderCode: e.ProviderCode,
		DataCode: e.DataCode, WaybillNumber: e.WaybillNumber, TrackingNumber: e.TrackingNumber,
		CustomerCode: e.CustomerCode, TrackNodeCode: e.TrackNodeCode, ProcessTime: e.ProcessTime,
		Payload: e.Payload, EnvelopeMeta: e.EnvelopeMeta, Status: e.Status,
		RetryCount: e.RetryCount, MaxRetries: e.MaxRetries, LastError: e.LastError,
		ProcessedAt: e.ProcessedAt, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func toEntity(m *rawEventModel) *entity.RawEvent {
	return &entity.RawEvent{
		ID: m.ID, IdempotencyKey: m.IdempotencyKey, ProviderCode: m.ProviderCode,
		DataCode: m.DataCode, WaybillNumber: m.WaybillNumber, TrackingNumber: m.TrackingNumber,
		CustomerCode: m.CustomerCode, TrackNodeCode: m.TrackNodeCode, ProcessTime: m.ProcessTime,
		Payload: m.Payload, EnvelopeMeta: m.EnvelopeMeta, Status: m.Status,
		RetryCount: m.RetryCount, MaxRetries: m.MaxRetries, LastError: m.LastError,
		ProcessedAt: m.ProcessedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}
