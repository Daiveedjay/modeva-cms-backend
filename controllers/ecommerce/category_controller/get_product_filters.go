package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetProductFilters godoc
// @Summary Get available product filters
// @Description Get all available filters for products (categories, sizes, price range)
// @Tags store
// @Produce json
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /store/products/filters [get]
func GetProductFilters(c *gin.Context) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	filters := models.ProductFilters{}

	// Get categories with product counts (only subcategories that have products)
	catQuery := `
		SELECT 
			c.id::text AS value,
			c.name AS label,
			COUNT(p.id)::int AS count
		FROM categories c
		LEFT JOIN products p ON p.sub_category_id = c.id AND p.status = 'Active'
		WHERE c.status = 'Active'
		GROUP BY c.id, c.name
		HAVING COUNT(p.id) > 0
		ORDER BY c.name ASC
	`

	var categories []models.FilterOption
	if err := config.CmsGorm.WithContext(ctx).Raw(catQuery).Scan(&categories).Error; err == nil {
		filters.Categories = categories
	} else {
		filters.Categories = []models.FilterOption{}
	}

	// Get unique sizes from variants
	sizeQuery := `
		SELECT DISTINCT 
			variant->>'size' AS value,
			variant->>'size' AS label,
			COUNT(*)::int AS count
		FROM products p, jsonb_array_elements(p.variants) AS variant
		WHERE p.status = 'Active' AND variant->>'size' IS NOT NULL
		GROUP BY variant->>'size'
		ORDER BY variant->>'size' ASC
	`

	var sizes []models.FilterOption
	if err := config.CmsGorm.WithContext(ctx).Raw(sizeQuery).Scan(&sizes).Error; err == nil {
		filters.Sizes = sizes
	} else {
		filters.Sizes = []models.FilterOption{}
	}

	// Get price range (use COALESCE for safety)
	priceQuery := `
		SELECT 
			COALESCE(MIN(price), 0)::float8 AS min, 
			COALESCE(MAX(price), 0)::float8 AS max
		FROM products
		WHERE status = 'Active'
	`

	var priceRange models.PriceRange
	if err := config.CmsGorm.WithContext(ctx).Raw(priceQuery).Scan(&priceRange).Error; err == nil {
		filters.PriceRange = priceRange
	} else {
		filters.PriceRange = models.PriceRange{Min: 0, Max: 0}
	}

	// Availability options
	filters.Availability = []models.FilterOption{
		{Label: "In Stock", Value: "in_stock"},
		{Label: "Out of Stock", Value: "out_of_stock"},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Filters fetched successfully", filters))
}
