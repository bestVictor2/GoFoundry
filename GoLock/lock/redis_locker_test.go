package lock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newLocker(t *testing.T) (*RedisLocker, *redis.Client, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := NewRedisLocker(client, WithPrefix("test:"))
	return locker, client, mr
}

func TestTryLockConflictAndUnlock(t *testing.T) {
	locker, client, _ := newLocker(t)
	defer client.Close()

	ctx := context.Background()
	leaseA, err := locker.TryLock(ctx, "job:1", time.Second)
	if err != nil {
		t.Fatalf("try lock failed: %v", err)
	}

	_, err = locker.TryLock(ctx, "job:1", time.Second)
	if !errors.Is(err, ErrNotAcquired) {
		t.Fatalf("expected ErrNotAcquired, got %v", err)
	}

	if err = leaseA.Unlock(ctx); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}

	if _, err = locker.TryLock(ctx, "job:1", time.Second); err != nil {
		t.Fatalf("lock should be acquirable after unlock, got %v", err)
	}
}

func TestUnlockNotOwner(t *testing.T) {
	locker, client, _ := newLocker(t)
	defer client.Close()

	ctx := context.Background()
	leaseA, err := locker.TryLock(ctx, "job:2", time.Second)
	if err != nil {
		t.Fatalf("try lock failed: %v", err)
	}

	fakeLease := &Lease{
		key:    leaseA.key,
		token:  "other-owner",
		ttl:    leaseA.ttl,
		locker: locker,
	}

	if err = fakeLease.Unlock(ctx); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func TestRefreshExtendsTTL(t *testing.T) {
	locker, client, mr := newLocker(t)
	defer client.Close()

	ctx := context.Background()
	leaseA, err := locker.TryLock(ctx, "job:3", 150*time.Millisecond)
	if err != nil {
		t.Fatalf("try lock failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if err = leaseA.Refresh(ctx, 600*time.Millisecond); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	ttl := mr.TTL("test:job:3")
	if ttl < 300*time.Millisecond {
		t.Fatalf("ttl should be extended, got %v", ttl)
	}
}

func TestLockWaitUntilReleased(t *testing.T) {
	locker, client, _ := newLocker(t)
	defer client.Close()

	ctx := context.Background()
	leaseA, err := locker.TryLock(ctx, "job:4", time.Second)
	if err != nil {
		t.Fatalf("try lock failed: %v", err)
	}

	go func() {
		time.Sleep(120 * time.Millisecond)
		_ = leaseA.Unlock(context.Background())
	}()

	start := time.Now()
	leaseB, err := locker.Lock(ctx, "job:4", time.Second, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("lock with wait failed: %v", err)
	}
	defer leaseB.Unlock(ctx)

	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Fatalf("expected blocking wait, got %v", elapsed)
	}
}

func TestKeepAlive(t *testing.T) {
	locker, client, _ := newLocker(t)
	defer client.Close()

	ctx := context.Background()
	leaseA, err := locker.TryLock(ctx, "job:5", 120*time.Millisecond)
	if err != nil {
		t.Fatalf("try lock failed: %v", err)
	}

	keepCtx, cancel := context.WithCancel(context.Background())
	errCh := leaseA.KeepAlive(keepCtx, 40*time.Millisecond)
	time.Sleep(260 * time.Millisecond)
	cancel()

	select {
	case keepErr, ok := <-errCh:
		if ok && keepErr != nil {
			t.Fatalf("unexpected keepalive err: %v", keepErr)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("keepalive channel should close quickly")
	}

	if _, err = locker.TryLock(ctx, "job:5", time.Second); !errors.Is(err, ErrNotAcquired) {
		t.Fatalf("lock should still be alive after keepalive, got %v", err)
	}
}

func TestDo(t *testing.T) {
	locker, client, _ := newLocker(t)
	defer client.Close()

	ctx := context.Background()
	called := 0
	err := locker.Do(ctx, "job:do", time.Second, 200*time.Millisecond, func(ctx context.Context) error {
		called++
		return nil
	})
	if err != nil {
		t.Fatalf("do failed: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected callback called once, got %d", called)
	}
	if _, err = locker.TryLock(ctx, "job:do", time.Second); err != nil {
		t.Fatalf("lock should be released after Do, got %v", err)
	}
}
