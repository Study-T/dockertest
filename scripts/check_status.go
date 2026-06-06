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

func main() {
	// 检查 Redis 队列
	rds := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 0})
	length, _ := rds.LLen(context.Background(), "queue:yun_express_webhook_track").Result()
	fmt.Printf("Redis queue length: %d\n\n", length)

	// 检查数据库
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("DB connect failed: %v\n", err)
		return
	}

	var details []TrackingDetail
	db.Order("id DESC").Limit(10).Find(&details)

	fmt.Println("=== tracking_details ===")
	for _, d := range details {
		fmt.Printf("tracking_number: %s, status: %d, synced_at: %s\n", d.TrackingNumber, d.Status, d.SyncedAt)
	}

	// 检查 YT2612500701966827
	var count int64
	db.Model(&TrackingDetail{}).Where("tracking_number = ?", "YT2612500701966827").Count(&count)
	fmt.Printf("\nYT2612500701966827 count: %d\n", count)
}
