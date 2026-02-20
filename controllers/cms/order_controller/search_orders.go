package order_controller

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

func parseTimeFlexible(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t, nil
	}

	// Try date only (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("invalid date format (expected RFC3339 or YYYY-MM-DD): %q", s)
}

// SearchOrders godoc
// @Summary Search orders (CMS)
// @Description Search orders by customer name/email, order number, status, price (exact/range), date range. Supports pagination.
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string false "Generic search (matches order number, customer name, email)"
// @Param order_number query string false "Order number (partial match)"
// @Param customer query string false "Customer name (partial match)"
// @Param email query string false "Customer email (partial match)"
// @Param status query string false "Status (pending|processing|shipped|completed|cancelled)"
// @Param price query number false "Exact total amount"
// @Param min_price query number false "Min total amount"
// @Param max_price query number false "Max total amount"
// @Param created_from query string false "Created from (RFC3339 or YYYY-MM-DD)"
// @Param created_to query string false "Created to (RFC3339 or YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 50)" default(10)
// @Success 200 {object} models.ApiResponse{data=object{orders=[]models.CMSOrderListRow},meta=models.Pagination}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /admin/orders/search [get]
func SearchOrders(c *gin.Context) {
	log.Printf("[admin.orders.search] start path=%s method=%s rawQuery=%s", c.FullPath(), c.Request.Method, c.Request.URL.RawQuery)

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Filters
	qTerm := strings.TrimSpace(c.Query("q"))
	orderNumber := strings.TrimSpace(c.Query("order_number"))
	customer := strings.TrimSpace(c.Query("customer"))
	email := strings.TrimSpace(c.Query("email"))
	status := strings.TrimSpace(strings.ToLower(c.Query("status")))

	var price *float64
	if s := strings.TrimSpace(c.Query("price")); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			price = &v
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid price"))
			return
		}
	}

	var minPrice *float64
	if s := strings.TrimSpace(c.Query("min_price")); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			minPrice = &v
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid min_price"))
			return
		}
	}

	var maxPrice *float64
	if s := strings.TrimSpace(c.Query("max_price")); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			maxPrice = &v
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid max_price"))
			return
		}
	}

	createdFromRaw := strings.TrimSpace(c.Query("created_from"))
	createdToRaw := strings.TrimSpace(c.Query("created_to"))
	createdFrom, err := parseTimeFlexible(createdFromRaw)
	if err != nil && createdFromRaw != "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid created_from (use RFC3339 or YYYY-MM-DD)"))
		return
	}
	createdTo, err := parseTimeFlexible(createdToRaw)
	if err != nil && createdToRaw != "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid created_to (use RFC3339 or YYYY-MM-DD)"))
		return
	}

	// If created_to is date-only, make it inclusive (end of day)
	if createdTo != nil && len(createdToRaw) == len("2006-01-02") {
		t := createdTo.Add(24*time.Hour - time.Nanosecond)
		createdTo = &t
	}

	log.Printf("[admin.orders.search] params page=%d limit=%d offset=%d q=%q order_number=%q customer=%q email=%q status=%q price=%v min_price=%v max_price=%v created_from=%v created_to=%v",
		page, limit, offset, qTerm, orderNumber, customer, email, status, price, minPrice, maxPrice, createdFrom, createdTo)

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Build WHERE conditions
	whereConditions := []string{}
	whereArgs := []interface{}{}

	// Generic search q matches order_number OR customer name OR email
	if qTerm != "" {
		whereConditions = append(whereConditions, "(o.order_number ILIKE ? OR u.name ILIKE ? OR u.email ILIKE ?)")
		like := "%" + qTerm + "%"
		whereArgs = append(whereArgs, like, like, like)
	}

	if orderNumber != "" {
		whereConditions = append(whereConditions, "o.order_number ILIKE ?")
		whereArgs = append(whereArgs, "%"+orderNumber+"%")
	}

	if customer != "" {
		whereConditions = append(whereConditions, "u.name ILIKE ?")
		whereArgs = append(whereArgs, "%"+customer+"%")
	}

	if email != "" {
		whereConditions = append(whereConditions, "u.email ILIKE ?")
		whereArgs = append(whereArgs, "%"+email+"%")
	}

	if status != "" {
		whereConditions = append(whereConditions, "o.status = ?")
		whereArgs = append(whereArgs, status)
	}

	if price != nil {
		whereConditions = append(whereConditions, "o.total_amount = ?")
		whereArgs = append(whereArgs, *price)
	}

	if minPrice != nil {
		whereConditions = append(whereConditions, "o.total_amount >= ?")
		whereArgs = append(whereArgs, *minPrice)
	}

	if maxPrice != nil {
		whereConditions = append(whereConditions, "o.total_amount <= ?")
		whereArgs = append(whereArgs, *maxPrice)
	}

	if createdFrom != nil {
		whereConditions = append(whereConditions, "o.created_at >= ?")
		whereArgs = append(whereArgs, *createdFrom)
	}

	if createdTo != nil {
		whereConditions = append(whereConditions, "o.created_at <= ?")
		whereArgs = append(whereArgs, *createdTo)
	}

	whereSQL := "1=1"
	if len(whereConditions) > 0 {
		whereSQL = strings.Join(whereConditions, " AND ")
	}

	log.Printf("[admin.orders.search] whereSQL=%s args=%v", whereSQL, whereArgs)

	// Count
	countSQL := `
		SELECT COUNT(DISTINCT o.id)
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE ` + whereSQL

	log.Printf("[admin.orders.search] countSQL=%s", strings.ReplaceAll(countSQL, "\n", " "))

	var total int64
	if err := config.EcommerceGorm.WithContext(ctx).Raw(countSQL, whereArgs...).Scan(&total).Error; err != nil {
		log.Printf("[admin.orders.search] ERROR count query failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch orders"))
		return
	}

	// Data query
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
		WHERE ` + whereSQL + `
		GROUP BY o.id, o.order_number, u.id, u.name, u.email, o.created_at, o.total_amount, o.status
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`

	dataArgs := append(whereArgs, limit, offset)
	log.Printf("[admin.orders.search] dataSQL=%s", strings.ReplaceAll(dataSQL, "\n", " "))
	log.Printf("[admin.orders.search] dataArgs=%v", dataArgs)

	out := make([]models.CMSOrderListRow, 0)
	if err := config.EcommerceGorm.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&out).Error; err != nil {
		log.Printf("[admin.orders.search] ERROR data query failed err=%v", err)
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

	log.Printf("[admin.orders.search] done returned=%d total=%d page=%d/%d", len(out), total, page, totalPages)

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Orders retrieved successfully",
		out,
		meta,
	))
}
