package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ════════════════════════════════════════════════════════════
// Configuration Maps
// ════════════════════════════════════════════════════════════

// pathToResourceType maps URL paths to resource types
var pathToResourceType = map[string]string{
	"categories": models.ResourceTypeCategory,
	"products":   models.ResourceTypeProduct,
	"orders":     models.ResourceTypeOrder,
	"customers":  models.ResourceTypeCustomer,
	"admins":     models.ResourceTypeAdmin,
}

// resourceTypeToNameField maps resource types to their name field
var resourceTypeToNameField = map[string]string{
	models.ResourceTypeCategory: "name",
	models.ResourceTypeProduct:  "name",
	models.ResourceTypeCustomer: "email",
	models.ResourceTypeOrder:    "id",
	models.ResourceTypeAdmin:    "email",
}

// methodToActionVerb maps HTTP methods to action verbs
var methodToActionVerb = map[string]string{
	"POST":   "created",
	"PATCH":  "updated",
	"PUT":    "updated",
	"DELETE": "deleted",
}

// ════════════════════════════════════════════════════════════
// Activity Logging Middleware
// ════════════════════════════════════════════════════════════

// ActivityLoggingMiddleware logs admin actions automatically
// Must be used AFTER AdminAuthMiddleware (which sets adminID and adminEmail)
func ActivityLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip GET requests - we only log non-GET (POST, PATCH, PUT, DELETE)
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// Extract admin info from context (set by AdminAuthMiddleware)
		adminIDRaw, adminIDExists := c.Get("adminID")
		adminEmailRaw, adminEmailExists := c.Get("adminEmail")

		if !adminIDExists || !adminEmailExists {
			log.Printf("[activity-logging] warning: admin info not in context")
			c.Next()
			return
		}

		adminID := uuid.UUID{}
		if id, ok := adminIDRaw.(uuid.UUID); ok {
			adminID = id
		} else if idStr, ok := adminIDRaw.(string); ok {
			parsedID, err := uuid.Parse(idStr)
			if err != nil {
				log.Printf("[activity-logging] failed to parse admin ID: %v", err)
				c.Next()
				return
			}
			adminID = parsedID
		}

		adminEmail := adminEmailRaw.(string)

		// Extract resource type from URL path
		resourceType := extractResourceType(c.Request.URL.Path)
		if resourceType == "" {
			log.Printf("[activity-logging] could not determine resource type from path: %s", c.Request.URL.Path)
			c.Next()
			return
		}

		// Extract resource ID from URL params
		resourceID := c.Param("id")
		if resourceID == "" {
			// Some routes might use different param names, but "id" is standard
			log.Printf("[activity-logging] warning: no :id param found for %s", c.Request.URL.Path)
		}

		// Determine action from HTTP method
		actionVerb := methodToActionVerb[c.Request.Method]
		if actionVerb == "" {
			log.Printf("[activity-logging] unknown HTTP method: %s", c.Request.Method)
			c.Next()
			return
		}

		// Build full action name (e.g., "created_product", "updated_category")
		action := actionVerb + "_" + resourceType

		// Fetch "before" object from DB (only for updates and deletes)
		var beforeObject interface{}
		if c.Request.Method != "POST" && resourceID != "" {
			beforeObject = fetchResourceFromDB(resourceType, resourceID)
		}

		// Extract resource name from before object (for updates/deletes)
		resourceName := extractResourceName(resourceType, beforeObject)

		// Store in context for use in response handler
		c.Set("activityAction", action)
		c.Set("activityResourceType", resourceType)
		c.Set("activityResourceID", resourceID)
		c.Set("activityResourceName", resourceName)
		c.Set("activityBeforeObject", beforeObject)
		c.Set("activityAdminID", adminID)
		c.Set("activityAdminEmail", adminEmail)

		// Execute the handler
		c.Next()

		// After handler execution, determine if successful and log
		statusCode := c.Writer.Status()
		isSuccess := statusCode >= 200 && statusCode < 300

		if isSuccess {
			// Fetch "after" object from DB
			var afterObject interface{}
			if resourceID != "" {
				afterObject = fetchResourceFromDB(resourceType, resourceID)
			}

			// Extract updated resource name
			updatedResourceName := extractResourceName(resourceType, afterObject)

			// Log success
			services.LogActivity(services.LogActivityRequest{
				AdminID:      adminID,
				AdminEmail:   adminEmail,
				Action:       action,
				ResourceType: resourceType,
				ResourceID:   resourceID,
				ResourceName: updatedResourceName,
				Changes:      services.CreateChanges(beforeObject, afterObject),
				Status:       models.StatusSuccess,
				Context:      c,
			})

			log.Printf("[activity-logging] success: %s by %s", action, adminEmail)
		} else {
			// Log failure - extract error message from response if possible
			errorMsg := "Request failed with status " + http.StatusText(statusCode)

			services.LogActivity(services.LogActivityRequest{
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

			log.Printf("[activity-logging] failed: %s by %s - status %d", action, adminEmail, statusCode)
		}
	}
}

// ════════════════════════════════════════════════════════════
// Helper Functions
// ════════════════════════════════════════════════════════════

// extractResourceType extracts resource type from URL path
// e.g., "/api/v1/admin/categories/123" → "category"
func extractResourceType(path string) string {
	// Split path by "/"
	parts := strings.Split(path, "/")

	// Find the resource type (usually second to last part before ID)
	// e.g., "/api/v1/admin/categories/:id" → parts = ["", "api", "v1", "admin", "categories", ":id"]
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && !isIDParam(parts[i]) {
			// Found a potential resource type
			singular := strings.TrimSuffix(parts[i], "s") // Remove trailing 's' for plural
			if resourceType, exists := pathToResourceType[parts[i]]; exists {
				return resourceType
			}
			if resourceType, exists := pathToResourceType[singular]; exists {
				return resourceType
			}
		}
	}

	return ""
}

// isIDParam checks if a path segment is an ID parameter
func isIDParam(segment string) bool {
	// Check if it looks like a UUID or numeric ID
	if segment == ":id" || segment == "" {
		return true
	}
	// Try to parse as UUID
	if _, err := uuid.Parse(segment); err == nil {
		return true
	}
	return false
}

// fetchResourceFromDB fetches a resource from the database
func fetchResourceFromDB(resourceType, resourceID string) interface{} {
	ctx, cancel := config.WithTimeout()
	defer cancel()

	switch resourceType {
	case models.ResourceTypeProduct:
		var product models.Product
		if err := config.CmsGorm.WithContext(ctx).First(&product, "id = ?", resourceID).Error; err != nil {
			log.Printf("[activity-logging] failed to fetch product %s: %v", resourceID, err)
			return nil
		}
		return product

	case models.ResourceTypeCategory:
		var category models.Category
		if err := config.CmsGorm.WithContext(ctx).First(&category, "id = ?", resourceID).Error; err != nil {
			log.Printf("[activity-logging] failed to fetch category %s: %v", resourceID, err)
			return nil
		}
		return category

	case models.ResourceTypeOrder:
		var order models.Order
		if err := config.CmsGorm.WithContext(ctx).First(&order, "id = ?", resourceID).Error; err != nil {
			log.Printf("[activity-logging] failed to fetch order %s: %v", resourceID, err)
			return nil
		}
		return order

	case models.ResourceTypeCustomer:
		var customer models.User
		if err := config.CmsGorm.WithContext(ctx).First(&customer, "id = ?", resourceID).Error; err != nil {
			log.Printf("[activity-logging] failed to fetch customer %s: %v", resourceID, err)
			return nil
		}
		return customer

	case models.ResourceTypeAdmin:
		var admin models.Admin
		if err := config.CmsGorm.WithContext(ctx).First(&admin, "id = ?", resourceID).Error; err != nil {
			log.Printf("[activity-logging] failed to fetch admin %s: %v", resourceID, err)
			return nil
		}
		return admin

	default:
		log.Printf("[activity-logging] unknown resource type: %s", resourceType)
		return nil
	}
}

// extractResourceName extracts the name/identifier from a resource object
func extractResourceName(resourceType string, obj interface{}) string {
	if obj == nil {
		return ""
	}

	// Convert to map for easy field access
	data, err := json.Marshal(obj)
	if err != nil {
		return ""
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return ""
	}

	// Get the field name for this resource type
	fieldName := resourceTypeToNameField[resourceType]
	if fieldName == "" {
		return ""
	}

	// Extract the value
	if value, exists := resourceMap[fieldName]; exists {
		return toString(value)
	}

	return ""
}

// toString converts any value to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		// Convert float64 to string
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		data, err := json.Marshal(val)
		if err != nil {
			return ""
		}
		return string(data)
	}
}
