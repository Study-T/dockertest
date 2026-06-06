package entity

import "time"

// TrackingDetail maps to the tracking_details table.
// Go writes this; Ruby reads it via fetch_cache.
type TrackingDetail struct {
	ID              int64                  `json:"id"`
	TrackingNumber  string                 `json:"tracking_number"`
	ServiceClass    string                 `json:"service_class"`
	Status          int                    `json:"status"`
	Detail          map[string]interface{} `json:"detail"`      // JSONB: {response: {Item: {...}}, synced_at: "..."}
	LastDetail      map[string]interface{} `json:"last_detail"` // JSONB: deep copy backup
	SyncedAt        time.Time              `json:"synced_at"`
	ErrorMessage    string                 `json:"error_message"`
	TrackingLogID   int64                  `json:"tracking_log_id"`
	AutoDeliveredAt *time.Time             `json:"auto_delivered_at"` // Go must NOT write this
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// TrackStatus枚举 (tracking_logs.track_status)
const (
	TrackStatusToReceive     = 0
	TrackStatusInTransit     = 1
	TrackStatusDelivered     = 2
	TrackStatusTrackException = 3
)
