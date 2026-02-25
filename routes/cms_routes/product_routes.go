package cms_routes

import (
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/product_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(rg *gin.RouterGroup) {
	product := rg.Group("/products")

	// ════════════════════════════════════════════════════════════
	// Public Routes
	// ════════════════════════════════════════════════════════════

	// Static routes first
	product.GET("", product_controller.GetProducts)
	product.GET("/stats", product_controller.GetProductStats)
	product.GET("/search", product_controller.SearchProducts)

	// Wildcard after
	product.GET("/:id", product_controller.GetProductByID)

	// ════════════════════════════════════════════════════════════
	// Protected Routes
	// ════════════════════════════════════════════════════════════
	protected := product.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	protected.Use(middleware.ActivityLoggingMiddleware())
	{
		protected.POST("", product_controller.CreateProduct)
		protected.POST("/cleanup-folder", product_controller.CleanupOrphanedFolder)
		protected.PATCH("/:id", product_controller.UpdateProduct)
		protected.DELETE("/:id", product_controller.DeleteProduct)
	}
}
