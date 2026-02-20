package middleware

import (
	"net/http"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token from cookie or Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try to get token from cookie first
		cookieToken, err := c.Cookie("auth_token")
		if err == nil && cookieToken != "" {
			token = cookieToken
		} else {
			// Fallback to Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Authorization header required"))
				c.Abort()
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Invalid authorization header format"))
				c.Abort()
				return
			}

			token = parts[1]
		}

		// Validate token
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Invalid or expired token"))
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userName", claims.Name)

		c.Next()
	}
}

// Helper functions remain the same
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("userEmail")
	if !exists {
		return "", false
	}
	return email.(string), true
}
