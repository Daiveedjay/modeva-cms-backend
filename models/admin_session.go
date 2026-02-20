package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminSession represents an active admin session
type AdminSession struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	AdminID        uuid.UUID `json:"admin_id" gorm:"type:uuid;not null;index"`
	TokenHash      string    `json:"-" gorm:"not null;uniqueIndex"` // Hash of JWT token
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;index"`
	LastActivityAt time.Time `json:"last_activity_at" gorm:"index"` // Last request time
	ExpiresAt      time.Time `json:"expires_at" gorm:"index"`       // When session expires
	IsActive       bool      `json:"is_active" gorm:"default:true;index"`
}

// BeforeCreate hook - auto-generate UUID v7
func (as *AdminSession) BeforeCreate(tx *gorm.DB) error {
	if as.ID == uuid.Nil {
		as.ID = uuid.Must(uuid.NewV7())
	}
	// Set default expiry to 24 hours if not set
	if as.ExpiresAt.IsZero() {
		as.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	// Set initial last activity
	if as.LastActivityAt.IsZero() {
		as.LastActivityAt = time.Now()
	}
	return nil
}

// TableName specifies the table name
func (AdminSession) TableName() string {
	return "admin_sessions"
}

// IsExpired checks if session has expired
func (as *AdminSession) IsExpired() bool {
	return time.Now().After(as.ExpiresAt)
}

// AdminSessionResponse is the public response for session data
type AdminSessionResponse struct {
	ID             uuid.UUID `json:"id"`
	AdminID        uuid.UUID `json:"admin_id"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
}

// ToResponse converts AdminSession to response
func (as *AdminSession) ToResponse() AdminSessionResponse {
	return AdminSessionResponse{
		ID:             as.ID,
		AdminID:        as.AdminID,
		IPAddress:      as.IPAddress,
		CreatedAt:      as.CreatedAt,
		LastActivityAt: as.LastActivityAt,
		ExpiresAt:      as.ExpiresAt,
		IsActive:       as.IsActive,
	}
}
