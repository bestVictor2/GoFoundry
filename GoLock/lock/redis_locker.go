package lock

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Locker interface {
	TryLock(ctx context.Context, key string, ttl time.Duration) (*Lease, error)
	Lock(ctx context.Context, key string, ttl, wait time.Duration) (*Lease, error)
	Do(ctx context.Context, key string, ttl, wait time.Duration, fn func(context.Context) error) error
}

type RedisLocker struct {
	client        redis.UniversalClient // 单机 redis 哨兵 cluster
	prefix        string
	retryInterval time.Duration
	tokenGen      TokenGenerator
}

func NewRedisLocker(client redis.UniversalClient, opts ...Option) *RedisLocker {
	if client == nil {
		panic("redis client is nil")
	}

	locker := &RedisLocker{
		client:        client,
		prefix:        defaultKeyPrefix,
		retryInterval: defaultRetryInterval,
		tokenGen:      NewRandomTokenGenerator(defaultTokenLength),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(locker)
		}
	}
	return locker
}

func (l *RedisLocker) TryLock(ctx context.Context, key string, ttl time.Duration) (*Lease, error) {
	if err := validateLockArgs(key, ttl); err != nil {
		return nil, err
	} // 参数校验的地方

	token, err := l.tokenGen.NextToken()
	if err != nil {
		return nil, err
	}

	ok, err := l.client.SetNX(ctx, l.formatKey(key), token, ttl).Result() // 原子化
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotAcquired
	} // 未抢到锁

	return &Lease{
		key:    key,
		token:  token,
		ttl:    ttl,
		locker: l,
	}, nil
}

func (l *RedisLocker) Lock(ctx context.Context, key string, ttl, wait time.Duration) (*Lease, error) {
	if wait <= 0 {
		return l.TryLock(ctx, key, ttl)
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()
	ticker := time.NewTicker(l.retryInterval)
	defer ticker.Stop()

	for {
		lease, err := l.TryLock(ctx, key, ttl)
		if err == nil {
			return lease, nil
		}
		if !errors.Is(err, ErrNotAcquired) {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, ErrNotAcquired // 时间耗尽
		case <-ticker.C: //到达重试时间
		}
	}
}

func (l *RedisLocker) Do(ctx context.Context, key string, ttl, wait time.Duration, fn func(context.Context) error) error {
	if fn == nil {
		return nil
	}
	lease, err := l.Lock(ctx, key, ttl, wait)
	if err != nil {
		return err
	}
	defer func() {
		_ = lease.Unlock(context.Background())
	}()
	return fn(ctx)
}

func (l *RedisLocker) formatKey(key string) string {
	return l.prefix + key
}

func validateLockArgs(key string, ttl time.Duration) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if ttl <= 0 {
		return ErrTTLInvalid
	}
	return nil
}
