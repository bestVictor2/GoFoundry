package lock

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func BenchmarkTryLockUnlock(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("start miniredis failed: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	locker := NewRedisLocker(client, WithPrefix("bench:"))
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lease, lockErr := locker.TryLock(ctx, "bench-key", time.Second)
		if lockErr != nil {
			b.Fatalf("try lock failed: %v", lockErr)
		}
		if unlockErr := lease.Unlock(ctx); unlockErr != nil {
			b.Fatalf("unlock failed: %v", unlockErr)
		}
	}
}
