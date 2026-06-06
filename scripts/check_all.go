package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TrackingLog struct {
	ID                   int64  `gorm:"column:id"`
	SourceTrackingNumber string `gorm:"column:source_tracking_number"`
	TrackStatus          int    `gorm:"column:track_status"`
	SyncedAt             string `gorm:"column:synced_at"`
	ReceivedAt           string `gorm:"column:received_at"`
	DeliveredAt          string `gorm:"column:delivered_at"`
}

func (TrackingLog) TableName() string { return "tracking_logs" }

type RawEvent struct {
	ID             int64  `gorm:"column:id"`
	WaybillNumber  string `gorm:"column:waybill_number"`
	TrackNodeCode  string `gorm:"column:track_node_code"`
	Status         string `gorm:"column:status"`
}

func (RawEvent) TableName() string { return "raw_events" }

type TrackingDetail struct {
	ID             int64  `gorm:"column:id"`
	TrackingNumber string `gorm:"column:tracking_number"`
	Status         int    `gorm:"column:status"`
	SyncedAt       string `gorm:"column:synced_at"`
}

func (TrackingDetail) TableName() string { return "tracking_details" }

func main() {
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// Check Redis cache
	rds := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 0})
	length, _ := rds.LLen(context.Background(), "queue:yun_express_webhook_track").Result()
	fmt.Printf("=== Redis Queue ===\nLength: %d\n\n", length)

	// Check tracking_logs
	fmt.Println("=== tracking_logs ===")
	var logs []TrackingLog
	db.Where("source_tracking_number = ?", "YT2612500701966827").Find(&logs)
	for _, l := range logs {
		fmt.Printf("id=%d, source=%s, status=%d, synced_at=%s, received_at=%s, delivered_at=%s\n",
			l.ID, l.SourceTrackingNumber, l.TrackStatus, l.SyncedAt, l.ReceivedAt, l.DeliveredAt)
	}

	// Check raw_events
	fmt.Println("\n=== raw_events ===")
	var events []RawEvent
	db.Where("waybill_number = ?", "YT2612500701966827").Order("id DESC").Find(&events)
	for _, e := range events {
		fmt.Printf("id=%d, waybill=%s, node=%s, status=%s\n", e.ID, e.WaybillNumber, e.TrackNodeCode, e.Status)
	}

	// Check tracking_details
	fmt.Println("\n=== tracking_details ===")
	var details []TrackingDetail
	db.Where("tracking_number = ?", "YT2612500701966827").Find(&details)
	for _, d := range details {
		fmt.Printf("id=%d, tracking=%s, status=%d, synced_at=%s\n", d.ID, d.TrackingNumber, d.Status, d.SyncedAt)
	}
}
