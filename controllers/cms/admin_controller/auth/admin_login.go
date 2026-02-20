package admin_auth_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminLogin godoc
// @Summary Login as admin
// @Description Authenticate admin with email and password. Returns JWT token and creates session
// @Tags Admin - Auth
// @Accept json
// @Produce json
// @Param loginRequest body models.AdminLoginRequest true "Email and password"
// @Success 200 {object} models.ApiResponse{data=models.AdminLoginResponse}
// @Failure 400 {object} models.ApiResponse "Invalid credentials"
// @Failure 403 {object} models.ApiResponse "Account suspended"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /admin/login [post]
func AdminLogin(c *gin.Context) {
	log.Printf("[admin.login] attempt")

	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Find admin by email
	var admin models.Admin
	if err := config.CmsGorm.WithContext(ctx).
		Where("email = ?", req.Email).
		First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[admin.login] user not found: %s", req.Email)
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid email or password"))
		} else {
			log.Printf("[admin.login] database error: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		}
		return
	}

	// Check if suspended
	if admin.Status == "suspended" {
		log.Printf("[admin.login] suspended account attempt: %s", req.Email)
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Account is suspended"))
		return
	}

	// Verify password
	authService := services.GetAdminAuthService()
	if !authService.VerifyPassword(admin.PasswordHash, req.Password) {
		log.Printf("[admin.login] invalid password: %s", req.Email)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid email or password"))
		return
	}

	// Update last login
	now := time.Now()
	if err := config.CmsGorm.WithContext(ctx).
		Model(&admin).
		Update("last_login_at", now).Error; err != nil {
		log.Printf("[admin.login] failed to update last login: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}
	admin.LastLoginAt = &now

	// Calculate current status
	admin.Status = authService.GetAdminStatus(admin.Status, admin.LastLoginAt)

	// Generate JWT token
	token, err := services.GenerateAdminJWT(admin.ID.String(), admin.Email)
	if err != nil {
		log.Printf("[admin.login] failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// ✅ CREATE SESSION
	sessionService := services.GetAdminSessionService()
	_, err = sessionService.CreateSession(
		ctx,
		admin.ID,
		token,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		log.Printf("[admin.login] failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// ✅ SET TOKEN IN HTTP COOKIE
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"admin_token",
		token,
		24*60*60,
		"/",
		"",
		false,
		true,
	)
	log.Printf("[admin.login] token set in cookie with SameSite=Lax")

	log.Printf("[admin.login] success: %s (%s)", admin.Email, admin.ID)

	response := models.AdminLoginResponse{
		Admin: admin.ToResponse(),
		Token: token,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Login successful", response))
}
