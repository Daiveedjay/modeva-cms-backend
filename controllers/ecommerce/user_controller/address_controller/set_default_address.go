package address_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SetDefaultAddress godoc
// @Summary Set default address
// @Description Set an address as the default for the user
// @Tags User - Addresses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Address ID"
// @Success 200 {object} models.ApiResponse{data=object{id=string}} "Default address updated successfully"
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Failure 404 {object} models.ApiResponse "Address not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/addresses/{id}/default [patch]
func SetDefaultAddress(c *gin.Context) {
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

	addressIDStr := c.Param("id")
	if addressIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Address ID is required"))
		return
	}

	// Parse addressID to UUID
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid address ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Find address and verify ownership
	var address models.Address
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ? AND status = ?", addressID, "active").
		First(&address).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Address not found"))
		return
	}

	// Verify ownership
	if address.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "You don't have permission to modify this address"))
		return
	}

	// Use transaction to ensure atomicity
	err = config.EcommerceGorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset all other defaults for this user
		if err := tx.Model(&models.Address{}).
			Where("user_id = ? AND id != ?", userID, addressID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set this address as default
		if err := tx.Model(&address).
			Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("❌ Failed to set default address: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to set default address"))
		return
	}

	log.Printf("✅ Default address set: %s for user: %s", addressID, userID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Default address updated successfully",
		map[string]string{"id": addressID.String()},
	))
}
