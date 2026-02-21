package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category represents a CMS category
type Category struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey" db:"id"`
	Name        string     `json:"name" gorm:"not null" db:"name"`
	Description string     `json:"description" gorm:"not null" db:"description"`
	Status      string     `json:"status" gorm:"type:varchar(20);default:'Inactive';check:status IN ('Active', 'Inactive')" db:"status"`
	ParentID    *uuid.UUID `json:"parent_id" gorm:"type:uuid;index" db:"parent_id"`
	ParentName  *string    `json:"parent_name" gorm:"type:text" db:"parent_name"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`

	// Relationships (GORM will handle these automatically)
	Parent   *Category   `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Children []*Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// CategoryWithProducts extends Category with product count
type CategoryWithProducts struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	ParentID    *uuid.UUID             `json:"parent_id"`
	ParentName  *string                `json:"parent_name"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Products    int                    `json:"products"`
	Children    []CategoryWithProducts `json:"children,omitempty"`
}

// BeforeCreate hook - runs automatically before creating a record
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	// Auto-generate UUID v7 if not set
	if c.ID == uuid.Nil {
		c.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}

// AfterUpdate hook - update children's parent_name when parent name changes
func (c *Category) AfterUpdate(tx *gorm.DB) error {
	// If name changed, update all children
	if tx.Statement.Changed("Name") {
		return tx.Model(&Category{}).
			Where("parent_id = ?", c.ID).
			Update("parent_name", c.Name).Error
	}
	return nil
}

// TableName specifies the table name (optional, GORM auto-pluralizes)
func (Category) TableName() string {
	return "categories"
}

// CategoryRequest is used when creating a category or subcategory
type CategoryRequest struct {
	Name        string     `json:"name" binding:"required" example:"Electronics"`
	Description string     `json:"description" binding:"required" example:"Devices and gadgets"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" example:"null"`
}

// UpdateCategoryRequest is used when updating a category
type UpdateCategoryRequest struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// UpdateCategoryStatusRequest is used when updating a category's status
type UpdateCategoryStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=Active Inactive" example:"Active"`
}

type Reassignment struct {
	ChildID     uuid.UUID `json:"child_id" example:"018d6cc9-94b0-450f-8ce3-a7892c1752c7"`      // Changed to uuid.UUID
	NewParentID uuid.UUID `json:"new_parent_id" example:"018d6cc9-0038-4d58-b52c-354e7c9a6656"` // Changed to uuid.UUID
}

type DeleteCategoryOptions struct {
	Mode          string         `json:"mode" binding:"required,oneof=cascade reassign" example:"cascade"`
	Reassignments []Reassignment `json:"reassignments,omitempty"`
}

// CategoryStatsResponseItem defines one of the items in the stats array
type CategoryStatsResponseItem struct {
	TotalCategories               int     `json:"total_categories"`
	ParentCategories              int     `json:"parent_categories"`
	SubCategories                 int     `json:"sub_categories"`
	ActiveCategories              int     `json:"active_categories"`
	ActiveParentCategories        int     `json:"active_parent_categories"`
	ActiveSubCategories           int     `json:"active_sub_categories"`
	PercentageActiveCategories    float64 `json:"percentage_active_categories"`
	PercentageActiveParents       float64 `json:"percentage_active_parents"`
	PercentageActiveSubCategories float64 `json:"percentage_active_sub_categories"`
}

type CategoryWithPath struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	CategoryPath string     `json:"category_path"` // e.g., "Electronics → Laptops"
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CategoryStats represents the cached statistics table
type CategoryStats struct {
	ID                     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	TotalCategories        int       `json:"total_categories"`
	ParentCategories       int       `json:"parent_categories"`
	SubCategories          int       `json:"sub_categories"`
	ActiveCategories       int       `json:"active_categories"`
	ActiveParentCategories int       `json:"active_parent_categories"`
	ActiveSubCategories    int       `json:"active_sub_categories"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (CategoryStats) TableName() string {
	return "category_stats"
}
