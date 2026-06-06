package main

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TrackingLog struct {
	ID                   int64  `gorm:"column:id"`
	SourceTrackingNumber string `gorm:"column:source_tracking_number"`
	TrackStatus          int    `gorm:"column:track_status"`
	SyncedAt             *time.Time `gorm:"column:synced_at"`
	ReceivedAt           *time.Time `gorm:"column:received_at"`
	DeliveredAt          *time.Time `gorm:"column:delivered_at"`
	TrackedAt            *time.Time `gorm:"column:tracked_at"`
}

func (TrackingLog) TableName() string { return "tracking_logs" }

func main() {
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	waybill := "YT2612500702088888"

	// 模拟 syncTrackingLog 逻辑
	fmt.Println("=== Testing syncTrackingLog ===")

	// 1. 查找 tracking_log
	var log TrackingLog
	result := db.Where("source_tracking_number = ?", waybill).Order("id DESC").First(&log)
	if result.Error != nil {
		fmt.Printf("Find failed: %v\n", result.Error)
		return
	}
	fmt.Printf("Found tracking_log: id=%d\n", log.ID)

	// 2. 更新 tracking_log
	now := time.Now()
	receivedAt := now.Add(-24 * time.Hour) // 模拟 received_at
	deliveredAt := now                      // 模拟 delivered_at
	trackedAt := now                        // 模拟 tracked_at

	updateResult := db.Exec(`
		UPDATE tracking_logs
		SET track_status = ?, synced_at = ?, received_at = ?, delivered_at = ?, tracked_at = ?
		WHERE source_tracking_number = ?
	`, 2, now, &receivedAt, &deliveredAt, &trackedAt, waybill)

	if updateResult.Error != nil {
		fmt.Printf("Update failed: %v\n", updateResult.Error)
		return
	}
	fmt.Printf("Updated %d rows\n", updateResult.RowsAffected)

	// 3. 验证更新
	var updatedLog TrackingLog
	db.Where("source_tracking_number = ?", waybill).First(&updatedLog)
	
	syncedAtStr := ""
	if updatedLog.SyncedAt != nil {
		syncedAtStr = updatedLog.SyncedAt.Format(time.RFC3339)
	}
	receivedAtStr := ""
	if updatedLog.ReceivedAt != nil {
		receivedAtStr = updatedLog.ReceivedAt.Format(time.RFC3339)
	}
	deliveredAtStr := ""
	if updatedLog.DeliveredAt != nil {
		deliveredAtStr = updatedLog.DeliveredAt.Format(time.RFC3339)
	}
	trackedAtStr := ""
	if updatedLog.TrackedAt != nil {
		trackedAtStr = updatedLog.TrackedAt.Format(time.RFC3339)
	}

	fmt.Printf("\nAfter update:\n")
	fmt.Printf("track_status: %d\n", updatedLog.TrackStatus)
	fmt.Printf("synced_at: %s\n", syncedAtStr)
	fmt.Printf("received_at: %s\n", receivedAtStr)
	fmt.Printf("delivered_at: %s\n", deliveredAtStr)
	fmt.Printf("tracked_at: %s\n", trackedAtStr)
}
