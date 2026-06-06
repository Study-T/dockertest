package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Connect failed: %v\n", err)
		return
	}

	// 检查 tracking_logs 表的记录数
	var count int64
	db.Raw("SELECT COUNT(*) FROM tracking_logs").Scan(&count)
	fmt.Printf("tracking_logs records: %d\n", count)

	// 检查 tracking_details 表的记录数
	var detailCount int64
	db.Raw("SELECT COUNT(*) FROM tracking_details").Scan(&detailCount)
	fmt.Printf("tracking_details records: %d\n", detailCount)

	// 检查 raw_events 表的记录数
	var rawCount int64
	db.Raw("SELECT COUNT(*) FROM raw_events").Scan(&rawCount)
	fmt.Printf("raw_events records: %d\n", rawCount)

	// 如果 tracking_logs 为空，检查表是否存在
	if count == 0 {
		fmt.Println("\n=== tracking_logs is EMPTY ===")
		fmt.Println("Possible reasons:")
		fmt.Println("1. Table was truncated")
		fmt.Println("2. Records were deleted")
		fmt.Println("3. Different database connection")
	}
}
