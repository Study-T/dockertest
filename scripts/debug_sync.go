package main

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TrackingLog struct {
	ID                   int64  `gorm:"column:id"`
	SourceTrackingNumber string `gorm:"column:source_tracking_number"`
}

func (TrackingLog) TableName() string { return "tracking_logs" }

func main() {
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	waybill := "YT2612500702088888"

	// 测试查询
	var log TrackingLog
	result := db.WithContext(context.Background()).
		Where("source_tracking_number = ?", waybill).
		Order("id DESC").First(&log)

	if result.Error != nil {
		fmt.Printf("Query failed: %v\n", result.Error)
	} else {
		fmt.Printf("Found: id=%d, source=%s\n", log.ID, log.SourceTrackingNumber)
	}
}
