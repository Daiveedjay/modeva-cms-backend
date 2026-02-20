package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Email           string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Name            string     `json:"name" gorm:"type:varchar(255);not null"`
	GoogleID        string     `json:"googleId" gorm:"column:google_id;type:varchar(255);uniqueIndex;not null"`
	Provider        string     `json:"provider" gorm:"type:varchar(50);default:'google'"`
	Phone           *string    `json:"phone,omitempty" gorm:"type:varchar(50);index:idx_users_phone,where:phone IS NOT NULL"`
	Status          string     `json:"status" gorm:"type:varchar(50);default:'active';index"`
	EmailVerified   bool       `json:"emailVerified" gorm:"column:email_verified;default:true"`
	Avatar          *string    `json:"avatar,omitempty" gorm:"type:text"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"autoCreateTime;index"`
	UpdatedAt       time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
	BanReason       *string    `json:"banReason,omitempty" gorm:"column:ban_reason;type:text"`
	SuspendedUntil  *time.Time `json:"suspendedUntil,omitempty" gorm:"column:suspended_until"`
	SuspendedReason *string    `json:"suspendedReason,omitempty" gorm:"column:suspended_reason;type:text"`

	// Relationships
	Addresses      []Address           `json:"addresses,omitempty" gorm:"foreignKey:UserID"`
	PaymentMethods []UserPaymentMethod `json:"paymentMethods,omitempty" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}

// UserResponse is the public-facing user data
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Phone         *string   `json:"phone"`
	Provider      string    `json:"provider"`
	EmailVerified bool      `json:"email_verified"`
	Avatar        *string   `json:"avatar,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Phone:         u.Phone, // ✅ keep pointer
		Provider:      u.Provider,
		EmailVerified: u.EmailVerified,
		Avatar:        u.Avatar, // ✅ keep pointer
		CreatedAt:     u.CreatedAt,
	}
}

// GoogleUserInfo represents data from Google OAuth
type GoogleUserInfo struct {
	Sub           string `json:"sub"` // Google user ID
	ID            string `json:"id"`  // Alternative field name
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// AuthResponse is returned after successful authentication
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// CreateUserRequest for user registration
type CreateUserRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Phone    *string `json:"phone"`
	Avatar   *string `json:"avatar"`
	GoogleID *string `json:"googleId"`
}

// UpdateUserRequest for profile updates
type UpdateUserRequest struct {
	Name   *string `json:"name"`
	Phone  *string `json:"phone"`
	Avatar *string `json:"avatar"`
}

// UserProfileSummary for overview
type UserProfileSummary struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Avatar   *string   `json:"avatar"`
	Phone    *string   `json:"phone"`
	JoinedAt time.Time `json:"joined_at"`
}

// RecentOrderSummary for user overview
type RecentOrderSummary struct {
	ID          uuid.UUID `json:"id"`
	OrderNumber string    `json:"order_number"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserOverviewResponse represents the complete user overview
type UserOverviewResponse struct {
	Profile         UserProfileSummary   `json:"profile"`
	TotalPurchases  float64              `json:"total_purchases"`
	TotalOrders     int                  `json:"total_orders"`
	CompletedOrders int                  `json:"completed_orders"`
	LoyaltyPoints   int                  `json:"loyalty_points"`
	RecentOrders    []RecentOrderSummary `json:"recent_orders"`
}
