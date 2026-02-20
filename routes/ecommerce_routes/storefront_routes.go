package ecommerce_routes

import (
	store_category "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/category_controller"
	store_filter "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/filter_controller"
	store_product "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/ecommerce/product_controller"
	"github.com/gin-gonic/gin"
)

func SetupStorefrontRoutes(router *gin.RouterGroup) {
	// Storefront routes (public, no auth required)
	store := router.Group("/store")

	// Product routes
	products := store.Group("/products")
	{
		products.GET("", store_product.GetStorefrontProducts) // List with filters

		products.GET("/filters", store_category.GetProductFilters)   // Get available filters
		products.GET("/:id", store_product.GetStorefrontProductByID) // Single product
	}

	// Category routes
	categories := store.Group("/categories")
	{
		categories.GET("", store_category.GetCategories)       // List all
		categories.GET("/:id", store_category.GetCategoryByID) // Single category

	}

	store.GET("/filters/metadata", store_filter.GetFilterMetadata)
}
