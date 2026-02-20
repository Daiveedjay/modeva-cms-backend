package admin_controller

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// SearchAdminActivityLogs godoc
// @Summary Search admin activities with filters
// @Description Search and filter activity logs by query, admin email, action, status, resource type, and date range
// @Tags Admin - Activity Logs
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param query query string false "Search by query (admin name, resource name, keywords)"
// @Param admin_email query string false "Filter by admin email"
// @Param action query string false "Filter by action (e.g., created, updated, deleted)"
// @Param status query string false "Filter by status (success, failed)"
// @Param resource_type query string false "Filter by resource type (category, product, order, admin, customer)"
// @Param created_from query string false "Filter from date (YYYY-MM-DD)"
// @Param created_to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} models.ApiResponse{data=map[string]interface{}}
// @Failure 400 {object} models.ApiResponse "Bad request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /admin/activity-logs [get]
func SearchAdminActivityLogs(c *gin.Context) {
	log.Printf("[admin.search-activity] search request")

	// ===== Pagination =====
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

	// ===== Filters =====
	query := c.Query("query")                // Free text search
	adminEmail := c.Query("admin_email")     // Exact/partial match on admin email
	action := c.Query("action")              // Exact match on action
	status := c.Query("status")              // Exact match on status (success/failed)
	resourceType := c.Query("resource_type") // Exact match on resource type
	createdFrom := c.Query("created_from")   // Date filter (from)
	createdTo := c.Query("created_to")       // Date filter (to)

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Build query
	dbQuery := config.CmsGorm.WithContext(ctx)

	// Free text search - search across resource_name and description
	if query != "" {
		dbQuery = dbQuery.Where(
			"resource_name ILIKE ? OR admin_email ILIKE ?",
			"%"+query+"%",
			"%"+query+"%",
		)
	}

	// Admin email filter
	if adminEmail != "" {
		dbQuery = dbQuery.Where("admin_email ILIKE ?", "%"+adminEmail+"%")
	}

	// Action filter
	if action != "" && action != "all" {
		if action == "created" {
			dbQuery = dbQuery.Where("action LIKE ?", "%created%")
		} else if action == "updated" {
			dbQuery = dbQuery.Where("action LIKE ?", "%updated%")
		} else if action == "deleted" {
			dbQuery = dbQuery.Where("action LIKE ?", "%deleted%")
		} else {
			dbQuery = dbQuery.Where("action = ?", action)
		}
	}

	// Status filter
	if status != "" && status != "all" {
		dbQuery = dbQuery.Where("status = ?", status)
	}

	// Resource type filter
	if resourceType != "" && resourceType != "all" {
		dbQuery = dbQuery.Where("resource_type = ?", resourceType)
	}

	// Date range filters
	if createdFrom != "" {
		// Parse date string (YYYY-MM-DD)
		fromDate, err := time.Parse("2006-01-02", createdFrom)
		if err == nil {
			dbQuery = dbQuery.Where("created_at >= ?", fromDate)
		}
	}

	if createdTo != "" {
		// Parse date string (YYYY-MM-DD) and add 1 day to include entire day
		toDate, err := time.Parse("2006-01-02", createdTo)
		if err == nil {
			toDate = toDate.AddDate(0, 0, 1) // Add 1 day
			dbQuery = dbQuery.Where("created_at < ?", toDate)
		}
	}

	// Get activity logs
	var activityLogs []models.ActivityLog
	var total int64

	if err := dbQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activityLogs).Error; err != nil {
		log.Printf("[admin.search-activity] failed to fetch logs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Get total count (before pagination)
	if err := dbQuery.Model(&models.ActivityLog{}).Count(&total).Error; err != nil {
		log.Printf("[admin.search-activity] failed to count logs: %v", err)
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

	log.Printf(
		"[admin.search-activity] search completed: query=%s, admin_email=%s, action=%s, status=%s, resource_type=%s, result_count=%d, page=%d/%d, total=%d",
		query,
		adminEmail,
		action,
		status,
		resourceType,
		len(responses),
		page,
		totalPages,
		total,
	)

	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Activity logs retrieved", response, meta))
}
