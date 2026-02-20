package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/analytics_controller"
	"github.com/gin-gonic/gin"
)

func SetupAnalyticsRoutes(rg *gin.RouterGroup) {
	analytics := rg.Group("/analytics")

	analytics.GET("/overview", analytics_controller.GetAnalyticsOverview)
	analytics.GET("/top-products", analytics_controller.GetTopProducts)
	analytics.GET("/monthly-revenue", analytics_controller.GetMonthlyRevenue)
	analytics.GET("/sales-metrics", analytics_controller.GetSalesMetrics)
	analytics.GET("/geographic-data", analytics_controller.GetGeographicData)
	analytics.GET("/device-analytics", analytics_controller.GetDeviceAnalytics)
}
