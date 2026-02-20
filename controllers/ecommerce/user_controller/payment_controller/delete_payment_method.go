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

// DeletePaymentMethod godoc
// @Summary Delete a payment method
// @Description Soft deletes a user's payment method by marking it as 'deleted'. If deleting the default payment method, the oldest remaining active method becomes the new default.
// @Tags User - Payment Methods
// @Security BearerAuth
// @Produce json
// @Param id path string true "Payment method ID"
// @Success 200 {object} models.ApiResponse "Payment method deleted successfully"
// @Failure 400 {object} models.ApiResponse "Payment method ID missing"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "User does not own this payment method"
// @Failure 404 {object} models.ApiResponse "Payment method not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/payment-methods/{id} [delete]
func DeletePaymentMethod(c *gin.Context) {
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

	// Find payment method and verify ownership
	var paymentMethod models.UserPaymentMethod
	if err := config.EcommerceGorm.
		Where("id = ?", paymentMethodID).
		First(&paymentMethod).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Payment method not found"))
		return
	}

	// Verify ownership
	if paymentMethod.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "You don't have permission to delete this payment method"))
		return
	}

	// Use transaction to handle deletion and default reassignment atomically
	err = config.EcommerceGorm.Transaction(func(tx *gorm.DB) error {
		wasDefault := paymentMethod.IsDefault

		// Soft delete (set status to 'deleted')
		if err := tx.Model(&paymentMethod).
			Update("status", "deleted").Error; err != nil {
			return err
		}

		// If we just deleted the default payment method, find and set a new default
		if wasDefault {
			var newDefault models.UserPaymentMethod

			// Find the oldest active payment method for this user (excluding the one we just deleted)
			err := tx.Where("user_id = ? AND status = ? AND id != ?", userID, "active", paymentMethodID).
				Order("created_at ASC").
				First(&newDefault).Error

			if err == nil {
				// Found a replacement, set it as default
				if err := tx.Model(&newDefault).
					Update("is_default", true).Error; err != nil {
					return err
				}
				log.Printf("✅ New default payment method set: %s (oldest) for user: %s", newDefault.ID, userID)
			} else if err != gorm.ErrRecordNotFound {
				// Real error (not just "no other payment methods")
				return err
			}
			// If no other payment methods exist, that's fine - no error
		}

		return nil
	})
	if err != nil {
		log.Printf("❌ Failed to delete payment method: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to delete payment method"))
		return
	}

	log.Printf("✅ Payment method deleted: %s for user: %s", paymentMethodID, userID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Payment method deleted successfully",
		nil,
	))
}
