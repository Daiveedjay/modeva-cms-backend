package customer_controller

import (
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

// SearchCustomers godoc
// @Summary Search customers (CMS)
// @Description Search customers with advanced filters: name, email, country, status, join date range/exact, and spending range/exact.
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 50)" default(5)
// @Param q query string false "Search by name"
// @Param email query string false "Search by email"
// @Param country query string false "Filter by country"
// @Param status query string false "Filter by status" Enums(active,suspended,deleted,banned)
// @Param joined_from query string false "Joined from date (YYYY-MM-DD)"
// @Param joined_to query string false "Joined to date (YYYY-MM-DD)"
// @Param spending_exact query number false "Exact spending amount"
// @Param spending_min query number false "Minimum spending amount"
// @Param spending_max query number false "Maximum spending amount"
// @Success 200 {object} models.ApiResponse{data=[]models.CMSCustomerListRow,meta=models.Pagination}
// @Failure 400 {object} models.ApiResponse "Bad request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /admin/customers [get]
func SearchCustomers(c *gin.Context) {
	log.Printf("[admin.search-customers] start path=%s method=%s rawQuery=%s",
		c.FullPath(), c.Request.Method, c.Request.URL.RawQuery)

	// ================================
	// Pagination
	// ================================
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 5
	}
	offset := (page - 1) * limit

	// ================================
	// Filters
	// ================================
	name := strings.TrimSpace(c.Query("q"))
	email := strings.TrimSpace(c.Query("email"))
	country := strings.TrimSpace(c.Query("country"))
	status := strings.TrimSpace(strings.ToLower(c.Query("status")))
	joinedFrom := strings.TrimSpace(c.Query("joined_from"))
	joinedTo := strings.TrimSpace(c.Query("joined_to"))
	spendingExact := c.Query("spending_exact")
	spendingMin := c.Query("spending_min")
	spendingMax := c.Query("spending_max")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// ================================
	// Validate date formats
	// ================================
	var parsedJoinedFrom, parsedJoinedTo *time.Time

	if joinedFrom != "" {
		if t, err := time.Parse("2006-01-02", joinedFrom); err == nil {
			parsedJoinedFrom = &t
		} else {
			log.Printf("[admin.search-customers] ERROR invalid joined_from date=%s err=%v", joinedFrom, err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid joined_from date format (use YYYY-MM-DD)"))
			return
		}
	}

	if joinedTo != "" {
		if t, err := time.Parse("2006-01-02", joinedTo); err == nil {
			parsedJoinedTo = &t
		} else {
			log.Printf("[admin.search-customers] ERROR invalid joined_to date=%s err=%v", joinedTo, err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid joined_to date format (use YYYY-MM-DD)"))
			return
		}
	}

	// ================================
	// Validate spending values
	// ================================
	var parsedSpendingExact, parsedSpendingMin, parsedSpendingMax *float64

	if spendingExact != "" {
		val := 0.0
		if _, err := strconv.ParseFloat(spendingExact, 64); err == nil {
			parsed, _ := strconv.ParseFloat(spendingExact, 64)
			val = parsed
			parsedSpendingExact = &val
		} else {
			log.Printf("[admin.search-customers] ERROR invalid spending_exact=%s err=%v", spendingExact, err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid spending_exact value"))
			return
		}
	}

	if spendingMin != "" {
		val := 0.0
		if _, err := strconv.ParseFloat(spendingMin, 64); err == nil {
			parsed, _ := strconv.ParseFloat(spendingMin, 64)
			val = parsed
			parsedSpendingMin = &val
		}
	}

	if spendingMax != "" {
		val := 0.0
		if _, err := strconv.ParseFloat(spendingMax, 64); err == nil {
			parsed, _ := strconv.ParseFloat(spendingMax, 64)
			val = parsed
			parsedSpendingMax = &val
		}
	}

	// ================================
	// Validate status
	// ================================
	if status != "" {
		switch status {
		case "active", "suspended", "banned", "deleted":
			// Valid
		default:
			log.Printf("[admin.search-customers] ERROR invalid status=%s", status)
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid status"))
			return
		}
	}

	// ================================
	// Count query with filters
	// ================================
	countSQL := `SELECT COUNT(*) FROM users u`
	countArgs := []interface{}{}

	countConditions := []string{}

	if name != "" {
		countConditions = append(countConditions, "(u.name ILIKE ?)")
		countArgs = append(countArgs, "%"+name+"%")
	}

	if email != "" {
		countConditions = append(countConditions, "(u.email ILIKE ?)")
		countArgs = append(countArgs, "%"+email+"%")
	}

	if country != "" {
		countConditions = append(countConditions, `
			EXISTS (
				SELECT 1 FROM addresses a 
				WHERE a.user_id = u.id AND a.is_default = true AND a.status = 'active' 
				AND a.country ILIKE ?
			)
		`)
		countArgs = append(countArgs, "%"+country+"%")
	}

	if status != "" {
		countConditions = append(countConditions, "LOWER(u.status) = ?")
		countArgs = append(countArgs, status)
	}

	if parsedJoinedFrom != nil {
		countConditions = append(countConditions, "u.created_at >= ?")
		countArgs = append(countArgs, parsedJoinedFrom)
	}

	if parsedJoinedTo != nil {
		countConditions = append(countConditions, "u.created_at <= ?")
		countArgs = append(countArgs, parsedJoinedTo.Add(24*time.Hour))
	}

	if parsedSpendingExact != nil {
		countConditions = append(countConditions, `
			EXISTS (
				SELECT 1 FROM (
					SELECT user_id, COALESCE(SUM(total_amount), 0)::float8 AS total
					FROM orders WHERE status = 'completed'
					GROUP BY user_id
				) os WHERE os.user_id = u.id AND os.total = ?
			)
		`)
		countArgs = append(countArgs, *parsedSpendingExact)
	}

	if parsedSpendingMin != nil {
		countConditions = append(countConditions, `
			EXISTS (
				SELECT 1 FROM (
					SELECT user_id, COALESCE(SUM(total_amount), 0)::float8 AS total
					FROM orders WHERE status = 'completed'
					GROUP BY user_id
				) os WHERE os.user_id = u.id AND os.total >= ?
			)
		`)
		countArgs = append(countArgs, *parsedSpendingMin)
	}

	if parsedSpendingMax != nil {
		countConditions = append(countConditions, `
			EXISTS (
				SELECT 1 FROM (
					SELECT user_id, COALESCE(SUM(total_amount), 0)::float8 AS total
					FROM orders WHERE status = 'completed'
					GROUP BY user_id
				) os WHERE os.user_id = u.id AND os.total <= ?
			)
		`)
		countArgs = append(countArgs, *parsedSpendingMax)
	}

	if len(countConditions) > 0 {
		countSQL += " WHERE " + strings.Join(countConditions, " AND ")
	}

	var total int64
	if err := config.EcommerceGorm.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&total).Error; err != nil {
		log.Printf("[admin.search-customers] ERROR count failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch customers"))
		return
	}

	// ================================
	// Data query with CTEs and filters
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

	// Rebuild WHERE clause for the data query
	dataConditions := []string{}
	dataArgs := []interface{}{}

	if name != "" {
		dataConditions = append(dataConditions, "(u.name ILIKE ?)")
		dataArgs = append(dataArgs, "%"+name+"%")
	}

	if email != "" {
		dataConditions = append(dataConditions, "(u.email ILIKE ?)")
		dataArgs = append(dataArgs, "%"+email+"%")
	}

	if country != "" {
		dataConditions = append(dataConditions, `
			EXISTS (
				SELECT 1 FROM addresses a 
				WHERE a.user_id = u.id AND a.is_default = true AND a.status = 'active' 
				AND a.country ILIKE ?
			)
		`)
		dataArgs = append(dataArgs, "%"+country+"%")
	}

	if status != "" {
		dataConditions = append(dataConditions, "LOWER(u.status) = ?")
		dataArgs = append(dataArgs, status)
	}

	if parsedJoinedFrom != nil {
		dataConditions = append(dataConditions, "u.created_at >= ?")
		dataArgs = append(dataArgs, parsedJoinedFrom)
	}

	if parsedJoinedTo != nil {
		dataConditions = append(dataConditions, "u.created_at <= ?")
		dataArgs = append(dataArgs, parsedJoinedTo.Add(24*time.Hour))
	}

	if parsedSpendingExact != nil {
		dataConditions = append(dataConditions, `
			EXISTS (
				SELECT 1 FROM (
					SELECT user_id, COALESCE(SUM(total_amount), 0)::float8 AS total
					FROM orders WHERE status = 'completed'
					GROUP BY user_id
				) os WHERE os.user_id = u.id AND os.total = ?
			)
		`)
		dataArgs = append(dataArgs, *parsedSpendingExact)
	}

	if parsedSpendingMin != nil {
		dataConditions = append(dataConditions, `
			EXISTS (
				SELECT 1 FROM (
					SELECT user_id, COALESCE(SUM(total_amount), 0)::float8 AS total
					FROM orders WHERE status = 'completed'
					GROUP BY user_id
				) os WHERE os.user_id = u.id AND os.total >= ?
			)
		`)
		dataArgs = append(dataArgs, *parsedSpendingMin)
	}

	if parsedSpendingMax != nil {
		dataConditions = append(dataConditions, `
			EXISTS (
				SELECT 1 FROM (
					SELECT user_id, COALESCE(SUM(total_amount), 0)::float8 AS total
					FROM orders WHERE status = 'completed'
					GROUP BY user_id
				) os WHERE os.user_id = u.id AND os.total <= ?
			)
		`)
		dataArgs = append(dataArgs, *parsedSpendingMax)
	}

	if len(dataConditions) > 0 {
		dataSQL += " WHERE " + strings.Join(dataConditions, " AND ")
	}

	dataSQL += " ORDER BY u.created_at DESC LIMIT ? OFFSET ?"
	dataArgs = append(dataArgs, limit, offset)

	if err := config.EcommerceGorm.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&out).Error; err != nil {
		log.Printf("[admin.search-customers] ERROR data query failed err=%v", err)
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

	log.Printf("[admin.search-customers] respond 200 total=%d page=%d", total, page)

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Customers retrieved successfully",
		out,
		meta,
	))
}
