package admin_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAdmin godoc
// @Summary Get admin details
// @Description Get details of a specific admin
// @Tags Admin - Management
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID"
// @Success 200 {object} models.ApiResponse{data=models.AdminResponse}
// @Failure 404 {object} models.ApiResponse "Admin not found"
// @Router /admin/admins/:id [get]
func GetAdmin(c *gin.Context) {
	adminID := c.Param("id")
	log.Printf("[admin.get] request for admin: %s", adminID)

	ctx, cancel := config.WithTimeout()
	defer cancel()

	var admin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("id = ?", adminID).
		First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Admin not found"))
		} else {
			log.Printf("[admin.get] database error: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		}
		return
	}

	// Calculate status
	authService := services.GetAdminAuthService()
	admin.Status = authService.GetAdminStatus(admin.Status, admin.LastLoginAt)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Admin retrieved", admin.ToResponse()))
}
