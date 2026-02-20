package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UpdateCategoryStatus godoc
// @Summary Update category status
// @Description Change a category's status (Active or Inactive) and cascade to subcategories
// @Tags CMS - Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param status body models.UpdateCategoryStatusRequest true "New status"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /api/v1/admin/categories/{id}/status [patch]
func UpdateCategoryStatus(c *gin.Context) {
	// Step 1: Parse and validate category ID
	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid category ID"))
		return
	}

	// Step 2: Parse and validate request body
	var input models.UpdateCategoryStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, err.Error()))
		return
	}

	// Note: Gin's binding tag already validates this, but double-check for clarity
	if input.Status != "Active" && input.Status != "Inactive" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Status must be either 'Active' or 'Inactive'"))
		return
	}

	// Step 3: Start transaction
	var category models.Category
	err = config.CmsGorm.Transaction(func(tx *gorm.DB) error {
		// Find and update the category
		if err := tx.First(&category, "id = ?", categoryID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return err
			}
			return err
		}

		// Update parent category status
		if err := tx.Model(&category).Update("status", input.Status).Error; err != nil {
			return err
		}

		// If status is being set to Inactive, cascade to all subcategories
		if input.Status == "Inactive" {
			if err := tx.Model(&models.Category{}).
				Where("parent_id = ?", categoryID).
				Update("status", "Inactive").Error; err != nil {
				return err
			}
		}

		// Reload category to get updated data
		if err := tx.First(&category, "id = ?", categoryID).Error; err != nil {
			return err
		}

		return nil // Commit transaction
	})
	// Step 4: Handle transaction result
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Category not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update category status"))
		}
		return
	}

	// Step 5: Return appropriate success message
	message := "Category status updated successfully"
	if input.Status == "Inactive" {
		message = "Category and all subcategories set to inactive"
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, message, category))
}
