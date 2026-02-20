package admin_controller

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetSingleAdminActivityLogs godoc
// @Summary Get admin activity logs
// @Description Get activity logs for a specific admin with pagination
// @Tags Admin - Management
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} models.ApiResponse{data=map[string]interface{}}
// @Failure 404 {object} models.ApiResponse "Admin not found"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Router /admin/admins/:id/activity [get]
func GetSingleAdminActivityLogs(c *gin.Context) {
	adminID := c.Param("id")
	log.Printf("[admin.activity] request for admin: %s", adminID)

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

	// Verify admin exists
	var admin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("id = ?", adminID).
		First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Admin not found"))
		} else {
			log.Printf("[admin.activity] database error: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		}
		return
	}

	// Get activity logs
	var activityLogs []models.ActivityLog
	var total int64

	if err := config.CmsGorm.WithContext(ctx).
		Where("admin_id = ?", adminID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activityLogs).Error; err != nil {
		log.Printf("[admin.activity] failed to fetch logs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Get total count
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.ActivityLog{}).
		Where("admin_id = ?", adminID).
		Count(&total).Error; err != nil {
		log.Printf("[admin.activity] failed to count logs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Convert to response objects
	responses := make([]models.ActivityLogResponse, len(activityLogs))
	for i, log := range activityLogs {
		responses[i] = log.ToResponse()
	}

	// Step 6: Prepare pagination meta
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	// ✅ Include admin details in response
	response := gin.H{
		"admin": gin.H{
			"id":     admin.ID.String(),
			"name":   admin.Name,
			"email":  admin.Email,
			"avatar": admin.Avatar,
			"role":   admin.Role,
		},
		"logs": responses,
	}

	log.Printf("[admin.activity] retrieved %d logs for admin %s (page %d/%d)", len(responses), adminID, page, totalPages)
	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Activity logs retrieved", response, meta))
}
