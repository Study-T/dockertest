package service

import (
	"context"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	syncBatchSize    = 100
	cacheHours       = 4
	syncLogPrefix    = "sync_job"
)

// SyncService handles fallback synchronization for missed webhook events.
type SyncService struct {
	detailRepo repo.TrackingDetailRepo
	logRepo    repo.TrackingLogRepo
}

func NewSyncService(detailRepo repo.TrackingDetailRepo, logRepo repo.TrackingLogRepo) *SyncService {
	return &SyncService{detailRepo: detailRepo, logRepo: logRepo}
}

// SyncResult holds the outcome of a sync run.
type SyncResult struct {
	Scanned  int
	Synced   int
	Skipped  int
	Errors   int
}

// RunSync finds tracking_logs that need fallback sync and returns candidates.
// Actual OMS API pull is handled by the caller (cron job) since it needs HTTP client.
func (s *SyncService) FindSyncCandidates(ctx context.Context) ([]*entity.TrackingLog, error) {
	// Find tracking_logs that:
	// 1. source_tracking_number is not empty
	// 2. fulfill_at is not empty (order has been fulfilled)
	// 3. track_status != delivered (2)
	// 4. tracking_detail not updated within CACHE_HOURS
	logs, err := s.logRepo.FindSyncCandidates(ctx, syncBatchSize, cacheHours)
	if err != nil {
		return nil, err
	}
	logx.Infof("%s: found %d sync candidates", syncLogPrefix, len(logs))
	return logs, nil
}

// MarkSynced updates the tracking_log after successful sync.
func (s *SyncService) MarkSynced(ctx context.Context, logID int64) error {
	now := time.Now()
	return s.logRepo.UpdateTrackingFields(ctx, logID, entity.TrackStatusInTransit, now, nil, nil, nil)
}
