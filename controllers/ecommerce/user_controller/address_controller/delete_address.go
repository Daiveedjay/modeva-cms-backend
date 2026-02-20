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

// DeleteAddress godoc
// @Summary Delete address
// @Description Soft delete an address (sets status to 'deleted'). If deleting the default address, the oldest remaining active address becomes the new default.
// @Tags User - Addresses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Address ID"
// @Success 200 {object} models.ApiResponse "Address deleted successfully"
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Failure 404 {object} models.ApiResponse "Address not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/addresses/{id} [delete]
func DeleteAddress(c *gin.Context) {
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

	// Find address and verify ownership
	var address models.Address
	if err := config.EcommerceGorm.
		Where("id = ?", addressID).
		First(&address).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Address not found"))
		return
	}

	// Verify ownership
	if address.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "You don't have permission to delete this address"))
		return
	}

	// Use transaction to handle deletion and default reassignment atomically
	err = config.EcommerceGorm.Transaction(func(tx *gorm.DB) error {
		wasDefault := address.IsDefault

		// Soft delete - update status to 'deleted'
		if err := tx.Model(&address).
			Update("status", "deleted").Error; err != nil {
			return err
		}

		// If we just deleted the default address, find and set a new default
		if wasDefault {
			var newDefault models.Address

			// Find the oldest active address for this user (excluding the one we just deleted)
			err := tx.Where("user_id = ? AND status = ? AND id != ?", userID, "active", addressID).
				Order("created_at ASC").
				First(&newDefault).Error

			if err == nil {
				// Found a replacement, set it as default
				if err := tx.Model(&newDefault).
					Update("is_default", true).Error; err != nil {
					return err
				}
				log.Printf("✅ New default address set: %s (oldest) for user: %s", newDefault.ID, userID)
			} else if err != gorm.ErrRecordNotFound {
				// Real error (not just "no other addresses")
				return err
			}
			// If no other addresses exist, that's fine - no error
		}

		return nil
	})
	if err != nil {
		log.Printf("❌ Failed to delete address: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to delete address"))
		return
	}

	log.Printf("✅ Address deleted: %s for user: %s", addressID, userID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Address deleted successfully",
		nil,
	))
}
