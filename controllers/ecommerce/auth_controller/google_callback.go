// ════════════════════════════════════════════════════════════
// Path: controllers/store/auth_controller/google_callback.go
// Google OAuth Callback Handler
// ════════════════════════════════════════════════════════════

package auth_controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/utils"
	"github.com/gin-gonic/gin"
)

// GoogleCallback godoc
// @Summary Google OAuth callback
// @Description Handles the callback from Google OAuth. Verifies the state token, exchanges the authorization code, retrieves user info, creates/updates the user in the database, issues a JWT cookie, and redirects the user back to the frontend.
// @Tags Auth - Google OAuth
// @Produce json
// @Success 307 "Redirect to frontend after successful login"
// @Failure 400 {object} models.ApiResponse "Invalid state or missing authorization code"
// @Failure 401 {object} models.ApiResponse "Unauthorized or token exchange failure"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /auth/google/callback [get]
func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	savedState, err := c.Cookie("oauth_state")
	if err != nil || state != savedState {
		log.Printf("❌ State mismatch")
		redirectToFrontendWithError(c, "Invalid state token")
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	code := c.Query("code")
	if code == "" {
		log.Printf("❌ No code")
		redirectToFrontendWithError(c, "No authorization code")
		return
	}

	log.Printf("🔄 Exchanging code for token...")
	token, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("❌ Exchange failed: %v", err)
		redirectToFrontendWithError(c, "Failed to exchange token")
		return
	}

	log.Printf("🔄 Getting user info...")
	client := config.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("❌ Failed to get user info: %v", err)
		redirectToFrontendWithError(c, "Failed to get user info")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response: %v", err)
		redirectToFrontendWithError(c, "Failed to read user info")
		return
	}

	log.Println("════════════════════════════════════════════════════════")
	log.Println("📋 FULL RAW GOOGLE API RESPONSE:")
	log.Println(string(body))
	log.Println("════════════════════════════════════════════════════════")

	var googleUser models.GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		log.Printf("❌ Decode failed: %v", err)
		redirectToFrontendWithError(c, "Failed to decode user info")
		return
	}

	googleID := googleUser.Sub
	if googleID == "" {
		googleID = googleUser.ID
	}

	if googleID == "" {
		log.Printf("❌ No Google ID")
		redirectToFrontendWithError(c, "Google ID not found")
		return
	}

	emailVerified := googleUser.EmailVerified || googleUser.VerifiedEmail
	log.Printf("🔍 Email verification - EmailVerified: %v, VerifiedEmail: %v, Final: %v",
		googleUser.EmailVerified, googleUser.VerifiedEmail, emailVerified)
	log.Printf("✅ Got user: %s (Google ID: %s, Verified: %v)", googleUser.Email, googleID, emailVerified)

	user, err := createOrUpdateUser(c, &googleUser, googleID, emailVerified)
	if err != nil {
		log.Printf("❌ DB error: %v", err)
		redirectToFrontendWithError(c, fmt.Sprintf("Database error: %v", err))
		return
	}

	// Log login event
	if err := utils.LogLoginEvent(c, user.ID); err != nil {
		log.Printf("⚠️  Failed to log login event: %v", err)
	}

	// Generate JWT token
	jwtToken, err := utils.GenerateJWT(user.ID, user.Email, user.Name)
	if err != nil {
		log.Printf("❌ JWT error: %v", err)
		redirectToFrontendWithError(c, "Failed to generate token")
		return
	}

	// Log env and computed frontend URL before redirect
	frontendEnv := os.Getenv("ECOMMERCE_FRONTEND_URL")
	chosenFrontend := config.GetFrontendURL()

	log.Printf("🔎 Backend ENV FRONTEND URL: %q", frontendEnv)
	log.Printf("➡️ Backend will redirect to frontend URL: %q", chosenFrontend)

	// Set HTTP-only cookie with the token
	isProd := os.Getenv("ENV") == "production"
	c.SetCookie(
		"auth_token",
		jwtToken,
		24*60*60, // 24 hours
		"/",
		"",
		isProd,
		true, // httpOnly
	)

	log.Println("✓ Auth cookie set")

	// Prepare user response
	userResponse := models.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Phone:         user.Phone,
		Provider:      user.Provider,
		EmailVerified: user.EmailVerified,
		Avatar:        user.Avatar,
		CreatedAt:     user.CreatedAt,
	}

	// Set temporary cookie with user data (for popup to read)
	userJSON, _ := json.Marshal(userResponse)
	c.SetCookie(
		"user_data",
		string(userJSON),
		60, // 1 minute (just for transfer)
		"/",
		"",
		isProd,
		false, // NOT httpOnly (popup needs to read it)
	)

	log.Printf("✅ Login successful: %s (verified: %v)", user.Email, emailVerified)

	// Redirect to frontend callback (NO token in URL)
	frontendURL := config.GetFrontendURL()

	redirectURL := fmt.Sprintf("%s/auth-popup", frontendURL)

	log.Printf("Using redirect URI: %s", config.GoogleOAuthConfig.RedirectURL)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
