package admin_controller

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
)

// GetAdmins godoc
// @Summary List all admins
// @Description Get list of all admins with their status (paginated)
// @Tags Admin - Management
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} models.ApiResponse{data=map[string]interface{}}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Router /admin/admins [get]
func GetAdmins(c *gin.Context) {
	log.Printf("[admin.list] request")

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

	// Base query
	baseQuery := config.CmsGorm.WithContext(ctx).Model(&models.Admin{})

	// Total count
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		log.Printf("[admin.list] failed to count admins: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Fetch admins
	var admins []models.Admin
	if err := baseQuery.
		Order("joined_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&admins).Error; err != nil {
		log.Printf("[admin.list] database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Convert to response objects and calculate status
	authService := services.GetAdminAuthService()
	responses := make([]models.AdminResponse, len(admins))
	for i, admin := range admins {
		admin.Status = authService.GetAdminStatus(admin.Status, admin.LastLoginAt)
		responses[i] = admin.ToResponse()
	}

	// Prepare pagination meta
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	log.Printf("[admin.list] retrieved %d admins (page %d/%d, total: %d)", len(responses), page, totalPages, total)

	// Keep response shape consistent with your other paginated endpoint
	log.Printf("[admin.list] meta: %+v", meta)

	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Admins retrieved", responses, meta))
}
