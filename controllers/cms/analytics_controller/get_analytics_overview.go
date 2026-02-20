package analytics_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetAnalyticsOverview godoc
// @Summary Get analytics overview
// @Description Returns overview stats: total revenue, orders, inventory, active customers with month-over-month comparisons
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=models.AnalyticsOverview}
// @Failure 500 {object} models.ApiResponse
// @Router /admin/analytics/overview [get]
func GetAnalyticsOverview(c *gin.Context) {
	log.Printf("[admin.analytics-overview] start")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := monthStart.AddDate(0, -1, 0)

	// ================================
	// Total Revenue (Current Month)
	// ================================
	var currentMonthRevenue float64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ? AND created_at >= ?", "completed", monthStart).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&currentMonthRevenue).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR current month revenue err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch analytics"))
		return
	}

	// ================================
	// Total Revenue (Last Month)
	// ================================
	var lastMonthRevenue float64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", "completed", lastMonthStart, monthStart).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&lastMonthRevenue).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR last month revenue err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch analytics"))
		return
	}

	// Calculate revenue growth percentage
	revenueGrowthPercent := 0.0
	if lastMonthRevenue > 0 {
		revenueGrowthPercent = ((currentMonthRevenue - lastMonthRevenue) / lastMonthRevenue) * 100
	} else if currentMonthRevenue > 0 {
		revenueGrowthPercent = 100.0
	}

	// ================================
	// Orders Count (Current Month)
	// ================================
	var currentMonthOrders int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ? AND created_at >= ?", "completed", monthStart).
		Count(&currentMonthOrders).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR current month orders err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch analytics"))
		return
	}

	// ================================
	// Orders Count (Last Month)
	// ================================
	var lastMonthOrders int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", "completed", lastMonthStart, monthStart).
		Count(&lastMonthOrders).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR last month orders err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch analytics"))
		return
	}

	// Calculate orders growth percentage
	ordersGrowthPercent := 0.0
	if lastMonthOrders > 0 {
		ordersGrowthPercent = ((float64(currentMonthOrders) - float64(lastMonthOrders)) / float64(lastMonthOrders)) * 100
	} else if currentMonthOrders > 0 {
		ordersGrowthPercent = 100.0
	}

	// ================================
	// Inventory Count (sum of all quantities in inventory JSONB)
	// ================================
	var currentInventory int64
	if err := config.CmsGorm.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(CAST(elem->>'quantity' AS INTEGER)), 0)
			FROM products, LATERAL jsonb_array_elements(inventory) AS elem
			WHERE status = ?
		`, "Active").
		Scan(&currentInventory).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR current inventory err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch analytics"))
		return
	}

	// ================================
	// Inventory Count (Last 30 days - approximation)
	// ================================
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	var lastMonthInventory int64
	// Since we don't have historical inventory, we'll estimate using products updated > 30 days ago
	// In production, you'd want to track inventory_snapshots or use logs
	if err := config.CmsGorm.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(CAST(elem->>'quantity' AS INTEGER)), 0)
			FROM products, LATERAL jsonb_array_elements(inventory) AS elem
			WHERE status = ? AND updated_at < ?
		`, "Active", thirtyDaysAgo).
		Scan(&lastMonthInventory).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR last month inventory err=%v", err)
		// Don't fail, just use current as fallback
		lastMonthInventory = currentInventory
	}

	// Calculate inventory growth percentage
	inventoryGrowthPercent := 0.0
	if lastMonthInventory > 0 {
		inventoryGrowthPercent = ((float64(currentInventory) - float64(lastMonthInventory)) / float64(lastMonthInventory)) * 100
	}

	// ================================
	// Active Customers (Last 90 days)
	// ================================
	ninetyDaysAgo := now.AddDate(0, 0, -90)
	var activeCustomers int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.User{}).
		Where("status = ? AND EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id AND orders.created_at >= ?)", "active", ninetyDaysAgo).
		Count(&activeCustomers).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR active customers err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch analytics"))
		return
	}

	// ================================
	// Active Customers Last Month (60-90 days ago)
	// ================================
	sixtyDaysAgo := now.AddDate(0, 0, -60)
	var lastMonthActiveCustomers int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.User{}).
		Where("status = ? AND EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id AND orders.created_at >= ? AND orders.created_at < ?)", "active", sixtyDaysAgo, ninetyDaysAgo).
		Count(&lastMonthActiveCustomers).Error; err != nil {
		log.Printf("[admin.analytics-overview] ERROR last month active customers err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch analytics"))
		return
	}

	// Calculate active customers growth percentage
	activeCustomersGrowthPercent := 0.0
	if lastMonthActiveCustomers > 0 {
		activeCustomersGrowthPercent = ((float64(activeCustomers) - float64(lastMonthActiveCustomers)) / float64(lastMonthActiveCustomers)) * 100
	} else if activeCustomers > 0 {
		activeCustomersGrowthPercent = 100.0
	}

	// ================================
	// Build Response
	// ================================
	overview := models.AnalyticsOverview{
		TotalRevenue:                 currentMonthRevenue,
		RevenueGrowthPercent:         revenueGrowthPercent,
		TotalOrders:                  int(currentMonthOrders),
		OrdersGrowthPercent:          ordersGrowthPercent,
		TotalInventory:               int(currentInventory),
		InventoryGrowthPercent:       inventoryGrowthPercent,
		ActiveCustomers:              int(activeCustomers),
		ActiveCustomersGrowthPercent: activeCustomersGrowthPercent,
	}

	log.Printf("[admin.analytics-overview] respond 200 revenue=%.2f orders=%d inventory=%d active_customers=%d",
		currentMonthRevenue, currentMonthOrders, currentInventory, activeCustomers)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Analytics overview retrieved successfully", overview))
}
