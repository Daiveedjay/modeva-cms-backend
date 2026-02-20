package order_controller

import (
	"log"
	"net/http"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetOrderDetailsByID godoc
// @Summary Get order details
// @Description Retrieve full order details including customer, address snapshot, items and product images
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} models.ApiResponse{data=models.CMSOrderDetailsResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /admin/orders/{id} [get]
func GetOrderDetailsByID(c *gin.Context) {
	orderIDStr := strings.TrimSpace(c.Param("id"))
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Order ID is required"))
		return
	}

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid order ID"))
		return
	}

	log.Printf("[admin.order-details] Fetching order: %s", orderID.String())

	ecomDB := config.EcommerceGorm
	cmsDB := config.CmsGorm

	var res models.CMSOrderDetailsResponse

	// =====================================
	// 1. Order + customer + address snapshot
	// =====================================
	err = ecomDB.Raw(`
		SELECT
			o.id::text AS id,
			o.order_number,
			o.status,
			o.created_at,

			u.id::text AS customer_id,
			COALESCE(NULLIF(u.name, ''), u.email) AS customer_name,
			u.email AS customer_email,

			o.payment_method_type,
			o.payment_method_last4,

			o.subtotal,
			o.shipping_cost,
			o.tax,
			o.discount,
			o.total_amount,

			o.customer_notes,
			o.admin_notes,
			o.address_snapshot::text AS address_snapshot,

			(o.address_snapshot->>'label')      AS label,
			(o.address_snapshot->>'first_name') AS first_name,
			(o.address_snapshot->>'last_name')  AS last_name,
			(o.address_snapshot->>'phone')      AS phone,
			(o.address_snapshot->>'street')     AS street,
			(o.address_snapshot->>'city')       AS city,
			(o.address_snapshot->>'state')      AS state,
			(o.address_snapshot->>'zip')        AS zip,
			(o.address_snapshot->>'country')    AS country
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE o.id = $1
		LIMIT 1
	`, orderID).Scan(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[admin.order-details] Order not found: %s", orderID.String())
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Order not found"))
		} else {
			log.Printf("[admin.order-details] ERROR fetching order: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch order"))
		}
		return
	}

	log.Printf("[admin.order-details] Order found: %s", res.OrderNumber)

	// =====================================
	// 2. Payment label (card-only)
	// =====================================
	res.PaymentMethodLabel = "Credit Card"
	if res.PaymentMethodLast4 != nil && *res.PaymentMethodLast4 != "" {
		res.PaymentMethodLabel = "Credit Card •••• " + *res.PaymentMethodLast4
	}

	// =====================================
	// 3. Order items
	// =====================================
	items := make([]models.OrderItemWithImage, 0)

	if err := ecomDB.
		Table("order_items").
		Select(`
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
		`).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Scan(&items).Error; err != nil {
		log.Printf("[admin.order-details] ERROR fetching order items: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch order items"))
		return
	}

	log.Printf("[admin.order-details] Found %d items", len(items))

	// =====================================
	// 4. Collect unique product IDs
	// =====================================
	productIDs := make([]string, 0)
	seen := make(map[string]struct{})

	for _, item := range items {
		if _, ok := seen[item.ProductID]; !ok {
			seen[item.ProductID] = struct{}{}
			productIDs = append(productIDs, item.ProductID)
		}
	}

	// =====================================
	// 5. Fetch product images (CMS DB)
	// =====================================
	imageByProductID := make(map[string]*string)

	if len(productIDs) > 0 {
		type row struct {
			ID         string
			PrimaryURL *string
		}

		var rows []row

		if err := cmsDB.Raw(`
			SELECT
				id::text AS id,
				NULLIF(media->'primary'->>'url', '') AS primary_url
			FROM products
			WHERE id::text = ANY($1)
		`, productIDs).Scan(&rows).Error; err == nil {
			for _, r := range rows {
				imageByProductID[r.ID] = r.PrimaryURL
			}
			log.Printf("[admin.order-details] Fetched images for %d products", len(rows))
		} else {
			log.Printf("[admin.order-details] WARN failed to fetch product images: %v", err)
		}
	}

	// =====================================
	// 6. Attach images
	// =====================================
	for i := range items {
		if url, ok := imageByProductID[items[i].ProductID]; ok {
			items[i].ProductImage = url
		}
	}

	res.Items = items

	log.Printf("[admin.order-details] Responding with order %s", res.OrderNumber)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Order details retrieved successfully",
		res,
	))
}
