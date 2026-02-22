// controllers/store/auth_controller/logout.go
package auth_controller

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Logout godoc
// @Summary Logout user
// @Description Logs out the authenticated user by clearing the auth_token and user_data cookies.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Logged out"
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	isProd := os.Getenv("APP_ENV") == "production"
	cookieDomain := ""
	if isProd {
		cookieDomain = ".modeva.shop"
	}

	// SameSite=None requires Secure=true (HTTPS), only valid in production
	// Locally over HTTP, use Lax so cookies are cleared by the browser
	if isProd {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie("auth_token", "", -1, "/", cookieDomain, isProd, true)
	c.SetCookie("user_data", "", -1, "/", cookieDomain, isProd, false)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
