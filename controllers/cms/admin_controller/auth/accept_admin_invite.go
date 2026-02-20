package admin_auth_controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AcceptAdminInvite godoc
// @Summary Accept admin invitation
// @Description Accept an admin invitation, create the admin account with password, and return admin profile
// @Tags Admin - Auth
// @Accept json
// @Produce json
// @Param request body AcceptAdminInviteRequest true "Accept invite request"
// @Success 201 {object} models.ApiResponse{data=models.AdminResponse}
// @Failure 400 {object} models.ApiResponse "Invalid request or invalid token"
// @Failure 404 {object} models.ApiResponse "Invitation not found or expired"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /admin/accept-invite [post]
func AcceptAdminInvite(c *gin.Context) {
	log.Printf("[admin.accept-invite] request")

	var req AcceptAdminInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[admin.accept-invite] validation failed: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request"))
		return
	}

	// Validate inputs
	if err := validateAcceptInviteRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, err.Error()))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Find the MOST RECENT unused invite for this email
	var invite models.AdminInvite
	if err := config.CmsGorm.WithContext(ctx).
		Where("email = ? AND used = ?", req.Email, false).
		Order("created_at DESC"). // Get the newest invite
		First(&invite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[admin.accept-invite] invite not found for %s", req.Email)
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Invitation not found or already used"))
			return
		}
		log.Printf("[admin.accept-invite] database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Verify token matches (hash the incoming token and compare)
	authService := services.GetAdminAuthService()
	tokenHash := authService.HashToken(req.Token)
	if tokenHash != invite.TokenHash {
		log.Printf("[admin.accept-invite] invalid token for %s", req.Email)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid or expired invitation token"))
		return
	}

	// Check if token is expired
	if authService.IsInviteExpired(invite.ExpiresAt) {
		log.Printf("[admin.accept-invite] token expired for %s", req.Email)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invitation has expired"))
		return
	}

	// Check if admin already exists with this email
	var existingAdmin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("email = ?", req.Email).
		First(&existingAdmin).Error; err == nil {
		log.Printf("[admin.accept-invite] admin already exists for %s", req.Email)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Admin account already exists"))
		return
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("[admin.accept-invite] database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Hash password
	passwordHash, err := authService.HashPassword(req.Password)
	if err != nil {
		log.Printf("[admin.accept-invite] password hashing failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Create admin account
	admin := &models.Admin{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: passwordHash,
		Role:         "admin", // New invites are regular admins, not super admins
		Status:       "active",
	}

	// Start transaction
	tx := config.CmsGorm.WithContext(ctx).Begin()

	// Create admin
	if err := tx.Create(admin).Error; err != nil {
		tx.Rollback()
		log.Printf("[admin.accept-invite] failed to create admin: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to create admin account"))
		return
	}

	// Mark invite as used
	if err := tx.Model(&invite).Updates(map[string]interface{}{
		"used":    true,
		"used_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("[admin.accept-invite] failed to mark invite as used: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to complete invitation"))
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("[admin.accept-invite] transaction commit failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Update last login
	now := time.Now()
	if err := config.CmsGorm.WithContext(ctx).
		Model(&admin).
		Update("last_login_at", now).Error; err != nil {
		log.Printf("[admin.accept-invite] failed to update last login: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}
	admin.LastLoginAt = &now

	// Calculate current status
	admin.Status = authService.GetAdminStatus(admin.Status, admin.LastLoginAt)

	// Generate JWT token
	token, err := services.GenerateAdminJWT(admin.ID.String(), admin.Email)
	if err != nil {
		log.Printf("[admin.accept-invite] failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// ✅ CREATE SESSION - Mark admin as online
	sessionService := services.GetAdminSessionService()
	_, err = sessionService.CreateSession(
		ctx,
		admin.ID,
		token,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		log.Printf("[admin.accept-invite] failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// ✅ SET TOKEN IN HTTP COOKIE with proper SameSite handling
	// For localhost (HTTP), use SameSite=Lax
	// For production (HTTPS), change to Secure=true and SameSite=Strict
	c.SetSameSite(http.SameSiteLaxMode) // ← Allows cross-site requests
	c.SetCookie(
		"admin_token", // cookie name
		token,         // cookie value
		24*60*60,      // max age in seconds (1 day)
		"/",           // path
		"",            // domain (empty = current domain)
		false,         // secure (false for localhost, true for HTTPS)
		true,          // httpOnly (prevent JS access)
	)
	log.Printf("[admin.accept-invite] token set in cookie with SameSite=Lax")

	log.Printf("[admin.accept-invite] admin account created: %s (%s)", admin.ID, admin.Email)

	// ✅ LOG THE ACTIVITY - Admin accepted invite and account created
	changes := map[string]interface{}{
		"email":       admin.Email,
		"name":        admin.Name,
		"role":        admin.Role,
		"status":      admin.Status,
		"created_at":  admin.JoinedAt,
		"invite_used": true,
	}
	changesJSON, _ := json.Marshal(changes)

	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      admin.ID,
		AdminEmail:   admin.Email,
		Action:       models.ActionAcceptAdminInvite,
		ResourceType: models.ResourceTypeAdminInvite,
		ResourceID:   invite.ID.String(),
		ResourceName: admin.Email,
		Changes:      datatypes.JSON(changesJSON),
		Status:       models.StatusSuccess,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[admin.accept-invite] failed to log activity: %v", err)
		// Don't fail the request, just log the error
	}

	// Return admin response (same format as /admin/me endpoint)
	c.JSON(http.StatusCreated, models.SuccessResponse(c, "Admin account created successfully", admin.ToResponse()))
}

// AcceptAdminInviteRequest represents the accept invite request
type AcceptAdminInviteRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Token    string `json:"token" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// validateAcceptInviteRequest validates the accept invite request
func validateAcceptInviteRequest(req *AcceptAdminInviteRequest) error {
	if req.Email == "" {
		return errInvalidEmail
	}

	if req.Token == "" {
		return errMissingToken
	}

	if req.Name == "" {
		return errMissingName
	}

	if len(req.Password) < 8 {
		return errPasswordTooShort
	}

	return nil
}

// Error definitions
var (
	errInvalidEmail     = gormErrorString("Invalid email format")
	errMissingToken     = gormErrorString("Token is required")
	errMissingName      = gormErrorString("Name is required")
	errPasswordTooShort = gormErrorString("Password must be at least 8 characters")
)

type gormErrorString string

func (e gormErrorString) Error() string {
	return string(e)
}
