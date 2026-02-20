package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Address struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Label     string    `json:"label" gorm:"type:varchar(50);not null"`
	FirstName string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName  string    `json:"last_name" gorm:"type:varchar(100);not null"`
	Street    string    `json:"street" gorm:"type:varchar(255);not null"`
	City      string    `json:"city" gorm:"type:varchar(100);not null"`
	State     string    `json:"state" gorm:"type:varchar(100);not null"`
	Zip       string    `json:"zip" gorm:"type:varchar(20);not null"`
	Country   string    `json:"country" gorm:"type:varchar(100);not null"`
	Phone     *string   `json:"phone,omitempty" gorm:"type:varchar(20)"`
	IsDefault bool      `json:"is_default" gorm:"default:false;index:idx_addresses_is_default,where:is_default = true"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'active';index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationship
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Address) TableName() string {
	return "addresses"
}

func (a *Address) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}

// Rest of your existing structs...
type AddressResponse struct {
	ID        uuid.UUID `json:"id"`
	Label     string    `json:"label"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Street    string    `json:"street"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	Zip       string    `json:"zip"`
	Country   string    `json:"country"`
	Phone     *string   `json:"phone,omitempty"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddAddressRequest struct {
	Label     string  `json:"label" binding:"required"`
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name" binding:"required"`
	Street    string  `json:"street" binding:"required"`
	City      string  `json:"city" binding:"required"`
	State     string  `json:"state" binding:"required"`
	Zip       string  `json:"zip" binding:"required"`
	Country   string  `json:"country" binding:"required"`
	Phone     *string `json:"phone,omitempty"`
	IsDefault bool    `json:"is_default"`
}

type UpdateAddressRequest struct {
	Label     *string `json:"label,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Street    *string `json:"street,omitempty"`
	City      *string `json:"city,omitempty"`
	State     *string `json:"state,omitempty"`
	Zip       *string `json:"zip,omitempty"`
	Country   *string `json:"country,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}
