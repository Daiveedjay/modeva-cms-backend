package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AdminJWTClaims represents the JWT claims for admin tokens
type AdminJWTClaims struct {
	AdminID string `json:"admin_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and verification
type JWTService struct {
	secretKey string
}

var jwtService *JWTService

// InitJWTService initializes the JWT service with a secret key
func InitJWTService(secretKey string) error {
	if secretKey == "" {
		return errors.New("JWT secret key cannot be empty")
	}
	jwtService = &JWTService{
		secretKey: secretKey,
	}
	return nil
}

// GetJWTService returns the initialized JWT service
func GetJWTService() *JWTService {
	if jwtService == nil {
		// Fallback to environment variable if not initialized
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "dev-secret-key-change-in-production"
		}
		jwtService = &JWTService{secretKey: secretKey}
	}
	return jwtService
}

// GenerateAdminJWT creates a new JWT token for an admin
// Token expires in 7 days
func (j *JWTService) GenerateAdminJWT(adminID, email string) (string, error) {
	if adminID == "" || email == "" {
		return "", errors.New("adminID and email cannot be empty")
	}

	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days

	claims := AdminJWTClaims{
		AdminID: adminID,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "modeva-cms",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyAdminJWT verifies and parses a JWT token
// Returns claims if valid, error if invalid or expired
func (j *JWTService) VerifyAdminJWT(tokenString string) (*AdminJWTClaims, error) {
	claims := &AdminJWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token has required claims
	if claims.AdminID == "" || claims.Email == "" {
		return nil, errors.New("token missing required claims")
	}

	return claims, nil
}

// Convenience functions that use the global service

// GenerateAdminJWT generates a JWT token using the global JWT service
func GenerateAdminJWT(adminID, email string) (string, error) {
	return GetJWTService().GenerateAdminJWT(adminID, email)
}

// VerifyAdminJWT verifies a JWT token using the global JWT service
func VerifyAdminJWT(tokenString string) (*AdminJWTClaims, error) {
	return GetJWTService().VerifyAdminJWT(tokenString)
}
