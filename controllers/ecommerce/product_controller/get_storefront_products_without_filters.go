package product_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetStorefrontProductsBasic godoc
// @Summary Get storefront products (no filters)
// @Description Retrieve active storefront products using only pagination and sorting (no category, size, colour, or price filters).
// @Tags store
// @Produce json
// @Param sortBy query string false "Sort by field (newest, price, etc.)" default(newest)
// @Param sortOrder query string false "Sort order (asc | desc)" default(desc)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(12)
// @Success 200 {object} models.ApiResponse "Products fetched successfully"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /store/products/basic [get]
func getStorefrontProductsWithoutFilters(c *gin.Context) {
	page, limit := parsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "newest")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	whereClause := "p.status = 'Active'"
	orderClause := buildStorefrontOrderClause(sortBy, sortOrder)

	products, totalCount, err := fetchStorefrontProductsFromDB(
		c,
		whereClause,
		orderClause,
		nil, // no filter args
		page,
		limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch products"))
		return
	}

	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, models.PaginatedResponse(
		c,
		"Products fetched successfully",
		products,
		&models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      totalCount,
			TotalPages: totalPages,
		},
	))
}
