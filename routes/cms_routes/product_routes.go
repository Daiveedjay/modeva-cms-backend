package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/product_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(rg *gin.RouterGroup) {
	product := rg.Group("/products")

	// ════════════════════════════════════════════════════════════
	// Public Routes (No Auth Required)
	// ════════════════════════════════════════════════════════════
	product.GET("", product_controller.GetProducts)
	product.GET("/:id", product_controller.GetProductByID)
	product.GET("/stats", product_controller.GetProductStats)
	product.GET("/search", product_controller.SearchProducts)

	// ════════════════════════════════════════════════════════════
	// Protected Routes (Auth + Activity Logging)
	// ════════════════════════════════════════════════════════════
	protected := product.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		// Create
		protected.POST("", product_controller.CreateProduct)

		// Update
		protected.PATCH("/:id", product_controller.UpdateProduct)

		// Delete
		protected.DELETE("/:id", product_controller.DeleteProduct)

		// Utility (cleanup - still needs auth + logging)
		protected.POST("/cleanup-folder", product_controller.CleanupOrphanedFolder)
	}
}
