package admin_auth_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminLogout godoc
// @Summary Logout admin
// @Description Logout the current admin and deactivate session
// @Tags Admin - Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse
// @Router /admin/logout [post]
func AdminLogout(c *gin.Context) {
	adminIDStr, exists := c.Get("adminID")
	if exists {
		log.Printf("[admin.logout] admin logging out: %s", adminIDStr)

		// ✅ DEACTIVATE SESSION
		ctx, cancel := config.WithTimeout()
		defer cancel()

		adminID, err := uuid.Parse(adminIDStr.(string))
		if err == nil {
			sessionService := services.GetAdminSessionService()
			if err := sessionService.DeactivateSession(ctx, adminID); err != nil {
				log.Printf("[admin.logout] failed to deactivate session: %v", err)
				// Don't fail the logout even if session deactivation fails
			}
		}
	}

	// ✅ CLEAR TOKEN COOKIE
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"admin_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
	log.Printf("[admin.logout] token cleared from cookie")

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Logout successful", nil))
}
