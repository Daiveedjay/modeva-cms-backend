package order_controller

import (
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetOrderDetails godoc
// @Summary Get order details
// @Description Retrieve complete order details including all items
// @Tags User - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} models.ApiResponse{data=models.OrderWithItems}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Permission denied"
// @Failure 404 {object} models.ApiResponse "Order not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/orders/{id} [get]
func GetOrderDetails(c *gin.Context) {
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

	orderIDStr := c.Param("id")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Order ID is required"))
		return
	}

	// Parse orderID to UUID
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid order ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Get order details using raw SQL (to handle all nullable fields properly)
	var order models.Order
	err = config.EcommerceGorm.WithContext(ctx).Raw(`
		SELECT 
			id::text AS id, 
			user_id::text AS user_id, 
			order_number, 
			payment_method_id::text AS payment_method_id, 
			address_id::text AS address_id,
			payment_method_type, 
			payment_method_last4, 
			address_snapshot,
			subtotal, 
			tax, 
			shipping_cost, 
			discount, 
			total_amount, 
			status,
			customer_notes, 
			admin_notes, 
			created_at, 
			updated_at,
			confirmed_at, 
			shipped_at, 
			delivered_at
		FROM orders
		WHERE id = ?
	`, orderID).Scan(&order).Error
	if err != nil {
		log.Printf("❌ Failed to fetch order: %v", err)
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Order not found"))
		return
	}

	// Check if order was found
	if order.OrderNumber == "" {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Order not found"))
		return
	}

	// Verify ownership
	if order.UserID != userID.String() {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Permission denied"))
		return
	}

	// Get order items
	var items []models.OrderItem
	err = config.EcommerceGorm.WithContext(ctx).Raw(`
		SELECT 
			id::text AS id, 
			order_id::text AS order_id, 
			user_id::text AS user_id, 
			product_id::text AS product_id, 
			product_name,
			variant_size, 
			variant_color, 
			price, 
			quantity, 
			subtotal,
			status, 
			created_at, 
			updated_at
		FROM order_items
		WHERE order_id = ?
		ORDER BY created_at ASC
	`, orderID).Scan(&items).Error
	if err != nil {
		log.Printf("❌ Failed to fetch order items: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch order items"))
		return
	}

	// Combine order and items
	orderWithItems := models.OrderWithItems{
		Order: order,
		Items: items,
	}

	log.Printf("✅ Fetched order %s with %d items", order.OrderNumber, len(items))

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Order details retrieved successfully",
		orderWithItems,
	))
}
