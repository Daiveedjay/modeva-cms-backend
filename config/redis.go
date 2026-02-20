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

	// CRITICAL DEBUG - Shows what Railway is actually injecting
	log.Printf("🔍 DEBUG: REDIS_URL = '%s'", redisURL)
	log.Printf("🔍 DEBUG: REDIS_URL length = %d", len(redisURL))
	log.Printf("🔍 DEBUG: REDIS_URL is empty = %v", redisURL == "")

	if redisURL == "" {
		// Default to local Redis for development
		redisURL = "redis://localhost:6379"
		log.Println("⚠️  REDIS_URL not set, using local Redis:", redisURL)
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(fmt.Sprintf("❌ invalid REDIS_URL: %v", err))
	}

	// DEBUG - Show parsed connection details
	log.Printf("🔍 DEBUG: Connecting to address: %s", opt.Addr)
	log.Printf("🔍 DEBUG: Using password: %s", opt.Password)

	RedisClient = redis.NewClient(opt)

	// test connection
	res, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("❌ failed to connect to Redis: %v", err))
	}
	fmt.Println("✅ Connected to Redis:", res)
}
