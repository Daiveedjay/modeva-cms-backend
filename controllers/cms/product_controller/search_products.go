package product_controller

import (
	"math"
	"net/http"
	"strconv"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SearchProducts godoc
// @Summary Search products
// @Description Search products by name, description, or tags (case-insensitive). Returns paginated results with subcategory info.
// @Tags CMS - Products
// @Produce json
// @Param query query string true "Search keyword"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Router /api/v1/admin/products/search [get]
func SearchProducts(c *gin.Context) {
	// Step 1: Parse query parameter
	queryParam := c.Query("query")
	if queryParam == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Query parameter 'query' is required"))
		return
	}

	// Step 2: Parse and validate pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Step 3: Build search query
	searchPattern := "%" + queryParam + "%"

	// Count total matches (using Raw SQL for JSONB array search)
	var total int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.Product{}).
		Where(`
			name ILIKE ? OR 
			description ILIKE ? OR 
			EXISTS (
				SELECT 1 FROM jsonb_array_elements_text(tags) AS tag
				WHERE tag ILIKE ?
			)
		`, searchPattern, searchPattern, searchPattern).
		Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count products"))
		return
	}

	// Step 4: Early return if no results
	if total == 0 {
		meta := &models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      0,
			TotalPages: 0,
		}
		c.JSON(http.StatusOK, models.PaginatedResponse(c, "No results found", make([]models.ProductResponse, 0), meta))
		return
	}

	// Step 5: Fetch matching products with subcategory
	products := make([]models.Product, 0)
	if err := config.CmsGorm.WithContext(ctx).
		Where(`
			name ILIKE ? OR 
			description ILIKE ? OR 
			EXISTS (
				SELECT 1 FROM jsonb_array_elements_text(tags) AS tag
				WHERE tag ILIKE ?
			)
		`, searchPattern, searchPattern, searchPattern).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("SubCategory", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, parent_id, parent_name")
		}).
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch products"))
		return
	}

	// Step 6: Transform to ProductResponse format
	responses := make([]models.ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, models.ProductResponse{
			BasicInfo: models.ProductBase{
				ID:              p.ID,
				Name:            p.Name,
				Description:     p.Description,
				Composition:     []models.Composition(p.Composition),
				Price:           p.Price,
				SubCategoryID:   p.SubCategoryID,
				SubCategoryName: p.SubCategoryName,
				Status:          p.Status,
				Tags:            []string(p.Tags),
				CreatedAt:       p.CreatedAt,
				UpdatedAt:       p.UpdatedAt,
			},
			SEO:       p.SEO,
			Media:     p.Media,
			Variants:  []models.ProductVariant(p.Variants),
			Inventory: []models.InventoryField(p.Inventory),
		})
	}

	// Step 7: Prepare pagination meta
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Search results", responses, meta))
}
