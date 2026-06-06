package cron

import (
	"context"
	"time"

	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/infrastructure/metrics"

	"github.com/zeromicro/go-zero/core/logx"
)

// SyncJob runs fallback sync for missed webhook events.
// Currently a stub — finds candidates but does not pull OMS API.
// TODO: implement OMS API pull with GE_YUN_EXPRESS_TOKEN.
type SyncJob struct {
	syncService *service.SyncService
}

func NewSyncJob(syncService *service.SyncService) *SyncJob {
	return &SyncJob{syncService: syncService}
}

func (j *SyncJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logx.Info("sync_job: starting fallback scan")
	start := time.Now()

	candidates, err := j.syncService.FindSyncCandidates(ctx)
	if err != nil {
		metrics.SyncJobTotal.WithLabelValues("error").Inc()
		logx.Errorf("sync_job: find candidates failed: %v", err)
		return
	}

	metrics.SyncJobTotal.WithLabelValues("success").Inc()
	logx.Infof("sync_job: found %d candidates, elapsed=%v", len(candidates), time.Since(start))

	for _, c := range candidates {
		logx.Infof("sync_job: candidate waybill=%s track_status=%d", c.SourceTrackingNumber, c.TrackStatus)
	}
}
