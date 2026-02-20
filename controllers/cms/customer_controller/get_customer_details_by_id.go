package customer_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCustomerDetailsByID godoc
// @Summary Get customer profile details
// @Description Fetch detailed profile information for a specific customer including orders, addresses, and payment methods
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Success 200 {object} models.ApiResponse{data=models.CMSCustomerDetail}
// @Failure 400 {object} models.ApiResponse "Bad request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 404 {object} models.ApiResponse "Customer not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /admin/customers/{id} [get]
func GetCustomerDetailsByID(c *gin.Context) {
	log.Printf("[admin.customer-details] start path=%s method=%s", c.FullPath(), c.Request.Method)

	customerIDStr := c.Param("id")
	if customerIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Customer ID is required"))
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid customer ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// =================================
	// Step 1: Fetch customer with order stats
	// =================================
	var customer models.CMSCustomerDetail

	querySQL := `
		WITH order_summary AS (
			SELECT
				user_id,
				COUNT(id)::int AS order_count,
				COALESCE(SUM(total_amount), 0)::float8 AS total_amount,
				COALESCE(AVG(total_amount), 0)::float8 AS avg_amount,
				MAX(created_at) AS last_order_date
			FROM orders
			WHERE status = 'completed'
			GROUP BY user_id
		)
		SELECT
			u.id::text AS id,
			u.name,
			u.email,
			u.phone,
			u.avatar,
			u.status,
			u.created_at AS join_date,
			u.ban_reason,
			u.suspended_until,
			u.suspended_reason,
			COALESCE(os.order_count, 0)::int AS orders,
			COALESCE(os.total_amount, 0)::float8 AS total_spent,
			COALESCE(os.avg_amount, 0)::float8 AS avg_order_value,
			os.last_order_date
		FROM users u
		LEFT JOIN order_summary os ON os.user_id = u.id
		WHERE u.id = ?
	`

	result := config.EcommerceGorm.WithContext(ctx).Raw(querySQL, customerID).Scan(&customer)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch customer"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Customer not found"))
		return
	}

	// =================================
	// Step 2: Fetch default address
	// =================================
	type AddressResult struct {
		ID      string
		Phone   *string
		Street  *string
		City    *string
		State   *string
		Zip     *string
		Country *string
	}

	var addrResult AddressResult

	addressSQL := `
		SELECT DISTINCT ON (a.user_id)
			a.id::text AS id,
			a.phone,
			a.street,
			NULLIF(a.city, '') AS city,
			a.state,
			a.zip,
			NULLIF(a.country, '') AS country
		FROM addresses a
		WHERE a.user_id = ?
			AND a.is_default = true
			AND a.status = 'active'
		ORDER BY a.user_id, a.updated_at DESC, a.created_at DESC
		LIMIT 1
	`

	addrResultQuery := config.EcommerceGorm.WithContext(ctx).
		Raw(addressSQL, customerID).
		Scan(&addrResult)

	if addrResultQuery.Error == nil && addrResultQuery.RowsAffected > 0 {
		customer.Address = &models.CustomerAddress{
			ID:      addrResult.ID,
			Phone:   addrResult.Phone,
			Street:  addrResult.Street,
			City:    addrResult.City,
			State:   addrResult.State,
			Zip:     addrResult.Zip,
			Country: addrResult.Country,
		}

		// Build location
		city := ""
		country := ""

		if addrResult.City != nil {
			city = *addrResult.City
		}
		if addrResult.Country != nil {
			country = *addrResult.Country
		}

		if city != "" && country != "" {
			customer.Location = city + ", " + country
		} else if city != "" {
			customer.Location = city
		} else if country != "" {
			customer.Location = country
		} else {
			customer.Location = "No address yet"
		}
	} else {
		customer.Address = nil
		customer.Location = "No address yet"
	}

	// =================================
	// Step 3: Recent orders (unchanged)
	// =================================
	var recentOrders []models.CustomerOrder
	err = config.EcommerceGorm.WithContext(ctx).
		Table("orders").
		Select("order_number, id::text AS id, total_amount, created_at, status").
		Where("user_id = ? AND status = ?", customerID, "completed").
		Order("created_at DESC").
		Limit(3).
		Find(&recentOrders).Error

	if err != nil {
		customer.RecentOrders = []models.CustomerOrder{}
	} else {
		customer.RecentOrders = recentOrders
	}

	// Remaining steps unchanged …

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Customer details retrieved successfully",
		customer,
	))
}
