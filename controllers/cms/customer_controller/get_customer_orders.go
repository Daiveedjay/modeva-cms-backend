package customer_controller

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCustomerOrders godoc
// @Summary Get customer orders (CMS)
// @Description Fetch all orders for a specific customer. Includes all statuses. Supports pagination.
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 50)" default(10)
// @Success 200 {object} models.ApiResponse{data=[]models.CMSCustomerOrderRow,meta=models.Pagination}
// @Failure 400 {object} models.ApiResponse
// @Failure 401 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /admin/customers/{id}/orders [get]
func GetCustomerOrders(c *gin.Context) {
	log.Printf("[admin.customer-orders] start path=%s method=%s rawQuery=%s",
		c.FullPath(), c.Request.Method, c.Request.URL.RawQuery)

	// ================================
	// Path param
	// ================================
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid customer ID"))
		return
	}

	// ================================
	// Pagination
	// ================================
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// ================================
	// Ensure customer exists
	// ================================
	var exists bool
	if err := config.EcommerceGorm.WithContext(ctx).
		Table("users").
		Select("count(*) > 0").
		Where("id = ?", customerID).
		Find(&exists).Error; err != nil || !exists {

		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Customer not found"))
		return
	}

	// ================================
	// Count
	// ================================
	var total int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Table("orders").
		Where("user_id = ?", customerID).
		Count(&total).Error; err != nil {

		log.Printf("[admin.customer-orders] ERROR count failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch orders"))
		return
	}

	// ================================
	// Data query
	// ================================
	var out []models.CMSCustomerOrderRow

	dataSQL := `
	SELECT
		o.id::text AS id,
		o.order_number,
		o.status,
		o.total_amount,
		o.created_at,
		u.name AS customer_name
	FROM orders o
	JOIN users u ON u.id = o.user_id
	WHERE o.user_id = ?
	ORDER BY o.created_at DESC
	LIMIT ? OFFSET ?
`

	if err := config.EcommerceGorm.WithContext(ctx).
		Raw(dataSQL, customerID, limit, offset).
		Scan(&out).Error; err != nil {

		log.Printf("[admin.customer-orders] ERROR data query failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch orders"))
		return
	}

	// ================================
	// Meta
	// ================================
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	log.Printf("[admin.customer-orders] respond 200 customer=%s total=%d page=%d",
		customerID, total, page)

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Customer orders retrieved successfully",
		out,
		meta,
	))
}
