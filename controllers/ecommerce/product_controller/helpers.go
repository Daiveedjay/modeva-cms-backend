package product_controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────

// buildStorefrontOrderClause builds the ORDER BY clause shared by handlers.
func buildStorefrontOrderClause(sortBy, sortOrder string) string {
	order := "DESC"
	if strings.ToUpper(sortOrder) == "ASC" {
		order = "ASC"
	}

	switch sortBy {
	case "price":
		return fmt.Sprintf("p.price %s", order)
	case "name":
		return fmt.Sprintf("p.name %s", order)
	case "newest":
		return fmt.Sprintf("p.created_at %s", order)
	default:
		return "p.created_at DESC"
	}
}

func parsePagination(c *gin.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "12"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	return page, limit
}

// ─────────────────────────────────────────────────────────────
// Database fetcher (THIN RESPONSE)
// ─────────────────────────────────────────────────────────────

func fetchStorefrontProductsFromDB(
	c *gin.Context,
	whereClause string,
	orderClause string,
	args []interface{},
	page int,
	limit int,
) ([]models.StorefrontProductResponse, int, error) {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	offset := (page - 1) * limit

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT p.id)
		FROM products p
		WHERE %s
	`, whereClause)

	var totalCount int64
	if err := config.CmsGorm.
		WithContext(ctx).
		Raw(countQuery, args...).
		Scan(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Data query (ONLY required fields)
	dataQuery := fmt.Sprintf(`
	SELECT 
		p.id::text AS id,
		p.name,
		p.price,
		COALESCE(p.media->'primary'->>'url', '') AS image
	FROM products p
	WHERE %s
	ORDER BY %s
	LIMIT ? OFFSET ?
`, whereClause, orderClause)

	dataArgs := append(args, limit, offset)

	products := make([]models.StorefrontProductResponse, 0)

	if err := config.CmsGorm.
		WithContext(ctx).
		Raw(dataQuery, dataArgs...).
		Scan(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, int(totalCount), nil
}
