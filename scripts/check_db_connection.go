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

	// 检查所有表
	fmt.Println("=== Tables in database ===")
	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables)
	for _, t := range tables {
		fmt.Printf("- %s\n", t)
	}

	// 检查 tracking_logs 表结构
	fmt.Println("\n=== tracking_logs columns ===")
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}
	db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'tracking_logs'").Scan(&columns)
	for _, c := range columns {
		fmt.Printf("- %s: %s\n", c.ColumnName, c.DataType)
	}
}
