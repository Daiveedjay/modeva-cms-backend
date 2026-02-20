package customer_controller

import (
	"log"
	"net/http"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateCustomerDetails godoc
// @Summary Update customer information
// @Description Update customer name, phone, and account status (active, suspended, banned, deleted)
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Param customer body models.UpdateCustomerRequest true "Customer update data"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse "Bad request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 404 {object} models.ApiResponse "Customer not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /admin/customers/{id} [patch]
func UpdateCustomerDetails(c *gin.Context) {
	log.Printf("[admin.update-customer] start path=%s method=%s", c.FullPath(), c.Request.Method)

	customerIDStr := c.Param("id")
	if customerIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Customer ID is required"))
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid customer ID"))
		return
	}

	var input models.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, err.Error()))
		return
	}

	if input.Status != nil {
		status := strings.ToLower(*input.Status)

		if status == "banned" {
			if input.BanReason == nil || strings.TrimSpace(*input.BanReason) == "" {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "ban_reason is required when status is 'banned'"))
				return
			}
		}

		if status == "suspended" {
			if input.SuspendedUntil == nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "suspended_until is required when status is 'suspended'"))
				return
			}
			if input.SuspendedReason == nil || strings.TrimSpace(*input.SuspendedReason) == "" {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "suspended_reason is required when status is 'suspended'"))
				return
			}
		}
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	var existingCustomer struct {
		ID     string
		Name   string
		Status string
	}

	err = config.EcommerceGorm.WithContext(ctx).
		Table("users").
		Select("id::text AS id, name, status").
		Where("id = ?", customerID).
		First(&existingCustomer).Error
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Customer not found"))
		return
	}

	updates := make(map[string]interface{})

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name != "" {
			updates["name"] = name
		}
	}

	// ✅ Correct phone handling (USER phone only)
	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		if phone == "" {
			updates["phone"] = nil // store NULL
		} else {
			updates["phone"] = phone
		}
	}

	if input.Status != nil {
		newStatus := strings.ToLower(*input.Status)
		updates["status"] = newStatus

		switch newStatus {
		case "banned":
			updates["ban_reason"] = *input.BanReason
			updates["suspended_until"] = nil
			updates["suspended_reason"] = nil

		case "suspended":
			updates["suspended_until"] = *input.SuspendedUntil
			updates["suspended_reason"] = *input.SuspendedReason
			updates["ban_reason"] = nil

		case "active":
			updates["ban_reason"] = nil
			updates["suspended_until"] = nil
			updates["suspended_reason"] = nil

		case "deleted":
			if input.BanReason != nil {
				updates["ban_reason"] = *input.BanReason
			}
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "No fields to update"))
		return
	}

	result := config.EcommerceGorm.WithContext(ctx).
		Table("users").
		Where("id = ?", customerID).
		Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update customer"))
		return
	}

	var updatedCustomer struct {
		ID              string  `gorm:"column:id"`
		Name            string  `gorm:"column:name"`
		Email           string  `gorm:"column:email"`
		Phone           *string `gorm:"column:phone"`
		Status          string  `gorm:"column:status"`
		BanReason       *string `gorm:"column:ban_reason"`
		SuspendedUntil  *string `gorm:"column:suspended_until"`
		SuspendedReason *string `gorm:"column:suspended_reason"`
	}

	_ = config.EcommerceGorm.WithContext(ctx).
		Table("users").
		Select(`
			id::text AS id,
			name,
			email,
			phone,
			status,
			ban_reason,
			suspended_until::text AS suspended_until,
			suspended_reason
		`).
		Where("id = ?", customerID).
		First(&updatedCustomer)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Customer updated successfully",
		updatedCustomer,
	))
}
