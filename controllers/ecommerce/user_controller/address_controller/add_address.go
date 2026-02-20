package address_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AddAddress godoc
// @Summary Add new address
// @Description Add a new address for the authenticated user
// @Tags User - Addresses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param address body models.AddAddressRequest true "Address details"
// @Success 201 {object} models.ApiResponse{data=object{id=string}} "Address added successfully"
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/addresses [post]
func AddAddress(c *gin.Context) {
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

	var req models.AddAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, err.Error()))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// If this is set as default, unset other defaults first
	if req.IsDefault {
		if err := config.EcommerceGorm.WithContext(ctx).
			Model(&models.Address{}).
			Where("user_id = ? AND is_default = ?", userID, true).
			Update("is_default", false).Error; err != nil {
			log.Printf("❌ Failed to unset other default addresses: %v", err)
		}
	}

	// Create new address
	address := models.Address{
		UserID:    userID,
		Label:     req.Label,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Street:    req.Street,
		City:      req.City,
		State:     req.State,
		Zip:       req.Zip,
		Country:   req.Country,
		Phone:     req.Phone,
		IsDefault: req.IsDefault,
		Status:    "active",
	}

	if err := config.EcommerceGorm.WithContext(ctx).Create(&address).Error; err != nil {
		log.Printf("❌ Failed to add address: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to add address"))
		return
	}

	log.Printf("✅ Address added: %s (default: %v) for user: %s", address.ID, req.IsDefault, userID)

	c.JSON(http.StatusCreated, models.SuccessResponse(
		c,
		"Address added successfully",
		map[string]string{"id": address.ID.String()},
	))
}
