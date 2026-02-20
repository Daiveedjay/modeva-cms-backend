package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ActivityLog represents an admin action log entry
type ActivityLog struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	AdminID      uuid.UUID      `json:"admin_id" gorm:"type:uuid;not null;index:idx_activity_admin_date,sort:desc"`
	AdminEmail   string         `json:"admin_email" gorm:"not null"`
	Action       string         `json:"action" gorm:"not null;index"`                                             // created_product, updated_order, deleted_category, etc.
	ResourceType string         `json:"resource_type" gorm:"not null;index:idx_activity_resource_date,sort:desc"` // product, category, order
	ResourceID   string         `json:"resource_id" gorm:"not null;index"`                                        // UUID or identifier
	ResourceName string         `json:"resource_name"`                                                            // Human readable: product name, order ID, etc.
	Changes      datatypes.JSON `json:"changes" gorm:"type:jsonb"`                                                // {before: {...}, after: {...}}
	Status       string         `json:"status" gorm:"not null"`                                                   // success, failed
	ErrorMessage string         `json:"error_message"`                                                            // Error details if failed
	IPAddress    string         `json:"ip_address"`                                                               // Client IP
	UserAgent    string         `json:"user_agent"`                                                               // Browser/client info
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime;index:idx_activity_admin_date,sort:desc;index:idx_activity_resource_date,sort:desc"`
}

// BeforeCreate hook - auto-generate UUID v7
func (al *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.Must(uuid.NewV7())
	}
	// Default status to success if not set
	if al.Status == "" {
		al.Status = "success"
	}
	return nil
}

// TableName specifies the table name
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// ════════════════════════════════════════════════════════════
// Changes Structure
// ════════════════════════════════════════════════════════════

// ActivityChanges represents the before/after changes
type ActivityChanges struct {
	Before map[string]interface{} `json:"before"`
	After  map[string]interface{} `json:"after"`
}

// MarshalJSON converts ActivityChanges to JSON
func (ac ActivityChanges) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"before": ac.Before,
		"after":  ac.After,
	})
}

// UnmarshalJSON parses JSON into ActivityChanges
func (ac *ActivityChanges) UnmarshalJSON(data []byte) error {
	var m map[string]map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	ac.Before = m["before"]
	ac.After = m["after"]
	return nil
}

// ════════════════════════════════════════════════════════════
// Request/Response Models
// ════════════════════════════════════════════════════════════

// ActivityLogResponse is the response for activity log data
type ActivityLogResponse struct {
	ID           uuid.UUID              `json:"id"`
	AdminID      uuid.UUID              `json:"admin_id"`
	AdminEmail   string                 `json:"admin_email"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	ResourceName string                 `json:"resource_name"`
	Changes      map[string]interface{} `json:"changes"`
	Status       string                 `json:"status"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ToResponse converts ActivityLog to ActivityLogResponse
func (al *ActivityLog) ToResponse() ActivityLogResponse {
	changes := make(map[string]interface{})
	if al.Changes != nil {
		_ = json.Unmarshal(al.Changes, &changes)
	}

	return ActivityLogResponse{
		ID:           al.ID,
		AdminID:      al.AdminID,
		AdminEmail:   al.AdminEmail,
		Action:       al.Action,
		ResourceType: al.ResourceType,
		ResourceID:   al.ResourceID,
		ResourceName: al.ResourceName,
		Changes:      changes,
		Status:       al.Status,
		ErrorMessage: al.ErrorMessage,
		IPAddress:    al.IPAddress,
		UserAgent:    al.UserAgent,
		CreatedAt:    al.CreatedAt,
	}
}

// ════════════════════════════════════════════════════════════
// Action Constants
// ════════════════════════════════════════════════════════════

const (
	// Product Actions
	ActionCreateProduct = "created_product"
	ActionUpdateProduct = "updated_product"
	ActionDeleteProduct = "deleted_product"

	// Category Actions
	ActionCreateCategory = "created_category"
	ActionUpdateCategory = "updated_category"
	ActionDeleteCategory = "deleted_category"

	// Order Actions
	ActionUpdateOrder = "updated_order"

	// Customer Actions
	ActionUpdateCustomer    = "updated_customer"
	ActionBanCustomer       = "banned_customer"
	ActionUnbanCustomer     = "unbanned_customer"
	ActionDeleteCustomer    = "deleted_customer"
	ActionSendCustomerEmail = "sent_customer_email"
	ActionSuspendCustomer   = "suspended_customer"
	ActionUnsuspendCustomer = "unsuspended_customer"

	// Admin Actions
	ActionCreateAdminInvite  = "created_admin_invite"
	ActionAcceptAdminInvite  = "accepted_admin_invite"
	ActionSuspendAdmin       = "suspended_admin"
	ActionUnsuspendAdmin     = "unsuspended_admin"
	ActionUpdateAdminProfile = "updated_admin_profile"

	// Resource Types
	ResourceTypeProduct     = "product"
	ResourceTypeCategory    = "category"
	ResourceTypeOrder       = "order"
	ResourceTypeCustomer    = "customer"
	ResourceTypeAdmin       = "admin"
	ResourceTypeAdminInvite = "admin_invite" // ← Add this line

	// Status
	StatusSuccess = "success"
	StatusFailed  = "failed"
)
