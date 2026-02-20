package analytics_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetTopProducts godoc
// @Summary Get top performing products
// @Description Returns top 6 best selling products this month with sales count, revenue, and revenue percentage
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=[]models.TopProduct}
// @Failure 500 {object} models.ApiResponse
// @Router /admin/analytics/top-products [get]
func GetTopProducts(c *gin.Context) {
	log.Printf("[admin.analytics-top-products] start")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// ================================
	// Get total revenue this month (for percentage calculation)
	// ================================
	var totalRevenue float64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ? AND created_at >= ?", "completed", monthStart).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenue).Error; err != nil {
		log.Printf("[admin.analytics-top-products] ERROR total revenue err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch top products"))
		return
	}

	// ================================
	// Get top 6 products this month
	// ================================
	var topProducts []models.TopProduct
	if err := config.EcommerceGorm.WithContext(ctx).
		Raw(`
			SELECT 
				oi.product_id::text AS product_id,
				oi.product_name,
				COUNT(DISTINCT oi.order_id) AS order_count,
				SUM(oi.quantity) AS sales_count,
				SUM(oi.subtotal)::float8 AS revenue
			FROM order_items oi
			INNER JOIN orders o ON oi.order_id = o.id
			WHERE o.status = ? AND o.created_at >= ?
			GROUP BY oi.product_id, oi.product_name
			ORDER BY revenue DESC
			LIMIT 6
		`, "completed", monthStart).
		Scan(&topProducts).Error; err != nil {
		log.Printf("[admin.analytics-top-products] ERROR query top products err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch top products"))
		return
	}

	// ================================
	// Calculate revenue percentage for each product
	// ================================
	for i := range topProducts {
		if totalRevenue > 0 {
			topProducts[i].RevenuePercent = (topProducts[i].Revenue / totalRevenue) * 100
		} else {
			topProducts[i].RevenuePercent = 0
		}
	}

	log.Printf("[admin.analytics-top-products] respond 200 products=%d total_revenue=%.2f",
		len(topProducts), totalRevenue)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Top products retrieved successfully", topProducts))
}
