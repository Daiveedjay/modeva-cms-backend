package payment_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UpdatePaymentMethodRequest struct {
	IsDefault      *bool   `json:"is_default,omitempty"`
	CardholderName *string `json:"cardholder_name,omitempty"`
	// Can't update sensitive fields like card numbers, expiry, etc.
}

// UpdatePaymentMethod godoc
// @Summary Update a payment method
// @Description Updates non-sensitive fields of a user's payment method, such as is_default and cardholder_name. Sensitive card data cannot be updated via this endpoint.
// @Tags User - Payment Methods
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Payment method ID"
// @Param payload body UpdatePaymentMethodRequest true "Payment method fields to update"
// @Success 200 {object} models.ApiResponse "Payment method updated successfully"
// @Failure 400 {object} models.ApiResponse "Invalid request body or no fields to update"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "User does not own this payment method"
// @Failure 404 {object} models.ApiResponse "Payment method not found"
// @Failure 500 {object} models.ApiResponse "Failed to update payment method"
// @Router /user/payment-methods/{id} [patch]
func UpdatePaymentMethod(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Unauthorized"))
		return
	}

	// Parse userID to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(c, "Invalid user ID"))
		return
	}

	paymentMethodIDStr := c.Param("id")
	if paymentMethodIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Payment method ID is required"))
		return
	}

	// Parse payment method ID to UUID
	paymentMethodID, err := uuid.Parse(paymentMethodIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid payment method ID"))
		return
	}

	var req UpdatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, err.Error()))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Find payment method and verify ownership
	var paymentMethod models.UserPaymentMethod
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ?", paymentMethodID).
		First(&paymentMethod).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Payment method not found"))
		return
	}

	// Verify ownership
	if paymentMethod.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "You don't have permission to update this payment method"))
		return
	}

	// Build update map with only provided fields
	updates := make(map[string]interface{})

	if req.CardholderName != nil {
		updates["cardholder_name"] = *req.CardholderName
	}

	if len(updates) == 0 && req.IsDefault == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "No fields to update"))
		return
	}

	// Handle is_default separately if provided (requires transaction)
	if req.IsDefault != nil && *req.IsDefault {
		// Use transaction to handle default logic
		err = config.EcommerceGorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Unset all other defaults for this user
			if err := tx.Model(&models.UserPaymentMethod{}).
				Where("user_id = ? AND id != ?", userID, paymentMethodID).
				Update("is_default", false).Error; err != nil {
				return err
			}

			// Set this as default and apply other updates
			updates["is_default"] = true
			if err := tx.Model(&paymentMethod).Updates(updates).Error; err != nil {
				return err
			}

			return nil
		})
	} else {
		// Simple update without default logic
		if req.IsDefault != nil {
			updates["is_default"] = *req.IsDefault
		}
		err = config.EcommerceGorm.WithContext(ctx).
			Model(&paymentMethod).
			Updates(updates).Error
	}

	if err != nil {
		log.Printf("❌ Failed to update payment method: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update payment method"))
		return
	}

	log.Printf("✅ Payment method updated: %s for user: %s", paymentMethodID, userID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Payment method updated successfully",
		gin.H{"id": paymentMethodID.String()},
	))
}
