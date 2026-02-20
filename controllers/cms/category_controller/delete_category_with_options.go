package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeleteCategoryWithOptions godoc
// @Summary Delete a category with options
// @Description Delete a category and either cascade delete its subcategories or reassign them to new parents
// @Tags CMS - Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param body body models.DeleteCategoryOptions true "Delete options"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /api/v1/admin/categories/{id}/delete-with-options [post]
func DeleteCategoryWithOptions(c *gin.Context) {
	// Step 1: Parse category ID
	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid category ID"))
		return
	}

	// Step 2: Parse request body
	var input models.DeleteCategoryOptions
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request body"))
		return
	}

	// Step 3: Check if category exists
	var category models.Category
	if err := config.CmsGorm.First(&category, "id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Category not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	// Step 4: Start transaction
	err = config.CmsGorm.Transaction(func(tx *gorm.DB) error {
		switch input.Mode {
		case "cascade":
			// Delete all children first
			if err := tx.Where("parent_id = ?", categoryID).Delete(&models.Category{}).Error; err != nil {
				return err
			}

		case "reassign":
			// Reassign each child to new parent
			for _, reassignment := range input.Reassignments {
				// Validate new parent exists
				var newParent models.Category
				if err := tx.First(&newParent, "id = ?", reassignment.NewParentID).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return gorm.ErrRecordNotFound // Will be caught below
					}
					return err
				}

				// Update child's parent
				if err := tx.Model(&models.Category{}).
					Where("id = ? AND parent_id = ?", reassignment.ChildID, categoryID).
					Updates(map[string]interface{}{
						"parent_id":   reassignment.NewParentID,
						"parent_name": newParent.Name,
					}).Error; err != nil {
					return err
				}
			}

		default:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid mode. Use 'cascade' or 'reassign'"))
			return gorm.ErrInvalidData
		}

		// Delete the parent category
		if err := tx.Delete(&category).Error; err != nil {
			return err
		}

		return nil // Commit transaction
	})
	// Step 5: Handle transaction result
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "One or more new parent categories not found"))
		} else if err == gorm.ErrInvalidData {
			// Already handled in switch
			return
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to delete category"))
		}
		return
	}

	// Step 6: Success response with appropriate message
	message := "Category deleted successfully"
	switch input.Mode {
	case "cascade":
		message = "Category and its subcategories deleted successfully"
	case "reassign":
		message = "Category deleted and subcategories reassigned successfully"
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, message, nil))
}
