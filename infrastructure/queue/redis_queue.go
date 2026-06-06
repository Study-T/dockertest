package queue

import (
	"context"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/repo"

	"github.com/redis/go-redis/v9"
)

const (
	defaultQueueKey = "queue:yun_express_webhook_track"
	defaultTimeout  = 0 // 0 = permanent block
)

// QueueConfig holds Redis queue configuration.
type QueueConfig struct {
	Key     string
	Timeout int64
}

// RedisQueue implements repo.QueueRepo using Redis BRPOP.
type RedisQueue struct {
	client *redis.Client
	config QueueConfig
}

// NewRedisQueue creates a new Redis queue instance.
func NewRedisQueue(addr, password string, db int, config QueueConfig) *RedisQueue {
	if config.Key == "" {
		config.Key = defaultQueueKey
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisQueue{
		client: client,
		config: config,
	}
}

var _ repo.QueueRepo = (*RedisQueue)(nil)

func (q *RedisQueue) BRPop(ctx context.Context, timeout int64) (*repo.QueueMessage, error) {
	if timeout <= 0 {
		timeout = q.config.Timeout
	}

	val, err := q.client.BRPop(ctx, time.Duration(timeout)*time.Second, q.config.Key).Result()
	if err != nil {
		return nil, fmt.Errorf("brpop failed: %w", err)
	}

	if len(val) < 2 {
		return nil, fmt.Errorf("invalid brpop response")
	}

	return &repo.QueueMessage{
		RawData: []byte(val[1]),
	}, nil
}

func (q *RedisQueue) LLen(ctx context.Context) (int64, error) {
	length, err := q.client.LLen(ctx, q.config.Key).Result()
	if err != nil {
		return 0, fmt.Errorf("llen failed: %w", err)
	}
	return length, nil
}
