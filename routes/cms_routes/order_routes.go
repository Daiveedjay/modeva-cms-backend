package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/order_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupOrderRoutes(rg *gin.RouterGroup) {
	order := rg.Group("/orders")

	// ════════════════════════════════════════════════════════════
	// Public Routes
	// ════════════════════════════════════════════════════════════

	// Static routes first
	order.GET("", order_controller.GetOrders)
	order.GET("/stats", order_controller.GetOrderStats)
	order.GET("/search", order_controller.SearchOrders)

	// Wildcard after
	order.GET("/:id", order_controller.GetOrderDetailsByID)

	// ════════════════════════════════════════════════════════════
	// Protected Routes
	// ════════════════════════════════════════════════════════════
	protected := order.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		protected.PATCH("/:id/status", order_controller.UpdateOrderStatus)
	}
}
