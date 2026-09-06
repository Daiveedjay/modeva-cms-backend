package filter_controller

import (
	"net/http"
	"sync"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetFilterMetadata godoc
// @Summary Get all filter metadata
// @Description Returns availability counts, categories, and price range for storefront filters
// @Tags store
// @Produce json
// @Success 200 {object} models.ApiResponse{data=models.FilterMetadata}
// @Failure 500 {object} models.ApiResponse
// @Router /store/filters/metadata [get]
func GetFilterMetadata(c *gin.Context) {
	db := config.CmsGorm

	// Use WaitGroup to run queries concurrently for better performance
	var wg sync.WaitGroup
	var mu sync.Mutex

	metadata := &models.FilterMetadata{}
	var errs []error

	// 1. Get availability counts (variants with quantity > 0 vs = 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		availability, err := getAvailabilityCounts(db)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metadata.Availability = availability
		}
	}()

	// 2. Get all categories with subcategories
	wg.Add(1)
	go func() {
		defer wg.Done()
		categories, err := getCategoriesWithSubcategories(db)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metadata.Categories = categories
		}
	}()

	// 3. Get price range
	wg.Add(1)
	go func() {
		defer wg.Done()
		priceRange, err := getPriceRange(db)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metadata.PriceRange = priceRange
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if there were any errors
	if len(errs) > 0 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch filter metadata"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Filter metadata fetched", metadata))
}

// getAvailabilityCounts counts products with at least one variant in stock vs all variants out of stock
func getAvailabilityCounts(db *gorm.DB) (*models.AvailabilityData, error) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	query := `
		SELECT 
			COUNT(DISTINCT p.id) FILTER (
				WHERE EXISTS (
					SELECT 1 
					FROM jsonb_array_elements(p.inventory) AS item
					WHERE (item->>'quantity')::int > 0
				)
			)::int as in_stock,
			COUNT(DISTINCT p.id) FILTER (
				WHERE NOT EXISTS (
					SELECT 1 
					FROM jsonb_array_elements(p.inventory) AS item
					WHERE (item->>'quantity')::int > 0
				)
			)::int as out_of_stock
		FROM products p
		WHERE p.status = 'Active'
	`

	var data models.AvailabilityData
	err := db.WithContext(ctx).Raw(query).Scan(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// getCategoriesWithSubcategories fetches all categories and their subcategories in tree structure
func getCategoriesWithSubcategories(db *gorm.DB) ([]models.CategoryData, error) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	query := `
		SELECT 
			id::text AS id,
			name,
			parent_id::text AS parent_id
		FROM categories
		WHERE status = 'Active'
		ORDER BY created_at ASC
	`

	var allCategories []struct {
		ID       string  `gorm:"column:id"`
		Name     string  `gorm:"column:name"`
		ParentID *string `gorm:"column:parent_id"`
	}

	err := db.WithContext(ctx).Raw(query).Scan(&allCategories).Error
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]*models.CategoryData)
	var categoryList []*models.CategoryData

	// First pass: collect all categories
	for _, cat := range allCategories {
		categoryData := &models.CategoryData{
			ID:   cat.ID,
			Name: cat.Name,
		}
		if cat.ParentID != nil {
			categoryData.ParentID = *cat.ParentID
		}

		categoryMap[cat.ID] = categoryData
		categoryList = append(categoryList, categoryData)
	}

	// Second pass: build tree structure
	var parentCategories []models.CategoryData

	for _, cat := range categoryList {
		if cat.ParentID == "" {
			// This is a parent category
			category := *cat
			category.Subcategories = []models.CategoryData{}

			// Find and attach subcategories
			for _, subCat := range categoryList {
				if subCat.ParentID == cat.ID {
					category.Subcategories = append(category.Subcategories, *subCat)
				}
			}

			parentCategories = append(parentCategories, category)
		}
	}

	return parentCategories, nil
}

// getPriceRange fetches the minimum and maximum product prices
func getPriceRange(db *gorm.DB) (*models.PriceRangeData, error) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	query := `
		SELECT 
			COALESCE(MIN(price), 0)::float8 as min,
			COALESCE(MAX(price), 1000)::float8 as max
		FROM products
		WHERE status = 'Active'
			AND price > 0
	`

	var priceRange models.PriceRangeData
	err := db.WithContext(ctx).Raw(query).Scan(&priceRange).Error
	if err != nil {
		return nil, err
	}

	return &priceRange, nil
}
