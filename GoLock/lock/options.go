package lock

import "time"

const (
	defaultKeyPrefix     = "gofoundry:lock:"
	defaultRetryInterval = 50 * time.Millisecond
	defaultTokenLength   = 16
)

type Option func(*RedisLocker)

func WithPrefix(prefix string) Option {
	return func(l *RedisLocker) {
		if prefix != "" {
			l.prefix = prefix
		}
	}
}

func WithRetryInterval(interval time.Duration) Option {
	return func(l *RedisLocker) {
		if interval > 0 {
			l.retryInterval = interval
		}
	}
}

func WithTokenGenerator(gen TokenGenerator) Option {
	return func(l *RedisLocker) {
		if gen != nil {
			l.tokenGen = gen
		}
	}
}
