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
    "waybill_number": "YT2612500701966827",
    "tracking_number": "00340434525230659090",
    "customer_order_number": "BG260505090437249440",
    "package_status": "D",
    "product_code": "THZXR",
    "product_name": "云途全球专线挂号（特惠带电）",
    "channel_code": "DEPostFTP",
    "check_in_time": "2026-05-09T20:32:09Z",
    "check_out_time": "2026-05-09T22:29:46Z",
    "pick_up_time": "0001-01-01T00:00:00Z",
    "customer_code": "CNHC318962",
    "origin_code": "CN",
    "destination_code": "DE",
    "postal_code": "",
    "actual_weight": 0.114,
    "interval_day": 8.6,
    "interval_work_day": 5.5,
    "last_mile_site": "https://www.deutschepost.de/",
    "last_mile_name": "DeutschePost",
    "phone_number": "",
    "pod_url": "",
    "IsSignature": false,
    "EstimatedDeliveryToDateZone": "",
    "EstimatedDeliveryFromDateZone": "",
    "track_events": [
      {
        "process_time": "2026-05-09T20:32:09Z",
        "process_utc_time": "2026-05-09T20:32:09Z",
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
        "process_time": "2026-05-10T08:15:00Z",
        "process_utc_time": "2026-05-10T08:15:00Z",
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
        "process_time": "2026-05-12T14:00:00Z",
        "process_utc_time": "2026-05-12T14:00:00Z",
        "process_content": "已抵达目的国",
        "process_country": "DE",
        "process_province": "",
        "process_city": "法兰克福",
        "process_location": "法兰克福转运中心",
        "track_node_code": "MAIN_LINE_ARRIVE",
        "track_node_description": "已抵达目的国",
        "pod_url": ""
      },
      {
        "process_time": "2026-05-18T09:30:00Z",
        "process_utc_time": "2026-05-18T09:30:00Z",
        "process_content": "已交付",
        "process_country": "DE",
        "process_province": "",
        "process_city": "柏林",
        "process_location": "柏林邮局",
        "track_node_code": "DELIVERED",
        "track_node_description": "已送达",
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
