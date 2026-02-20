package auth_controller

import (
	"fmt"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func createOrUpdateUser(
	c *gin.Context,
	googleUser *models.GoogleUserInfo,
	googleID string,
	emailVerified bool,
) (*models.User, error) {
	var user models.User

	// Try to find existing user by email
	result := config.EcommerceGorm.
		Where("email = ?", googleUser.Email).
		First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// First-time Google login, create user
			user = models.User{
				Email:         googleUser.Email,
				Name:          googleUser.Name,
				GoogleID:      googleID,
				Provider:      "google",
				EmailVerified: emailVerified,
				Avatar:        &googleUser.Picture,
				Status:        "active",
			}

			if err := config.EcommerceGorm.Create(&user).Error; err != nil {
				return nil, err
			}

			return &user, nil
		}

		return nil, result.Error
	}

	// Existing user: update safe fields only
	updates := map[string]interface{}{
		"avatar":         googleUser.Picture,
		"email_verified": emailVerified,
	}

	// Only set name if user never had one
	if user.Name == "" {
		updates["name"] = googleUser.Name
	}

	// Attach Google account if not already linked
	if user.GoogleID == "" {
		updates["google_id"] = googleID
		updates["provider"] = "google"
	}

	if err := config.EcommerceGorm.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Sync struct with DB updates
	if user.Name == "" {
		user.Name = googleUser.Name
	}
	user.Avatar = &googleUser.Picture
	user.EmailVerified = emailVerified

	return &user, nil
}

func redirectToFrontendWithError(c *gin.Context, errorMsg string) {
	frontendURL := config.GetFrontendURL()
	redirectURL := fmt.Sprintf("%s/auth/error?message=%s", frontendURL, errorMsg)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
