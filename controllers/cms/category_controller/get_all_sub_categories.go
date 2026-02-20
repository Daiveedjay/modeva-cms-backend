package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetAllSubCategories godoc
// @Summary Get all sub categories
// @Description Retrieve categories that have a parent (sub-categories only) with full path
// @Tags CMS - Categories
// @Produce json
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /api/v1/admin/categories/children [get]
func GetAllSubCategories(c *gin.Context) {
	var subCategories []models.Category

	// Fetch all subcategories with their parent information using Preload
	if err := config.CmsGorm.
		Where("parent_id IS NOT NULL").
		Preload("Parent"). // This loads the parent relationship
		Order("parent_name ASC, name ASC").
		Find(&subCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch sub-categories"))
		return
	}

	// Transform to CategoryWithPath response
	response := make([]models.CategoryWithPath, 0, len(subCategories))
	for _, cat := range subCategories {
		// Build full path
		categoryPath := cat.Name
		if cat.ParentName != nil && *cat.ParentName != "" {
			categoryPath = *cat.ParentName + " → " + cat.Name
		}

		response = append(response, models.CategoryWithPath{
			ID:           cat.ID,
			Name:         cat.Name,
			CategoryPath: categoryPath,
			Description:  cat.Description,
			Status:       cat.Status,
			ParentID:     cat.ParentID,
			CreatedAt:    cat.CreatedAt,
			UpdatedAt:    cat.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Sub-categories fetched", response))
}
