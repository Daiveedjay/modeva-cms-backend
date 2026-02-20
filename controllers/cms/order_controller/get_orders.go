package order_controller

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetOrders godoc
// @Summary Get orders (CMS)
// @Description Retrieve all orders for CMS (admin) with customer details and pagination. Supports optional filtering by status and search.
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 50)" default(10)
// @Param status query string false "Filter by order status (pending, confirmed, processing, shipped, delivered, cancelled, refunded)"
// @Param q query string false "Search by order number, customer email, or customer name"
// @Success 200 {object} models.ApiResponse{data=[]models.CMSOrderListRow,meta=models.Pagination}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /admin/orders [get]
func GetOrders(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		log.Printf("[admin.orders] WARN invalid page=%q err=%v -> default 1", c.Query("page"), err)
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		log.Printf("[admin.orders] WARN invalid limit=%q err=%v -> default 10", c.Query("limit"), err)
		limit = 10
	}

	if page < 1 {
		log.Printf("[admin.orders] WARN page < 1 (%d) -> set 1", page)
		page = 1
	}
	if limit < 1 || limit > 50 {
		log.Printf("[admin.orders] WARN limit out of range (%d) -> set 10", limit)
		limit = 10
	}
	offset := (page - 1) * limit

	status := strings.TrimSpace(c.Query("status"))
	q := strings.TrimSpace(c.Query("q"))

	log.Printf("[admin.orders] params page=%d limit=%d offset=%d status=%q q=%q", page, limit, offset, status, q)

	db := config.EcommerceGorm.Table("orders o").
		Joins("LEFT JOIN users u ON u.id = o.user_id")

	// Apply filters
	if status != "" {
		db = db.Where("o.status = ?", status)
		log.Printf("[admin.orders] filter status=%q", status)
	}

	if q != "" {
		like := "%" + q + "%"
		db = db.Where("o.order_number ILIKE ? OR u.email ILIKE ? OR u.name ILIKE ?", like, like, like)
		log.Printf("[admin.orders] filter q=%q", like)
	}

	// Count total orders
	var total int64
	if err := db.Count(&total).Error; err != nil {
		log.Printf("[admin.orders] ERROR count failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count orders"))
		return
	}

	log.Printf("[admin.orders] count OK total=%d", total)

	// Fetch orders with aggregates
	dataSQL := `
		SELECT
			o.id::text AS id,
			o.order_number,
			u.id::text AS customer_id,
			COALESCE(NULLIF(u.name, ''), u.email) AS customer_name,
			u.email AS customer_email,
			o.created_at,
			COUNT(oi.id)::int AS item_count,
			COALESCE(SUM(oi.quantity), 0)::int AS total_quantity,
			o.total_amount,
			o.status
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		LEFT JOIN order_items oi ON oi.order_id = o.id
	`

	// Build WHERE clause
	whereConditions := []string{}
	whereArgs := []interface{}{}

	if status != "" {
		whereConditions = append(whereConditions, "o.status = ?")
		whereArgs = append(whereArgs, status)
	}

	if q != "" {
		like := "%" + q + "%"
		whereConditions = append(whereConditions, "(o.order_number ILIKE ? OR u.email ILIKE ? OR u.name ILIKE ?)")
		whereArgs = append(whereArgs, like, like, like)
	}

	if len(whereConditions) > 0 {
		dataSQL += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	dataSQL += `
		GROUP BY o.id, o.order_number, u.id, u.name, u.email, o.created_at, o.total_amount, o.status
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`

	whereArgs = append(whereArgs, limit, offset)

	log.Printf("[admin.orders] dataSQL=%s", strings.ReplaceAll(dataSQL, "\n", " "))
	log.Printf("[admin.orders] dataArgs=%v", whereArgs)

	result := make([]models.CMSOrderListRow, 0, limit)

	if err := config.EcommerceGorm.Raw(dataSQL, whereArgs...).Scan(&result).Error; err != nil {
		log.Printf("[admin.orders] ERROR data query failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch orders"))
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	log.Printf("[admin.orders] respond 200 meta=%+v", *meta)

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Orders retrieved successfully",
		result,
		meta,
	))
}
