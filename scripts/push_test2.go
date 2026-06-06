package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	defer client.Close()

	testData := `{
  "data_code": "tisPushData",
  "data": {
    "waybill_number": "YT2612500702088888",
    "tracking_number": "00340434525230666666",
    "customer_order_number": "BG260505090437250000",
    "package_status": "T",
    "product_code": "THZXR",
    "product_name": "云途全球专线挂号（特惠带电）",
    "channel_code": "DEPostFTP",
    "check_in_time": "2026-06-01T10:00:00Z",
    "check_out_time": "2026-06-01T12:00:00Z",
    "pick_up_time": "0001-01-01T00:00:00Z",
    "customer_code": "CNHC318962",
    "origin_code": "CN",
    "destination_code": "US",
    "postal_code": "",
    "actual_weight": 0.5,
    "interval_day": 3,
    "interval_work_day": 2,
    "last_mile_site": "https://www.usps.com/",
    "last_mile_name": "USPS",
    "phone_number": "",
    "pod_url": "",
    "IsSignature": false,
    "EstimatedDeliveryToDateZone": "",
    "EstimatedDeliveryFromDateZone": "",
    "track_events": [
      {
        "process_time": "2026-06-01T10:00:00Z",
        "process_utc_time": "2026-06-01T10:00:00Z",
        "process_content": "已收到发货信息",
        "process_country": "",
        "process_province": "",
        "process_city": "",
        "process_location": "",
        "track_node_code": "ORDER_CREATION",
        "track_node_description": "已收到货物信息",
        "pod_url": ""
      },
      {
        "process_time": "2026-06-01T15:00:00Z",
        "process_utc_time": "2026-06-01T15:00:00Z",
        "process_content": "已到达始发地设施",
        "process_country": "CN",
        "process_province": "广东",
        "process_city": "深圳",
        "process_location": "深圳集货中心",
        "track_node_code": "FIRST_MILE_ARRIVE",
        "track_node_description": "已抵达始发地设施",
        "pod_url": ""
      },
      {
        "process_time": "2026-06-03T08:00:00Z",
        "process_utc_time": "2026-06-03T08:00:00Z",
        "process_content": "运输中",
        "process_country": "US",
        "process_province": "",
        "process_city": "洛杉矶",
        "process_location": "洛杉矶转运中心",
        "track_node_code": "IN_TRANSIT",
        "track_node_description": "运输中",
        "pod_url": ""
      }
    ]
  }
}`

	err := client.LPush(context.Background(), "queue:yun_express_webhook_track", testData).Err()
	if err != nil {
		fmt.Printf("Push failed: %v\n", err)
		return
	}

	length, _ := client.LLen(context.Background(), "queue:yun_express_webhook_track").Result()
	fmt.Printf("Push success! Queue length: %d\n", length)
}
