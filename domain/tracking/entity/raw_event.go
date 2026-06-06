package entity

import "time"

// RawEvent represents a single track event from a webhook push.
// One tisPushData may produce multiple RawEvents (one per track_event).
type RawEvent struct {
	ID             int64      `json:"id"`
	IdempotencyKey string     `json:"idempotency_key"`
	ProviderCode   string     `json:"provider_code"`
	DataCode       string     `json:"data_code"`
	WaybillNumber  string     `json:"waybill_number"`
	TrackingNumber string     `json:"tracking_number"`
	CustomerCode   string     `json:"customer_code"`
	TrackNodeCode  string     `json:"track_node_code"`
	ProcessTime    string     `json:"process_time"`
	Payload        string     `json:"payload"`        // single track_event JSON
	EnvelopeMeta   string     `json:"envelope_meta"`  // top-level fields JSON
	Status         string     `json:"status"`
	RetryCount     int        `json:"retry_count"`
	MaxRetries     int        `json:"max_retries"`
	LastError      string     `json:"last_error"`
	ProcessedAt    *time.Time `json:"processed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

const (
	RawEventPending      = "pending"
	RawEventProcessed    = "processed"
	RawEventFailed       = "failed"
	RawEventDeadLettered = "dead_lettered"
	DefaultMaxRetries    = 5
)
