package category_controller

import (
	"math"
	"net/http"
	"strconv"

	category_cache "github.com/Modeva-Ecommerce/modeva-cms-backend/cache"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCategories godoc
// @Summary Get paginated categories with subcategories
// @Description Retrieve parent categories and their subcategories with pagination and product counts
// @Tags CMS - Categories
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} models.ApiResponse
// @Router /api/v1/admin/categories [get]
func GetCategories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Try cache first
	parents, productCounts, ok := category_cache.GetTree()
	if !ok {
		// Cache miss — fetch from DB
		if err := config.CmsGorm.
			Where("parent_id IS NULL").
			Order("created_at ASC").
			Preload("Children", func(db *gorm.DB) *gorm.DB {
				return db.Order("created_at ASC")
			}).
			Find(&parents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch categories"))
			return
		}

		// Collect all IDs for product count query
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

	// Apply pagination in-memory (data is already cached)
	total := int64(len(parents))
	start := offset
	end := offset + limit
	if start > len(parents) {
		start = len(parents)
	}
	if end > len(parents) {
		end = len(parents)
	}
	paginated := parents[start:end]

	// Build response
	response := make([]models.CategoryWithProducts, len(paginated))
	for i, parent := range paginated {
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

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(c, "Categories fetched", response, meta))
}
