package cache

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ConfigCache struct {
	redis  *redis.Redis
	prefix string
}

func NewConfigCache(redisClient *redis.Redis) *ConfigCache {
	return &ConfigCache{
		redis:  redisClient,
		prefix: "system:config:",
	}
}

func (c *ConfigCache) Get(ctx context.Context, key string) (string, error) {
	return c.redis.GetCtx(ctx, c.prefix+key)
}

func (c *ConfigCache) Set(ctx context.Context, key, value string, ttl int) error {
	return c.redis.SetexCtx(ctx, c.prefix+key, value, ttl)
}

func (c *ConfigCache) Delete(ctx context.Context, key string) error {
	_, err := c.redis.DelCtx(ctx, c.prefix+key)
	return err
}

func (c *ConfigCache) HealthCheck(ctx context.Context) bool {
	return c.redis.PingCtx(ctx)
}

func (c *ConfigCache) BuildKey(key string) string {
	return fmt.Sprintf("%s%s", c.prefix, key)
}
