package service

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// NodeCodeMapping maps track_node_code → TrackingStatus (39 nodes).
// Source: OMS real samples + 云途基础数据表.xlsx cross-validation.
var NodeCodeMapping = map[string]int{
	"ORDER_CREATION":          10,
	"PICKED_UP":               20,
	"PRE_ADVICING":            20,
	"PRE_INFO":                20,
	"FIRST_MILE_ARRIVE":       20,
	"FIRST_MILE_DEPART":       20,
	"TRANSIT_OUT":             20,
	"TRANSIT_IN":              20,
	"DEPART_CONFIRM_OC":       20,
	"ARRIVE_CONFIRM_OC":       20,
	"ARRIVE_ORIN_CUSTOMS":     20,
	"EXPORT_CUSTOMS_COMPLETE": 20,
	"AIRPORT_RELEASE":         20,
	"MAIN_LINE_DEPART":        20,
	"MAIN_LINE_ARRIVE":        20,
	"CUSTOMS_PROCESSING":      20,
	"CUSTOMS_COMPLETE":        20,
	"CUSTOMS_RELEASE":         20,
	"READY_FOR_PICKUP":        20,
	"PICKUP_CARGO_TERMINAL":   20,
	"IN_TRANSIT":              20,
	"READY_FOR_OUTBOUND":      20,
	"TRANSITHUB_ARRIVE":       20,
	"CHANGE_WAYBILL_INFO":     20,
	"EDD":                     20,
	"CARRIER_PICKUP":          20,
	"IN_TRANSIT_CARRIER":      20,
	"DELIVERY_ATTEMPT":        40,
	"DELIVERED":               50,
	"DELIVERY_FAILURE":        60,
	"CUSTOMS_DELAY":           60,
	"CUSTOMS_HOLD":            60,
	"CUSTOMS_INSPCTION":       60,
	"AIRPORT_HOLD":            60,
	"AIRPORT_INSPECTION":      60,
	"PACKAGE_EXCEPTION":       60,
	"PACKAGE_LOST":            60,
	"OTHER":                   60,
	"RETURNED":                90,
	"RETURNED_BACK":           90,
	"RETURNED_TO_SENDER":      90,
	// Chinese codes from OpenAPI Webhook
	"YB/XD": 10,
	"XD/ZC": 20,
	"XD/CS": 20,
	"MD/DD": 30,
	"MD/TT": 50,
}

// PackageStateMapping maps TrackingStatus → PackageState.
var PackageStateMapping = map[int]string{
	0: "0", 10: "4", 20: "2", 30: "2",
	40: "6", 50: "3", 60: "6", 70: "6",
	80: "6", 90: "7", 100: "5",
}

// PackageStatusMapping maps TrackingStatus → package_status letter.
var PackageStatusMapping = map[int]string{
	0: "N", 10: "N", 20: "T", 30: "T",
	40: "T", 50: "D", 60: "E", 70: "E",
	80: "E", 90: "R", 100: "C",
}

// NormalizedResult is the output of normalization.
type NormalizedResult struct {
	WaybillNumber          string
	TrackingNumber         string
	TopLevelTrackingStatus int
	PackageState           string
	OrderTrackingDetails   []map[string]interface{}
	LastMileTrackingNumber string
	LastMileCarrierName    string
	LastMileTrackingURL    string
	LastMileCarrierPhone   string
	ProviderName           string
	ProviderSite           string
	ProvicerTelephone      string // note: missing 'd' — Ruby compat
	CountryCode            string
	OriginCountryCode      string
	ProductCode            string
	ProductName            string
	ChannelCode            string
	CustomerCode           string
	ShippingAgent          string
	ShippingChannel        string
	ChannelAlias           string
}

// Normalizer converts Webhook payload to OMS-equivalent format.
type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// Normalize converts Webhook data to OMS-equivalent format.
func (n *Normalizer) Normalize(data map[string]interface{}) (*NormalizedResult, error) {
	result := &NormalizedResult{
		WaybillNumber:     extractStr(data, "waybill_number"),
		TrackingNumber:    extractStr(data, "tracking_number"),
		CountryCode:       extractStr(data, "destination_code"),
		OriginCountryCode: extractStr(data, "origin_code"),
		ProductCode:       extractStr(data, "product_code"),
		ProductName:       extractStr(data, "product_name"),
		ChannelCode:       extractStr(data, "channel_code"),
		CustomerCode:      extractStr(data, "customer_code"),
		ShippingAgent:     "YunExpress",
		ShippingChannel:   extractStr(data, "product_code"),
		ChannelAlias:      extractStr(data, "product_name"),
	}

	events := n.extractEvents(data)
	details := make([]map[string]interface{}, 0, len(events))

	for _, evt := range events {
		detail, _ := n.normalizeEvent(evt)
		details = append(details, detail)
	}

	// Sort by ProcessDate ascending — technical spec requires "按 ProcessDate 排序后最后一条"
	sort.Slice(details, func(i, j int) bool {
		return extractStr(details[i], "ProcessDate") < extractStr(details[j], "ProcessDate")
	})

	lastStatus := 0
	if len(details) > 0 {
		if s, ok := details[len(details)-1]["TrackingStatus"].(string); ok {
			fmt.Sscanf(s, "%d", &lastStatus)
		}
	}

	// Customs inspection: 80 → 20
	if lastStatus == 80 {
		lastStatus = 20
	}
	for _, d := range details {
		if s, ok := d["TrackingStatus"].(int); ok && s == 80 {
			d["TrackingStatus"] = 20
		}
	}

	result.TopLevelTrackingStatus = lastStatus
	result.PackageState = PackageStateMapping[lastStatus]
	if result.PackageState == "" {
		result.PackageState = "0"
	}
	result.OrderTrackingDetails = details

	// Last mile info
	result.LastMileCarrierName = extractStr(data, "last_mile_name")
	result.LastMileTrackingURL = extractStr(data, "last_mile_site")
	result.LastMileCarrierPhone = extractStr(data, "phone_number")
	result.ProviderSite = extractStr(data, "last_mile_site")
	result.ProvicerTelephone = extractStr(data, "phone_number")

	return result, nil
}

func (n *Normalizer) normalizeEvent(evt map[string]interface{}) (map[string]interface{}, int) {
	nodeCode := extractStr(evt, "track_node_code")
	status := NodeCodeMapping[nodeCode] // 0 if not found
	if status == 0 {
		status = 20 // default
	}
	// Customs: 80 → 20
	if status == 80 {
		status = 20
	}

	processTime := extractStr(evt, "process_time")
	processUTCTime := extractStr(evt, "process_utc_time")

	detail := map[string]interface{}{
		"ProcessDate":          processTime,
		"ProcessContent":       extractStr(evt, "process_content"),
		"ProcessLocation":      extractStr(evt, "process_location"),
		"TrackingStatus":       fmt.Sprintf("%d", status), // string for Ruby compat
		"TrackNodeCode":        nodeCode,
		"TrackCodeDescription": extractStr(evt, "track_node_description"),
		"ProcessTimezone":      n.calculateTimezone(processTime, processUTCTime),
		"ProcessCountry":       extractStr(evt, "process_country"),
		"ProcessProvince":      extractStr(evt, "process_province"),
		"ProcessCity":          extractStr(evt, "process_city"),
	}

	return detail, status
}

func (n *Normalizer) calculateTimezone(processTime, processUTCTime string) string {
	if processTime == "" || processUTCTime == "" {
		return ""
	}
	local, err1 := time.Parse("2006-01-02T15:04:05", processTime)
	utc, err2 := time.Parse("2006-01-02T15:04:05", processUTCTime)
	if err1 != nil || err2 != nil {
		return ""
	}
	diff := local.Sub(utc)
	hours := int(math.Round(diff.Hours()))
	if hours >= 0 {
		return fmt.Sprintf("UTC+%02d:00", hours)
	}
	return fmt.Sprintf("UTC-%02d:00", -hours)
}

func (n *Normalizer) extractEvents(data map[string]interface{}) []map[string]interface{} {
	raw, ok := data["track_events"].([]interface{})
	if !ok {
		return nil
	}
	events := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if evt, ok := item.(map[string]interface{}); ok {
			events = append(events, evt)
		}
	}
	return events
}

// extractStr extracts string value from map.
func extractStr(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

// extractFloat extracts float64 value from map.
func extractFloat(data map[string]interface{}, key string) float64 {
	if v, ok := data[key].(float64); ok {
		return v
	}
	return 0
}

// extractInt extracts int value from map.
func extractInt(data map[string]interface{}, key string) int {
	if v, ok := data[key].(float64); ok {
		return int(v)
	}
	return 0
}

// extractBool extracts bool value from map.
func extractBool(data map[string]interface{}, key string) bool {
	if v, ok := data[key].(bool); ok {
		return v
	}
	return false
}

// extractStringArray extracts string array from map.
func extractStringArray(data map[string]interface{}, key string) []string {
	if v, ok := data[key].([]interface{}); ok {
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
