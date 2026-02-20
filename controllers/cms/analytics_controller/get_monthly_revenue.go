package analytics_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetMonthlyRevenue godoc
// @Summary Get monthly revenue for last 12 months
// @Description Returns revenue data for the last 12 months for chart visualization
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=[]models.MonthlyRevenueData}
// @Failure 500 {object} models.ApiResponse
// @Router /admin/analytics/monthly-revenue [get]
func GetMonthlyRevenue(c *gin.Context) {
	log.Printf("[admin.analytics-monthly-revenue] start")

	ctx, cancel := config.WithTimeout()
	defer cancel()

	now := time.Now()

	// ================================
	// Get revenue for last 12 months
	// ================================
	var monthlyData []models.MonthlyRevenueData
	if err := config.EcommerceGorm.WithContext(ctx).
		Raw(`
			SELECT 
				TO_CHAR(date_trunc('month', created_at), 'Mon') AS month,
				EXTRACT(MONTH FROM created_at)::int AS month_number,
				COALESCE(SUM(total_amount), 0)::float8 AS revenue
			FROM orders
			WHERE status = ? AND created_at >= ?
			GROUP BY date_trunc('month', created_at), TO_CHAR(date_trunc('month', created_at), 'Mon'), EXTRACT(MONTH FROM created_at)
			ORDER BY date_trunc('month', created_at) ASC
		`, "completed", now.AddDate(0, -12, 0)).
		Scan(&monthlyData).Error; err != nil {
		log.Printf("[admin.analytics-monthly-revenue] ERROR query monthly revenue err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch monthly revenue"))
		return
	}

	// ================================
	// Ensure all 12 months are present (fill missing months with 0)
	// ================================
	monthlyMap := make(map[int]models.MonthlyRevenueData)
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for _, data := range monthlyData {
		monthlyMap[data.MonthNumber] = data
	}

	// Build complete 12-month data with current and previous 11 months
	completeData := []models.MonthlyRevenueData{}
	startMonth := now.AddDate(0, -11, 0) // Start from 11 months ago

	for i := 0; i < 12; i++ {
		currentMonth := startMonth.AddDate(0, i, 0)
		monthNum := int(currentMonth.Month())
		monthName := monthNames[monthNum-1]

		if data, exists := monthlyMap[monthNum]; exists {
			completeData = append(completeData, models.MonthlyRevenueData{
				Month:       data.Month,
				MonthNumber: data.MonthNumber,
				Revenue:     data.Revenue,
			})
		} else {
			completeData = append(completeData, models.MonthlyRevenueData{
				Month:       monthName,
				MonthNumber: monthNum,
				Revenue:     0,
			})
		}
	}

	log.Printf("[admin.analytics-monthly-revenue] respond 200 months=%d", len(completeData))

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Monthly revenue retrieved successfully", completeData))
}
