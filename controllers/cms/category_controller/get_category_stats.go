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

	var totalCategories int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Category{}).
		Count(&totalCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count categories"))
		return
	}

	var parentCategories int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Category{}).
		Where("parent_id IS NULL").
		Count(&parentCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count parent categories"))
		return
	}

	var subCategories int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Category{}).
		Where("parent_id IS NOT NULL").
		Count(&subCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count subcategories"))
		return
	}

	var activeCategories int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Category{}).
		Where("status = ?", "Active").
		Count(&activeCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count active categories"))
		return
	}

	var activeParentCategories int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Category{}).
		Where("parent_id IS NULL AND status = ?", "Active").
		Count(&activeParentCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count active parent categories"))
		return
	}

	var activeSubCategories int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Category{}).
		Where("parent_id IS NOT NULL AND status = ?", "Active").
		Count(&activeSubCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count active subcategories"))
		return
	}

	computePct := func(numerator, denominator int64) float64 {
		if denominator == 0 {
			return 0
		}
		return (float64(numerator) / float64(denominator)) * 100
	}

	stats := models.CategoryStatsResponseItem{
		TotalCategories:               int(totalCategories),
		ParentCategories:              int(parentCategories),
		SubCategories:                 int(subCategories),
		ActiveCategories:              int(activeCategories),
		ActiveParentCategories:        int(activeParentCategories),
		ActiveSubCategories:           int(activeSubCategories),
		PercentageActiveCategories:    computePct(activeCategories, totalCategories),
		PercentageActiveParents:       computePct(activeParentCategories, parentCategories),
		PercentageActiveSubCategories: computePct(activeSubCategories, subCategories),
	}

	c.JSON(
		http.StatusOK,
		models.SuccessResponse(c, "Category stats fetched successfully", stats),
	)
}
