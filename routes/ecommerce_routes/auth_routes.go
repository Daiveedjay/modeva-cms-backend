package ecommerce_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/auth_controller"
	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes sets up all authentication routes
func SetupAuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		// Google OAuth routes
		auth.GET("/google", auth_controller.GoogleLogin)
		auth.GET("/google/callback", auth_controller.GoogleCallback)

		auth.POST("/logout", auth_controller.Logout)
	}
}
