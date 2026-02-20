// ════════════════════════════════════════════════════════════
// Path: utils/login_tracker.go
// Track user login events
// ════════════════════════════════════════════════════════════

package utils

import (
	"log"
	"net"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogLoginEvent records a login event to the database
func LogLoginEvent(c *gin.Context, userID uuid.UUID) error {
	ctx := c.Request.Context()

	// Get IP address
	ipAddress := c.ClientIP()

	// Get user agent
	userAgent := c.GetHeader("User-Agent")

	// Parse device info (basic)
	deviceType := parseDeviceType(userAgent)
	browser := parseBrowser(userAgent)
	os := parseOS(userAgent)

	query := `
		INSERT INTO login_events (
			id, user_id, logged_in_at, ip_address, user_agent,
			device_type, browser, os
		) VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7)
	`

	_, err := config.EcommerceDB.Exec(ctx, query,
		uuid.New().String(),
		userID.String(), // Convert UUID to string for database
		ipAddress,
		userAgent,
		deviceType,
		browser,
		os,
	)
	if err != nil {
		log.Printf("❌ Failed to log login event: %v", err)
		return err
	}

	log.Printf("✅ Login event logged for user: %s from IP: %s", userID.String(), ipAddress)
	return nil
}

// parseDeviceType determines if the request is from mobile, tablet, or desktop
func parseDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") {
		return "mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	return "desktop"
}

// parseBrowser extracts browser name from user agent
func parseBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "edg") {
		return "Edge"
	}
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		return "Chrome"
	}
	if strings.Contains(ua, "firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Safari"
	}
	return "Other"
}

// parseOS extracts operating system from user agent
func parseOS(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "windows") {
		return "Windows"
	}
	if strings.Contains(ua, "mac os") {
		return "macOS"
	}
	if strings.Contains(ua, "linux") {
		return "Linux"
	}
	if strings.Contains(ua, "android") {
		return "Android"
	}
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		return "iOS"
	}
	return "Other"
}

// GetClientIP gets the real client IP (handles proxies)
func GetClientIP(c *gin.Context) string {
	// Try X-Forwarded-For first (if behind proxy)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Try X-Real-IP
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}
