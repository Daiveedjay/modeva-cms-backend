package payment_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AddPaymentMethod godoc
// @Summary Add a new payment method
// @Description Adds a payment method (card only) for the authenticated user
// @Tags User - Payment Methods
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body models.AddPaymentMethodRequest true "Payment method payload"
// @Success 201 {object} models.ApiResponse "Payment method added successfully"
// @Failure 400 {object} models.ApiResponse "Invalid or missing request fields"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Failed to add payment method"
// @Router /user/payment-methods [post]
func AddPaymentMethod(c *gin.Context) {
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

	var req models.AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request payload"))
		return
	}

	// Validate card type (already validated by binding, but double check)
	if req.CardType != "credit" && req.CardType != "debit" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid card type (must be 'credit' or 'debit')"))
		return
	}

	// Validate expiration
	if req.ExpMonth < 1 || req.ExpMonth > 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid expiration month"))
		return
	}

	currentYear := time.Now().Year()
	if req.ExpYear < currentYear {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid expiration year"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Use transaction to handle default payment method logic
	var paymentMethod models.UserPaymentMethod
	err = config.EcommerceGorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If this is set as default, unset other defaults first
		if req.IsDefault {
			if err := tx.Model(&models.UserPaymentMethod{}).
				Where("user_id = ? AND is_default = ?", userID, true).
				Update("is_default", false).Error; err != nil {
				log.Printf("❌ Failed to unset other default payment methods: %v", err)
				return err
			}
		}

		// Create payment method
		paymentMethod = models.UserPaymentMethod{
			UserID:                  userID,
			Type:                    "card",
			IsDefault:               req.IsDefault,
			Provider:                req.Provider,
			ProviderPaymentMethodID: req.ProviderPaymentMethodID,
			CardType:                req.CardType,
			CardBrand:               req.CardBrand,
			CardNumber:              req.CardNumber,
			ExpMonth:                req.ExpMonth,
			ExpYear:                 req.ExpYear,
			CVV:                     &req.CVV,
			CardholderName:          req.CardholderName,
			Status:                  "active",
		}

		if err := tx.Create(&paymentMethod).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("❌ Failed to add payment method: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to add payment method"))
		return
	}

	log.Printf("✅ Payment method added: %s (card) for user: %s", paymentMethod.ID, userID)

	// Return masked response (never expose card number or CVV)
	c.JSON(http.StatusCreated, models.SuccessResponse(
		c,
		"Payment method added successfully",
		paymentMethod.ToResponse(),
	))
}
