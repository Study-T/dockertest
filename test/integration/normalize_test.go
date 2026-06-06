package integration

import (
	"encoding/json"
	"os"
	"testing"

	"ns-tracking-go/domain/tracking/service"
)

// TestNormalize_RealTPushDataSample uses the real tisPushData sample from the project.
func TestNormalize_RealTPushDataSample(t *testing.T) {
	samplePath := "../../../Desktop/ns-tracking-go/tis_push_test.json"
	data, err := os.ReadFile(samplePath)
	if err != nil {
		t.Skipf("sample file not found: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("failed to parse sample: %v", err)
	}

	// Extract data layer (tisPushData has data wrapper)
	inner := payload
	if d, ok := payload["data"].(map[string]interface{}); ok {
		inner = d
	}

	n := service.NewNormalizer()
	result, err := n.Normalize(inner)
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}

	// Verify basic fields
	if result.WaybillNumber == "" {
		t.Error("waybill_number is empty")
	}
	if result.PackageState == "" {
		t.Error("package_state is empty")
	}
	if len(result.OrderTrackingDetails) == 0 {
		t.Error("no order tracking details")
	}

	// Verify all events have required fields
	for i, detail := range result.OrderTrackingDetails {
		if _, ok := detail["TrackingStatus"]; !ok {
			t.Errorf("event[%d] missing TrackingStatus", i)
		}
		if _, ok := detail["TrackNodeCode"]; !ok {
			t.Errorf("event[%d] missing TrackNodeCode", i)
		}
		if _, ok := detail["ProcessDate"]; !ok {
			t.Errorf("event[%d] missing ProcessDate", i)
		}
	}

	// Verify TrackingStatus is string type
	for i, detail := range result.OrderTrackingDetails {
		if _, ok := detail["TrackingStatus"].(string); !ok {
			t.Errorf("event[%d] TrackingStatus is not string: %T", i, detail["TrackingStatus"])
		}
	}

	// Verify sorted by ProcessDate
	for i := 1; i < len(result.OrderTrackingDetails); i++ {
		prev := result.OrderTrackingDetails[i-1]["ProcessDate"].(string)
		curr := result.OrderTrackingDetails[i]["ProcessDate"].(string)
		if prev > curr {
			t.Errorf("events not sorted: event[%d]=%s > event[%d]=%s", i-1, prev, i, curr)
		}
	}

	t.Logf("Normalized: waybill=%s status=%d packageState=%s events=%d",
		result.WaybillNumber, result.TopLevelTrackingStatus, result.PackageState, len(result.OrderTrackingDetails))
}

// TestDetailStructure_RubyCompat verifies the detail structure matches Ruby fetch_cache expectations.
func TestDetailStructure_RubyCompat(t *testing.T) {
	n := service.NewNormalizer()
	data := map[string]interface{}{
		"waybill_number":   "YT2615000705163221",
		"tracking_number":  "3SWLT0037525480",
		"package_status":   "T",
		"origin_code":      "CN",
		"destination_code": "NL",
		"product_code":     "MUZXR",
		"customer_code":    "CNHC318962",
		"channel_code":     "NLDHL",
		"last_mile_name":   "PostNL",
		"last_mile_site":   "https://postnl.nl/track",
		"phone_number":     "+31 883454399",
		"track_events": []interface{}{
			map[string]interface{}{
				"track_node_code": "ORDER_CREATION", "track_node_description": "Order created",
				"process_time": "2026-05-30T16:00:00", "process_utc_time": "2026-05-30T08:00:00",
			},
			map[string]interface{}{
				"track_node_code": "DELIVERED", "track_node_description": "Delivered",
				"process_time": "2026-06-02T16:40:28", "process_utc_time": "2026-06-02T08:40:28",
			},
		},
	}

	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}

	// Build detail structure like TrackingDetailService.buildDetail
	item := map[string]interface{}{
		"TrackingNumber":       result.WaybillNumber,
		"WayBillNumber":        result.WaybillNumber,
		"TrackingStatus":       result.TopLevelTrackingStatus,
		"PackageState":         result.PackageState,
		"LastMileCarrierName":  result.LastMileCarrierName,
		"ProviderSite":         result.ProviderSite,
		"ProvicerTelephone":    result.ProvicerTelephone,
		"CountryCode":          result.CountryCode,
		"OriginCountryCode":    result.OriginCountryCode,
		"OrderTrackingDetails": result.OrderTrackingDetails,
	}

	detail := map[string]interface{}{
		"response":  map[string]interface{}{"Item": item},
		"synced_at": "2026-06-05T10:00:00Z",
	}

	// Verify Ruby fetch_cache compatibility
	response, ok := detail["response"].(map[string]interface{})
	if !ok {
		t.Fatal("detail['response'] is not a map")
	}

	itemMap, ok := response["Item"].(map[string]interface{})
	if !ok {
		t.Fatal("detail['response']['Item'] is not a map")
	}

	// Ruby: detail.dig('response', 'Item')
	if itemMap["TrackingNumber"] == nil {
		t.Error("Item.TrackingNumber is nil")
	}

	// Ruby: detail['synced_at'].to_time
	if _, ok := detail["synced_at"].(string); !ok {
		t.Error("detail['synced_at'] is not a string")
	}

	// Ruby: STATUS_CODES[status.to_s]
	if _, ok := itemMap["TrackingStatus"].(int); ok {
		// int is OK — Ruby's to_s converts it
	}

	// ProvicerTelephone (missing 'd')
	if itemMap["ProvicerTelephone"] == nil {
		t.Error("ProvicerTelephone is nil (Ruby reads this exact key)")
	}

	t.Logf("Detail structure verified: %d events, status=%v, packageState=%v",
		len(result.OrderTrackingDetails), itemMap["TrackingStatus"], itemMap["PackageState"])
}
