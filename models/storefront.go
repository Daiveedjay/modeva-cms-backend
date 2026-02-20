// ════════════════════════════════════════════════════════════
// STOREFRONT MODELS (FINAL - INVENTORY AS JSONB)
// File: models/storefront.go
// ════════════════════════════════════════════════════════════

package models

import (
	"encoding/json"
	"time"
)

// StorefrontProduct represents a product in the storefront (customer-facing)
type StorefrontProduct struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Price         float64         `json:"price"`
	Inventory     json.RawMessage `json:"inventory,omitempty"`       // Hidden if not set
	Status        string          `json:"status,omitempty"`          // Hidden if not set
	SubCategoryID string          `json:"sub_category_id,omitempty"` // Hidden if not set
	CategoryName  string          `json:"category_name,omitempty"`   // Hidden if not set
	Media         json.RawMessage `json:"media,omitempty"`           // Hidden if not set
	Variants      json.RawMessage `json:"variants,omitempty"`        // Hidden if not set
	Views         int             `json:"views,omitempty"`           // Hidden if 0
	CreatedAt     time.Time       `json:"created_at,omitempty"`      // Hidden if not set
	UpdatedAt     time.Time       `json:"updated_at,omitempty"`      // Hidden if not set
}

type StorefrontProductResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Image string  `json:"image"`
	Price float64 `json:"price"`
}

// StorefrontCategory represents a category in the storefront
type StorefrontCategory struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	ParentID      *string              `json:"parent_id"`
	ProductCount  int                  `json:"product_count"`
	Subcategories []StorefrontCategory `json:"subcategories,omitempty"`
}

// ProductFilters represents available filters for products
type ProductFilters struct {
	Categories   []FilterOption `json:"categories"`
	Sizes        []FilterOption `json:"sizes"`
	PriceRange   PriceRange     `json:"price_range"`
	Availability []FilterOption `json:"availability"`
}

// FilterOption represents a single filter option
type FilterOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Count int    `json:"count"`
}

// PriceRange represents min and max price
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
