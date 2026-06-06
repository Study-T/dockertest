package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TrackingLog struct {
	ID                   int64  `gorm:"column:id"`
	TrackingNumber       string `gorm:"column:tracking_number"`
	SourceTrackingNumber string `gorm:"column:source_tracking_number"`
	TrackStatus          int    `gorm:"column:track_status"`
	SyncedAt             string `gorm:"column:synced_at"`
	ReceivedAt           string `gorm:"column:received_at"`
	DeliveredAt          string `gorm:"column:delivered_at"`
	TrackedAt            string `gorm:"column:tracked_at"`
}

func (TrackingLog) TableName() string { return "tracking_logs" }

func main() {
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	fmt.Println("=== tracking_logs ===")
	var logs []TrackingLog
	db.Where("source_tracking_number IN ?", []string{"YT2612500701966827", "YT2612500702088888"}).Find(&logs)

	for _, l := range logs {
		fmt.Printf("\nid=%d\n", l.ID)
		fmt.Printf("tracking_number: %s\n", l.TrackingNumber)
		fmt.Printf("source_tracking_number: %s\n", l.SourceTrackingNumber)
		fmt.Printf("track_status: %d\n", l.TrackStatus)
		fmt.Printf("synced_at: %s\n", l.SyncedAt)
		fmt.Printf("received_at: %s\n", l.ReceivedAt)
		fmt.Printf("delivered_at: %s\n", l.DeliveredAt)
		fmt.Printf("tracked_at: %s\n", l.TrackedAt)
	}
}
