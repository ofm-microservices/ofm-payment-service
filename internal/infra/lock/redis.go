package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const releaseScript = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`

type redisLock struct {
	client *redis.Client
	ttl    time.Duration
	wait   time.Duration
}

// NewRedisConnectAccountLock creates a Redis-backed per-user onboarding lock.
func NewRedisConnectAccountLock(client *redis.Client) (*redisLock, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return &redisLock{client: client, ttl: 30 * time.Second, wait: 10 * time.Second}, nil
}

func (l *redisLock) Acquire(ctx context.Context, userID string) (func(), error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is empty")
	}
	key := "payment:connect-onboarding:" + userID
	token := uuid.NewString()
	deadline := time.NewTimer(l.wait)
	defer deadline.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		ok, err := l.client.SetNX(ctx, key, token, l.ttl).Result()
		if err != nil {
			return nil, fmt.Errorf("acquire connect onboarding lock: %w", err)
		}
		if ok {
			return func() {
				_ = l.client.Eval(context.Background(), releaseScript, []string{key}, token).Err()
			}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline.C:
			return nil, fmt.Errorf("connect onboarding lock timeout for user %s", userID)
		case <-ticker.C:
		}
	}
}
