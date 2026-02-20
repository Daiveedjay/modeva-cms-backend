// models/filter_models.go
package models

// FilterMetadata represents all filter data for the storefront
type FilterMetadata struct {
	Availability *AvailabilityData `json:"availability"`
	Categories   []CategoryData    `json:"categories"`
	PriceRange   *PriceRangeData   `json:"priceRange"`
}

// AvailabilityData represents product availability counts
type AvailabilityData struct {
	InStock    int `json:"inStock"`
	OutOfStock int `json:"outOfStock"`
}

// CategoryData represents a category with optional subcategories
type CategoryData struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	ParentID      string         `json:"parentId,omitempty"`
	Subcategories []CategoryData `json:"subcategories,omitempty"`
}

// PriceRangeData represents the minimum and maximum price in the store
type PriceRangeData struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
