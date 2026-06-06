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
}

func (TrackingLog) TableName() string { return "tracking_logs" }

func main() {
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	fmt.Println("=== All tracking_logs ===")
	var logs []TrackingLog
	db.Find(&logs)

	for _, l := range logs {
		fmt.Printf("id=%d, tracking=%s, source=%s\n", l.ID, l.TrackingNumber, l.SourceTrackingNumber)
	}
}
