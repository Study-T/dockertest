package main

import (
	"fmt"

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
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Connect failed: %v\n", err)
		return
	}

	var details []TrackingDetail
	db.Order("id DESC").Limit(10).Find(&details)

	fmt.Println("=== tracking_details ===")
	for _, d := range details {
		fmt.Printf("tracking_number: %s, status: %d, synced_at: %s\n", d.TrackingNumber, d.Status, d.SyncedAt)
	}
}
