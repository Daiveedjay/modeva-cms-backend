package category_controller

import (
	"net/http"

	category_cache "github.com/Modeva-Ecommerce/modeva-cms-backend/cache"
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
	// Try cache first — shares the same tree as GetCategories
	parents, productCounts, ok := category_cache.GetTree()
	if !ok {
		// Cache miss — fetch from DB
		if err := config.CmsGorm.
			Where("parent_id IS NULL").
			Order("created_at ASC").
			Preload("Children").
			Find(&parents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch parent categories"))
			return
		}

		categoryIDs := make([]string, 0, len(parents)*4)
		for _, p := range parents {
			categoryIDs = append(categoryIDs, p.ID.String())
			for _, child := range p.Children {
				categoryIDs = append(categoryIDs, child.ID.String())
			}
		}

		productCounts = make(map[string]int)
		if len(categoryIDs) > 0 {
			type CountResult struct {
				SubCategoryID string `gorm:"column:sub_category_id"`
				Count         int    `gorm:"column:count"`
			}
			var counts []CountResult
			if err := config.CmsGorm.Table("products").
				Select("sub_category_id, COUNT(*) as count").
				Where("sub_category_id IN ?", categoryIDs).
				Group("sub_category_id").
				Scan(&counts).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count products"))
				return
			}
			for _, cr := range counts {
				productCounts[cr.SubCategoryID] = cr.Count
			}
		}

		category_cache.SetTree(parents, productCounts)
	}

	// Build response (parents only, no children in this endpoint)
	response := make([]models.CategoryWithProducts, len(parents))
	for i, parent := range parents {
		parentProductCount := 0
		for _, child := range parent.Children {
			parentProductCount += productCounts[child.ID.String()]
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
