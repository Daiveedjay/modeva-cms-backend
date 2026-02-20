// package order_controller

// import (
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	"gorm.io/gorm"
// )

// // DownloadOrderInvoicePDF godoc
// // @Summary Download order invoice PDF
// // @Description Generate and download an invoice PDF for the order
// // @Tags Orders
// // @Produce octet-stream
// // @Security BearerAuth
// // @Param orderId path string true "Order ID"
// // @Success 200 "PDF file"
// // @Failure 400 {object} models.ApiResponse "Invalid order ID"
// // @Failure 404 {object} models.ApiResponse "Order not found"
// // @Failure 500 {object} models.ApiResponse "Server error"
// // @Router /orders/:id/download-invoice [get]
// func DownloadOrderInvoicePDF(c *gin.Context) {
// 	orderId := c.Param("id")
// 	log.Printf("[order.download-invoice] request for order: %s", orderId)

// 	// Validate order ID
// 	if _, err := uuid.Parse(orderId); err != nil {
// 		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid order ID"))
// 		return
// 	}

// 	ctx, cancel := config.WithTimeout()
// 	defer cancel()

// 	// Get the order (from ecommerce database)
// 	var order models.Order
// 	if err := config.EcommerceGorm.WithContext(ctx).
// 		Where("id = ?", orderId).
// 		First(&order).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			log.Printf("[order.download-invoice] order not found: %s", orderId)
// 			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Order not found"))
// 			return
// 		}
// 		log.Printf("[order.download-invoice] database error: %v", err)
// 		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
// 		return
// 	}

// 	// Get order items (from ecommerce database)
// 	var orderItems []models.OrderItem
// 	if err := config.EcommerceGorm.WithContext(ctx).
// 		Where("order_id = ?", orderId).
// 		Find(&orderItems).Error; err != nil {
// 		log.Printf("[order.download-invoice] failed to fetch order items: %v", err)
// 		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
// 		return
// 	}

// 	// Get customer details (from ecommerce database)
// 	var customer struct {
// 		Email string
// 		Name  string
// 	}
// 	if err := config.EcommerceGorm.WithContext(ctx).
// 		Table("users").
// 		Select("email, name").
// 		Where("id = ?", order.UserID).
// 		Scan(&customer).Error; err != nil {
// 		log.Printf("[order.download-invoice] failed to fetch customer: %v", err)
// 		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
// 		return
// 	}

// 	// Generate PDF in memory
// 	pdfBuffer := generateOrderInvoicePDF(&order, orderItems, customer.Name, customer.Email)

// 	// Set response headers for file download
// 	filename := fmt.Sprintf("invoice-%s.pdf", order.OrderNumber)
// 	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
// 	c.Header("Content-Type", "application/pdf")
// 	c.Header("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

// 	// Write PDF to response
// 	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())

// 	log.Printf("[order.download-invoice] invoice PDF downloaded for order %s", orderId)
// }

package order_controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DownloadOrderInvoicePDF godoc
// @Summary Download order invoice PDF
// @Description Generate and download an invoice PDF for the order
// @Tags Orders
// @Produce octet-stream
// @Security BearerAuth
// @Param orderId path string true "Order ID"
// @Success 200 "PDF file"
// @Failure 400 {object} models.ApiResponse "Invalid order ID"
// @Failure 404 {object} models.ApiResponse "Order not found"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /orders/:id/download-invoice [get]
func DownloadOrderInvoicePDF(c *gin.Context) {
	orderId := c.Param("id")
	log.Printf("[order.download-invoice] request for order: %s", orderId)

	// Validate order ID
	if _, err := uuid.Parse(orderId); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid order ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Get the order (from ecommerce database)
	var order models.Order
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("id = ?", orderId).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[order.download-invoice] order not found: %s", orderId)
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Order not found"))
			return
		}
		log.Printf("[order.download-invoice] database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Get order items (from ecommerce database)
	var orderItems []models.OrderItem
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("order_id = ?", orderId).
		Find(&orderItems).Error; err != nil {
		log.Printf("[order.download-invoice] failed to fetch order items: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Get customer details (from ecommerce database)
	var customer struct {
		Email string
		Name  string
	}
	if err := config.EcommerceGorm.WithContext(ctx).
		Table("users").
		Select("email, name").
		Where("id = ?", order.UserID).
		Scan(&customer).Error; err != nil {
		log.Printf("[order.download-invoice] failed to fetch customer: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Generate PDF in memory
	pdfBuffer := generateOrderInvoicePDF(&order, orderItems, customer.Name, customer.Email)

	// Set response headers for file download
	filename := fmt.Sprintf("invoice-%s.pdf", order.OrderNumber)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// CORS headers for PDF download
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Write PDF to response
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())

	log.Printf("[order.download-invoice] invoice PDF downloaded for order %s", orderId)
}
