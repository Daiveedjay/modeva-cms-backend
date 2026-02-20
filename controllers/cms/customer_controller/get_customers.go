package customer_controller

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

// GetCustomers godoc
// @Summary Get customers (CMS)
// @Description Fetch customers for CMS table view. Includes location, orders count, total spent, and activity status.
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 50)" default(10)
// @Param q query string false "Search by name or email"
// @Param status query string false "Filter by status" Enums(active,suspended,deleted,banned)
// @Success 200 {object} models.ApiResponse{data=[]models.CMSCustomerListRow,meta=models.Pagination}
// @Failure 400 {object} models.ApiResponse
// @Failure 401 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /admin/customers [get]
func GetCustomers(c *gin.Context) {
	log.Printf("[admin.customers] start path=%s method=%s rawQuery=%s",
		c.FullPath(), c.Request.Method, c.Request.URL.RawQuery)

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

	// ================================
	// Filters
	// ================================
	q := strings.TrimSpace(c.Query("q"))
	status := strings.TrimSpace(strings.ToLower(c.Query("status")))

	ctx, cancel := config.WithTimeout()
	defer cancel()

	db := config.EcommerceGorm.WithContext(ctx).Table("users u")

	// Apply search filter
	if q != "" {
		db = db.Where("u.name ILIKE ? OR u.email ILIKE ?", "%"+q+"%", "%"+q+"%")
	}

	// Apply status filter
	if status != "" {
		switch status {
		case "active", "suspended", "deleted", "banned":
			db = db.Where("LOWER(u.status) = ?", status)
		default:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid status"))
			return
		}
	}

	// ================================
	// Count
	// ================================
	var total int64
	if err := db.Count(&total).Error; err != nil {
		log.Printf("[admin.customers] ERROR count failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch customers"))
		return
	}

	// ================================
	// Data query with CTEs
	// ================================
	var out []models.CMSCustomerListRow

	dataSQL := `
		WITH default_addr AS (
			SELECT DISTINCT ON (a.user_id)
				a.user_id,
				NULLIF(a.city, '')    AS city,
				NULLIF(a.country, '') AS country
			FROM addresses a
			WHERE a.is_default = true AND a.status = 'active'
			ORDER BY a.user_id, a.updated_at DESC, a.created_at DESC
		),
		order_summary AS (
			SELECT
				user_id,
				COUNT(id)::int AS order_count,
				COALESCE(SUM(total_amount), 0)::float8 AS total_amount,
				MAX(created_at) AS last_order_date
			FROM orders
			WHERE status = 'completed'
			GROUP BY user_id
		)
		SELECT
			u.id::text AS id,
			u.name,
			u.email,
			CASE
				WHEN da.user_id IS NULL THEN 'No address yet'
				WHEN da.city IS NOT NULL AND da.country IS NOT NULL THEN da.city || ', ' || da.country
				WHEN da.city IS NOT NULL THEN da.city
				WHEN da.country IS NOT NULL THEN da.country
				ELSE 'No address yet'
			END AS location,
			COALESCE(os.order_count, 0)::int AS orders,
			COALESCE(os.total_amount, 0)::float8 AS total_spent,
			CASE
				WHEN os.last_order_date IS NULL THEN 'inactive'
				WHEN NOW() - os.last_order_date > INTERVAL '12 hours' THEN 'inactive'
				ELSE 'active'
			END AS activity,
			u.status,
			u.created_at AS join_date,
			u.avatar,
			u.ban_reason,
			u.suspended_until,
			u.suspended_reason
		FROM users u
		LEFT JOIN default_addr da ON da.user_id = u.id
		LEFT JOIN order_summary os ON os.user_id = u.id
	`

	// Rebuild WHERE clause for the CTE query
	whereConditions := []string{}
	whereArgs := []interface{}{}

	if q != "" {
		whereConditions = append(whereConditions, "(u.name ILIKE ? OR u.email ILIKE ?)")
		whereArgs = append(whereArgs, "%"+q+"%", "%"+q+"%")
	}

	if status != "" {
		whereConditions = append(whereConditions, "LOWER(u.status) = ?")
		whereArgs = append(whereArgs, status)
	}

	if len(whereConditions) > 0 {
		dataSQL += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	dataSQL += " ORDER BY u.created_at DESC LIMIT ? OFFSET ?"
	whereArgs = append(whereArgs, limit, offset)

	if err := config.EcommerceGorm.WithContext(ctx).Raw(dataSQL, whereArgs...).Scan(&out).Error; err != nil {
		log.Printf("[admin.customers] ERROR data query failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch customers"))
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

	log.Printf("[admin.customers] respond 200 total=%d page=%d", total, page)

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Customers retrieved successfully",
		out,
		meta,
	))
}
