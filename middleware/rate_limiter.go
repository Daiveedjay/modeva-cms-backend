package middleware

import (
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath() // /api/v1/categories, /api/v1/categories/:id, etc.
		method := c.Request.Method

		// Key is per-IP, per-method, per-endpoint
		key := "rl:" + ip + ":" + method + ":" + endpoint
		resetKey := key + ":resetAt"

		// Increment request count
		count, err := config.RedisClient.Incr(config.Ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Redis error"))
			c.Abort()
			return
		}

		// First request → set expiry and stable resetAt
		if count == 1 {
			config.RedisClient.Expire(config.Ctx, key, window)
			resetAt := time.Now().Add(window)
			config.RedisClient.Set(config.Ctx, resetKey, resetAt.Unix(), window)
		}

		// Get stable resetAt from Redis
		resetAtUnix, _ := config.RedisClient.Get(config.Ctx, resetKey).Int64()
		resetAt := time.Unix(resetAtUnix, 0)

		// Calculate remaining requests (clamped at 0)
		remaining := maxRequests - int(count)
		if remaining < 0 {
			remaining = 0
		}

		// Reset in seconds (clamped at 0)
		resetInSeconds := int(time.Until(resetAt).Seconds())
		if resetInSeconds < 0 {
			resetInSeconds = 0
		}

		rate := &models.RateLimiter{
			Limit:          maxRequests,
			Remaining:      remaining,
			ResetAt:        resetAt,
			ResetInSeconds: resetInSeconds,
		}

		// Store in context for controllers
		c.Set("rateLimiter", rate)

		// If limit exceeded → block request
		if int(count) > maxRequests {
			c.JSON(http.StatusTooManyRequests, models.ApiResponse{
				Message: "Too many requests",
				Error:   true,
				Rate:    rate,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
