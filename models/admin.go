package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ════════════════════════════════════════════════════════════
// Database Models
// ════════════════════════════════════════════════════════════

// Admin represents an admin user in the CMS
type Admin struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	Name         string     `json:"name" gorm:"not null"`
	Avatar       string     `json:"avatar" gorm:"type:text"` // Cloudinary URL
	PhoneNumber  string     `json:"phone_number"`
	Country      string     `json:"country"`
	PasswordHash string     `json:"-" gorm:"not null"`            // Never expose in JSON
	Role         string     `json:"role" gorm:"not null;index"`   // super_admin, admin
	Status       string     `json:"status" gorm:"not null;index"` // active, inactive, suspended
	LastLoginAt  *time.Time `json:"last_login_at"`
	JoinedAt     time.Time  `json:"joined_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate hook - auto-generate UUID v7
func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.Must(uuid.NewV7())
	}
	// Set default status if not provided
	if a.Status == "" {
		a.Status = "active"
	}
	// Set default role if not provided
	if a.Role == "" {
		a.Role = "admin"
	}
	return nil
}

// TableName specifies the table name
func (Admin) TableName() string {
	return "admins"
}

// AdminInvite represents an invitation to become an admin
type AdminInvite struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null"`
	TokenHash string     `json:"-" gorm:"not null;index"` // Hashed token (SHA256)
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	Used      bool       `json:"used" gorm:"default:false"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate hook - auto-generate UUID v7
func (ai *AdminInvite) BeforeCreate(tx *gorm.DB) error {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}

// TableName specifies the table name
func (AdminInvite) TableName() string {
	return "admin_invites"
}

// ════════════════════════════════════════════════════════════
// Request Models
// ════════════════════════════════════════════════════════════

// AdminLoginRequest is the request to login
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=1"`
}

// AcceptInviteRequest is the request to accept an invite and create account
type AcceptInviteRequest struct {
	Token    string `json:"token" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=1"`
	Password string `json:"password" binding:"required,min=8"`
}

// CreateAdminInviteRequest is the request to invite a new admin
type CreateAdminInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// UpdateAdminProfileRequest is the request to update admin profile
type UpdateAdminProfileRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1"`
	PhoneNumber *string `json:"phone_number" binding:"omitempty"`
	Country     *string `json:"country" binding:"omitempty"`
	Avatar      *string `json:"avatar" binding:"omitempty"` // Cloudinary URL
}

// SuspendAdminRequest is the request to suspend an admin
type SuspendAdminRequest struct {
	AdminID uuid.UUID `json:"admin_id" binding:"required"`
	Reason  string    `json:"reason"`
}

// UnsuspendAdminRequest is the request to unsuspend an admin
type UnsuspendAdminRequest struct {
	AdminID uuid.UUID `json:"admin_id" binding:"required"`
}

// ════════════════════════════════════════════════════════════
// Response Models
// ════════════════════════════════════════════════════════════

// AdminResponse is the public response for admin data (no password hash)
type AdminResponse struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Avatar      string     `json:"avatar"`
	PhoneNumber string     `json:"phone_number"`
	Country     string     `json:"country"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	JoinedAt    time.Time  `json:"joined_at"`
}

// AdminLoginResponse is the response after login
type AdminLoginResponse struct {
	Admin AdminResponse `json:"admin"`
	Token string        `json:"token"`
}

// ToResponse converts an Admin model to AdminResponse
func (a *Admin) ToResponse() AdminResponse {
	return AdminResponse{
		ID:          a.ID,
		Email:       a.Email,
		Name:        a.Name,
		Avatar:      a.Avatar,
		PhoneNumber: a.PhoneNumber,
		Country:     a.Country,
		Role:        a.Role,
		Status:      a.Status,
		LastLoginAt: a.LastLoginAt,
		JoinedAt:    a.JoinedAt,
	}
}
