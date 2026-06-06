package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

var ErrLockHeld = errors.New("lock already held")

type RedisLock struct {
	redis *redis.Redis
}

func NewRedisLock(rds *redis.Redis) *RedisLock {
	return &RedisLock{redis: rds}
}

func (l *RedisLock) Lock(ctx context.Context, key string, ttl time.Duration) (string, error) {
	sec := int(ttl.Seconds())
	if sec < 1 {
		sec = 1
	}
	token := generateToken()
	ok, err := l.redis.SetnxExCtx(ctx, key, token, sec)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrLockHeld
	}
	return token, nil
}

var ErrNotLockOwner = errors.New("not lock owner")

func (l *RedisLock) Unlock(ctx context.Context, key, token string) error {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	val, err := l.redis.EvalCtx(ctx, script, []string{key}, token)
	if err != nil {
		return err
	}
	if val == "0" {
		return ErrNotLockOwner
	}
	return nil
}

func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}
