package services

import (
	"encoding/json"
	"log"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ActivityLogService handles activity logging
type ActivityLogService struct{}

// NewActivityLogService creates a new activity log service
func NewActivityLogService() *ActivityLogService {
	return &ActivityLogService{}
}

// LogActivityRequest contains the parameters for logging an activity
type LogActivityRequest struct {
	AdminID      uuid.UUID              // Who performed the action
	AdminEmail   string                 // Admin's email
	Action       string                 // ActionCreateProduct, ActionUpdateOrder, etc.
	ResourceType string                 // ResourceTypeProduct, ResourceTypeCustomer, etc.
	ResourceID   string                 // ID of the resource (product_id, customer_id, order_id, etc.)
	ResourceName string                 // Human readable name (product name, customer email, order ID, etc.)
	Changes      map[string]interface{} // {before: {...}, after: {...}}
	Status       string                 // StatusSuccess or StatusFailed
	ErrorMessage string                 // Error details if failed
	Context      *gin.Context           // For IP and User-Agent extraction
}

// LogActivity logs an admin action to the database
// Automatically captures IP address and User-Agent from context
func (s *ActivityLogService) LogActivity(req LogActivityRequest) error {
	if req.AdminID == uuid.Nil {
		log.Printf("[activity-log] warning: AdminID is nil for action %s", req.Action)
		return nil // Don't fail the request if logging fails
	}

	// Extract IP address
	ipAddress := extractClientIP(req.Context)

	// Extract User-Agent
	userAgent := ""
	if req.Context != nil {
		userAgent = req.Context.GetHeader("User-Agent")
	}

	// Marshal changes to JSONB
	var changesJSON []byte
	if req.Changes != nil {
		data, err := json.Marshal(req.Changes)
		if err != nil {
			log.Printf("[activity-log] failed to marshal changes: %v", err)
			changesJSON = []byte("{}")
		} else {
			changesJSON = data
		}
	}

	// Set default status
	if req.Status == "" {
		req.Status = models.StatusSuccess
	}

	// Create activity log entry
	activityLog := models.ActivityLog{
		AdminID:      req.AdminID,
		AdminEmail:   req.AdminEmail,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		ResourceName: req.ResourceName,
		Changes:      changesJSON,
		Status:       req.Status,
		ErrorMessage: req.ErrorMessage,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	// Log to database
	ctx, cancel := config.WithTimeout()
	defer cancel()

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[activity-log] failed to create activity log: %v", err)
		// Don't fail the request if logging fails - return nil
		return nil
	}

	log.Printf("[activity-log] %s: %s/%s/%s by %s", req.Action, req.ResourceType, req.ResourceID, req.ResourceName, req.AdminEmail)
	return nil
}

// extractClientIP extracts the client IP address from the request
// Checks X-Forwarded-For, X-Real-IP, then RemoteAddr
func extractClientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}

	// Check X-Forwarded-For header (from reverse proxy)
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}

	// Check X-Real-IP header (from reverse proxy)
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr from request
	return c.RemoteIP()
}

// Global instance
var activityLogService *ActivityLogService

// GetActivityLogService returns the global activity log service
func GetActivityLogService() *ActivityLogService {
	if activityLogService == nil {
		activityLogService = NewActivityLogService()
	}
	return activityLogService
}

// Convenience function
// LogActivity logs an activity using the global service
func LogActivity(req LogActivityRequest) error {
	return GetActivityLogService().LogActivity(req)
}

// Helper function to create changes map
func CreateChanges(before, after interface{}) map[string]interface{} {
	return map[string]interface{}{
		"before": before,
		"after":  after,
	}
}

// Helper to log success
func LogActivitySuccess(adminID uuid.UUID, adminEmail string, action, resourceType, resourceID, resourceName string, changes map[string]interface{}, c *gin.Context) error {
	return LogActivity(LogActivityRequest{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Changes:      changes,
		Status:       models.StatusSuccess,
		Context:      c,
	})
}

// Helper to log failure
func LogActivityFailed(adminID uuid.UUID, adminEmail string, action, resourceType, resourceID, resourceName, errorMsg string, c *gin.Context) error {
	return LogActivity(LogActivityRequest{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Status:       models.StatusFailed,
		ErrorMessage: errorMsg,
		Context:      c,
	})
}
