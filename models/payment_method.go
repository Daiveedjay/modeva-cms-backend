package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserPaymentMethod struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `json:"-" gorm:"type:uuid;not null;index"`
	Type      string    `json:"type" gorm:"type:varchar(20);default:'card'"`
	IsDefault bool      `json:"is_default" gorm:"default:false;index:idx_user_payment_methods_is_default,where:is_default = true"`

	// Provider info
	Provider                *string `json:"provider,omitempty" gorm:"type:varchar(50)"`
	ProviderPaymentMethodID *string `json:"-" gorm:"column:provider_payment_method_id;type:varchar(255)"`

	// Card details (ENCRYPT IN PRODUCTION!)
	CardType       string  `json:"card_type" gorm:"type:varchar(10);not null"`  // 'credit' or 'debit'
	CardBrand      string  `json:"card_brand" gorm:"type:varchar(20);not null"` // 'visa', 'mastercard', etc.
	CardNumber     string  `json:"-" gorm:"type:varchar(255);not null"`         // Full card number - NEVER expose in JSON!
	ExpMonth       int     `json:"exp_month" gorm:"not null"`
	ExpYear        int     `json:"exp_year" gorm:"not null"`
	CVV            *string `json:"-" gorm:"type:varchar(4)"` // NEVER expose in JSON!
	CardholderName string  `json:"cardholder_name" gorm:"type:varchar(255);not null"`

	Status    string    `json:"status" gorm:"type:varchar(20);default:'active';index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationship
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserPaymentMethod) TableName() string {
	return "user_payment_methods"
}

func (pm *UserPaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == uuid.Nil {
		pm.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}

// GetLast4 returns last 4 digits of card for display
func (pm *UserPaymentMethod) GetLast4() string {
	if len(pm.CardNumber) >= 4 {
		return pm.CardNumber[len(pm.CardNumber)-4:]
	}
	return ""
}

// GetMaskedCardNumber returns masked card number for display
func (pm *UserPaymentMethod) GetMaskedCardNumber() string {
	if len(pm.CardNumber) >= 4 {
		return "•••• •••• •••• " + pm.GetLast4()
	}
	return "•••• •••• •••• ••••"
}

type PaymentMethodResponse struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`        // always 'card'
	CardType   string    `json:"card_type"`   // 'credit' or 'debit'
	CardNumber string    `json:"card_number"` // "•••• •••• •••• 4242"
	CardBrand  string    `json:"card_brand"`  // 'visa', 'mastercard'
	CardHolder string    `json:"card_holder"`
	Expiry     string    `json:"expiry"` // "12/26"
	IsDefault  bool      `json:"is_default"`
	Status     string    `json:"status"`
}

type AddPaymentMethodRequest struct {
	CardType                string  `json:"card_type" binding:"required,oneof=credit debit"`
	CardBrand               string  `json:"card_brand" binding:"required"`
	CardNumber              string  `json:"card_number" binding:"required,min=13,max=19"`
	ExpMonth                int     `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear                 int     `json:"exp_year" binding:"required,min=2025"`
	CVV                     string  `json:"cvv" binding:"required,min=3,max=4"`
	CardholderName          string  `json:"cardholder_name" binding:"required"`
	IsDefault               bool    `json:"is_default"`
	Provider                *string `json:"provider,omitempty"`
	ProviderPaymentMethodID *string `json:"provider_payment_method_id,omitempty"`
}

func (pm *UserPaymentMethod) ToResponse() PaymentMethodResponse {
	return PaymentMethodResponse{
		ID:         pm.ID,
		Type:       pm.Type,
		CardType:   pm.CardType,
		CardNumber: pm.GetMaskedCardNumber(),
		CardBrand:  pm.CardBrand,
		CardHolder: pm.CardholderName,
		Expiry:     fmt.Sprintf("%02d/%02d", pm.ExpMonth, pm.ExpYear%100),
		IsDefault:  pm.IsDefault,
		Status:     pm.Status,
	}
}
