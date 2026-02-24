package customer_controller

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// UnbanCustomer godoc
// @Summary Unban a customer
// @Description Reinstate a banned customer account, restoring full access
// @Tags CMS - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Param unbanRequest body models.UnbanCustomerRequest true "Unban reason"
// @Success 200 {object} models.ApiResponse{data=models.UnbanCustomerResponse}
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 404 {object} models.ApiResponse "Customer not found"
// @Failure 409 {object} models.ApiResponse "Customer not banned"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /api/v1/admin/customers/{id}/unban [post]
func UnbanCustomer(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid customer ID"))
		return
	}

	var req models.UnbanCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Fetch customer
	var customer models.User
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ?", customerID).
		First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Customer not found"))
		return
	}

	// Check if not banned
	if !customer.IsBanned {
		c.JSON(http.StatusConflict, models.ErrorResponse(c, "Customer is not banned"))
		return
	}

	// Check if deleted
	if customer.DeletedAt != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(c, "Cannot unban a deleted customer"))
		return
	}

	// Unban the customer
	unbannedAt := time.Now()
	if err := config.EcommerceGorm.WithContext(ctx).Model(&customer).Updates(map[string]interface{}{
		"is_banned":  false,
		"ban_reason": nil,
		"banned_at":  nil,
		"status":     "active",
	}).Error; err != nil {
		log.Printf("[customer.unban] failed to unban customer: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to unban customer"))
		return
	}

	adminIDStr, _ := c.Get("adminID")
	adminEmail, _ := c.Get("adminEmail")

	// ✅ LOG THE ACTIVITY
	changes := map[string]interface{}{
		"unban_reason": req.Reason,
		"unbanned_at":  unbannedAt,
	}
	changesJSON, _ := json.Marshal(changes)

	adminID, _ := uuid.Parse(adminIDStr.(string))
	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      adminID,
		AdminEmail:   adminEmail.(string),
		Action:       "customer_unbanned",
		ResourceType: models.ResourceTypeCustomer,
		ResourceID:   customerID.String(),
		ResourceName: customer.Name,
		Changes:      datatypes.JSON(changesJSON),
		Status:       models.StatusSuccess,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[customer.unban] failed to log activity: %v", err)
	}

	// Send unban notification email
	supportEmail := os.Getenv("RESEND_FROM_EMAIL")
	if supportEmail == "" {
		supportEmail = "support@modeva.shop"
	}
	go sendUnbanNotificationEmail(customer.Email, customer.Name, req.Reason, supportEmail)

	log.Printf("[customer.unban] customer %s unbanned by admin %s", customerID, adminIDStr)

	response := models.UnbanCustomerResponse{
		ID:          customerID.String(),
		Name:        customer.Name,
		Email:       customer.Email,
		UnbannedAt:  unbannedAt.Format(time.RFC3339),
		UnbanReason: req.Reason,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Customer unbanned successfully", response))
}

func sendUnbanNotificationEmail(email, name, reason, supportEmail string) {
	resendClient := services.NewResendClient()

	emailData := services.CustomerUnbanEmailData{
		CustomerName:  name,
		CustomerEmail: email,
		UnbanReason:   reason,
		SupportEmail:  supportEmail,
	}

	if err := resendClient.SendCustomerUnbanEmail(emailData); err != nil {
		log.Printf("[customer.unban] failed to send unban email to %s: %v", email, err)
	} else {
		log.Printf("[customer.unban] unban notification sent to %s", email)
	}
}
