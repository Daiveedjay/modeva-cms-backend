package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func ConnectRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		panic("❌ REDIS_ADDR is not set")
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test connection
	res, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		panic("❌ Failed to connect to Redis at " + addr + ": " + err.Error())
	}
	fmt.Println("✅ Connected to Redis:", res)
}
