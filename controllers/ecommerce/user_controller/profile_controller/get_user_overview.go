package profile_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserOverview godoc
// @Summary Get user overview
// @Description Get comprehensive user overview including profile, purchase stats, loyalty points, and recent orders
// @Tags User - Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=models.UserOverviewResponse}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 404 {object} models.ApiResponse "User not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/overview [get]
func GetUserOverview(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized"))
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Invalid user ID"))
		return
	}

	var overview models.UserOverviewResponse

	// =====================================
	// 1. Fetch user profile
	// =====================================
	var user models.User
	if err := config.EcommerceGorm.
		Select("id, name, email, avatar, phone, created_at").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "User not found"))
		return
	}

	overview.Profile = models.UserProfileSummary{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Phone:    user.Phone,
		JoinedAt: user.CreatedAt,
	}

	// =====================================
	// 2. Fetch purchase statistics
	// =====================================
	type PurchaseStats struct {
		TotalOrders     int     `json:"total_orders"`
		TotalSpent      float64 `json:"total_spent"`
		CompletedOrders int     `json:"completed_orders"`
	}

	var stats PurchaseStats
	err = config.EcommerceGorm.Raw(`
		SELECT
			COUNT(*)::int AS total_orders,
			COALESCE(SUM(CASE WHEN status IN ('completed', 'delivered') THEN total_amount ELSE 0 END), 0)::float8 AS total_spent,
			COUNT(CASE WHEN status IN ('completed', 'delivered') THEN 1 END)::int AS completed_orders
		FROM orders
		WHERE user_id = $1
	`, userID).Scan(&stats).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch purchase stats"))
		return
	}

	// Calculate loyalty points (2.5% of total spent)
	loyaltyPoints := int(stats.TotalSpent * 0.025)

	overview.TotalPurchases = stats.TotalSpent
	overview.TotalOrders = stats.TotalOrders
	overview.CompletedOrders = stats.CompletedOrders
	overview.LoyaltyPoints = loyaltyPoints

	// =====================================
	// 3. Fetch recent orders (last 3)
	// =====================================
	recentOrders := []models.RecentOrderSummary{}
	err = config.EcommerceGorm.
		Table("orders").
		Select(`
			id,
			order_number,
			total_amount,
			status,
			created_at
		`).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(3).
		Scan(&recentOrders).Error
	if err != nil {
		// Not critical, just set empty array
		recentOrders = []models.RecentOrderSummary{}
	}

	overview.RecentOrders = recentOrders

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"User overview retrieved successfully",
		overview,
	))
}
