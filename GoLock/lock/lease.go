package lock

import (
	"context"
	"time"
)

type Lease struct {
	key    string
	token  string
	ttl    time.Duration
	locker *RedisLocker
}

type LeaseInfo struct {
	Key   string        `json:"key"`
	Token string        `json:"token"`
	TTL   time.Duration `json:"ttl"`
}

func (l *Lease) Key() string {
	if l == nil {
		return ""
	}
	return l.key
}

func (l *Lease) Token() string {
	if l == nil {
		return ""
	}
	return l.token
}

func (l *Lease) Info() LeaseInfo {
	return LeaseInfo{Key: l.Key(), Token: l.Token(), TTL: l.ttl}
}

func (l *Lease) Unlock(ctx context.Context) error {
	if err := l.validate(); err != nil {
		return err
	}

	key := l.locker.formatKey(l.key)
	n, err := unlockScript.Run(ctx, l.locker.client, []string{key}, l.token).Int64()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotOwner
	}
	return nil
}

func (l *Lease) Refresh(ctx context.Context, ttl time.Duration) error {
	if err := l.validate(); err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = l.ttl
	}
	if ttl <= 0 {
		return ErrTTLInvalid
	}

	key := l.locker.formatKey(l.key)
	n, err := refreshScript.Run(ctx, l.locker.client, []string{key}, l.token, ttl.Milliseconds()).Int64()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotOwner
	}
	l.ttl = ttl
	return nil
}

func (l *Lease) KeepAlive(ctx context.Context, interval time.Duration) <-chan error {
	errCh := make(chan error, 1)
	if err := l.validate(); err != nil {
		errCh <- err
		close(errCh)
		return errCh
	}

	if interval <= 0 {
		interval = l.ttl / 3
	}
	if interval <= 0 {
		errCh <- ErrTTLInvalid
		close(errCh)
		return errCh
	}

	go func() {
		defer close(errCh)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := l.Refresh(ctx, l.ttl); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return errCh
}

func (l *Lease) validate() error {
	if l == nil || l.locker == nil {
		return ErrLockerNil
	}
	if l.token == "" {
		return ErrTokenMissing
	}
	return nil
}
