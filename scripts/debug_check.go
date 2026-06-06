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
	TrackingNumber       string `gorm:"column:tracking_number"`
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

func main() {
	rds := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 0})
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	waybill := "YT2612500702088888"

	// 1. 检查 Redis 队列
	length, _ := rds.LLen(context.Background(), "queue:yun_express_webhook_track").Result()
	fmt.Printf("1. Redis queue length: %d\n\n", length)

	// 2. 检查 tracking_logs 表是否存在
	fmt.Println("2. Checking tracking_logs table...")
	var count int64
	db.Model(&TrackingLog{}).Count(&count)
	fmt.Printf("   Total tracking_logs records: %d\n", count)

	// 3. 查找特定运单号
	fmt.Printf("\n3. Looking for %s in tracking_logs...\n", waybill)
	var log TrackingLog
	result := db.Where("source_tracking_number = ?", waybill).First(&log)
	if result.Error != nil {
		fmt.Printf("   Not found: %v\n", result.Error)
	} else {
		fmt.Printf("   Found: id=%d, track_status=%d, synced_at=%s\n", log.ID, log.TrackStatus, log.SyncedAt)
		fmt.Printf("   received_at=%s, delivered_at=%s, tracked_at=%s\n", log.ReceivedAt, log.DeliveredAt, log.TrackedAt)
	}

	// 4. 检查 tracking_details
	fmt.Printf("\n4. Checking tracking_details for %s...\n", waybill)
	var detail TrackingDetail
	db.Where("tracking_number = ?", waybill).First(&detail)
	fmt.Printf("   id=%d, status=%d, synced_at=%s\n", detail.ID, detail.Status, detail.SyncedAt)
}
