package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/event/define"
	"ns-tracking-go/domain/tracking/repo"
)

const (
	serviceClassYunExpress = "YunExpressService"
	trackingCacheTTL       = 24 * time.Hour
)

// TrackingDetailService handles writing normalized data to tracking_details.
type TrackingDetailService struct {
	detailRepo repo.TrackingDetailRepo
	logRepo    repo.TrackingLogRepo
	cacheRepo  repo.TrackingCache
	normalizer *Normalizer
	eventBus   define.Dispatcher
}

func NewTrackingDetailService(
	detailRepo repo.TrackingDetailRepo,
	logRepo repo.TrackingLogRepo,
	cacheRepo repo.TrackingCache,
	normalizer *Normalizer,
	eventBus define.Dispatcher,
) *TrackingDetailService {
	return &TrackingDetailService{
		detailRepo: detailRepo,
		logRepo:    logRepo,
		cacheRepo:  cacheRepo,
		normalizer: normalizer,
		eventBus:   eventBus,
	}
}

// QueryByOrderNumber queries tracking detail by order number and returns formatted result.
// First checks cache, then falls back to database.
func (s *TrackingDetailService) QueryByOrderNumber(ctx context.Context, orderNumber string) (*TrackingResult, error) {
	// 1. Try cache first
	cached, err := s.cacheRepo.GetTrackingDetail(ctx, orderNumber)
	if err == nil && cached != "" {
		var result TrackingResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	// 2. Query from database
	detail, err := s.detailRepo.FindByTrackingNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}

	result := s.buildTrackingResult(detail)

	// 3. Write to cache asynchronously
	if data, err := json.Marshal(result); err == nil {
		_ = s.cacheRepo.SetTrackingDetail(ctx, orderNumber, string(data), trackingCacheTTL)
	}

	return result, nil
}

func (s *TrackingDetailService) buildTrackingResult(detail *entity.TrackingDetail) *TrackingResult {
	detailData := detail.Detail
	if detailData == nil {
		detailData = make(map[string]interface{})
	}

	response, _ := detailData["response"].(map[string]interface{})
	item, _ := response["Item"].(map[string]interface{})

	// 提取 track_events
	var trackEvents []TrackEvent
	if events, ok := item["OrderTrackingDetails"].([]interface{}); ok {
		for _, evt := range events {
			if evtMap, ok := evt.(map[string]interface{}); ok {
				trackEvents = append(trackEvents, TrackEvent{
					ProcessTime:     extractStr(evtMap, "ProcessDate"),
					ProcessUTCTime:  extractStr(evtMap, "ProcessDate"),
					ProcessContent:  extractStr(evtMap, "ProcessContent"),
					ProcessCountry:  extractStr(evtMap, "ProcessCountry"),
					ProcessProvince: extractStr(evtMap, "ProcessProvince"),
					ProcessCity:     extractStr(evtMap, "ProcessCity"),
					ProcessLocation: extractStr(evtMap, "ProcessLocation"),
					TrackNodeCode:   extractStr(evtMap, "TrackNodeCode"),
					TrackNodeDesc:   extractStr(evtMap, "TrackCodeDescription"),
					PodURL:          extractStr(evtMap, "PodURL"),
				})
			}
		}
	}

	// 获取 package_status
	packageStatus := PackageStatusMapping[detail.Status]
	if packageStatus == "" {
		packageStatus = "N"
	}

	return &TrackingResult{
		OrderNumber:   detail.TrackingNumber,
		PackageStatus: packageStatus,
		TrackInfo: TrackInfo{
			WaybillNumber:       detail.TrackingNumber,
			TrackingNumber:      extractStr(item, "TrackingNumber"),
			CustomerOrderNumber: extractStr(item, "CustomerOrderNumber"),
			ProductCode:         extractStr(detailData, "product_code"),
			ProductName:         extractStr(detailData, "product_name"),
			ChannelCode:         extractStr(detailData, "channel_code"),
			CheckInTime:         extractStr(item, "CheckInTime"),
			CheckOutTime:        extractStr(item, "CheckOutTime"),
			PickUpTime:          extractStr(item, "PickUpTime"),
			CustomerCode:        extractStr(detailData, "customer_code"),
			OriginCode:          extractStr(item, "OriginCountryCode"),
			DestinationCode:     extractStr(item, "CountryCode"),
			PostalCode:          extractStr(item, "PostalCode"),
			ActualWeight:        extractFloat(item, "ActualWeight"),
			IntervalDay:         extractInt(item, "IntervalDay"),
			IntervalWorkDay:     extractInt(item, "IntervalWorkDay"),
			LastMileSite:        extractStr(detailData, "last_mile_tracking_url"),
			LastMileName:        extractStr(detailData, "last_mile_carrier_name"),
			PhoneNumber:         extractStr(detailData, "last_mile_carrier_phone"),
			TrackEvents:         trackEvents,
			PodURL:              extractStr(item, "PodURL"),
			PodURLs:             extractStringArray(item, "PodURLs"),
			IsSignature:         extractBool(item, "IsSignature"),
			SignatureURLs:       extractStringArray(item, "SignatureUrls"),
			EstimatedDeliveryToDateZone:   extractStr(item, "EstimatedDeliveryToDateZone"),
			EstimatedDeliveryFromDateZone: extractStr(item, "EstimatedDeliveryFromDateZone"),
		},
	}
}

// SaveFromWebhook normalizes payload and writes to tracking_details + tracking_logs.
func (s *TrackingDetailService) SaveFromWebhook(ctx context.Context, data map[string]interface{}) error {
	normalized, err := s.normalizer.Normalize(data)
	if err != nil {
		return fmt.Errorf("normalize failed: %w", err)
	}

	// Query existing record for last_detail backup
	existing, _ := s.detailRepo.FindByTrackingNumber(ctx, normalized.WaybillNumber)

	// Build new detail
	detail := s.buildDetail(normalized)

	// Backup old detail to last_detail
	var lastDetail map[string]interface{}
	if existing != nil {
		lastDetail = deepCopyMap(existing.Detail)
	}

	// Get tracking times from events
	receivedAt := FindReceivedAt(normalized.OrderTrackingDetails)
	deliveredAt := FindDeliveredAt(normalized.OrderTrackingDetails)
	trackedAt := FindTrackedAt(normalized.OrderTrackingDetails)
	trackStatus := CalculateTrackStatus(normalized.TopLevelTrackingStatus, "")

	// Debug: log the extracted times
	fmt.Printf("Extracted times: receivedAt=%v, deliveredAt=%v, trackedAt=%v\n", receivedAt, deliveredAt, trackedAt)
	fmt.Printf("OrderTrackingDetails count: %d\n", len(normalized.OrderTrackingDetails))
	if len(normalized.OrderTrackingDetails) > 0 {
		fmt.Printf("First event: %v\n", normalized.OrderTrackingDetails[0])
	}

	// Create tracking_log record
	now := time.Now()
	trackingLog := &entity.TrackingLog{
		SourceTrackingNumber: normalized.WaybillNumber,
		TrackingNumber:       normalized.TrackingNumber,
		ChannelAlias:         normalized.ChannelAlias,
		ShippingAgent:        normalized.ShippingAgent,
		ShippingChannel:      normalized.ShippingChannel,
		CountryCode:          normalized.CountryCode,
		TrackStatus:          trackStatus,
		SyncedAt:             &now,
		ReceivedAt:           receivedAt,
		DeliveredAt:          deliveredAt,
		TrackedAt:            trackedAt,
	}
	if createErr := s.logRepo.Create(ctx, trackingLog); createErr != nil {
		fmt.Printf("warning: create tracking_log failed: %v\n", createErr)
	} else {
		fmt.Printf("Created tracking_log: id=%d\n", trackingLog.ID)
	}

	detailEntity := &entity.TrackingDetail{
		TrackingNumber: normalized.WaybillNumber,
		ServiceClass:   serviceClassYunExpress,
		Status:         normalized.TopLevelTrackingStatus,
		Detail:         detail,
		LastDetail:     lastDetail,
		SyncedAt:       time.Now(),
		TrackingLogID:  trackingLog.ID,
	}

	// Save to database (upsert)
	if err := s.detailRepo.Save(ctx, detailEntity); err != nil {
		return fmt.Errorf("save tracking detail failed: %w", err)
	}

	// Update Redis cache
	s.writeToCache(ctx, detailEntity)

	return nil
}

func (s *TrackingDetailService) writeToCache(ctx context.Context, detail *entity.TrackingDetail) {
	result := s.buildTrackingResult(detail)
	data, err := json.Marshal(result)
	if err != nil {
		return
	}
	_ = s.cacheRepo.SetTrackingDetail(ctx, detail.TrackingNumber, string(data), trackingCacheTTL)
}

func (s *TrackingDetailService) buildDetail(nr *NormalizedResult) map[string]interface{} {
	// Convert []map[string]interface{} to []interface{} for JSON compatibility
	trackingDetails := make([]interface{}, len(nr.OrderTrackingDetails))
	for i, d := range nr.OrderTrackingDetails {
		trackingDetails[i] = d
	}

	item := map[string]interface{}{
		"TrackingNumber":       nr.WaybillNumber,
		"WayBillNumber":        nr.WaybillNumber,
		"TrackingStatus":       fmt.Sprintf("%d", nr.TopLevelTrackingStatus), // string for Ruby
		"PackageState":         nr.PackageState,
		"LastMileCarrierName":  nr.LastMileCarrierName,
		"ProviderSite":         nr.ProviderSite,
		"ProvicerTelephone":    nr.ProvicerTelephone,
		"ProviderName":         nr.ProviderName,
		"CountryCode":          nr.CountryCode,
		"OriginCountryCode":    nr.OriginCountryCode,
		"OrderTrackingDetails": trackingDetails,
	}

	return map[string]interface{}{
		"response":                   map[string]interface{}{"Item": item},
		"synced_at":                  time.Now().UTC().Format(time.RFC3339),
		"last_mile_tracking_number":  nr.LastMileTrackingNumber,
		"last_mile_carrier_name":     nr.LastMileCarrierName,
		"last_mile_tracking_url":     nr.LastMileTrackingURL,
		"last_mile_carrier_phone":    nr.LastMileCarrierPhone,
		"product_code":               nr.ProductCode,
		"product_name":               nr.ProductName,
		"channel_code":               nr.ChannelCode,
		"customer_code":              nr.CustomerCode,
	}
}

// CalculateTrackStatus maps TrackingStatus int → track_status enum.
func CalculateTrackStatus(status int, errorMessage string) int {
	if errorMessage != "" && status != 50 {
		return entity.TrackStatusTrackException
	}
	switch status {
	case 10:
		return entity.TrackStatusToReceive
	case 20, 30:
		return entity.TrackStatusInTransit
	case 50:
		return entity.TrackStatusDelivered
	case 40, 60, 70, 80, 90, 100:
		return entity.TrackStatusTrackException
	default:
		return entity.TrackStatusToReceive
	}
}

// FindReceivedAt finds first event with TrackingStatus == "10".
// Behavior fix: not using the buggy OR logic from Ruby.
func FindReceivedAt(events []map[string]interface{}) *time.Time {
	for _, evt := range events {
		if s, ok := evt["TrackingStatus"].(string); ok && s == "10" {
			return parseEventTime(evt["ProcessDate"])
		}
	}
	return nil
}

// FindDeliveredAt finds first event with TrackingStatus == "50".
func FindDeliveredAt(events []map[string]interface{}) *time.Time {
	for _, evt := range events {
		if s, ok := evt["TrackingStatus"].(string); ok && s == "50" {
			return parseEventTime(evt["ProcessDate"])
		}
	}
	return nil
}

// FindTrackedAt returns the last event's ProcessDate.
func FindTrackedAt(events []map[string]interface{}) *time.Time {
	if len(events) == 0 {
		return nil
	}
	return parseEventTime(events[len(events)-1]["ProcessDate"])
}

func parseEventTime(v interface{}) *time.Time {
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}

	// Try multiple formats
	formats := []string{
		time.RFC3339,              // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z",   // UTC
		"2006-01-02T15:04:05",    // No timezone
		"2006-01-02 15:04:05",    // Space separated
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return &t
		}
	}

	return nil
}

func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	b, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var dst map[string]interface{}
	json.Unmarshal(b, &dst)
	return dst
}
