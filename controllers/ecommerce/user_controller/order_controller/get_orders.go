package order_controller

import (
	"math"
	"net/http"
	"strconv"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetOrders godoc
// @Summary Get order history
// @Description Retrieve all orders for the authenticated user with pagination
// @Tags User - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 50)" default(10)
// @Success 200 {object} models.ApiResponse{data=[]models.OrderHistoryResponse,meta=models.Pagination}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/orders [get]
func GetOrders(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized"))
		return
	}

	// Parse userID to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Invalid user ID"))
		return
	}

	// Parse query params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Context with timeout
	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Count total orders for this user
	var total int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Table("orders").
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch orders"))
		return
	}

	// Fetch paginated orders with item count
	var orders []models.OrderHistoryResponse
	err = config.EcommerceGorm.WithContext(ctx).Raw(`
		SELECT 
			o.id::text AS id,
			o.order_number,
			o.status,
			o.total_amount,
			o.created_at,
			COUNT(oi.id)::int AS item_count
		FROM orders o
		LEFT JOIN order_items oi ON o.id = oi.order_id
		WHERE o.user_id = ?
		GROUP BY o.id, o.order_number, o.status, o.total_amount, o.created_at
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset).Scan(&orders).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch orders"))
		return
	}

	// Pagination meta
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Orders retrieved successfully",
		orders,
		meta,
	))
}
