package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"
	"ns-tracking-go/infrastructure/metrics"
	"ns-tracking-go/infrastructure/webhook"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const (
	providerCodeYuntu = "yuntu"
	contentTypeJSON   = "application/json"
)

func WebhookHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(405, "method not allowed", nil))
			return
		}
		if !isJSONContent(r) {
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(415, "unsupported content type", nil))
			return
		}

		maxSize := svcCtx.Config.Webhook.MaxPayloadSize
		if maxSize <= 0 {
			maxSize = 1 << 20
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, maxSize))
		if err != nil {
			logx.WithContext(r.Context()).Errorf("webhook: read body failed: %v", err)
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(400, "failed to read body", nil))
			return
		}

		if err := svcCtx.WebhookVerifier.Verify(
			r.Header.Get("X-Openapi-Request-Timestamp"),
			r.Header.Get("X-Openapi-Signature"),
			body,
		); err != nil {
			metrics.WebhookRequests.WithLabelValues("error").Inc()
			logx.WithContext(r.Context()).Errorf("webhook signature failed: ip=%s err=%v", r.RemoteAddr, err)
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(401, "invalid signature", nil))
			return
		}

		envelope, err := svcCtx.WebhookParser.Parse(body, svcCtx.Config.Webhook.EncryptKey)
		if err != nil {
			metrics.WebhookRequests.WithLabelValues("error").Inc()
			logx.WithContext(r.Context()).Errorf("webhook parse failed: %v", err)
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(400, "invalid payload", nil))
			return
		}

		// Grayscale check
		gsData := resolveDataLayer(envelope.Body)
		gsWaybill := extractString(gsData, "waybill_number")
		if !svcCtx.GrayscaleService.ShouldProcessByGo(gsWaybill) {
			metrics.GrayscaleDecisions.WithLabelValues(svcCtx.Config.Grayscale.Mode, "skipped").Inc()
			logx.WithContext(r.Context()).Infof("webhook: skipped by grayscale waybill=%s", gsWaybill)
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(200, "skipped by grayscale", map[string]string{"waybill_number": gsWaybill}))
			return
		}
		metrics.GrayscaleDecisions.WithLabelValues(svcCtx.Config.Grayscale.Mode, "processed").Inc()

		// 1. Save raw events
		results, err := saveTrackEvents(r.Context(), svcCtx, envelope)
		if err != nil {
			metrics.WebhookRequests.WithLabelValues("error").Inc()
			logx.WithContext(r.Context()).Errorf("webhook save raw events failed: %v", err)
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(500, "save raw events failed", nil))
			return
		}
		logResults(r.Context(), results)

		// Check duplicate status
		allDuplicate := true
		var lastRawEventID int64
		for _, res := range results {
			if res.IsNew {
				allDuplicate = false
			}
			if res.Event != nil {
				lastRawEventID = res.Event.ID
			}
		}

		// 2. Normalize and write to tracking_details
		data := resolveDataLayer(envelope.Body)
		if err := svcCtx.TrackingDetailSvc.SaveFromWebhook(r.Context(), data); err != nil {
			metrics.WebhookRequests.WithLabelValues("error").Inc()
			logx.WithContext(r.Context()).Errorf("webhook normalize failed: %v", err)
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(500, "normalize failed: "+err.Error(), nil))
			return
		}
		metrics.WebhookRequests.WithLabelValues("success").Inc()

		status := "queued"
		if allDuplicate {
			status = "duplicate"
		}
		httpx.OkJsonCtx(r.Context(), w, types.NewResponse(200, status, map[string]interface{}{
			"waybill_number": gsWaybill,
			"raw_event_id":   lastRawEventID,
		}))
	}
}

func isJSONContent(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return ct == "" || strings.HasPrefix(ct, contentTypeJSON)
}

func saveTrackEvents(ctx context.Context, svcCtx *svc.ServiceContext, env *webhook.Envelope) ([]*repo.SaveResult, error) {
	data := resolveDataLayer(env.Body)
	meta := marshalMap(buildEnvelopeMeta(env.Body, data))
	events := extractTrackEvents(data)
	if len(events) == 0 {
		return nil, nil
	}

	waybill := extractString(data, "waybill_number")
	tracking := extractString(data, "tracking_number")
	customer := extractString(data, "customer_code")
	dataCode := extractString(env.Body, "data_code")

	results := make([]*repo.SaveResult, 0, len(events))
	for _, evt := range events {
		nodeCode := extractString(evt, "track_node_code")
		processTime := extractString(evt, "process_time")

		rawEvent := &entity.RawEvent{
			IdempotencyKey: buildIdempotencyKey(providerCodeYuntu, waybill, nodeCode, processTime),
			ProviderCode:   providerCodeYuntu,
			DataCode:       dataCode,
			WaybillNumber:  waybill,
			TrackingNumber: tracking,
			CustomerCode:   customer,
			TrackNodeCode:  nodeCode,
			ProcessTime:    processTime,
			Payload:        marshalMap(evt),
			EnvelopeMeta:   meta,
			Status:         entity.RawEventPending,
			MaxRetries:     entity.DefaultMaxRetries,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		affected, err := svcCtx.RawEventRepo.Save(ctx, rawEvent)
		if err != nil {
			return results, fmt.Errorf("save node=%s: %w", nodeCode, err)
		}

		if affected == 0 {
			existing, findErr := svcCtx.RawEventRepo.FindByIdempotencyKey(ctx, rawEvent.IdempotencyKey)
			if findErr == nil && existing != nil {
				rawEvent = existing
			}
		}

		results = append(results, &repo.SaveResult{Event: rawEvent, IsNew: affected > 0})
	}
	return results, nil
}

func resolveDataLayer(body map[string]interface{}) map[string]interface{} {
	if data, ok := body["data"].(map[string]interface{}); ok {
		return data
	}
	return body
}

func buildEnvelopeMeta(body, data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"package_status":   extractString(data, "package_status"),
		"origin_code":      extractString(data, "origin_code"),
		"destination_code": extractString(data, "destination_code"),
		"product_code":     extractString(data, "product_code"),
		"product_name":     extractString(data, "product_name"),
		"customer_code":    extractString(data, "customer_code"),
		"channel_code":     extractString(data, "channel_code"),
		"last_mile_site":   extractString(data, "last_mile_site"),
		"last_mile_name":   extractString(data, "last_mile_name"),
		"phone_number":     extractString(data, "phone_number"),
	}
}

func extractTrackEvents(data map[string]interface{}) []map[string]interface{} {
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

func buildIdempotencyKey(providerCode, waybill, nodeCode, processTime string) string {
	return fmt.Sprintf("%s:%s:%s:%s", providerCode, waybill, nodeCode, processTime)
}

func extractString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func marshalMap(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func logResults(ctx context.Context, results []*repo.SaveResult) {
	logger := logx.WithContext(ctx)
	newCount, dupCount := 0, 0
	for _, res := range results {
		if res.IsNew {
			newCount++
		} else {
			dupCount++
		}
	}
	logger.Infof("webhook: processed %d events (%d new, %d duplicate)", len(results), newCount, dupCount)
}
