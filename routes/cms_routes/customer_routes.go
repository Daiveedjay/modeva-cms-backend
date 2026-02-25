package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/customer_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCustomerRoutes(rg *gin.RouterGroup) {
	customer := rg.Group("/customers")

	// ── Static routes FIRST ──────────────────────────────────
	customer.GET("", customer_controller.GetCustomers)
	customer.GET("/search", customer_controller.SearchCustomers)
	customer.GET("/stats", customer_controller.GetCustomerStats)

	// ── Wildcard routes AFTER ────────────────────────────────
	customer.GET("/:id", customer_controller.GetCustomerDetailsByID)
	customer.GET("/:id/orders", customer_controller.GetCustomerOrders)

	// ── Protected ────────────────────────────────────────────
	protected := customer.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		protected.PATCH("/:id", customer_controller.UpdateCustomerDetails)
		protected.POST("/:id/send-email", customer_controller.SendCustomerEmail)
		protected.POST("/:id/ban", customer_controller.BanCustomer)
		protected.POST("/:id/unban", customer_controller.UnbanCustomer)
		protected.DELETE("/:id", customer_controller.DeleteCustomer)
	}
}
