package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	Status        string `gorm:"column:status"`
}

func (RawEvent) TableName() string { return "raw_events" }

func main() {
	// 检查 Redis 队列
	rds := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 0})
	length, _ := rds.LLen(context.Background(), "queue:yun_express_webhook_track").Result()
	fmt.Printf("Redis queue length: %d\n\n", length)

	// 检查数据库
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// 检查 tracking_details
	fmt.Println("=== tracking_details (YT2612500702088888) ===")
	var detail TrackingDetail
	db.Where("tracking_number = ?", "YT2612500702088888").First(&detail)
	fmt.Printf("id=%d, status=%d, synced_at=%s\n", detail.ID, detail.Status, detail.SyncedAt)

	// 检查 raw_events
	fmt.Println("\n=== raw_events (YT2612500702088888) ===")
	var events []RawEvent
	db.Where("waybill_number = ?", "YT2612500702088888").Order("id DESC").Find(&events)
	if len(events) == 0 {
		fmt.Println("No raw_events found")
	} else {
		for _, e := range events {
			fmt.Printf("id=%d, node=%s, status=%s\n", e.ID, e.TrackNodeCode, e.Status)
		}
	}
}
