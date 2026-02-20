package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetCategoryByID godoc
// @Summary Get a category by ID
// @Description Retrieve a single category by its ID, including its immediate children (if any)
// @Tags CMS - Categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /api/v1/admin/categories/{id} [get]
func GetCategoryByID(c *gin.Context) {
	// Step 1: Parse and validate category ID
	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid category ID"))
		return
	}

	// Step 2: Fetch category with its children
	var category models.Category
	if err := config.CmsGorm.
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&category, "id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Category not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Category fetched successfully", category))
}
