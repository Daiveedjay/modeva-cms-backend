package admin_controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UnsuspendAdmin godoc
// @Summary Unsuspend admin (Super admin only)
// @Description Unsuspend a suspended admin account. Super admin only.
// @Tags Admin - Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID to unsuspend"
// @Success 200 {object} models.ApiResponse
// @Failure 403 {object} models.ApiResponse "Super admin access required"
// @Failure 404 {object} models.ApiResponse "Admin not found"
// @Router /admin/admins/:id/unsuspend [post]
func UnsuspendAdmin(c *gin.Context) {
	adminIDStr, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized"))
		return
	}

	// Get admin email for logging
	adminEmail, _ := c.Get("adminEmail")

	// Middleware checked super_admin, but we double-check
	adminRole, _ := c.Get("adminRole")
	if adminRole != "super_admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Super admin access required"))
		return
	}

	adminID := c.Param("id")
	log.Printf("[admin.unsuspend] request: %s by %s", adminID, adminIDStr)

	ctx, cancel := config.WithTimeout()
	defer cancel()

	var admin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("id = ?", adminID).
		First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Admin not found"))
		} else {
			log.Printf("[admin.unsuspend] database error: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		}
		return
	}

	// Store old status for audit log
	oldStatus := admin.Status

	// Update status
	if err := config.CmsGorm.WithContext(ctx).
		Model(&admin).
		Update("status", "active").Error; err != nil {
		log.Printf("[admin.unsuspend] failed to unsuspend: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// ✅ LOG THE ACTIVITY
	changes := map[string]interface{}{
		"before": map[string]interface{}{"status": oldStatus},
		"after":  map[string]interface{}{"status": "active"},
	}
	changesJSON, _ := json.Marshal(changes)

	adminIDUUID, _ := uuid.Parse(adminIDStr.(string))
	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      adminIDUUID,
		AdminEmail:   adminEmail.(string),
		Action:       models.ActionUnsuspendAdmin,
		ResourceType: models.ResourceTypeAdmin,
		ResourceID:   adminID,
		ResourceName: admin.Email,
		Changes:      datatypes.JSON(changesJSON),
		Status:       models.StatusSuccess,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[admin.unsuspend] failed to log activity: %v", err)
		// Don't fail the request, just log the error
	}

	log.Printf("[admin.unsuspend] success: %s unsuspended by %s", adminID, adminIDStr)
	c.JSON(http.StatusOK, models.SuccessResponse(c, "Admin unsuspended", nil))
}
