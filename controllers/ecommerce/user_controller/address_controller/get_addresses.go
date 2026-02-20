package address_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAddresses godoc
// @Summary Get user addresses
// @Description Retrieve all active addresses for the authenticated user
// @Tags User - Addresses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=[]models.AddressResponse}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/addresses [get]
func GetAddresses(c *gin.Context) {
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

	var addresses []models.Address
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error; err != nil {
		log.Printf("❌ Failed to fetch addresses: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch addresses"))
		return
	}

	// Convert to response format
	addressResponses := make([]models.AddressResponse, len(addresses))
	for i, addr := range addresses {
		addressResponses[i] = models.AddressResponse{
			ID:        addr.ID,
			Label:     addr.Label,
			FirstName: addr.FirstName,
			LastName:  addr.LastName,
			Street:    addr.Street,
			City:      addr.City,
			State:     addr.State,
			Zip:       addr.Zip,
			Country:   addr.Country,
			Phone:     addr.Phone,
			IsDefault: addr.IsDefault,
			CreatedAt: addr.CreatedAt,
			UpdatedAt: addr.UpdatedAt,
		}
	}

	log.Printf("✅ Fetched %d addresses for user: %s", len(addressResponses), userID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Addresses retrieved successfully",
		addressResponses,
	))
}
