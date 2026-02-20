package category_controller

import (
	"math"
	"net/http"
	"strconv"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SearchCategories godoc
// @Summary Search categories (with subcategories)
// @Description Search parent categories by name or description (case-insensitive). Returns paginated parents with their subcategories and product counts.
// @Tags CMS - Categories
// @Produce json
// @Param query query string true "Search keyword"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Router /api/v1/admin/categories/search [get]
func SearchCategories(c *gin.Context) {
	// Step 1: Parse query params
	queryParam := c.Query("query")
	if queryParam == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Query parameter 'query' is required"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Step 2: Build search query
	searchPattern := "%" + queryParam + "%"

	// Count total matching parent categories
	var total int64
	if err := config.CmsGorm.Model(&models.Category{}).
		Where("parent_id IS NULL").
		Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to count categories"))
		return
	}

	// Step 3: Early return if no results
	if total == 0 {
		emptyParents := make([]models.CategoryWithProducts, 0)
		meta := &models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      0,
			TotalPages: 0,
		}
		c.JSON(http.StatusOK, models.PaginatedResponse(c, "No results found", emptyParents, meta))
		return
	}

	// Step 4: Fetch matching parent categories with their children
	parents := make([]models.Category, 0)
	if err := config.CmsGorm.
		Where("parent_id IS NULL").
		Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Find(&parents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch categories"))
		return
	}

	// Step 5: Get product counts for all categories
	productCounts := make(map[string]int)

	// Collect all category IDs (parents + children)
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

	// Step 6: Transform to CategoryWithProducts
	response := make([]models.CategoryWithProducts, len(parents))
	for i, parent := range parents {
		// Count products for parent (sum of all its children's products)
		parentProductCount := 0
		children := make([]models.CategoryWithProducts, len(parent.Children))

		for j, child := range parent.Children {
			childCount := productCounts[child.ID.String()]
			parentProductCount += childCount

			children[j] = models.CategoryWithProducts{
				ID:          child.ID,
				Name:        child.Name,
				Description: child.Description,
				Status:      child.Status,
				ParentID:    child.ParentID,
				ParentName:  child.ParentName,
				CreatedAt:   child.CreatedAt,
				UpdatedAt:   child.UpdatedAt,
				Products:    childCount,
			}
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
			Children:    children,
		}
	}

	// Step 7: Prepare pagination meta
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Search results", response, meta))
}
