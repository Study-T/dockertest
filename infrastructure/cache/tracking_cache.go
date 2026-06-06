package cache

import (
	"context"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/repo"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const trackingCachePrefix = "tracking:detail:"
const defaultTrackingTTL = 24 * time.Hour

// TrackingCacheImpl implements repo.TrackingCache.
type TrackingCacheImpl struct {
	redis  *redis.Redis
	prefix string
}

// NewTrackingCacheImpl creates a new tracking cache instance.
func NewTrackingCacheImpl(redisClient *redis.Redis) *TrackingCacheImpl {
	return &TrackingCacheImpl{
		redis:  redisClient,
		prefix: trackingCachePrefix,
	}
}

var _ repo.TrackingCache = (*TrackingCacheImpl)(nil)

func (c *TrackingCacheImpl) GetTrackingDetail(ctx context.Context, orderNumber string) (string, error) {
	key := c.buildKey(orderNumber)
	val, err := c.redis.GetCtx(ctx, key)
	if err != nil {
		return "", fmt.Errorf("cache get: %w", err)
	}
	return val, nil
}

func (c *TrackingCacheImpl) SetTrackingDetail(ctx context.Context, orderNumber, data string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = defaultTrackingTTL
	}
	key := c.buildKey(orderNumber)
	if err := c.redis.SetexCtx(ctx, key, data, int(ttl.Seconds())); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

func (c *TrackingCacheImpl) DeleteTrackingDetail(ctx context.Context, orderNumber string) error {
	key := c.buildKey(orderNumber)
	if _, err := c.redis.DelCtx(ctx, key); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}
	return nil
}

func (c *TrackingCacheImpl) buildKey(orderNumber string) string {
	return fmt.Sprintf("%s%s", c.prefix, orderNumber)
}
