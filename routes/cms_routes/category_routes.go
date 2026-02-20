package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/category_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCategoryRoutes(rg *gin.RouterGroup) {
	category := rg.Group("/categories")

	// ════════════════════════════════════════════════════════════
	// Public Routes (No Auth Required)
	// ════════════════════════════════════════════════════════════
	category.GET("", category_controller.GetCategories)
	category.GET("/parents", category_controller.GetAllParentCategories)
	category.GET("/children", category_controller.GetAllSubCategories)
	category.GET("/:id", category_controller.GetCategoryByID)
	category.GET("/search", category_controller.SearchCategories)
	category.GET("/stats", category_controller.GetCategoryStats)

	// ════════════════════════════════════════════════════════════
	// Protected Routes (Auth + Activity Logging)
	// ════════════════════════════════════════════════════════════
	protected := category.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		// Create
		protected.POST("", category_controller.CreateCategory)

		// Update
		protected.PATCH("/:id", category_controller.UpdateCategory)
		protected.PATCH("/:id/status", category_controller.UpdateCategoryStatus)

		// Delete
		protected.DELETE("/:id", category_controller.DeleteCategory)
		protected.POST("/:id/delete-with-options", category_controller.DeleteCategoryWithOptions)
	}
}
