package analytics_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetDeviceAnalytics godoc
// @Summary Get device analytics
// @Description Returns order distribution by device type (desktop, mobile, tablet) with percentages for the current month
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=[]models.DeviceAnalytics}
// @Failure 500 {object} models.ApiResponse
// @Router /admin/analytics/devices [get]
func GetDeviceAnalytics(c *gin.Context) {
	log.Printf("[admin.analytics-devices] start")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// ================================
	// Get total orders this month
	// ================================
	var totalOrders int64
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&models.Order{}).
		Where("status = ? AND created_at >= ?", "completed", monthStart).
		Count(&totalOrders).Error; err != nil {
		log.Printf("[admin.analytics-devices] ERROR total orders err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch device analytics"))
		return
	}

	// ================================
	// Get orders by device type this month
	// ================================
	var deviceData []models.DeviceAnalytics
	if err := config.EcommerceGorm.WithContext(ctx).
		Raw(`
			SELECT 
				COALESCE(device_type, 'desktop') AS device_type,
				COUNT(id)::int AS order_count
			FROM orders
			WHERE status = ? AND created_at >= ?
			GROUP BY device_type
			ORDER BY order_count DESC
		`, "completed", monthStart).
		Scan(&deviceData).Error; err != nil {
		log.Printf("[admin.analytics-devices] ERROR query device analytics err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch device analytics"))
		return
	}

	// ================================
	// Calculate percentages
	// ================================
	for i := range deviceData {
		if totalOrders > 0 {
			deviceData[i].Percentage = (float64(deviceData[i].OrderCount) / float64(totalOrders)) * 100
		} else {
			deviceData[i].Percentage = 0
		}
	}

	log.Printf("[admin.analytics-devices] respond 200 devices=%d total_orders=%d",
		len(deviceData), totalOrders)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Device analytics retrieved successfully", deviceData))
}
