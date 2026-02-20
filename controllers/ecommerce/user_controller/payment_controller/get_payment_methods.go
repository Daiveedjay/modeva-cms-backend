package payment_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetPaymentMethods godoc
// @Summary Get user's payment methods
// @Description Retrieves all active payment methods for the authenticated user (cards only, masked)
// @Tags User - Payment Methods
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.PaymentMethodResponse} "Payment methods retrieved successfully"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Failed to retrieve payment methods"
// @Router /user/payment-methods [get]
func GetPaymentMethods(c *gin.Context) {
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

	ctx, cancel := config.WithTimeout()
	defer cancel()

	var paymentMethods []models.UserPaymentMethod
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		log.Printf("❌ Failed to retrieve payment methods: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to retrieve payment methods"))
		return
	}

	// Transform to response format (masked)
	response := make([]models.PaymentMethodResponse, len(paymentMethods))
	for i, pm := range paymentMethods {
		response[i] = pm.ToResponse()
	}

	log.Printf("✅ Retrieved %d payment methods for user: %s", len(response), userID)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Payment methods retrieved successfully", response))
}
