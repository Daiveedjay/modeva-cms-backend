package product_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetProductByID godoc
// @Summary Get a product by ID
// @Description Retrieve a single product and its related details
// @Tags CMS - Products
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /api/v1/admin/products/{id} [get]
func GetProductByID(c *gin.Context) {
	// Step 1: Parse and validate product ID
	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid product ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Step 2: Fetch product with subcategory relationship
	var product models.Product
	if err := config.CmsGorm.WithContext(ctx).
		Preload("SubCategory", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, parent_id, parent_name")
		}).
		First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Product not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	// Step 3: Build response with structured data
	response := gin.H{
		"basic_info": models.ProductBase{
			ID:              product.ID,
			Name:            product.Name,
			Description:     product.Description,
			Composition:     []models.Composition(product.Composition),
			Price:           product.Price,
			SubCategoryID:   product.SubCategoryID,
			SubCategoryName: product.SubCategoryName,
			Status:          product.Status,
			Tags:            []string(product.Tags),
			CreatedAt:       product.CreatedAt,
			UpdatedAt:       product.UpdatedAt,
		},
		"seo":       product.SEO,
		"media":     product.Media,
		"variants":  []models.ProductVariant(product.Variants),
		"inventory": []models.InventoryField(product.Inventory),
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Product fetched successfully", response))
}
