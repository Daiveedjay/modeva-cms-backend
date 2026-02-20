// ════════════════════════════════════════════════════════════
// Path: controllers/ecommerce/user_controller/profile_controller/update_profile.go
// Update authenticated user's profile
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

// UpdateProfileRequest represents the update profile request
type UpdateProfileRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=2,max=255"`
	Phone *string `json:"phone" binding:"omitempty,min=10,max=20"`
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile (name, phone)
// @Tags User - Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Update request"
// @Success 200 {object} models.ApiResponse{data=models.UserResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 401 {object} models.ApiResponse
// @Router /user/profile [patch]
func UpdateProfile(c *gin.Context) {
	// Get user ID from JWT
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

	// Parse request
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request: "+err.Error()))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Find user
	var user models.User
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ? AND status = ?", userID, "active").
		First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "User not found"))
		return
	}

	// Build update map with only provided fields
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "No fields to update"))
		return
	}

	// Perform update (GORM will automatically update updated_at)
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&user).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update profile"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Profile updated", user.ToResponse()))
}
