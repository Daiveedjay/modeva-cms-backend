package product_controller

import (
	"github.com/gin-gonic/gin"
)

// GetStorefrontProducts godoc
// @Summary Get storefront products
// @Description Get paginated products for storefront with optional search and filtering
// @Tags store
// @Produce json
// @Param q query string false "Search query"
// @Param category query []string false "Category IDs (repeatable ?category=ID&category=ID)"
// @Param subcategory query []string false "Subcategory IDs (repeatable ?subcategory=ID&subcategory=ID)"
// @Param size query []string false "Sizes (repeatable ?size=XS&size=S)"
// @Param availability query string false "Availability filter" Enums(in_stock, out_of_stock, inStock, outOfStock)
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param sortBy query string false "Sort by field" Enums(price, name, newest, popular) default(newest)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /store/products [get]
func GetStorefrontProducts(c *gin.Context) {
	if hasStorefrontFilters(c) {
		getStorefrontProductsWithFilters(c)
	} else {
		getStorefrontProductsWithoutFilters(c)
	}
}

// hasStorefrontFilters checks if any filter-related query param is present.
// hasStorefrontFilters checks if any filter-related query param is present.
func hasStorefrontFilters(c *gin.Context) bool {
	if c.Query("q") != "" ||
		len(c.QueryArray("category")) > 0 ||
		len(c.QueryArray("subcategory")) > 0 ||
		len(c.QueryArray("size")) > 0 ||
		len(c.QueryArray("color")) > 0 ||
		c.Query("availability") != "" ||
		c.Query("minPrice") != "" ||
		c.Query("maxPrice") != "" ||
		c.Query("style") != "" { // Add this line
		return true
	}
	return false
}
