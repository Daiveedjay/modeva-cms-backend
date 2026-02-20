package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetCategories godoc
// @Summary Get storefront categories
// @Description Get all active categories with product counts for storefront
// @Tags store
// @Produce json
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /store/categories [get]
func GetCategories(c *gin.Context) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Get parent categories with product counts
	query := `
		SELECT 
			c.id::text AS id,
			c.name,
			c.description,
			c.parent_id::text AS parent_id,
			COUNT(DISTINCT p.id)::int AS product_count
		FROM categories c
		LEFT JOIN products p ON p.sub_category_id = c.id AND p.status = 'Active'
		WHERE c.status = 'Active'
		GROUP BY c.id, c.name, c.description, c.parent_id
		ORDER BY c.name ASC
	`

	var allCategories []models.StorefrontCategory
	if err := config.CmsGorm.WithContext(ctx).Raw(query).Scan(&allCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch categories"))
		return
	}

	// Build hierarchy
	categoriesMap := make(map[string]*models.StorefrontCategory)
	parentCategories := make([]*models.StorefrontCategory, 0)

	// First pass: Create map and identify parents
	for i := range allCategories {
		cat := &allCategories[i]
		categoriesMap[cat.ID] = cat

		// If no parent, it's a top-level category
		if cat.ParentID == nil {
			parentCategories = append(parentCategories, cat)
		}
	}

	// Second pass: Build hierarchy
	for _, cat := range categoriesMap {
		if cat.ParentID != nil {
			if parent, exists := categoriesMap[*cat.ParentID]; exists {
				if parent.Subcategories == nil {
					parent.Subcategories = []models.StorefrontCategory{}
				}
				parent.Subcategories = append(parent.Subcategories, *cat)
			}
		}
	}

	// Return the hierarchical categories
	c.JSON(http.StatusOK, models.SuccessResponse(c, "Categories fetched successfully", parentCategories))
}
