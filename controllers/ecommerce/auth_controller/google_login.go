// Path: controllers/store/auth_controller/google_login.go

package auth_controller

import (
	"log"
	"net/http"
	"os"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GoogleLogin godoc
// @Summary Redirect to Google OAuth
// @Description Starts the Google OAuth flow by generating a state token, storing it in a secure cookie, and redirecting the user to Google's OAuth consent page.
// @Tags Auth - Google OAuth
// @Produce json
// @Success 307 "Temporary redirect to Google OAuth"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /auth/google/login [get]
func GoogleLogin(c *gin.Context) {
	isProd := os.Getenv("APP_ENV") == "production"
	cookieDomain := ""
	if isProd {
		cookieDomain = ".modeva.shop"
	}

	state := uuid.New().String()

	// Lax is fine here - Google redirect back is a top-level navigation
	// Set SameSite based on environment
	if isProd {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("oauth_state", state, 3600, "/", cookieDomain, isProd, true)

	url := config.GoogleOAuthConfig.AuthCodeURL(state)
	log.Printf("🔗 Redirecting to Google OAuth")

	c.Redirect(http.StatusTemporaryRedirect, url)
}
