package webhook

import (
	"testing"
)

func TestBuildIdempotencyKey(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		waybill     string
		nodeCode    string
		processTime string
		want        string
	}{
		{"ORDER_CREATION", "yuntu", "YT123", "ORDER_CREATION", "2025-06-10T08:00:00Z", "yuntu:YT123:ORDER_CREATION:2025-06-10T08:00:00Z"},
		{"YB/XD", "yuntu", "YT123", "YB/XD", "2025-06-10T08:00:00Z", "yuntu:YT123:YB/XD:2025-06-10T08:00:00Z"},
		{"DELIVERED", "yuntu", "YT456", "DELIVERED", "2025-06-12T10:00:00Z", "yuntu:YT456:DELIVERED:2025-06-12T10:00:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildIdempotencyKey(tt.provider, tt.waybill, tt.nodeCode, tt.processTime)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractString(t *testing.T) {
	data := map[string]interface{}{
		"waybill_number": "YT123",
		"number_field":   42,
	}
	tests := []struct {
		key  string
		want string
	}{
		{"waybill_number", "YT123"},
		{"missing", ""},
		{"number_field", ""},
	}
	for _, tt := range tests {
		if got := extractString(data, tt.key); got != tt.want {
			t.Errorf("extractString(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestResolveDataLayer(t *testing.T) {
	tests := []struct {
		name   string
		body   map[string]interface{}
		wantKey string
		wantVal string
	}{
		{"tisPushData", map[string]interface{}{"data": map[string]interface{}{"waybill_number": "YT123"}}, "waybill_number", "YT123"},
		{"direct", map[string]interface{}{"waybill_number": "YT123"}, "waybill_number", "YT123"},
		{"nil data", map[string]interface{}{"data_code": "x", "data": nil}, "data_code", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := resolveDataLayer(tt.body)
			if data[tt.wantKey] != tt.wantVal {
				t.Errorf("got %v, want %v", data[tt.wantKey], tt.wantVal)
			}
		})
	}
}

func TestExtractTrackEvents(t *testing.T) {
	data := map[string]interface{}{
		"track_events": []interface{}{
			map[string]interface{}{"track_node_code": "ORDER_CREATION", "process_time": "t1"},
			map[string]interface{}{"track_node_code": "DELIVERED", "process_time": "t2"},
		},
	}
	events := extractTrackEvents(data)
	if len(events) != 2 {
		t.Fatalf("expected 2, got %d", len(events))
	}
	if events[0]["track_node_code"] != "ORDER_CREATION" {
		t.Error("first event wrong")
	}
	if events[1]["track_node_code"] != "DELIVERED" {
		t.Error("second event wrong")
	}
}

func TestExtractTrackEvents_Empty(t *testing.T) {
	if got := extractTrackEvents(map[string]interface{}{}); len(got) != 0 {
		t.Errorf("expected 0, got %d", len(got))
	}
}

func TestBuildEnvelopeMeta(t *testing.T) {
	body := map[string]interface{}{"data_code": "tisPushData"}
	data := map[string]interface{}{"package_status": "T", "origin_code": "CN"}
	meta := buildEnvelopeMeta(body, data)
	if meta["package_status"] != "T" {
		t.Error("package_status wrong")
	}
	if meta["origin_code"] != "CN" {
		t.Error("origin_code wrong")
	}
}

func TestMarshalMap(t *testing.T) {
	got := marshalMap(map[string]interface{}{"key": "value"})
	if got == "{}" || got == "null" {
		t.Errorf("expected non-empty, got %s", got)
	}
	if got := marshalMap(nil); got != "null" {
		t.Errorf("expected null, got %s", got)
	}
}
