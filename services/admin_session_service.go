package services

import (
	"context"
	"log"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/google/uuid"
)

// AdminSessionService handles admin session operations
type AdminSessionService struct{}

// NewAdminSessionService creates a new session service
func NewAdminSessionService() *AdminSessionService {
	return &AdminSessionService{}
}

// CreateSession creates a new admin session
func (s *AdminSessionService) CreateSession(
	ctx context.Context,
	adminID uuid.UUID,
	token string,
	ipAddress string,
	userAgent string,
) (*models.AdminSession, error) {
	authService := GetAdminAuthService()
	tokenHash := authService.HashToken(token)

	session := &models.AdminSession{
		ID:             uuid.Must(uuid.NewV7()),
		AdminID:        adminID,
		TokenHash:      tokenHash,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		IsActive:       true,
	}

	if err := config.CmsGorm.WithContext(ctx).Create(session).Error; err != nil {
		log.Printf("[session] failed to create session: %v", err)
		return nil, err
	}

	log.Printf("[session] created session %s for admin %s", session.ID, adminID)
	return session, nil
}

// UpdateSessionActivity updates the last activity timestamp for a session
func (s *AdminSessionService) UpdateSessionActivity(
	ctx context.Context,
	tokenHash string,
) error {
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.AdminSession{}).
		Where("token_hash = ? AND is_active = ?", tokenHash, true).
		Update("last_activity_at", time.Now()).Error; err != nil {
		log.Printf("[session] failed to update session activity: %v", err)
		return err
	}
	return nil
}

// DeactivateSession marks a session as inactive (logout)
func (s *AdminSessionService) DeactivateSession(
	ctx context.Context,
	adminID uuid.UUID,
) error {
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.AdminSession{}).
		Where("admin_id = ? AND is_active = ?", adminID, true).
		Update("is_active", false).Error; err != nil {
		log.Printf("[session] failed to deactivate session: %v", err)
		return err
	}

	log.Printf("[session] deactivated session for admin %s", adminID)
	return nil
}

// GetActiveSessionsByAdmin gets all active sessions for an admin
func (s *AdminSessionService) GetActiveSessionsByAdmin(
	ctx context.Context,
	adminID uuid.UUID,
) ([]models.AdminSession, error) {
	var sessions []models.AdminSession
	if err := config.CmsGorm.WithContext(ctx).
		Where("admin_id = ? AND is_active = ? AND expires_at > ?", adminID, true, time.Now()).
		Find(&sessions).Error; err != nil {
		log.Printf("[session] failed to get active sessions: %v", err)
		return nil, err
	}
	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions (run periodically)
func (s *AdminSessionService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result := config.CmsGorm.WithContext(ctx).
		Where("expires_at < ? OR (is_active = ? AND last_activity_at < ?)",
			time.Now(),
			false,
			time.Now().Add(-7*24*time.Hour), // Keep inactive sessions for 7 days
		).
		Delete(&models.AdminSession{})

	if result.Error != nil {
		log.Printf("[session] failed to cleanup expired sessions: %v", result.Error)
		return 0, result.Error
	}

	log.Printf("[session] cleaned up %d expired sessions", result.RowsAffected)
	return result.RowsAffected, nil
}

// CountActiveSessions counts total active sessions across all admins
func (s *AdminSessionService) CountActiveSessions(ctx context.Context) (int64, error) {
	var count int64
	if err := config.CmsGorm.WithContext(ctx).
		Model(&models.AdminSession{}).
		Where("is_active = ? AND expires_at > ?", true, time.Now()).
		Count(&count).Error; err != nil {
		log.Printf("[session] failed to count active sessions: %v", err)
		return 0, err
	}
	return count, nil
}

// Global instance
var adminSessionService *AdminSessionService

// GetAdminSessionService returns the global session service instance
func GetAdminSessionService() *AdminSessionService {
	if adminSessionService == nil {
		adminSessionService = NewAdminSessionService()
	}
	return adminSessionService
}
