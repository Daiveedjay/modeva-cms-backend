package admin_auth_controller

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreateAdminInvite godoc
// @Summary Create admin invite (Super admin only)
// @Description Generate and send an invite email to become an admin. Super admin only.
// @Tags Admin - Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param inviteRequest body models.CreateAdminInviteRequest true "Email to invite"
// @Success 201 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse "Invalid request or already invited"
// @Failure 403 {object} models.ApiResponse "Super admin access required"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /admin/invites [post]
func CreateAdminInvite(c *gin.Context) {
	log.Printf("[admin.invites.create] request")

	// Middleware already checked super_admin, but we double-check
	adminRole, exists := c.Get("adminRole")
	if !exists || adminRole != "super_admin" {
		log.Printf("[admin.invites.create] unauthorized - not super admin")
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Super admin access required"))
		return
	}

	var req models.CreateAdminInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Check if email already has an admin account
	var existingAdmin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("email = ?", req.Email).
		First(&existingAdmin).Error; err == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Admin with this email already exists"))
		return
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("[admin.invites.create] database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// TODO: Re-enable this check after email template is finalized
	// Check if email already has a pending invite
	// var existingInvite models.AdminInvite
	// if err := config.CmsGorm.WithContext(ctx).
	// 	Where("email = ? AND used = ?", req.Email, false).
	// 	First(&existingInvite).Error; err == nil {
	// 	c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "This email already has a pending invite"))
	// 	return
	// } else if err != gorm.ErrRecordNotFound {
	// 	log.Printf("[admin.invites.create] database error: %v", err)
	// 	c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
	// 	return
	// }

	// Generate invite token
	authService := services.GetAdminAuthService()
	token, err := authService.GenerateInviteToken()
	if err != nil {
		log.Printf("[admin.invites.create] failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	tokenHash := authService.HashToken(token)
	expiresAt := authService.GetInviteTokenExpiration()

	// Create invite record
	invite := models.AdminInvite{
		Email:     req.Email,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&invite).Error; err != nil {
		log.Printf("[admin.invites.create] failed to create invite: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to create invite"))
		return
	}

	adminIDStr, _ := c.Get("adminID")
	adminEmail, _ := c.Get("adminEmail")
	log.Printf("[admin.invites.create] invite created by %s for %s (expires: %v)", adminIDStr, req.Email, expiresAt)

	// ✅ LOG THE ACTIVITY - Admin invite created
	changes := map[string]interface{}{
		"email":      req.Email,
		"expires_at": expiresAt,
	}
	changesJSON, _ := json.Marshal(changes)

	adminID, _ := uuid.Parse(adminIDStr.(string))
	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      adminID,
		AdminEmail:   adminEmail.(string),
		Action:       models.ActionCreateAdminInvite,
		ResourceType: models.ResourceTypeAdminInvite,
		ResourceID:   invite.ID.String(),
		ResourceName: req.Email,
		Changes:      datatypes.JSON(changesJSON),
		Status:       models.StatusSuccess,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[admin.invites.create] failed to log activity: %v", err)
		// Don't fail the request, just log the error
	}

	// Send invitation email via Resend
	go sendAdminInviteEmail(req.Email, token)

	c.JSON(http.StatusCreated, models.SuccessResponse(c, "Invite created and email sent", map[string]interface{}{
		"email":   req.Email,
		"expires": expiresAt,
	}))
}

// sendAdminInviteEmail sends the invitation email (async)
func sendAdminInviteEmail(email string, token string) {
	resendClient := services.NewResendClient()
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // Dev default
	}

	inviteLink := frontendURL + "/accept-invite?email=" + email + "&token=" + token

	emailData := services.AdminInviteEmailData{
		AdminName:  email, // Will be updated when they accept invite
		AdminEmail: email,
		InviteLink: inviteLink,
	}

	if err := resendClient.SendAdminInviteEmail(emailData); err != nil {
		log.Printf("[admin.invites.create] failed to send email to %s: %v", email, err)
		// Don't fail the request - invite is already created
		// In production, you might want to queue this for retry
	} else {
		log.Printf("[admin.invites.create] invitation email sent to %s", email)
	}
}
