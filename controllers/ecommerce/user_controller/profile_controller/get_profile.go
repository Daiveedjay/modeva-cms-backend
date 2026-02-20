// ════════════════════════════════════════════════════════════
// Path: controllers/ecommerce/user_controller/profile_controller/get_profile.go
// Get authenticated user's profile
// ════════════════════════════════════════════════════════════

package profile_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's profile information
// @Tags User - Profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.ApiResponse{data=models.UserResponse}
// @Failure 401 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /user/profile [get]
func GetProfile(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized"))
		return
	}

	// Parse userID to UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Invalid user ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Fetch user from database
	var user models.User
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ? AND status = ?", userID, "active").
		First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "User not found"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Profile fetched", user.ToResponse()))
}
