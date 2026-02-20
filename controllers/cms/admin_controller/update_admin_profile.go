package admin_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UpdateAdminProfile godoc
// @Summary Update admin profile
// @Description Update current admin's profile information (name, phone, country, avatar)
// @Tags Admin - Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updateRequest body models.UpdateAdminProfileRequest true "Profile update"
// @Success 200 {object} models.ApiResponse{data=models.AdminResponse}
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Router /admin/profile [put]
func UpdateAdminProfile(c *gin.Context) {
	adminIDStr, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized"))
		return
	}

	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Invalid admin ID"))
		return
	}

	var req models.UpdateAdminProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Find admin
	var admin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("id = ?", adminID).
		First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Admin not found"))
		} else {
			log.Printf("[admin.update-profile] database error: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		}
		return
	}

	// Update fields if provided
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "No fields to update"))
		return
	}

	if err := config.CmsGorm.WithContext(ctx).
		Model(&admin).
		Updates(updates).Error; err != nil {
		log.Printf("[admin.update-profile] failed to update: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Reload admin
	config.CmsGorm.WithContext(ctx).First(&admin, "id = ?", adminID)

	authService := services.GetAdminAuthService()
	admin.Status = authService.GetAdminStatus(admin.Status, admin.LastLoginAt)

	log.Printf("[admin.update-profile] success: %s", adminID)
	c.JSON(http.StatusOK, models.SuccessResponse(c, "Profile updated", admin.ToResponse()))
}
