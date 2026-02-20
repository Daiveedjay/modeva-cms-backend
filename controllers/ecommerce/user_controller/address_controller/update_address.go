package address_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateAddress godoc
// @Summary Update address
// @Description Update specific fields of an address
// @Tags User - Addresses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Address ID"
// @Param address body models.UpdateAddressRequest true "Fields to update"
// @Success 200 {object} models.ApiResponse "Address updated successfully"
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Permission denied"
// @Failure 404 {object} models.ApiResponse "Address not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/addresses/{id} [patch]
func UpdateAddress(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Permission denied"))
		return
	}

	var req models.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, err.Error()))
		return
	}

	// Build update map with only provided fields
	updates := make(map[string]interface{})

	if req.Label != nil {
		updates["label"] = *req.Label
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Street != nil {
		updates["street"] = *req.Street
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.State != nil {
		updates["state"] = *req.State
	}
	if req.Zip != nil {
		updates["zip"] = *req.Zip
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "No fields to update"))
		return
	}

	// Perform update (GORM will automatically update updated_at)
	if err := config.EcommerceGorm.WithContext(ctx).
		Model(&address).
		Updates(updates).Error; err != nil {
		log.Printf("❌ Failed to update address: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update address"))
		return
	}

	log.Printf("✅ Address updated: %s", addressID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Address updated successfully",
		nil,
	))
}
