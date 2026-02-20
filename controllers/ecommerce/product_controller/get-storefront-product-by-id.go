package product_controller

import (
	"encoding/json"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetStorefrontProductByID godoc
// @Summary Get single product details for storefront
// @Description Get detailed product information by ID
// @Tags store
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /store/products/{id} [get]
func GetStorefrontProductByID(c *gin.Context) {
	productIDStr := c.Param("id")

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid product ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	query := `
		SELECT 
			p.id::text AS id,
			p.name,
			p.description,
			p.price,
			p.inventory,
			p.media,
			p.variants,
			c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON c.id = p.sub_category_id
		WHERE p.id = ? AND p.status = 'Active'
	`

	var result struct {
		ID           string  `gorm:"column:id"`
		Name         string  `gorm:"column:name"`
		Description  string  `gorm:"column:description"`
		Price        float64 `gorm:"column:price"`
		Inventory    []byte  `gorm:"column:inventory"`
		Media        []byte  `gorm:"column:media"`
		Variants     []byte  `gorm:"column:variants"`
		CategoryName *string `gorm:"column:category_name"`
	}

	err = config.CmsGorm.WithContext(ctx).Raw(query, productID).Scan(&result).Error
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Product not found"))
		return
	}

	// Check if product was actually found
	if result.ID == "" {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Product not found"))
		return
	}

	var media models.ProductMedia
	if err := json.Unmarshal(result.Media, &media); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to parse product media"))
		return
	}

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to serialize product media"))
		return
	}

	product := models.StorefrontProduct{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		Price:       result.Price,
		Inventory:   result.Inventory,
		Variants:    result.Variants,
		Media:       mediaJSON, // ✅ now RawMessage
	}
	// Optional: Increment view count
	go incrementProductViews(productID)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Product fetched successfully", product))
}

// incrementProductViews increments the view count for a product
func incrementProductViews(productID uuid.UUID) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	query := `
		UPDATE products 
		SET views = COALESCE(views, 0) + 1 
		WHERE id = ?
	`
	config.CmsGorm.WithContext(ctx).Exec(query, productID)
}
