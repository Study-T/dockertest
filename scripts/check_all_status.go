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
	TrackedAt            string `gorm:"column:tracked_at"`
}

func (TrackingLog) TableName() string { return "tracking_logs" }

type TrackingDetail struct {
	ID             int64  `gorm:"column:id"`
	TrackingNumber string `gorm:"column:tracking_number"`
	Status         int    `gorm:"column:status"`
	SyncedAt       string `gorm:"column:synced_at"`
}

func (TrackingDetail) TableName() string { return "tracking_details" }

type RawEvent struct {
	ID            int64  `gorm:"column:id"`
	WaybillNumber string `gorm:"column:waybill_number"`
	TrackNodeCode string `gorm:"column:track_node_code"`
}

func (RawEvent) TableName() string { return "raw_events" }

func main() {
	rds := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 0})
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	waybill := "YT2612500702088888"

	// Redis 队列
	length, _ := rds.LLen(context.Background(), "queue:yun_express_webhook_track").Result()
	fmt.Printf("Redis queue length: %d\n\n", length)

	// tracking_details
	fmt.Println("=== tracking_details ===")
	var detail TrackingDetail
	db.Where("tracking_number = ?", waybill).First(&detail)
	fmt.Printf("id=%d, status=%d, synced_at=%s\n", detail.ID, detail.Status, detail.SyncedAt)

	// tracking_logs
	fmt.Println("\n=== tracking_logs ===")
	var log TrackingLog
	result := db.Where("source_tracking_number = ?", waybill).First(&log)
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	} else {
		fmt.Printf("id=%d\n", log.ID)
		fmt.Printf("track_status: %d\n", log.TrackStatus)
		fmt.Printf("synced_at: %s\n", log.SyncedAt)
		fmt.Printf("received_at: %s\n", log.ReceivedAt)
		fmt.Printf("delivered_at: %s\n", log.DeliveredAt)
		fmt.Printf("tracked_at: %s\n", log.TrackedAt)
	}

	// raw_events
	fmt.Println("\n=== raw_events ===")
	var rawCount int64
	db.Model(&RawEvent{}).Where("waybill_number = ?", waybill).Count(&rawCount)
	fmt.Printf("Count: %d\n", rawCount)
}
