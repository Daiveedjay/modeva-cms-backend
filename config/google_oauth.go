// ════════════════════════════════════════════════════════════
// Path: config/google_oauth.go
// Google OAuth Configuration
// ════════════════════════════════════════════════════════════

package config

import (
	"context"
	"log"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	GoogleOAuthConfig *oauth2.Config
	OIDCVerifier      *oidc.IDTokenVerifier
)

// InitGoogleOAuth initializes Google OAuth configuration
func InitGoogleOAuth() {
	ctx := context.Background()

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" {
		log.Fatal("❌ GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set in .env")
	}

	if redirectURL == "" {
		redirectURL = "http://localhost:8081/api/v1/auth/google/callback"
		log.Printf("⚠️  GOOGLE_REDIRECT_URL not set, using default: %s", redirectURL)
	}

	// Configure OAuth2
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Setup OIDC provider for ID token verification (Google One Tap)
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		log.Fatalf("❌ Failed to create OIDC provider: %v", err)
	}

	OIDCVerifier = provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	log.Println("✅ Google OAuth initialized successfully")
}

// GetFrontendURL returns frontend URL from environment
func GetFrontendURL() string {
	urlFromEnv := os.Getenv("ECOMMERCE_FRONTEND_URL")
	log.Printf("🌍 ECOMMERCE_FRONTEND_URL env value: %q", urlFromEnv)

	if urlFromEnv == "" {
		defaultURL := "http://localhost:3001"
		log.Printf("⚠️  ECOMMERCE_FRONTEND_URL not set, using default: %s", defaultURL)
		return defaultURL
	}

	return urlFromEnv
}
