package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetAllParentCategories godoc
// @Summary Get all parent categories
// @Description Retrieve categories that have no parent (top-level categories only) with product counts
// @Tags CMS - Categories
// @Produce json
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /api/v1/admin/categories/parents [get]
func GetAllParentCategories(c *gin.Context) {
	parents := make([]models.Category, 0)

	// Step 1: Fetch all parent categories (where parent_id IS NULL)
	if err := config.CmsGorm.
		Where("parent_id IS NULL").
		Order("created_at ASC").
		Preload("Children").
		Find(&parents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch parent categories"))
		return
	}

	// Step 2: Get product counts for all categories
	productCounts := make(map[string]int)

	// Collect all category IDs (parents + their children)
	categoryIDs := make([]string, 0)
	for _, parent := range parents {
		categoryIDs = append(categoryIDs, parent.ID.String())
		for _, child := range parent.Children {
			categoryIDs = append(categoryIDs, child.ID.String())
		}
	}

	// Query product counts in one go
	type CountResult struct {
		SubCategoryID string `gorm:"column:sub_category_id"`
		Count         int    `gorm:"column:count"`
	}

	var counts []CountResult
	if len(categoryIDs) > 0 {
		if err := config.CmsGorm.Table("products").
			Select("sub_category_id, COUNT(*) as count").
			Where("sub_category_id IN ?", categoryIDs).
			Group("sub_category_id").
			Scan(&counts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count products"))
			return
		}

		// Map counts to category IDs
		for _, count := range counts {
			productCounts[count.SubCategoryID] = count.Count
		}
	}

	// Step 3: Transform to CategoryWithProducts
	response := make([]models.CategoryWithProducts, len(parents))
	for i, parent := range parents {
		// Count products for parent (sum of all its children's products)
		parentProductCount := 0

		for _, child := range parent.Children {
			childCount := productCounts[child.ID.String()]
			parentProductCount += childCount
		}

		response[i] = models.CategoryWithProducts{
			ID:          parent.ID,
			Name:        parent.Name,
			Description: parent.Description,
			Status:      parent.Status,
			ParentID:    parent.ParentID,
			ParentName:  parent.ParentName,
			CreatedAt:   parent.CreatedAt,
			UpdatedAt:   parent.UpdatedAt,
			Products:    parentProductCount,
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Parent categories fetched", response))
}
