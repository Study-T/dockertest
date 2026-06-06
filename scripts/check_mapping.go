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
	// 检查 Redis 缓存
	rds := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 0})
	val, err := rds.Get(context.Background(), "tracking:detail:YT2612500702088888").Result()
	if err != nil {
		fmt.Printf("Redis cache not found: %v\n", err)
	} else {
		fmt.Printf("=== Redis Cache ===\n%s\n\n", val)
	}

	// 检查数据库
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	var detail TrackingDetail
	db.Where("tracking_number = ?", "YT2612500702088888").First(&detail)

	fmt.Printf("=== Database ===\n")
	fmt.Printf("status: %d\n", detail.Status)
	fmt.Printf("synced_at: %s\n\n", detail.SyncedAt)

	// 状态映射说明
	fmt.Println("=== Status Mapping ===")
	fmt.Println("status 10 → package_status 'N' (InfoReceived)")
	fmt.Println("status 20 → package_status 'T' (InTransit)")
	fmt.Println("status 30 → package_status 'T' (AvailableForPickup)")
	fmt.Println("status 40 → package_status 'T' (DeliveryAttempt)")
	fmt.Println("status 50 → package_status 'D' (Delivered)")
	fmt.Println("status 60 → package_status 'E' (Exception)")
	fmt.Println("status 90 → package_status 'R' (Returned)")
}
