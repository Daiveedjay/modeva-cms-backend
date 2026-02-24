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

// DeleteCustomer godoc
// @Summary Delete a customer
// @Description Permanently delete a customer (soft delete)
// @Tags CMS - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Param deleteRequest body models.DeleteCustomerRequest true "Deletion reason"
// @Success 200 {object} models.ApiResponse{data=models.DeleteCustomerResponse}
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 404 {object} models.ApiResponse "Customer not found"
// @Failure 409 {object} models.ApiResponse "Customer already deleted"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /api/v1/admin/customers/{id} [delete]
func DeleteCustomer(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid customer ID"))
		return
	}

	var req models.DeleteCustomerRequest
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

	// Check if already deleted
	if customer.DeletedAt != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(c, "Customer is already deleted"))
		return
	}

	// Soft delete the customer
	deletedAt := time.Now()
	if err := config.EcommerceGorm.WithContext(ctx).Model(&customer).Updates(map[string]interface{}{
		"deleted_at":      deletedAt,
		"deletion_reason": req.Reason,
		"status":          "deleted",
	}).Error; err != nil {
		log.Printf("[customer.delete] failed to delete customer: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to delete customer"))
		return
	}

	adminIDStr, _ := c.Get("adminID")
	adminEmail, _ := c.Get("adminEmail")

	// ✅ LOG THE ACTIVITY
	changes := map[string]interface{}{
		"deletion_reason": req.Reason,
		"deleted_at":      deletedAt,
	}
	changesJSON, _ := json.Marshal(changes)

	adminID, _ := uuid.Parse(adminIDStr.(string))
	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      adminID,
		AdminEmail:   adminEmail.(string),
		Action:       "customer_deleted",
		ResourceType: models.ResourceTypeCustomer,
		ResourceID:   customerID.String(),
		ResourceName: customer.Name,
		Changes:      datatypes.JSON(changesJSON),
		Status:       models.StatusSuccess,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[customer.delete] failed to log activity: %v", err)
	}

	// Send deletion confirmation email
	supportEmail := os.Getenv("RESEND_FROM_EMAIL")
	if supportEmail == "" {
		supportEmail = "support@modeva.shop"
	}
	go sendDeletionConfirmationEmail(customer.Email, customer.Name, supportEmail)

	log.Printf("[customer.delete] customer %s deleted by admin %s", customerID, adminIDStr)

	response := models.DeleteCustomerResponse{
		ID:             customerID.String(),
		Name:           customer.Name,
		DeletedAt:      deletedAt.Format(time.RFC3339),
		DeletionReason: req.Reason,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Customer deleted successfully", response))
}

func sendDeletionConfirmationEmail(email, name, supportEmail string) {
	resendClient := services.NewResendClient()

	emailData := services.CustomerDeletionEmailData{
		CustomerName:  name,
		CustomerEmail: email,
		SupportEmail:  supportEmail,
	}

	if err := resendClient.SendCustomerDeletionEmail(emailData); err != nil {
		log.Printf("[customer.delete] failed to send deletion email to %s: %v", email, err)
	} else {
		log.Printf("[customer.delete] deletion confirmation sent to %s", email)
	}
}
