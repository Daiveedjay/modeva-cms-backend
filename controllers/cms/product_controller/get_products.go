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

// GetProducts godoc
// @Summary Get paginated products
// @Description Retrieve all products with pagination and optional filtering
// @Tags CMS - Products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status" Enums(Active, Draft)
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /api/v1/admin/products [get]
func GetProducts(c *gin.Context) {
	// Step 1: Parse and validate pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Step 2: Build query with optional filters
	query := config.CmsGorm.Model(&models.Product{})

	// Optional status filter
	if status := c.Query("status"); status != "" {
		if status == "Active" || status == "Draft" {
			query = query.Where("status = ?", status)
		}
	}

	// Step 3: Count total products
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count products"))
		return
	}

	// Step 4: Fetch products with subcategory info
	products := make([]models.Product, 0)
	if err := query.
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

	// Step 5: Transform products into structured response format
	productResponses := make([]gin.H, 0, len(products))
	for _, product := range products {
		// Build sub_category_path
		var subCategoryPath string
		if product.SubCategory != nil {
			if product.SubCategory.ParentName != nil {
				// Has parent: "Parent > Child"
				subCategoryPath = *product.SubCategory.ParentName + " -> " + product.SubCategory.Name
			} else {
				// No parent: just "Category"
				subCategoryPath = product.SubCategory.Name
			}
		}

		productResponses = append(productResponses, gin.H{
			"basic_info": models.ProductBase{
				ID:              product.ID,
				Name:            product.Name,
				Description:     product.Description,
				Composition:     []models.Composition(product.Composition),
				Price:           product.Price,
				SubCategoryID:   product.SubCategoryID,
				SubCategoryName: product.SubCategoryName,
				SubCategoryPath: &subCategoryPath,
				Status:          product.Status,
				Tags:            []string(product.Tags),
				CreatedAt:       product.CreatedAt,
				UpdatedAt:       product.UpdatedAt,
			},
			"seo":       product.SEO,
			"media":     product.Media,
			"variants":  []models.ProductVariant(product.Variants),
			"inventory": []models.InventoryField(product.Inventory),
		})
	}

	// Step 6: Prepare pagination meta
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Products fetched successfully", productResponses, meta))
}
