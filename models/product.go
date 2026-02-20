package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════════════════
// JSONB Type Definitions
// ═══════════════════════════════════════════════════════════

type MediaURL struct {
	URL   string `json:"url" binding:"required"`
	Order *int   `json:"order,omitempty"`
}

type ProductMedia struct {
	Primary MediaURL   `json:"primary" binding:"required"`
	Other   []MediaURL `json:"other,omitempty"`
}

type ProductVariant struct {
	Type    string   `json:"type" binding:"required" example:"Size"`
	Options []string `json:"options" binding:"required" example:"['Small', 'Medium', 'Large']"`
}

type InventoryField struct {
	Combo       []string `json:"combo" binding:"required" example:"['Small', 'Black']"`
	VariantName string   `json:"variant_name" binding:"required" example:"Small-Black"`
	Quantity    int      `json:"quantity" binding:"required,min=0" example:"100"`
}

// Create custom types for slices (so we can add methods)
type (
	CompositionList []Composition
	TagsList        []string
	VariantsList    []ProductVariant
	InventoryList   []InventoryField
)

// Use custom types as aliases for convenience
type Inventory = InventoryList

type Seo struct {
	SEOTitle       string `json:"seo_title" binding:"required" example:"Best Sample Product"`
	SEODescription string `json:"seo_description" binding:"required" example:"This is the best sample product."`
}

type Composition struct {
	Label   string `json:"label" binding:"required" example:"Material"`
	Content string `json:"content" binding:"required" example:"100% Cotton"`
}

// ═══════════════════════════════════════════════════════════
// Main Product Model (GORM)
// ═══════════════════════════════════════════════════════════

type Product struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	Name            string          `json:"name" gorm:"not null;index"`
	Description     string          `json:"description" gorm:"not null"`
	Composition     CompositionList `json:"composition" gorm:"type:jsonb;not null;default:'[]'"`
	Price           float64         `json:"price" gorm:"type:numeric(12,2);not null;check:price >= 0"`
	SubCategoryID   uuid.UUID       `json:"sub_category_id" gorm:"type:uuid;not null;index:idx_products_subcategory"`
	SubCategoryName *string         `json:"sub_category_name,omitempty" gorm:"-"` // Computed field
	SubCategory     *Category       `json:"sub_category,omitempty" gorm:"foreignKey:SubCategoryID;references:ID"`
	Status          string          `json:"status" gorm:"not null;check:status IN ('Active', 'Draft');index"`
	Tags            TagsList        `json:"tags" gorm:"type:jsonb;not null;default:'[]';index:,type:gin"`
	Media           ProductMedia    `json:"media" gorm:"type:jsonb;not null;default:'{}'"`
	Variants        VariantsList    `json:"variants" gorm:"type:jsonb;not null;default:'[]'"`
	Inventory       InventoryList   `json:"inventory" gorm:"type:jsonb;not null;default:'[]'"`
	SEO             Seo             `json:"seo" gorm:"type:jsonb;not null;default:'{}'"`
	Views           int             `json:"views" gorm:"default:0;index:idx_products_views,sort:desc"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate hook - auto-generate UUID v7
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}

// AfterFind hook - populate SubCategoryName from relationship
func (p *Product) AfterFind(tx *gorm.DB) error {
	if p.SubCategory != nil {
		p.SubCategoryName = &p.SubCategory.Name
	}
	return nil
}

// TableName specifies the table name
func (Product) TableName() string {
	return "products"
}

// ═══════════════════════════════════════════════════════════
// Request Models
// ═══════════════════════════════════════════════════════════

type ProductRequest struct {
	Name          string           `json:"name" binding:"required" example:"Sample Product"`
	Description   string           `json:"description" binding:"required" example:"This is a sample product"`
	Composition   []Composition    `json:"composition" binding:"required,dive"`
	Price         float64          `json:"price" binding:"required,min=0" example:"99.99"`
	SubCategoryID uuid.UUID        `json:"sub_category_id" binding:"required" example:"018d1234-5678-7abc-def0-123456789abc"`
	Status        string           `json:"status" binding:"required,oneof=Active Draft" example:"Draft"`
	Tags          []string         `json:"tags" binding:"required" example:"['cotton', 'summer']"`
	Media         ProductMedia     `json:"media" binding:"required"`
	Variants      []ProductVariant `json:"variants" binding:"required,dive"`
	Inventory     []InventoryField `json:"inventory" binding:"required,dive"`
	SEO           Seo              `json:"seo" binding:"required"`
}

type UpdateProductRequest struct {
	Name          *string           `json:"name"`
	Description   *string           `json:"description"`
	Composition   *[]Composition    `json:"composition"`
	Price         *float64          `json:"price" binding:"omitempty,min=0"`
	SubCategoryID *uuid.UUID        `json:"sub_category_id"`
	Status        *string           `json:"status" binding:"omitempty,oneof=Active Draft"`
	Tags          *[]string         `json:"tags"`
	Media         *ProductMedia     `json:"media"`
	Variants      *[]ProductVariant `json:"variants"`
	Inventory     *[]InventoryField `json:"inventory"`
	SEO           *Seo              `json:"seo"`
}

// ═══════════════════════════════════════════════════════════
// Response Models
// ═══════════════════════════════════════════════════════════

type ProductBase struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Composition     []Composition `json:"composition"`
	Price           float64       `json:"price"`
	SubCategoryID   uuid.UUID     `json:"sub_category_id"`
	SubCategoryName *string       `json:"sub_category_name,omitempty"`
	Status          string        `json:"status"`
	Tags            []string      `json:"tags"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	SubCategoryPath *string       `json:"sub_category_path,omitempty"`
}

type ProductResponse struct {
	BasicInfo ProductBase      `json:"basic_info"`
	SEO       Seo              `json:"seo"`
	Media     ProductMedia     `json:"media"`
	Variants  []ProductVariant `json:"variants"`
	Inventory []InventoryField `json:"inventory"`
}

type ProductStatsResponseItem struct {
	Type               string  `json:"type"`
	TotalProducts      int     `json:"total_products,omitempty"`
	ActiveProducts     int     `json:"active_products,omitempty"`
	DraftProducts      int     `json:"draft_products,omitempty"`
	PercentageActive   float64 `json:"percentage_active,omitempty"`
	AveragePrice       float64 `json:"average_price,omitempty"`
	TotalInventory     int     `json:"total_inventory,omitempty"`
	TaggedProducts     int     `json:"tagged_products,omitempty"`
	LowStockProducts   int     `json:"low_stock_products,omitempty"`
	PercentageLowStock float64 `json:"percentage_low_stock,omitempty"`
}

// ═══════════════════════════════════════════════════════════
// JSONB Scanner/Valuer for GORM (Custom slice types)
// ═══════════════════════════════════════════════════════════

// CompositionList methods
func (c *CompositionList) Scan(value interface{}) error {
	if value == nil {
		*c = make(CompositionList, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CompositionList")
	}
	return json.Unmarshal(bytes, c)
}

func (c CompositionList) Value() (driver.Value, error) {
	if c == nil {
		return json.Marshal([]Composition{})
	}
	return json.Marshal(c)
}

// TagsList methods
func (t *TagsList) Scan(value interface{}) error {
	if value == nil {
		*t = make(TagsList, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan TagsList")
	}
	return json.Unmarshal(bytes, t)
}

func (t TagsList) Value() (driver.Value, error) {
	if t == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(t)
}

// VariantsList methods
func (v *VariantsList) Scan(value interface{}) error {
	if value == nil {
		*v = make(VariantsList, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan VariantsList")
	}
	return json.Unmarshal(bytes, v)
}

func (v VariantsList) Value() (driver.Value, error) {
	if v == nil {
		return json.Marshal([]ProductVariant{})
	}
	return json.Marshal(v)
}

// InventoryList methods
func (i *InventoryList) Scan(value interface{}) error {
	if value == nil {
		*i = make(InventoryList, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan InventoryList")
	}
	return json.Unmarshal(bytes, i)
}

func (i InventoryList) Value() (driver.Value, error) {
	if i == nil {
		return json.Marshal([]InventoryField{})
	}
	return json.Marshal(i)
}

// ProductMedia methods
func (m *ProductMedia) Scan(value interface{}) error {
	if value == nil {
		*m = ProductMedia{Other: make([]MediaURL, 0)}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ProductMedia")
	}
	return json.Unmarshal(bytes, m)
}

func (m ProductMedia) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Seo methods
func (s *Seo) Scan(value interface{}) error {
	if value == nil {
		*s = Seo{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Seo")
	}
	return json.Unmarshal(bytes, s)
}

func (s Seo) Value() (driver.Value, error) {
	return json.Marshal(s)
}
