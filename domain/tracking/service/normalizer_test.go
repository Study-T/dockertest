package service

import (
	"testing"
)

func TestNormalize_FullPayload(t *testing.T) {
	n := NewNormalizer()
	data := map[string]interface{}{
		"waybill_number":   "YT2615000705163221",
		"tracking_number":  "3SWLT0037525480",
		"package_status":   "T",
		"origin_code":      "CN",
		"destination_code": "NL",
		"product_code":     "MUZXR",
		"product_name":     "云途全球化妆品类专线挂号",
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
				"track_node_code": "PICKED_UP", "track_node_description": "Picked up",
				"process_time": "2026-05-30T20:30:00", "process_utc_time": "2026-05-30T12:30:00",
			},
			map[string]interface{}{
				"track_node_code": "DELIVERED", "track_node_description": "Delivered",
				"process_time": "2026-06-02T16:40:28", "process_utc_time": "2026-06-02T08:40:28",
			},
		},
	}

	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.WaybillNumber != "YT2615000705163221" {
		t.Errorf("waybill = %q", result.WaybillNumber)
	}
	if result.TopLevelTrackingStatus != 50 {
		t.Errorf("status = %d, want 50", result.TopLevelTrackingStatus)
	}
	if result.PackageState != "3" {
		t.Errorf("packageState = %q, want 3", result.PackageState)
	}
	if len(result.OrderTrackingDetails) != 3 {
		t.Fatalf("details count = %d, want 3", len(result.OrderTrackingDetails))
	}
	// TrackingStatus should be string
	if s := result.OrderTrackingDetails[0]["TrackingStatus"].(string); s != "10" {
		t.Errorf("event[0] status = %q, want '10'", s)
	}
	if s := result.OrderTrackingDetails[1]["TrackingStatus"].(string); s != "20" {
		t.Errorf("event[1] status = %q, want '20'", s)
	}
	if s := result.OrderTrackingDetails[2]["TrackingStatus"].(string); s != "50" {
		t.Errorf("event[2] status = %q, want '50'", s)
	}
	if tz := result.OrderTrackingDetails[0]["ProcessTimezone"].(string); tz != "UTC+08:00" {
		t.Errorf("timezone = %q, want UTC+08:00", tz)
	}
	if result.ProvicerTelephone != "+31 883454399" {
		t.Errorf("phone = %q", result.ProvicerTelephone)
	}
}

func TestNormalize_SortsByProcessDate(t *testing.T) {
	// Events arrive out of order — should sort by ProcessDate
	n := NewNormalizer()
	data := map[string]interface{}{
		"waybill_number": "YT123",
		"track_events": []interface{}{
			map[string]interface{}{"track_node_code": "DELIVERED", "process_time": "2026-06-03T10:00:00"},
			map[string]interface{}{"track_node_code": "ORDER_CREATION", "process_time": "2026-05-30T08:00:00"},
			map[string]interface{}{"track_node_code": "PICKED_UP", "process_time": "2026-06-01T10:00:00"},
		},
	}
	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After sorting: ORDER_CREATION(10), PICKED_UP(20), DELIVERED(50)
	if result.TopLevelTrackingStatus != 50 {
		t.Errorf("status = %d, want 50 (sorted by date)", result.TopLevelTrackingStatus)
	}
	// First detail should be ORDER_CREATION
	if s := result.OrderTrackingDetails[0]["TrackNodeCode"].(string); s != "ORDER_CREATION" {
		t.Errorf("first event = %q, want ORDER_CREATION", s)
	}
}

func TestNormalize_LastEventNotMax(t *testing.T) {
	n := NewNormalizer()
	data := map[string]interface{}{
		"waybill_number": "YT123",
		"track_events": []interface{}{
			map[string]interface{}{"track_node_code": "DELIVERED", "process_time": "2026-06-01T10:00:00"},
			map[string]interface{}{"track_node_code": "RETURNED", "process_time": "2026-06-02T10:00:00"},
			map[string]interface{}{"track_node_code": "DELIVERED", "process_time": "2026-06-03T10:00:00"},
		},
	}
	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TopLevelTrackingStatus != 50 {
		t.Errorf("status = %d, want 50 (last event, not max)", result.TopLevelTrackingStatus)
	}
}

func TestNormalize_CustomsInspectionNode(t *testing.T) {
	n := NewNormalizer()
	data := map[string]interface{}{
		"waybill_number": "YT123",
		"track_events": []interface{}{
			map[string]interface{}{"track_node_code": "CUSTOMS_INSPCTION", "process_time": "2026-06-01T10:00:00"},
		},
	}
	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TopLevelTrackingStatus != 60 {
		t.Errorf("status = %d, want 60", result.TopLevelTrackingStatus)
	}
}

func TestNormalize_UnknownNodeDefaultsTo20(t *testing.T) {
	n := NewNormalizer()
	data := map[string]interface{}{
		"waybill_number": "YT123",
		"track_events": []interface{}{
			map[string]interface{}{"track_node_code": "UNKNOWN_NODE", "process_time": "2026-06-01T10:00:00"},
		},
	}
	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TopLevelTrackingStatus != 20 {
		t.Errorf("status = %d, want 20 (default)", result.TopLevelTrackingStatus)
	}
}

func TestNormalize_ChineseNodeCodes(t *testing.T) {
	n := NewNormalizer()
	data := map[string]interface{}{
		"waybill_number": "YT123",
		"track_events": []interface{}{
			map[string]interface{}{"track_node_code": "YB/XD", "process_time": "t1"},
			map[string]interface{}{"track_node_code": "XD/ZC", "process_time": "t2"},
			map[string]interface{}{"track_node_code": "MD/TT", "process_time": "t3"},
		},
	}
	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	details := result.OrderTrackingDetails
	if s := details[0]["TrackingStatus"].(string); s != "10" {
		t.Errorf("YB/XD = %q, want '10'", s)
	}
	if s := details[1]["TrackingStatus"].(string); s != "20" {
		t.Errorf("XD/ZC = %q, want '20'", s)
	}
	if s := details[2]["TrackingStatus"].(string); s != "50" {
		t.Errorf("MD/TT = %q, want '50'", s)
	}
	if result.TopLevelTrackingStatus != 50 {
		t.Errorf("top status = %d, want 50", result.TopLevelTrackingStatus)
	}
}

func TestNormalize_EmptyEvents(t *testing.T) {
	n := NewNormalizer()
	data := map[string]interface{}{"waybill_number": "YT123", "track_events": []interface{}{}}
	result, err := n.Normalize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TopLevelTrackingStatus != 0 {
		t.Errorf("status = %d, want 0", result.TopLevelTrackingStatus)
	}
}

func TestCalculateTrackStatus(t *testing.T) {
	tests := []struct {
		status   int
		errMsg   string
		expected int
	}{
		{10, "", 0}, {20, "", 1}, {30, "", 1}, {50, "", 2},
		{40, "", 3}, {60, "", 3}, {90, "", 3}, {100, "", 3},
		{50, "err", 2}, {20, "err", 3}, {999, "", 0},
	}
	for _, tt := range tests {
		got := CalculateTrackStatus(tt.status, tt.errMsg)
		if got != tt.expected {
			t.Errorf("CalculateTrackStatus(%d, %q) = %d, want %d", tt.status, tt.errMsg, got, tt.expected)
		}
	}
}

func TestCalculateTimezone(t *testing.T) {
	n := NewNormalizer()
	tests := []struct {
		local, utc, want string
	}{
		{"2026-06-10T16:30:16", "2026-06-10T08:30:16", "UTC+08:00"},
		{"2026-06-10T10:00:00", "2026-06-10T15:00:00", "UTC-05:00"},
		{"", "", ""}, {"bad", "bad", ""},
	}
	for _, tt := range tests {
		if got := n.calculateTimezone(tt.local, tt.utc); got != tt.want {
			t.Errorf("calculateTimezone(%q, %q) = %q, want %q", tt.local, tt.utc, got, tt.want)
		}
	}
}

func TestFindReceivedAt(t *testing.T) {
	events := []map[string]interface{}{
		{"TrackingStatus": "20", "ProcessDate": "2026-06-01T10:00:00"},
		{"TrackingStatus": "10", "ProcessDate": "2026-05-30T08:00:00"},
		{"TrackingStatus": "50", "ProcessDate": "2026-06-02T10:00:00"},
	}
	got := FindReceivedAt(events)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Year() != 2026 || got.Month() != 5 {
		t.Errorf("got %v", got)
	}
}

func TestFindReceivedAt_NotFound(t *testing.T) {
	events := []map[string]interface{}{
		{"TrackingStatus": "20", "ProcessDate": "2026-06-01T10:00:00"},
	}
	if got := FindReceivedAt(events); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestFindDeliveredAt(t *testing.T) {
	events := []map[string]interface{}{
		{"TrackingStatus": "20", "ProcessDate": "2026-06-01T10:00:00"},
		{"TrackingStatus": "50", "ProcessDate": "2026-06-02T10:00:00"},
	}
	got := FindDeliveredAt(events)
	if got == nil || got.Day() != 2 {
		t.Errorf("got %v", got)
	}
}

func TestFindTrackedAt(t *testing.T) {
	events := []map[string]interface{}{
		{"TrackingStatus": "10", "ProcessDate": "2026-05-30T08:00:00"},
		{"TrackingStatus": "20", "ProcessDate": "2026-06-01T10:00:00"},
	}
	got := FindTrackedAt(events)
	if got == nil || got.Day() != 1 {
		t.Errorf("got %v, want day=1", got)
	}
}

func TestFindTrackedAt_Empty(t *testing.T) {
	if got := FindTrackedAt(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestDeepCopyMap(t *testing.T) {
	src := map[string]interface{}{"key": "value", "nested": map[string]interface{}{"inner": 42}}
	dst := deepCopyMap(src)
	dst["key"] = "modified"
	if src["key"] != "value" {
		t.Error("deep copy failed")
	}
}

func TestDeepCopyMap_Nil(t *testing.T) {
	if got := deepCopyMap(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestNodeCodeMapping_All39(t *testing.T) {
	expectedCodes := []string{"ORDER_CREATION", "PICKED_UP", "DELIVERED", "DELIVERY_FAILURE", "RETURNED", "YB/XD", "XD/ZC", "MD/TT"}
	for _, code := range expectedCodes {
		if _, ok := NodeCodeMapping[code]; !ok {
			t.Errorf("missing node code: %s", code)
		}
	}
	if len(NodeCodeMapping) < 44 {
		t.Errorf("expected >= 44 mappings, got %d", len(NodeCodeMapping))
	}
}

func TestPackageStateMapping(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{0, "0"}, {10, "4"}, {20, "2"}, {50, "3"}, {60, "6"}, {90, "7"}, {100, "5"},
	}
	for _, tt := range tests {
		if got := PackageStateMapping[tt.status]; got != tt.want {
			t.Errorf("PackageStateMapping[%d] = %q, want %q", tt.status, got, tt.want)
		}
	}
}
