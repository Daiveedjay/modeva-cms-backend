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

// SetDefaultPaymentMethod godoc
// @Summary Set default payment method
// @Description Marks a specific active payment method as the default for the authenticated user. Other payment methods will be unset automatically.
// @Tags User - Payment Methods
// @Security BearerAuth
// @Produce json
// @Param id path string true "Payment method ID"
// @Success 200 {object} models.ApiResponse "Default payment method updated successfully"
// @Failure 400 {object} models.ApiResponse "Payment method ID is required"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "User does not own this payment method"
// @Failure 404 {object} models.ApiResponse "Payment method not found"
// @Failure 500 {object} models.ApiResponse "Failed to set default payment method"
// @Router /user/payment-methods/{id}/default [patch]
func SetDefaultPaymentMethod(c *gin.Context) {
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

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Find payment method and verify ownership
	var paymentMethod models.UserPaymentMethod
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ? AND status = ?", paymentMethodID, "active").
		First(&paymentMethod).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Payment method not found"))
		return
	}

	// Verify ownership
	if paymentMethod.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "You don't have permission to modify this payment method"))
		return
	}

	// Use transaction to ensure atomicity
	err = config.EcommerceGorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset all other defaults for this user
		if err := tx.Model(&models.UserPaymentMethod{}).
			Where("user_id = ? AND id != ?", userID, paymentMethodID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set this payment method as default
		if err := tx.Model(&paymentMethod).
			Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("❌ Failed to set default payment method: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to set default payment method"))
		return
	}

	log.Printf("✅ Default payment method set: %s for user: %s", paymentMethodID, userID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Default payment method updated successfully",
		gin.H{"id": paymentMethodID.String()},
	))
}
