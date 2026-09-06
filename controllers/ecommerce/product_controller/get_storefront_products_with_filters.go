package product_controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetStorefrontProducts godoc
// @Summary Get storefront products with filters
// @Description Retrieve active storefront products with optional search, category, subcategory, size, colour, availability, price range, and sorting filters.
// @Tags Storefront - Products
// @Produce json
// @Param q query string false "Search query (name or description)"
// @Param category query []string false "Parent category names (repeatable)"
// @Param subcategory query []string false "Subcategory IDs (repeatable)"
// @Param style query string false "Style filter (subcategory name)"
// @Param size query []string false "Sizes (repeatable)"
// @Param color query []string false "Colours (repeatable)"
// @Param availability query string false "Availability filter (in_stock | out_of_stock)"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param sortBy query string false "Sort by field (newest, price, etc.)" default(newest)
// @Param sortOrder query string false "Sort order (asc | desc)" default(desc)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(12)
// @Success 200 {object} models.ApiResponse "Products with filters fetched successfully"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /store/products [get]
func getStorefrontProductsWithFilters(c *gin.Context) {
	page, limit := parsePagination(c)

	log.Printf("=== FILTER DEBUG START ===")
	log.Printf("Page: %d, Limit: %d", page, limit)

	// Parse filters
	searchQuery := c.Query("q")
	categoryNames := c.QueryArray("category")
	subcategoryIDs := c.QueryArray("subcategory")
	style := c.Query("style")
	sizes := c.QueryArray("size")
	colors := c.QueryArray("color")
	availability := c.Query("availability")
	minPriceStr := c.Query("minPrice")
	maxPriceStr := c.Query("maxPrice")
	sortBy := c.DefaultQuery("sortBy", "newest")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	conditions := []string{"p.status = 'Active'"}
	args := []interface{}{}

	// Search query (name or description)
	if searchQuery != "" {
		conditions = append(conditions, "(p.name ILIKE ? OR p.description ILIKE ?)")
		args = append(args, "%"+searchQuery+"%", "%"+searchQuery+"%")
		log.Printf("Added search condition")
	}

	// Style filter (subcategory name match)
	if style != "" {
		cond := `p.sub_category_id IN (
			SELECT id FROM categories 
			WHERE LOWER(name) = LOWER(?) AND parent_id IS NOT NULL
		)`
		conditions = append(conditions, cond)
		args = append(args, strings.TrimSpace(style))
		log.Printf("Added style filter: %s", style)
	}

	// Category filter (parent categories by NAME - multiple)
	if len(categoryNames) > 0 {
		// Build LOWER() placeholders for case-insensitive matching
		lowerPlaceholders := make([]string, len(categoryNames))
		for i, name := range categoryNames {
			lowerPlaceholders[i] = "LOWER(?)"
			args = append(args, strings.TrimSpace(name))
		}
		// Find parent category IDs by name, then find all subcategories under those parents
		cond := fmt.Sprintf(
			`p.sub_category_id IN (
				SELECT id FROM categories 
				WHERE parent_id IN (
					SELECT id FROM categories 
					WHERE LOWER(name) IN (%s) AND parent_id IS NULL
				)
			)`,
			strings.Join(lowerPlaceholders, ","),
		)
		conditions = append(conditions, cond)
		log.Printf("Added category filter by names: %v", categoryNames)
	}

	// Subcategory filter (multiple subcategory IDs)
	if len(subcategoryIDs) > 0 {
		placeholders := make([]string, len(subcategoryIDs))
		for i := range subcategoryIDs {
			placeholders[i] = "?"
			args = append(args, subcategoryIDs[i])
		}
		cond := fmt.Sprintf(
			"p.sub_category_id IN (%s)",
			strings.Join(placeholders, ","),
		)
		conditions = append(conditions, cond)
		log.Printf("Added subcategory filter: %v", subcategoryIDs)
	}

	// Size filter
	if len(sizes) > 0 {
		log.Printf("Building size filter for: %v", sizes)
		sizePlaceholders := make([]string, len(sizes))
		for i, size := range sizes {
			sizePlaceholders[i] = "?"
			args = append(args, strings.TrimSpace(size))
		}
		cond := fmt.Sprintf(
			`EXISTS (
				SELECT 1 
				FROM jsonb_array_elements(p.variants) AS variant,
				     jsonb_array_elements_text(variant->'options') AS size_option
				WHERE variant->>'type' = 'Size' 
				  AND size_option IN (%s)
			)`,
			strings.Join(sizePlaceholders, ","),
		)
		conditions = append(conditions, cond)
		log.Printf("Added size filter")
	}

	// Color filter
	if len(colors) > 0 {
		log.Printf("Building color filter for: %v", colors)
		colorPlaceholders := make([]string, len(colors))
		for i, color := range colors {
			colorPlaceholders[i] = "?"
			args = append(args, strings.TrimSpace(color))
		}
		cond := fmt.Sprintf(
			`EXISTS (
				SELECT 1 
				FROM jsonb_array_elements(p.variants) AS variant,
				     jsonb_array_elements_text(variant->'options') AS color_option
				WHERE variant->>'type' = 'Color' 
				  AND color_option IN (%s)
			)`,
			strings.Join(colorPlaceholders, ","),
		)
		conditions = append(conditions, cond)
		log.Printf("Added color filter")
	}

	// Availability filter
	switch availability {
	case "in_stock", "inStock":
		cond := `EXISTS (
			SELECT 1 
			FROM jsonb_array_elements(p.inventory) AS item
			WHERE (item->>'quantity')::int > 0
		)`
		conditions = append(conditions, cond)
		log.Printf("Added availability condition: in_stock")
	case "out_of_stock", "outOfStock":
		cond := `NOT EXISTS (
			SELECT 1 
			FROM jsonb_array_elements(p.inventory) AS item
			WHERE (item->>'quantity')::int > 0
		)`
		conditions = append(conditions, cond)
		log.Printf("Added availability condition: out_of_stock")
	}

	// Price range filter
	if minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			conditions = append(conditions, "p.price >= ?")
			args = append(args, minPrice)
			log.Printf("Added minPrice condition = %.2f", minPrice)
		}
	}
	if maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			conditions = append(conditions, "p.price <= ?")
			args = append(args, maxPrice)
			log.Printf("Added maxPrice condition = %.2f", maxPrice)
		}
	}

	whereClause := strings.Join(conditions, " AND ")
	orderClause := buildStorefrontOrderClause(sortBy, sortOrder)

	products, totalCount, err := fetchStorefrontProductsFromDB(
		c,
		whereClause,
		orderClause,
		args,
		page,
		limit,
	)
	if err != nil {
		log.Printf("ERROR in fetchStorefrontProductsFromDB: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch products"))
		return
	}

	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Products with filters fetched successfully",
		products,
		&models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      totalCount,
			TotalPages: totalPages,
		},
	))
}
