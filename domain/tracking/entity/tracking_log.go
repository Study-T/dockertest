package entity

import "time"

// TrackingLog maps to the tracking_logs table (read-only for Go).
// Go queries this for channel context and updates sync fields.
type TrackingLog struct {
	ID                   int64      `json:"id"`
	TrackingNumber       string     `json:"tracking_number"`
	SourceTrackingNumber string     `json:"source_tracking_number"`
	ChannelAlias         string     `json:"channel_alias"`
	ShippingAgent        string     `json:"shipping_agent"`
	ShippingChannel      string     `json:"shipping_channel"`
	CountryCode          string     `json:"country_code"`
	TrackStatus          int        `json:"track_status"`
	SyncedAt             *time.Time `json:"synced_at"`
	ReceivedAt           *time.Time `json:"received_at"`
	DeliveredAt          *time.Time `json:"delivered_at"`
	TrackedAt            *time.Time `json:"tracked_at"`
	FulfillAt            *time.Time `json:"fulfill_at"`
}
