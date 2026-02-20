package admin_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetAdminStats godoc
// @Summary Get admin dashboard statistics
// @Description Get stats about admins, sessions, activities, and system health
// @Tags Admin - Stats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=AdminStats}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /admin/stats [get]
func GetAdminStats(c *gin.Context) {
	log.Printf("[admin.stats] request")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Get total admins count
	var totalAdmins int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Admin{}).
		Count(&totalAdmins).Error; err != nil {
		log.Printf("[admin.stats] failed to count total admins: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch stats"))
		return
	}

	// Get active admins count (status = "active")
	var activeAdmins int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Admin{}).
		Where("status = ?", "active").
		Count(&activeAdmins).Error; err != nil {
		log.Printf("[admin.stats] failed to count active admins: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch stats"))
		return
	}

	// Get active sessions count (sessions that are active and not expired, AND admin is not suspended)
	var activeSessions int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.AdminSession{}).
		Joins("JOIN admins ON admin_sessions.admin_id = admins.id").
		Where("admin_sessions.is_active = ? AND admin_sessions.expires_at > ? AND admins.status != ?", true, time.Now(), "suspended").
		Count(&activeSessions).Error; err != nil {
		log.Printf("[admin.stats] failed to count active sessions: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch stats"))
		return
	}

	// Get daily actions count (activity logs created today)
	startOfDay := time.Now().Truncate(24 * time.Hour)
	var dailyActions int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.ActivityLog{}).
		Where("created_at >= ?", startOfDay).
		Count(&dailyActions).Error; err != nil {
		log.Printf("[admin.stats] failed to count daily actions: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch stats"))
		return
	}

	stats := AdminStats{
		TotalAdmins:    int(totalAdmins),
		ActiveAdmins:   int(activeAdmins),
		ActiveSessions: int(activeSessions),
		DailyActions:   int(dailyActions),
		SystemStatus:   "Healthy",
	}

	log.Printf("[admin.stats] retrieved: total=%d, active=%d, sessions=%d, actions=%d",
		stats.TotalAdmins, stats.ActiveAdmins, stats.ActiveSessions, stats.DailyActions)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Admin stats retrieved", stats))
}

// AdminStats represents admin dashboard statistics
type AdminStats struct {
	TotalAdmins    int    `json:"total_admins"`
	ActiveAdmins   int    `json:"active_admins"`
	ActiveSessions int    `json:"active_sessions"`
	DailyActions   int    `json:"daily_actions"`
	SystemStatus   string `json:"system_status"` // "Healthy", "Degraded", "Unhealthy"
}
