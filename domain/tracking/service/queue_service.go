package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	queueLogPrefix     = "queue_worker"
	providerCodeYuntu  = "yuntu"
)

// QueueService handles processing messages from Redis queue.
type QueueService struct {
	queueRepo     repo.QueueRepo
	rawEventRepo  repo.RawEventRepo
	detailService *TrackingDetailService
}

// NewQueueService creates a new queue service.
func NewQueueService(
	queueRepo repo.QueueRepo,
	rawEventRepo repo.RawEventRepo,
	detailService *TrackingDetailService,
) *QueueService {
	return &QueueService{
		queueRepo:     queueRepo,
		rawEventRepo:  rawEventRepo,
		detailService: detailService,
	}
}

// StartListening starts BRPOP loop to consume queue messages.
func (s *QueueService) StartListening(ctx context.Context) error {
	logx.Infof("%s: starting queue listener", queueLogPrefix)

	for {
		select {
		case <-ctx.Done():
			logx.Infof("%s: stopping queue listener", queueLogPrefix)
			return nil
		default:
			s.processNext(ctx)
		}
	}
}

// StopListening stops the queue listener gracefully.
func (s *QueueService) StopListening() {
	logx.Infof("%s: stop signal received", queueLogPrefix)
}

func (s *QueueService) processNext(ctx context.Context) {
	// Use 5 second timeout to allow graceful shutdown
	msg, err := s.queueRepo.BRPop(ctx, 5)
	if err != nil {
		// Ignore timeout errors (normal when no messages)
		if ctx.Err() != nil {
			return
		}
		// Other errors - log and continue
		logx.Errorf("%s: brpop failed: %v", queueLogPrefix, err)
		time.Sleep(1 * time.Second)
		return
	}

	if err := s.processMessage(ctx, msg); err != nil {
		logx.Errorf("%s: process message failed: %v", queueLogPrefix, err)
	}
}

func (s *QueueService) processMessage(ctx context.Context, msg *repo.QueueMessage) error {
	// Parse raw message
 rawData, data, err := s.parseMessageWithRaw(msg.RawData)
	if err != nil {
		return fmt.Errorf("parse message failed: %w", err)
	}

	waybill := extractStr(data, "waybill_number")
	tracking := extractStr(data, "tracking_number")
	customer := extractStr(data, "customer_code")
	dataCode := extractStr(data, "data_code")

	logx.Infof("%s: processing waybill=%s", queueLogPrefix, waybill)

	// Save raw_events
	if err := s.saveRawEvents(ctx, rawData, data, waybill, tracking, customer, dataCode); err != nil {
		logx.Errorf("%s: save raw_events failed: %v", queueLogPrefix, err)
		// Continue processing even if raw_events save fails
	}

	// Save tracking_details and update cache
	if err := s.detailService.SaveFromWebhook(ctx, data); err != nil {
		return fmt.Errorf("save webhook failed: %w", err)
	}

	logx.Infof("%s: processed waybill=%s", queueLogPrefix, waybill)
	return nil
}

func (s *QueueService) parseMessageWithRaw(raw []byte) (map[string]interface{}, map[string]interface{}, error) {
	var rawData map[string]interface{}
	if err := json.Unmarshal(raw, &rawData); err != nil {
		return nil, nil, err
	}

	// Handle nested data structure
	data, ok := rawData["data"].(map[string]interface{})
	if !ok {
		data = rawData
	}

	return rawData, data, nil
}

func (s *QueueService) saveRawEvents(ctx context.Context, rawData, data map[string]interface{}, waybill, tracking, customer, dataCode string) error {
	events, ok := data["track_events"].([]interface{})
	if !ok {
		return nil
	}

	for _, evt := range events {
		evtMap, ok := evt.(map[string]interface{})
		if !ok {
			continue
		}

		nodeCode := extractStr(evtMap, "track_node_code")
		processTime := extractStr(evtMap, "process_time")

		rawEvent := &entity.RawEvent{
			IdempotencyKey: buildIdempotencyKey(providerCodeYuntu, waybill, nodeCode, processTime),
			ProviderCode:   providerCodeYuntu,
			DataCode:       dataCode,
			WaybillNumber:  waybill,
			TrackingNumber: tracking,
			CustomerCode:   customer,
			TrackNodeCode:  nodeCode,
			ProcessTime:    processTime,
			Payload:        marshalMap(evtMap),
			EnvelopeMeta:   marshalMap(rawData),
			Status:         entity.RawEventProcessed,
			MaxRetries:     entity.DefaultMaxRetries,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if _, err := s.rawEventRepo.Save(ctx, rawEvent); err != nil {
			logx.Errorf("%s: save raw_event failed: node=%s err=%v", queueLogPrefix, nodeCode, err)
		}
	}

	return nil
}

func (s *QueueService) parseMessage(raw []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	// Handle nested data structure
	if inner, ok := data["data"].(map[string]interface{}); ok {
		return inner, nil
	}

	return data, nil
}

// GetQueueLength returns current queue length for monitoring.
func (s *QueueService) GetQueueLength(ctx context.Context) (int64, error) {
	return s.queueRepo.LLen(ctx)
}

func buildIdempotencyKey(providerCode, waybill, nodeCode, processTime string) string {
	return fmt.Sprintf("%s:%s:%s:%s", providerCode, waybill, nodeCode, processTime)
}

func marshalMap(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}
