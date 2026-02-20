package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UpdateCategory godoc
// @Summary Update a category
// @Description Update category name, description, and optionally parent_id
// @Tags CMS - Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body models.UpdateCategoryRequest true "Update category"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /api/v1/admin/categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	// Step 1: Parse category ID
	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid category ID"))
		return
	}

	// Step 2: Parse request body
	var input models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request body"))
		return
	}

	// Step 3: Find existing category
	var existing models.Category
	if err := config.CmsGorm.First(&existing, "id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Category not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	// Step 4: Check if anything actually changed
	if !hasChanges(input, existing) {
		// Nothing to update - return existing category
		c.JSON(http.StatusOK, models.SuccessResponse(c, "No changes detected", existing))
		return
	}

	// Step 5: Validate parent if it's being changed
	if input.ParentID != nil {
		// Can't be its own parent
		if *input.ParentID == categoryID {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Category cannot be its own parent"))
			return
		}

		// Check if new parent exists and is top-level
		var parent models.Category
		if err := config.CmsGorm.First(&parent, "id = ?", *input.ParentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Parent category not found"))
			} else {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
			}
			return
		}

		// Parent must be top-level (no parent of its own)
		if parent.ParentID != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Selected parent category must be top-level"))
			return
		}

		// Update parent name
		existing.ParentName = &parent.Name
	}

	// Step 6: Apply updates (only fields that were provided)
	updates := map[string]interface{}{}

	if input.Name != nil {
		updates["name"] = *input.Name
		existing.Name = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
		existing.Description = *input.Description
	}
	if input.ParentID != nil {
		updates["parent_id"] = *input.ParentID
		updates["parent_name"] = existing.ParentName
		existing.ParentID = input.ParentID
	}

	// Step 7: Update in database
	if err := config.CmsGorm.Model(&existing).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update category"))
		return
	}

	// Step 8: Reload to get fresh data (with updated_at)
	if err := config.CmsGorm.First(&existing, "id = ?", categoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to reload category"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Category updated successfully", existing))
}

// hasChanges checks if any field in the request differs from existing
func hasChanges(input models.UpdateCategoryRequest, existing models.Category) bool {
	if input.Name != nil && *input.Name != existing.Name {
		return true
	}
	if input.Description != nil && *input.Description != existing.Description {
		return true
	}
	if input.ParentID != nil {
		if existing.ParentID == nil {
			return true
		}
		if *input.ParentID != *existing.ParentID {
			return true
		}
	}
	return false
}
