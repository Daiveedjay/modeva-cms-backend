package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func ConnectRedis() {
	// read Redis URL
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Default to local Redis for development
		redisURL = "redis://localhost:6379"
		log.Println("⚠️  REDIS_URL not set, using local Redis:", redisURL)
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(fmt.Sprintf("❌ invalid REDIS_URL: %v", err))
	}

	RedisClient = redis.NewClient(opt)

	// test connection
	res, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("❌ failed to connect to Redis: %v", err))
	}
	fmt.Println("✅ Connected to Redis:", res)
}
