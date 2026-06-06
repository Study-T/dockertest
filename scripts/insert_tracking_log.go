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
	ChannelAlias         string `gorm:"column:channel_alias"`
	ShippingAgent        string `gorm:"column:shipping_agent"`
	ShippingChannel      string `gorm:"column:shipping_channel"`
	CountryCode          string `gorm:"column:country_code"`
	TrackStatus          int    `gorm:"column:track_status"`
}

func (TrackingLog) TableName() string { return "tracking_logs" }

func main() {
	dsn := "postgres://postgres:123456@127.0.0.1:5432/ns_tracking?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Connect failed: %v\n", err)
		return
	}

	// 插入测试数据
	logs := []TrackingLog{
		{
			TrackingNumber:       "00340434525230659090",
			SourceTrackingNumber: "YT2612500701966827",
			ChannelAlias:         "云途全球专线挂号（特惠带电）",
			ShippingAgent:        "YunExpress",
			ShippingChannel:      "THZXR",
			CountryCode:          "DE",
			TrackStatus:          0,
		},
		{
			TrackingNumber:       "00340434525230666666",
			SourceTrackingNumber: "YT2612500702088888",
			ChannelAlias:         "云途全球专线挂号（特惠带电）",
			ShippingAgent:        "YunExpress",
			ShippingChannel:      "THZXR",
			CountryCode:          "US",
			TrackStatus:          0,
		},
	}

	for _, log := range logs {
		result := db.Create(&log)
		if result.Error != nil {
			fmt.Printf("Insert failed for %s: %v\n", log.SourceTrackingNumber, result.Error)
		} else {
			fmt.Printf("Inserted tracking_log: %s (id=%d)\n", log.SourceTrackingNumber, log.ID)
		}
	}
}
