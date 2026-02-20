package admin_controller

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetAllAdminActivityLogs godoc
// @Summary Get all admin activities
// @Description Get activity logs for all admins with pagination
// @Tags Admin - Management
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param admin_id query string false "Filter by admin ID"
// @Param action query string false "Filter by action (e.g., created_product, updated_order)"
// @Success 200 {object} models.ApiResponse{data=map[string]interface{}}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Router /admin/activity-logs [get]
func GetAllAdminActivityLogs(c *gin.Context) {
	log.Printf("[admin.all-activity] request for all activities")

	// Pagination
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 100 {
				parsed = 100 // Max 100 items per page
			}
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Build base query
	baseQuery := config.CmsGorm.WithContext(ctx)

	// Optional filters
	if adminID := c.Query("admin_id"); adminID != "" {
		baseQuery = baseQuery.Where("admin_id = ?", adminID)
	}

	if action := c.Query("action"); action != "" {
		baseQuery = baseQuery.Where("action = ?", action)
	}

	// Get activity logs
	var activityLogs []models.ActivityLog
	var total int64

	if err := baseQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activityLogs).Error; err != nil {
		log.Printf("[admin.all-activity] failed to fetch logs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Get total count
	if err := baseQuery.Model(&models.ActivityLog{}).Count(&total).Error; err != nil {
		log.Printf("[admin.all-activity] failed to count logs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Convert to response objects
	responses := make([]models.ActivityLogResponse, len(activityLogs))
	for i, log := range activityLogs {
		responses[i] = log.ToResponse()
	}

	// Prepare pagination meta
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	response := gin.H{
		"logs": responses,
	}

	log.Printf("[admin.all-activity] retrieved %d logs (page %d/%d, total: %d)", len(responses), page, totalPages, total)
	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Activity logs retrieved", response, meta))
}
