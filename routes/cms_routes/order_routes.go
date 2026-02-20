package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/order_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupOrderRoutes(rg *gin.RouterGroup) {
	order := rg.Group("/orders")

	// ════════════════════════════════════════════════════════════
	// Public Routes (No Auth Required)
	// ════════════════════════════════════════════════════════════
	order.GET("", order_controller.GetOrders)
	order.GET("/:id", order_controller.GetOrderDetailsByID)
	order.GET("/stats", order_controller.GetOrderStats)
	order.GET("/search", order_controller.SearchOrders)

	// ════════════════════════════════════════════════════════════
	// Protected Routes (Auth + Activity Logging)
	// ════════════════════════════════════════════════════════════
	protected := order.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		// Update order status (only write operation for orders)
		protected.PATCH("/:id/status", order_controller.UpdateOrderStatus)
	}
}
