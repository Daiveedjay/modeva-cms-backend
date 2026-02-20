package order_controller

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateOrderStatus godoc
// @Summary Update order status (CMS)
// @Description Update an order status. admin_notes is optional for all statuses, but required when status is cancelled (cancellation reason).
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)"
// @Param payload body models.UpdateOrderStatusRequest true "Update payload"
// @Success 200 {object} models.ApiResponse{data=models.UpdateOrderStatusResponse}
// @Failure 400 {object} models.ApiResponse "Bad request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Failure 404 {object} models.ApiResponse "Order not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /admin/orders/{id}/status [patch]
func UpdateOrderStatus(c *gin.Context) {
	log.Printf("[admin.order.update] start path=%s method=%s rawQuery=%s", c.FullPath(), c.Request.Method, c.Request.URL.RawQuery)

	// Optional admin guard (keep if your middleware sets it)
	if v, ok := c.Get("isAdmin"); ok {
		log.Printf("[admin.order.update] ctx isAdmin=%v (type=%T)", v, v)
		if isAdmin, ok2 := v.(bool); ok2 && !isAdmin {
			log.Printf("[admin.order.update] forbidden: isAdmin=false")
			c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Forbidden"))
			return
		}
	} else {
		log.Printf("[admin.order.update] ctx isAdmin missing (confirm admin middleware sets this)")
	}

	orderIDStr := strings.TrimSpace(c.Param("id"))
	if orderIDStr == "" {
		log.Printf("[admin.order.update] bad request: empty order id")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Order ID is required"))
		return
	}

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		log.Printf("[admin.order.update] bad request: invalid order id")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid order ID"))
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[admin.order.update] bad request: bind json err=%v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request body"))
		return
	}

	req.Status = strings.TrimSpace(strings.ToLower(req.Status))

	// admin_notes supported for all statuses, but required for cancelled
	if req.Status == "cancelled" {
		if req.AdminNotes == nil || strings.TrimSpace(*req.AdminNotes) == "" {
			log.Printf("[admin.order.update] bad request: cancelled without admin_notes")
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "admin_notes is required when cancelling an order"))
			return
		}
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	q := `
		UPDATE orders
		SET
			status = ?::text,
			admin_notes = CASE
				WHEN ?::text IS NULL THEN admin_notes
				ELSE ?::text
			END,
			updated_at = NOW(),
			confirmed_at = CASE
				WHEN ?::text = 'confirmed' AND confirmed_at IS NULL THEN NOW()
				ELSE confirmed_at
			END,
			shipped_at = CASE
				WHEN ?::text = 'shipped' AND shipped_at IS NULL THEN NOW()
				ELSE shipped_at
			END,
			delivered_at = CASE
				WHEN ?::text = 'delivered' AND delivered_at IS NULL THEN NOW()
				ELSE delivered_at
			END
		WHERE id = ?
		RETURNING id::text AS id, order_number, status, admin_notes
	`

	log.Printf("[admin.order.update] orderID=%s newStatus=%s adminNotesProvided=%v now=%s",
		orderID, req.Status, req.AdminNotes != nil, time.Now().Format(time.RFC3339))
	log.Printf("[admin.order.update] sql=%s", strings.ReplaceAll(q, "\n", " "))

	var out models.UpdateOrderStatusResponse
	err = config.EcommerceGorm.WithContext(ctx).Raw(
		q,
		req.Status,
		req.AdminNotes,
		req.AdminNotes,
		req.Status,
		req.Status,
		req.Status,
		orderID,
	).Scan(&out).Error
	if err != nil {
		log.Printf("[admin.order.update] ERROR update failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update order"))
		return
	}

	// Check if order was found
	if out.OrderNumber == "" {
		log.Printf("[admin.order.update] order not found id=%s", orderID)
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Order not found"))
		return
	}

	log.Printf("[admin.order.update] success order_number=%s status=%s", out.OrderNumber, out.Status)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Order updated successfully",
		out,
	))
}
