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

// BanCustomer godoc
// @Summary Ban a customer
// @Description Ban a customer account, preventing all access while preserving data
// @Tags CMS - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Param banRequest body models.BanCustomerRequest true "Ban reason"
// @Success 200 {object} models.ApiResponse{data=models.BanCustomerResponse}
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 404 {object} models.ApiResponse "Customer not found"
// @Failure 409 {object} models.ApiResponse "Customer already banned"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /api/v1/admin/customers/{id}/ban [post]
func BanCustomer(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid customer ID"))
		return
	}

	var req models.BanCustomerRequest
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

	// Check if already banned
	if customer.IsBanned {
		c.JSON(http.StatusConflict, models.ErrorResponse(c, "Customer is already banned"))
		return
	}

	// Check if deleted
	if customer.DeletedAt != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(c, "Cannot ban a deleted customer"))
		return
	}

	// Ban the customer
	bannedAt := time.Now()
	if err := config.EcommerceGorm.WithContext(ctx).Model(&customer).Updates(map[string]interface{}{
		"is_banned":  true,
		"ban_reason": req.Reason,
		"banned_at":  bannedAt,
		"status":     "banned",
	}).Error; err != nil {
		log.Printf("[customer.ban] failed to ban customer: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to ban customer"))
		return
	}

	adminIDStr, _ := c.Get("adminID")
	adminEmail, _ := c.Get("adminEmail")

	// ✅ LOG THE ACTIVITY
	changes := map[string]interface{}{
		"ban_reason": req.Reason,
		"banned_at":  bannedAt,
	}
	changesJSON, _ := json.Marshal(changes)

	adminID, _ := uuid.Parse(adminIDStr.(string))
	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      adminID,
		AdminEmail:   adminEmail.(string),
		Action:       "customer_banned",
		ResourceType: models.ResourceTypeCustomer,
		ResourceID:   customerID.String(),
		ResourceName: customer.Name,
		Changes:      datatypes.JSON(changesJSON),
		Status:       models.StatusSuccess,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[customer.ban] failed to log activity: %v", err)
	}

	// Send ban notification email
	supportEmail := os.Getenv("RESEND_FROM_EMAIL")
	if supportEmail == "" {
		supportEmail = "support@modeva.shop"
	}
	go sendBanNotificationEmail(customer.Email, customer.Name, req.Reason, supportEmail)

	log.Printf("[customer.ban] customer %s banned by admin %s", customerID, adminIDStr)

	response := models.BanCustomerResponse{
		ID:        customerID.String(),
		Name:      customer.Name,
		Email:     customer.Email,
		BannedAt:  bannedAt.Format(time.RFC3339),
		BanReason: req.Reason,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Customer banned successfully", response))
}

func sendBanNotificationEmail(email, name, reason, supportEmail string) {
	resendClient := services.NewResendClient()

	emailData := services.CustomerBanEmailData{
		CustomerName:  name,
		CustomerEmail: email,
		BanReason:     reason,
		SupportEmail:  supportEmail,
	}

	if err := resendClient.SendCustomerBanEmail(emailData); err != nil {
		log.Printf("[customer.ban] failed to send ban email to %s: %v", email, err)
	} else {
		log.Printf("[customer.ban] ban notification sent to %s", email)
	}
}
