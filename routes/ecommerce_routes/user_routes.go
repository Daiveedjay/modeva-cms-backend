package ecommerce_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/user_controller/address_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/user_controller/order_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/user_controller/payment_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/user_controller/profile_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupUserRoutes sets up all user profile routes
func SetupUserRoutes(router *gin.RouterGroup) {
	user := router.Group("/user")
	user.Use(middleware.AuthMiddleware()) // All routes require auth
	{

		user.GET("/", profile_controller.GetProfile)
		user.PATCH("/", profile_controller.UpdateProfile)
		user.GET("/overview", profile_controller.GetUserOverview)
		user.GET("/me", profile_controller.GetMe)

		// Payment methods
		user.GET("/payment-methods", payment_controller.GetPaymentMethods)
		user.POST("/payment-methods", payment_controller.AddPaymentMethod)
		user.PATCH("/payment-methods/:id", payment_controller.UpdatePaymentMethod)
		user.DELETE("/payment-methods/:id", payment_controller.DeletePaymentMethod)
		user.PATCH("/payment-methods/:id/default", payment_controller.SetDefaultPaymentMethod)

		// Addresses
		user.GET("/addresses", address_controller.GetAddresses)
		user.POST("/addresses", address_controller.AddAddress)
		user.PATCH("/addresses/:id", address_controller.UpdateAddress)
		user.DELETE("/addresses/:id", address_controller.DeleteAddress)
		user.PATCH("/addresses/:id/default", address_controller.SetDefaultAddress)

		// Orders
		user.GET("/orders", order_controller.GetOrders)
		user.GET("/orders/:id", order_controller.GetOrderDetails)
		user.POST("/orders", order_controller.CreateOrder)
	}
}
