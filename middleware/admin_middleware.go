package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware validates JWT token and checks admin authorization
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie first, then Authorization header
		token, err := c.Cookie("admin_token")
		if err != nil || token == "" {
			// Try Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized - no token provided"))
				c.Abort()
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized - invalid token format"))
				c.Abort()
				return
			}
			token = parts[1]
		}

		// Validate and parse JWT
		claims, err := services.VerifyAdminJWT(token)
		if err != nil {
			log.Printf("[auth] invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized - invalid token"))
			c.Abort()
			return
		}

		ctx, cancel := config.WithTimeout()
		defer cancel()

		// ✅ UPDATE SESSION ACTIVITY
		sessionService := services.GetAdminSessionService()
		authService := services.GetAdminAuthService()
		tokenHash := authService.HashToken(token)

		if err := sessionService.UpdateSessionActivity(ctx, tokenHash); err != nil {
			log.Printf("[auth] failed to update session activity: %v", err)
			// Don't abort - session update failure shouldn't block the request
		}

		// Set admin info in context
		c.Set("adminID", claims.AdminID)
		c.Set("adminEmail", claims.Email)

		// Fetch admin role from database
		var admin models.Admin
		if err := config.CmsGorm.WithContext(ctx).
			Select("role").
			Where("id = ?", claims.AdminID).
			First(&admin).Error; err != nil {
			log.Printf("[auth] failed to fetch admin role: %v", err)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized - admin not found"))
			c.Abort()
			return
		}

		c.Set("adminRole", admin.Role)

		c.Next()
	}
}

// RequireSuperAdminMiddleware checks if the admin is a super admin
func RequireSuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminRole, exists := c.Get("adminRole")
		if !exists {
			c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Forbidden - role not found"))
			c.Abort()
			return
		}

		if adminRole != "super_admin" {
			log.Printf("[auth] non-super-admin attempted restricted action")
			c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Forbidden - super admin access required"))
			c.Abort()
			return
		}

		c.Next()
	}
}
