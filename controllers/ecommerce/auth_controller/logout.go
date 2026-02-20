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
	isProd := os.Getenv("ENV") == "production"
	// delete auth_token (must match name, path, domain, secure, httpOnly)
	c.SetCookie(
		"auth_token",
		"",
		-1, // MaxAge < 0 -> delete
		"/",
		"",
		isProd,
		true, // HttpOnly (same as when set)
	)

	// also clear the user_data helper cookie
	c.SetCookie(
		"user_data",
		"",
		-1,
		"/",
		"",
		isProd,
		false, // same as when set (NOT HttpOnly)
	)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
