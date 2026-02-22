package main

import (
	"GoLock/lock"
	"context"
	"flag"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	var redisAddr string
	var key string
	var hold time.Duration
	flag.StringVar(&redisAddr, "redis", "localhost:6379", "redis address")
	flag.StringVar(&key, "key", "demo-lock", "lock key")
	flag.DurationVar(&hold, "hold", 5*time.Second, "how long to hold the lock")
	flag.Parse()

	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer client.Close()

	locker := lock.NewRedisLocker(client)
	ctx := context.Background()

	lk, err := locker.Lock(ctx, key, 10*time.Second, 3*time.Second)
	if err != nil {
		log.Fatalf("acquire lock failed: %v", err)
	}
	defer func() {
		if err := lk.Unlock(context.Background()); err != nil {
			log.Printf("unlock failed: %v", err)
		}
	}()

	log.Printf("lock acquired key=%s token=%s", lk.Key(), lk.Token())
	time.Sleep(hold)
	log.Printf("lock released key=%s", lk.Key())
}
