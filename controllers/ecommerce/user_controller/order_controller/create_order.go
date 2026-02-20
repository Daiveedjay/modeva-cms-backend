package order_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// detectDevice determines device type from User-Agent string
func detectDevice(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") ||
		strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipod") {
		return "mobile"
	}

	if strings.Contains(ua, "ipad") ||
		strings.Contains(ua, "tablet") ||
		strings.Contains(ua, "kindle") {
		return "tablet"
	}

	return "desktop"
}

// CreateOrder godoc
// @Summary Create new order (checkout)
// @Description Create a new order from cart items with payment and address
// @Tags User - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.CreateOrderRequest true "Order details"
// @Success 201 {object} models.ApiResponse{data=object{order_id=string,order_number=string}} "Order created successfully"
// @Failure 400 {object} models.ApiResponse "Invalid request"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 404 {object} models.ApiResponse "Payment method or address not found"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /user/orders [post]
func CreateOrder(c *gin.Context) {
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

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, err.Error()))
		return
	}

	// Validate cart items
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Cart cannot be empty"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Parse payment method ID and address ID
	paymentMethodID, err := uuid.Parse(req.PaymentMethodID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid payment method ID"))
		return
	}

	addressID, err := uuid.Parse(req.AddressID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid address ID"))
		return
	}

	// Verify payment method ownership
	var paymentMethod models.UserPaymentMethod
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ? AND status = ?", paymentMethodID, "active").
		First(&paymentMethod).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Payment method not found"))
		return
	}
	if paymentMethod.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Invalid payment method"))
		return
	}

	// Verify address ownership
	var address models.Address
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ? AND status = ?", addressID, "active").
		First(&address).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Address not found"))
		return
	}
	if address.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Invalid address"))
		return
	}

	// Detect device type from User-Agent
	deviceType := detectDevice(c.Request.UserAgent())

	// Start transaction
	var orderID uuid.UUID
	var orderNumber string
	var totalAmount float64

	err = config.EcommerceGorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create address snapshot
		addressSnapshot := map[string]interface{}{
			"label":      address.Label,
			"first_name": address.FirstName,
			"last_name":  address.LastName,
			"street":     address.Street,
			"city":       address.City,
			"state":      address.State,
			"zip":        address.Zip,
			"country":    address.Country,
			"phone":      address.Phone,
		}
		addressJSON, _ := json.Marshal(addressSnapshot)

		// Fetch current product prices from CMS DB
		productIDs := make([]uuid.UUID, len(req.Items))
		for i, item := range req.Items {
			pid, err := uuid.Parse(item.ProductID)
			if err != nil {
				return fmt.Errorf("invalid product ID: %s", item.ProductID)
			}
			productIDs[i] = pid
		}

		log.Printf("🔍 Querying CMS DB for products: %v", productIDs)

		var products []struct {
			ID    uuid.UUID `gorm:"column:id"`
			Name  string    `gorm:"column:name"`
			Price float64   `gorm:"column:price"`
		}

		if err := config.CmsGorm.WithContext(ctx).
			Table("products").
			Select("id, name, price").
			Where("id IN ? AND status = ?", productIDs, "Active").
			Find(&products).Error; err != nil {
			log.Printf("❌ Failed to fetch product prices: %v", err)
			return fmt.Errorf("failed to validate products")
		}

		// Build product map
		productPrices := make(map[string]ProductInfo)
		for _, p := range products {
			productPrices[p.ID.String()] = ProductInfo{Name: p.Name, Price: p.Price}
		}

		// Validate all products exist
		for _, item := range req.Items {
			if _, exists := productPrices[item.ProductID]; !exists {
				return fmt.Errorf("product %s not found or inactive", item.ProductID)
			}
		}

		// Calculate order totals
		var subtotal float64 = 0
		for _, item := range req.Items {
			productInfo := productPrices[item.ProductID]
			itemSubtotal := productInfo.Price * float64(item.Quantity)
			subtotal += itemSubtotal
		}

		// ✅ Calculate tax (10%) and shipping (5%)
		tax := subtotal * 0.10
		shippingCost := subtotal * 0.05
		discount := 0.0
		totalAmount = subtotal + tax + shippingCost - discount

		// Create order using raw SQL (to get order_number from trigger)
		orderID = uuid.Must(uuid.NewV7())
		last4 := paymentMethod.GetLast4()

		if err := tx.Exec(`
    INSERT INTO orders 
    (id, user_id, order_number, payment_method_id, address_id, 
     payment_method_type, payment_method_last4, address_snapshot,
     subtotal, tax, shipping_cost, discount, total_amount, 
     status, customer_notes, device_type, created_at, updated_at)
    VALUES (?, ?, '', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			orderID,
			userID,
			paymentMethodID,
			addressID,
			paymentMethod.Type,
			&last4,
			addressJSON,
			subtotal,
			tax,
			shippingCost,
			discount,
			totalAmount,
			"pending",
			req.CustomerNotes,
			deviceType, // ✅ Added device_type
		).Error; err != nil {
			log.Printf("❌ Failed to create order: %v", err)
			return fmt.Errorf("failed to create order")
		}

		// Create order items
		for _, item := range req.Items {
			productInfo := productPrices[item.ProductID]
			itemSubtotal := productInfo.Price * float64(item.Quantity)
			itemProductID, _ := uuid.Parse(item.ProductID)

			orderItem := struct {
				ID           uuid.UUID
				OrderID      uuid.UUID
				UserID       uuid.UUID
				ProductID    uuid.UUID
				ProductName  string
				VariantSize  *string
				VariantColor *string
				Price        float64
				Quantity     int
				Subtotal     float64
				Status       string
			}{
				ID:           uuid.Must(uuid.NewV7()),
				OrderID:      orderID,
				UserID:       userID,
				ProductID:    itemProductID,
				ProductName:  productInfo.Name,
				VariantSize:  item.VariantSize,
				VariantColor: item.VariantColor,
				Price:        productInfo.Price,
				Quantity:     item.Quantity,
				Subtotal:     itemSubtotal,
				Status:       "pending",
			}

			if err := tx.Table("order_items").Create(&orderItem).Error; err != nil {
				log.Printf("❌ Failed to create order item: %v", err)
				return fmt.Errorf("failed to create order items")
			}
		}

		// Get generated order number
		if err := tx.Raw(`SELECT order_number FROM orders WHERE id = ?`, orderID).Scan(&orderNumber).Error; err != nil {
			log.Printf("❌ Failed to fetch order number: %v", err)
			return fmt.Errorf("failed to create order")
		}

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, err.Error()))
		return
	}

	log.Printf("✅ Order created: %s (%s) for user: %s - Total: $%.2f - Device: %s",
		orderNumber, orderID, userID, totalAmount, deviceType)

	c.JSON(http.StatusCreated, models.SuccessResponse(
		c,
		"Order created successfully",
		gin.H{
			"order_id":     orderID.String(),
			"order_number": orderNumber,
			"total_amount": totalAmount,
		},
	))
}

// Helper struct for product info
type ProductInfo struct {
	Name  string
	Price float64
}
