package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/customer_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCustomerRoutes(rg *gin.RouterGroup) {
	customer := rg.Group("/customers")

	// ════════════════════════════════════════════════════════════
	// Public Routes (No Auth Required)
	// ════════════════════════════════════════════════════════════
	customer.GET("", customer_controller.GetCustomers)
	customer.GET("/:id", customer_controller.GetCustomerDetailsByID)
	customer.GET("/:id/orders", customer_controller.GetCustomerOrders)
	customer.GET("/search", customer_controller.SearchCustomers)
	customer.GET("/stats", customer_controller.GetCustomerStats)

	// ════════════════════════════════════════════════════════════
	// Protected Routes (Auth + Activity Logging)
	// ════════════════════════════════════════════════════════════
	protected := customer.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		// Update customer details
		protected.PATCH("/:id", customer_controller.UpdateCustomerDetails)

		// Add other write operations here as needed (ban, suspend, delete, etc.)
		// protected.POST("/:id/ban", customer_controller.BanCustomer)
		// protected.DELETE("/:id", customer_controller.DeleteCustomer)
	}
}
