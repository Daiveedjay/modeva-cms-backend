package product_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetProductStats godoc
// @Summary Get product statistics
// @Description Returns overall product stats including low-stock counts
// @Tags CMS - Products
// @Produce json
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /api/v1/admin/products/stats [get]
func GetProductStats(c *gin.Context) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Step 1: Count total products
	var totalProducts int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Product{}).
		Count(&totalProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count total products"))
		return
	}

	// Step 2: Count active products
	var activeProducts int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Product{}).
		Where("status = ?", "Active").
		Count(&activeProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count active products"))
		return
	}

	// Step 3: Count draft products
	var draftProducts int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Product{}).
		Where("status = ?", "Draft").
		Count(&draftProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count draft products"))
		return
	}

	// Step 4: Total inventory (sum of all quantities in inventory JSONB array)
	var totalInventory int
	if err := config.CmsGorm.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM((inv->>'quantity')::int), 0)
			FROM products, jsonb_array_elements(inventory) AS inv
		`).
		Scan(&totalInventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count total inventory"))
		return
	}

	// Step 5: Average product price
	var averagePrice float64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Product{}).
		Select("COALESCE(AVG(price), 0)").
		Scan(&averagePrice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to calculate average price"))
		return
	}

	// Step 6: Count tagged products (tags array has elements)
	var taggedProducts int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Product{}).
		Where("jsonb_array_length(tags) > 0").
		Count(&taggedProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count tagged products"))
		return
	}

	// Step 7: Count low stock products (any variant quantity < 5)
	var lowStockProducts int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Product{}).
		Where(`EXISTS (
			SELECT 1 
			FROM jsonb_array_elements(inventory) AS inv
			WHERE (inv->>'quantity')::int < 5
		)`).
		Count(&lowStockProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count low stock products"))
		return
	}

	// Compute percentages safely
	computePct := func(numerator int64, denominator int64) float64 {
		if denominator == 0 {
			return 0
		}
		return (float64(numerator) / float64(denominator)) * 100
	}

	stats := []models.ProductStatsResponseItem{
		{
			Type:               "total",
			TotalProducts:      int(totalProducts),
			ActiveProducts:     int(activeProducts),
			DraftProducts:      int(draftProducts),
			AveragePrice:       averagePrice,
			TotalInventory:     totalInventory,
			TaggedProducts:     int(taggedProducts),
			LowStockProducts:   int(lowStockProducts),
			PercentageLowStock: computePct(lowStockProducts, totalProducts),
			PercentageActive:   computePct(activeProducts, totalProducts),
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Product stats fetched successfully", stats))
}
