package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCategoryByID godoc
// @Summary Get category details
// @Description Get single category with subcategories and product count
// @Tags store
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /store/categories/{id} [get]
func GetCategoryByID(c *gin.Context) {
	categoryIDStr := c.Param("id")

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid category ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Get category with product count
	query := `
		SELECT 
			c.id::text AS id,
			c.name,
			c.description,
			c.parent_id::text AS parent_id,
			COUNT(DISTINCT p.id)::int AS product_count
		FROM categories c
		LEFT JOIN products p ON p.sub_category_id = c.id AND p.status = 'Active'
		WHERE c.id = ? AND c.status = 'Active'
		GROUP BY c.id, c.name, c.description, c.parent_id
	`

	var category models.StorefrontCategory
	err = config.CmsGorm.WithContext(ctx).Raw(query, categoryID).Scan(&category).Error
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Category not found"))
		return
	}

	// Check if category was actually found
	if category.ID == "" {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Category not found"))
		return
	}

	// Get subcategories
	subQuery := `
		SELECT 
			c.id::text AS id,
			c.name,
			c.description,
			c.parent_id::text AS parent_id,
			COUNT(DISTINCT p.id)::int AS product_count
		FROM categories c
		LEFT JOIN products p ON p.sub_category_id = c.id AND p.status = 'Active'
		WHERE c.parent_id = ? AND c.status = 'Active'
		GROUP BY c.id, c.name, c.description, c.parent_id
		ORDER BY c.name ASC
	`

	var subcategories []models.StorefrontCategory
	if err := config.CmsGorm.WithContext(ctx).Raw(subQuery, categoryID).Scan(&subcategories).Error; err == nil {
		category.Subcategories = subcategories
	} else {
		// If error fetching subcategories, set empty array
		category.Subcategories = []models.StorefrontCategory{}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Category fetched successfully", category))
}
