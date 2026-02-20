package customer_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetCustomerStats godoc
// @Summary Get customer statistics
// @Description Returns stats: total customers, new customers this month, active customers, avg order value, and comparisons
// @Tags Admin - Customers
// @Produce json
// @Success 200 {object} models.ApiResponse{data=models.CustomerStats}
// @Failure 500 {object} models.ApiResponse
// @Router /admin/customers/stats [get]
func GetCustomerStats(c *gin.Context) {
	log.Printf("[admin.customer-stats] start")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// ================================
	// Current Month Stats
	// ================================
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Total customers (all time)
	var totalCustomers int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.User{}).
		Where("status = ?", "active").
		Count(&totalCustomers).Error; err != nil {
		log.Printf("[admin.customer-stats] ERROR total customers count err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count total customers"))
		return
	}

	// New customers this month
	var newCustomersThisMonth int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.User{}).
		Where("status = ? AND created_at >= ?", "active", monthStart).
		Count(&newCustomersThisMonth).Error; err != nil {
		log.Printf("[admin.customer-stats] ERROR new customers count err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count new customers"))
		return
	}

	// New customers last month
	lastMonthStart := monthStart.AddDate(0, -1, 0)
	var newCustomersLastMonth int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.User{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", "active", lastMonthStart, monthStart).
		Count(&newCustomersLastMonth).Error; err != nil {
		log.Printf("[admin.customer-stats] ERROR last month customers count err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count last month customers"))
		return
	}

	// Calculate growth percentage
	var growthPercentage float64
	if newCustomersLastMonth == 0 {
		if newCustomersThisMonth > 0 {
			growthPercentage = 100 // 100% growth if no customers last month but some this month
		}
	} else {
		growthPercentage = ((float64(newCustomersThisMonth) - float64(newCustomersLastMonth)) / float64(newCustomersLastMonth)) * 100
	}

	// Active customers (inactive if no order in last 90 days)
	ninetyDaysAgo := now.AddDate(0, 0, -90)
	var activeCustomers int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.User{}).
		Where("status = ? AND EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id AND orders.created_at >= ?)", "active", ninetyDaysAgo).
		Count(&activeCustomers).Error; err != nil {
		log.Printf("[admin.customer-stats] ERROR active customers count err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count active customers"))
		return
	}

	// Calculate active percentage
	activePercentage := 0.0
	if totalCustomers > 0 {
		activePercentage = (float64(activeCustomers) / float64(totalCustomers)) * 100
	}

	// Average order value per customer
	var avgOrderValue float64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ?", "completed").
		Select("COALESCE(AVG(total_amount), 0)").
		Scan(&avgOrderValue).Error; err != nil {
		log.Printf("[admin.customer-stats] ERROR avg order value err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to calculate average order value"))
		return
	}

	// ================================
	// Build Response
	// ================================
	stats := models.CustomerStats{
		TotalCustomers:               int(totalCustomers),
		NewCustomersThisMonth:        int(newCustomersThisMonth),
		NewCustomersGrowthPercentage: growthPercentage,
		ActiveCustomers:              int(activeCustomers),
		ActiveCustomersPercentage:    activePercentage,
		AvgOrderValue:                avgOrderValue,
	}

	log.Printf("[admin.customer-stats] respond 200 total=%d new_this_month=%d active=%d avg_order=%.2f",
		stats.TotalCustomers, stats.NewCustomersThisMonth, stats.ActiveCustomers, stats.AvgOrderValue)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Customer stats fetched successfully", stats))
}
