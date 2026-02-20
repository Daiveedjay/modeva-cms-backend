// Path: controllers/store/auth_controller/google_login.go

package auth_controller

import (
	"log"
	"net/http"

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
	// Generate state token
	state := uuid.New().String()

	log.Printf("🔐 Setting state cookie: %s", state)

	// Set cookie with better settings
	c.SetCookie(
		"oauth_state", // name
		state,         // value
		3600,          // maxAge (1 hour)
		"/",           // path
		"",            // domain (empty = current domain)
		false,         // secure (false for localhost)
		true,          // httpOnly
	)

	// Also set as SameSite=Lax for better compatibility
	c.SetSameSite(http.SameSiteLaxMode)

	// Generate OAuth URL
	url := config.GoogleOAuthConfig.AuthCodeURL(state)

	log.Printf("🔗 Redirecting to: %s", url)
	log.Printf("🍪 State cookie should be set: %s", state)

	// Redirect to Google
	c.Redirect(http.StatusTemporaryRedirect, url)
}
