package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category by ID (will fail if it has children due to FK constraint)
// @Tags CMS - Categories
// @Param id path string true "Category ID"
// @Success 200 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 409 {object} models.ApiResponse
// @Router /api/v1/admin/categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	// Step 1: Parse category ID
	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid category ID"))
		return
	}
	// Step 2: Check if category exists
	var category models.Category
	if err := config.CmsGorm.First(&category, "id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Category not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	// Step 3: Check if category has children
	// var childCount int64
	// if err := config.CmsGorm.Model(&models.Category{}).
	// 	Where("parent_id = ?", categoryID).
	// 	Count(&childCount).Error; err != nil {
	// 	c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to check for children"))
	// 	return
	// }

	// if childCount > 0 {
	// 	c.JSON(http.StatusConflict, models.ErrorResponse(c, "Cannot delete category with subcategories"))
	// 	return
	// }

	// Step 4: Delete the category
	if err := config.CmsGorm.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to delete category"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Category deleted successfully", nil))
}
