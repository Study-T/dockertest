package main

import (
	"encoding/json"
	"fmt"

	"ns-tracking-go/domain/tracking/service"
)

func main() {
	// 模拟 Redis 队列中的数据
	jsonData := `{
  "data_code": "tisPushData",
  "data": {
    "waybill_number": "YT2612500702088888",
    "tracking_number": "00340434525230666666",
    "package_status": "D",
    "track_events": [
      {
        "process_time": "2026-06-01T10:00:00Z",
        "process_utc_time": "2026-06-01T10:00:00Z",
        "process_content": "已收到发货信息",
        "track_node_code": "ORDER_CREATION"
      },
      {
        "process_time": "2026-06-05T14:20:00Z",
        "process_utc_time": "2026-06-05T14:20:00Z",
        "process_content": "已签收",
        "track_node_code": "DELIVERED"
      }
    ]
  }
}`

	var rawData map[string]interface{}
	json.Unmarshal([]byte(jsonData), &rawData)

	// 提取 data 层
	data, _ := rawData["data"].(map[string]interface{})

	// 标准化
	normalizer := service.NewNormalizer()
	normalized, err := normalizer.Normalize(data)
	if err != nil {
		fmt.Printf("Normalize failed: %v\n", err)
		return
	}

	fmt.Printf("WaybillNumber: %s\n", normalized.WaybillNumber)
	fmt.Printf("TopLevelTrackingStatus: %d\n", normalized.TopLevelTrackingStatus)
	fmt.Printf("OrderTrackingDetails count: %d\n", len(normalized.OrderTrackingDetails))

	// 检查时间提取
	fmt.Println("\n=== OrderTrackingDetails ===")
	for i, evt := range normalized.OrderTrackingDetails {
		fmt.Printf("Event %d: ProcessDate=%s, TrackingStatus=%s\n",
			i, evt["ProcessDate"], evt["TrackingStatus"])
	}

	// 测试 FindReceivedAt、FindDeliveredAt、FindTrackedAt
	receivedAt := service.FindReceivedAt(normalized.OrderTrackingDetails)
	deliveredAt := service.FindDeliveredAt(normalized.OrderTrackingDetails)
	trackedAt := service.FindTrackedAt(normalized.OrderTrackingDetails)

	fmt.Printf("\nreceivedAt: %v\n", receivedAt)
	fmt.Printf("deliveredAt: %v\n", deliveredAt)
	fmt.Printf("trackedAt: %v\n", trackedAt)
}
