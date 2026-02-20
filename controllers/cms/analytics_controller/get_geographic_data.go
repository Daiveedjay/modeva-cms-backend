package analytics_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetGeographicData godoc
// @Summary Get geographic data
// @Description Returns order distribution by country with percentages for the current month
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=[]models.GeographicData}
// @Failure 500 {object} models.ApiResponse
// @Router /admin/analytics/geographic [get]
func GetGeographicData(c *gin.Context) {
	log.Printf("[admin.analytics-geographic] start")

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
		log.Printf("[admin.analytics-geographic] ERROR total orders err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch geographic data"))
		return
	}

	// ================================
	// Get orders by country this month
	// ================================
	var geographicData []models.GeographicData
	if err := config.EcommerceGorm.WithContext(ctx).
		Raw(`
			SELECT 
				a.country,
				COUNT(o.id)::int AS order_count
			FROM orders o
			LEFT JOIN addresses a ON o.address_id = a.id
			WHERE o.status = ? AND o.created_at >= ?
			GROUP BY a.country
			ORDER BY order_count DESC
		`, "completed", monthStart).
		Scan(&geographicData).Error; err != nil {
		log.Printf("[admin.analytics-geographic] ERROR query geographic data err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch geographic data"))
		return
	}

	// ================================
	// Calculate percentages
	// ================================
	for i := range geographicData {
		if totalOrders > 0 {
			geographicData[i].Percentage = (float64(geographicData[i].OrderCount) / float64(totalOrders)) * 100
		} else {
			geographicData[i].Percentage = 0
		}
	}

	log.Printf("[admin.analytics-geographic] respond 200 countries=%d total_orders=%d",
		len(geographicData), totalOrders)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Geographic data retrieved successfully", geographicData))
}
