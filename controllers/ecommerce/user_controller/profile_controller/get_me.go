package profile_controller

import (
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMe godoc
// @Summary Get current authenticated user
// @Description Check authentication status and return basic user info
// @Tags User - Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=models.UserResponse}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Router /user/me [get]
func GetMe(c *gin.Context) {
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

	// Fetch user from database (include status because we validate it below)
	var user models.User
	if err := config.EcommerceGorm.
		Select("id, name, email, avatar, phone, provider, email_verified, created_at, status").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "User not found"))
		return
	}

	// Check if user is active
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Account is not active"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Authenticated",
		user.ToResponse(),
	))
}
