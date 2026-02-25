package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/category_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCategoryRoutes(rg *gin.RouterGroup) {
	category := rg.Group("/categories")

	// ════════════════════════════════════════════════════════════
	// Public Routes
	// ════════════════════════════════════════════════════════════

	// Static routes first
	category.GET("", category_controller.GetCategories)
	category.GET("/parents", category_controller.GetAllParentCategories)
	category.GET("/children", category_controller.GetAllSubCategories)
	category.GET("/search", category_controller.SearchCategories)
	category.GET("/stats", category_controller.GetCategoryStats)

	// Wildcard after
	category.GET("/:id", category_controller.GetCategoryByID)

	// ════════════════════════════════════════════════════════════
	// Protected Routes
	// ════════════════════════════════════════════════════════════
	protected := category.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		protected.POST("", category_controller.CreateCategory)
		protected.PATCH("/:id", category_controller.UpdateCategory)
		protected.PATCH("/:id/status", category_controller.UpdateCategoryStatus)
		protected.DELETE("/:id", category_controller.DeleteCategory)
		protected.POST("/:id/delete-with-options", category_controller.DeleteCategoryWithOptions)
	}
}
