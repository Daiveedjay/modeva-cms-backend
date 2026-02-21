package category_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetCategoryStats godoc
// @Summary Get category statistics
// @Description Returns stats: total categories (parents & subcategories), active ones, percentages
// @Tags CMS - Categories
// @Produce json
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /api/v1/admin/categories/stats [get]
func GetCategoryStats(c *gin.Context) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	var stats models.CategoryStats
	if err := config.CmsGorm.WithContext(ctx).First(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch category stats"))
		return
	}

	computePct := func(numerator, denominator int) float64 {
		if denominator == 0 {
			return 0
		}
		return (float64(numerator) / float64(denominator)) * 100
	}

	response := models.CategoryStatsResponseItem{
		TotalCategories:               stats.TotalCategories,
		ParentCategories:              stats.ParentCategories,
		SubCategories:                 stats.SubCategories,
		ActiveCategories:              stats.ActiveCategories,
		ActiveParentCategories:        stats.ActiveParentCategories,
		ActiveSubCategories:           stats.ActiveSubCategories,
		PercentageActiveCategories:    computePct(stats.ActiveCategories, stats.TotalCategories),
		PercentageActiveParents:       computePct(stats.ActiveParentCategories, stats.ParentCategories),
		PercentageActiveSubCategories: computePct(stats.ActiveSubCategories, stats.SubCategories),
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Category stats fetched successfully", response))
}
