package models

import (
	"time"

	"github.com/gin-gonic/gin"
)

type ApiResponse struct {
	Message         string       `json:"message"`
	Data            any          `json:"data,omitempty"`
	Error           bool         `json:"error,omitempty"`
	Meta            *Pagination  `json:"meta"`
	Rate            *RateLimiter `json:"rate_limit,omitempty"`
	RequestedEntity string       `json:"requested_entity,omitempty"`
}

type Pagination struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	Total      int `json:"total" example:"42"`
	TotalPages int `json:"total_pages" example:"5"`
}

type RateLimiter struct {
	Limit          int       `json:"limit"`
	Remaining      int       `json:"remaining"`
	ResetAt        time.Time `json:"reset_at"`
	ResetInSeconds int       `json:"reset_in_seconds"`
}

// helper to fetch rate limiter info from Gin context
func getRateFromContext(c *gin.Context) *RateLimiter {
	if c == nil {
		return nil
	}
	if rate, exists := c.Get("rateLimiter"); exists {
		if rl, ok := rate.(*RateLimiter); ok {
			return rl
		}
	}
	return nil
}

func SuccessResponse(c *gin.Context, message string, data any) ApiResponse {
	return ApiResponse{
		Message:         message,
		Data:            data,
		Rate:            getRateFromContext(c),
		RequestedEntity: c.Request.Method + " " + c.FullPath(),
	}
}

func PaginatedResponse(c *gin.Context, message string, data any, meta *Pagination) ApiResponse {
	return ApiResponse{
		Message:         message,
		Data:            data,
		Meta:            meta,
		Rate:            getRateFromContext(c),
		RequestedEntity: c.Request.Method + " " + c.FullPath(),
	}
}

func ErrorResponse(c *gin.Context, message string) ApiResponse {
	return ApiResponse{
		Message:         message,
		Error:           true,
		Rate:            getRateFromContext(c),
		RequestedEntity: c.Request.Method + " " + c.FullPath(),
	}
}
