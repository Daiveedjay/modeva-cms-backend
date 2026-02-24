package customer_controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SendCustomerEmail godoc
// @Summary Send email to customer
// @Description Send a custom email to a specific customer
// @Tags CMS - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Param emailRequest body models.SendCustomerEmailRequest true "Email details"
// @Success 200 {object} models.ApiResponse{data=models.SendCustomerEmailResponse}
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 404 {object} models.ApiResponse "Customer not found"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /api/v1/admin/customers/{id}/send-email [post]
func SendCustomerEmail(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid customer ID"))
		return
	}

	var req models.SendCustomerEmailRequest
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

	// Send email asynchronously
	sentAt := time.Now()
	go sendCustomerEmailAsync(customer.Email, customer.Name, req.Subject, req.Message)

	adminIDStr, _ := c.Get("adminID")
	adminEmail, _ := c.Get("adminEmail")

	// ✅ LOG THE ACTIVITY - Customer email sent
	changes := map[string]interface{}{
		"subject":   req.Subject,
		"recipient": customer.Email,
		"sent_at":   sentAt,
	}
	changesJSON, _ := json.Marshal(changes)

	adminID, _ := uuid.Parse(adminIDStr.(string))
	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      adminID,
		AdminEmail:   adminEmail.(string),
		Action:       "customer_email_sent", // You might want to add this to your ActionType constants
		ResourceType: models.ResourceTypeCustomer,
		ResourceID:   customerID.String(),
		ResourceName: customer.Name,
		Changes:      datatypes.JSON(changesJSON),
		Status:       models.StatusSuccess,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[customer.send-email] failed to log activity: %v", err)
	}

	log.Printf("[customer.send-email] email sent to %s (customer: %s)", customer.Email, customerID)

	response := models.SendCustomerEmailResponse{
		Email:   customer.Email,
		Subject: req.Subject,
		SentAt:  sentAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Email sent successfully", response))
}

// sendCustomerEmailAsync sends the email asynchronously
func sendCustomerEmailAsync(email, name, subject, message string) {
	resendClient := services.NewResendClient()

	emailData := services.CustomerEmailData{
		CustomerName:  name,
		CustomerEmail: email,
		Subject:       subject,
		Message:       message,
	}

	if err := resendClient.SendCustomerEmail(emailData); err != nil {
		log.Printf("[customer.send-email] failed to send email to %s: %v", email, err)
	} else {
		log.Printf("[customer.send-email] email successfully sent to %s", email)
	}
}
