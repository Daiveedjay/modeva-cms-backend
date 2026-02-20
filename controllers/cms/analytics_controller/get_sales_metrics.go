package analytics_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetSalesMetrics godoc
// @Summary Get sales metrics
// @Description Returns sales metrics: average order value, customer lifetime value, return customer rate
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=models.SalesMetrics}
// @Failure 500 {object} models.ApiResponse
// @Router /admin/analytics/metrics [get]
func GetSalesMetrics(c *gin.Context) {
	log.Printf("[admin.analytics-metrics] start")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// ================================
	// Average Order Value (AOV)
	// ================================
	var avgOrderValue float64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ?", "completed").
		Select("COALESCE(AVG(total_amount), 0)").
		Scan(&avgOrderValue).Error; err != nil {
		log.Printf("[admin.analytics-metrics] ERROR avg order value err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch sales metrics"))
		return
	}

	// ================================
	// Customer Lifetime Value (CLV)
	// Sum of all orders per user, then average across all users
	// ================================
	var clv float64
	if err := config.EcommerceGorm.WithContext(ctx).
		Raw(`
			SELECT COALESCE(AVG(user_total), 0)::float8
			FROM (
				SELECT user_id, COALESCE(SUM(total_amount), 0) AS user_total
				FROM orders
				WHERE status = ?
				GROUP BY user_id
			) user_totals
		`, "completed").
		Scan(&clv).Error; err != nil {
		log.Printf("[admin.analytics-metrics] ERROR customer lifetime value err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch sales metrics"))
		return
	}

	// ================================
	// Return Customer Rate (RCR)
	// Percentage of users with 2 or more orders
	// ================================
	type CustomerCounts struct {
		TotalCustomers  int64
		RepeatCustomers int64
	}

	var counts CustomerCounts
	if err := config.EcommerceGorm.WithContext(ctx).
		Raw(`
			SELECT 
				COUNT(DISTINCT user_id) AS total_customers,
				COUNT(DISTINCT CASE WHEN order_count >= 2 THEN user_id END) AS repeat_customers
			FROM (
				SELECT user_id, COUNT(*) AS order_count
				FROM orders
				WHERE status = ?
				GROUP BY user_id
			) user_orders
		`, "completed").
		Scan(&counts).Error; err != nil {
		log.Printf("[admin.analytics-metrics] ERROR return customer rate err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch sales metrics"))
		return
	}

	// Calculate return customer rate percentage
	returnCustomerRate := 0.0
	if counts.TotalCustomers > 0 {
		returnCustomerRate = (float64(counts.RepeatCustomers) / float64(counts.TotalCustomers)) * 100
	}

	// ================================
	// Build Response
	// ================================
	metrics := models.SalesMetrics{
		AverageOrderValue:     avgOrderValue,
		CustomerLifetimeValue: clv,
		ReturnCustomerRate:    returnCustomerRate,
	}

	log.Printf("[admin.analytics-metrics] respond 200 aov=%.2f clv=%.2f rcr=%.1f%%",
		avgOrderValue, clv, returnCustomerRate)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Sales metrics retrieved successfully", metrics))
}
