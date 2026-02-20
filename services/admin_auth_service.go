package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminAuthService handles admin authentication operations
type AdminAuthService struct{}

// NewAdminAuthService creates a new admin auth service
func NewAdminAuthService() *AdminAuthService {
	return &AdminAuthService{}
}

// ════════════════════════════════════════════════════════════
// Password Management
// ════════════════════════════════════════════════════════════

// HashPassword hashes a password using bcrypt (cost: 12)
func (s *AdminAuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if a password matches its bcrypt hash
func (s *AdminAuthService) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if a password meets minimum requirements
// Minimum 8 characters
func (s *AdminAuthService) ValidatePassword(password string) bool {
	return len(password) >= 8
}

// ════════════════════════════════════════════════════════════
// Invite Token Management
// ════════════════════════════════════════════════════════════

// GenerateInviteToken generates a cryptographically secure random token
// Returns 64 character hex string (32 bytes)
func (s *AdminAuthService) GenerateInviteToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// HashToken hashes a token using SHA256 for storage in database
func (s *AdminAuthService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GetInviteTokenExpiration returns the expiration time for an invite token
// Expires in 48 hours
func (s *AdminAuthService) GetInviteTokenExpiration() time.Time {
	return time.Now().Add(48 * time.Hour)
}

// IsInviteExpired checks if an invite token has expired
func (s *AdminAuthService) IsInviteExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// ════════════════════════════════════════════════════════════
// Admin Status Management
// ════════════════════════════════════════════════════════════

// IsStatusInactive checks if an admin should be marked as inactive
// Inactive if last login is more than 7 days ago
func (s *AdminAuthService) IsStatusInactive(lastLoginAt *time.Time) bool {
	if lastLoginAt == nil {
		// Never logged in, not yet inactive
		return false
	}
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	return lastLoginAt.Before(sevenDaysAgo)
}

// GetAdminStatus calculates the current status based on last login
// Rules:
// - If suspended: stays suspended
// - If last login > 7 days ago: inactive
// - Otherwise: active
func (s *AdminAuthService) GetAdminStatus(currentStatus string, lastLoginAt *time.Time) string {
	// Suspended stays suspended
	if currentStatus == "suspended" {
		return "suspended"
	}

	// Check if should be inactive
	if s.IsStatusInactive(lastLoginAt) {
		return "inactive"
	}

	return "active"
}

// ════════════════════════════════════════════════════════════
// Global Instance
// ════════════════════════════════════════════════════════════

var adminAuthService *AdminAuthService

// GetAdminAuthService returns the global admin auth service instance
func GetAdminAuthService() *AdminAuthService {
	if adminAuthService == nil {
		adminAuthService = NewAdminAuthService()
	}
	return adminAuthService
}

// Convenience functions using global service

// HashPassword hashes a password using the global service
func HashAdminPassword(password string) (string, error) {
	return GetAdminAuthService().HashPassword(password)
}

// VerifyPassword verifies a password using the global service
func VerifyAdminPassword(hash, password string) bool {
	return GetAdminAuthService().VerifyPassword(hash, password)
}

// ValidateAdminPassword validates password requirements using the global service
func ValidateAdminPassword(password string) bool {
	return GetAdminAuthService().ValidatePassword(password)
}

// GenerateAdminInviteToken generates a token using the global service
func GenerateAdminInviteToken() (string, error) {
	return GetAdminAuthService().GenerateInviteToken()
}

// HashAdminToken hashes a token using the global service
func HashAdminToken(token string) string {
	return GetAdminAuthService().HashToken(token)
}

// GetAdminInviteTokenExpiration returns expiration time using the global service
func GetAdminInviteTokenExpiration() time.Time {
	return GetAdminAuthService().GetInviteTokenExpiration()
}

// IsAdminInviteExpired checks if invite is expired using the global service
func IsAdminInviteExpired(expiresAt time.Time) bool {
	return GetAdminAuthService().IsInviteExpired(expiresAt)
}

// GetCalculatedAdminStatus calculates status using the global service
func GetCalculatedAdminStatus(currentStatus string, lastLoginAt *time.Time) string {
	return GetAdminAuthService().GetAdminStatus(currentStatus, lastLoginAt)
}
