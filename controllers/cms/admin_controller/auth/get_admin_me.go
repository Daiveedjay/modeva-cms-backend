package admin_auth_controller

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

// GetAdminMe godoc
// @Summary Get current admin profile
// @Description Returns the current logged-in admin's profile. Used to check if admin is authenticated on page reload
// @Tags Admin - Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=models.AdminResponse}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Account suspended"
// @Router /admin/me [get]
func GetAdminMe(c *gin.Context) {
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

	ctx, cancel := config.WithTimeout()
	defer cancel()

	var admin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("id = ?", adminID).
		First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Admin not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	// Check if admin is suspended
	if admin.Status == "suspended" {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Admin account is suspended"))
		return
	}

	// Calculate current status (active/inactive)
	authService := services.GetAdminAuthService()
	admin.Status = authService.GetAdminStatus(admin.Status, admin.LastLoginAt)

	log.Printf("[admin.me] retrieved: %s", admin.Email)
	c.JSON(http.StatusOK, models.SuccessResponse(c, "Admin profile retrieved", admin.ToResponse()))
}
