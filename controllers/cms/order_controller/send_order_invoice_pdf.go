package order_controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SendOrderInvoicePDF godoc
// @Summary Send order invoice PDF to customer
// @Description Generate and send an invoice PDF to the customer
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param orderId path string true "Order ID"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse "Invalid order ID"
// @Failure 404 {object} models.ApiResponse "Order not found"
// @Failure 500 {object} models.ApiResponse "Server error"
// @Router /orders/:id/send-invoice [post]
func SendOrderInvoicePDF(c *gin.Context) {
	orderId := c.Param("id")
	log.Printf("[order.send-invoice] request for order: %s", orderId)

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
			log.Printf("[order.send-invoice] order not found: %s", orderId)
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Order not found"))
			return
		}
		log.Printf("[order.send-invoice] database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Get order items (from ecommerce database)
	var orderItems []models.OrderItem
	if err := config.EcommerceGorm.WithContext(ctx).
		Where("order_id = ?", orderId).
		Find(&orderItems).Error; err != nil {
		log.Printf("[order.send-invoice] failed to fetch order items: %v", err)
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
		log.Printf("[order.send-invoice] failed to fetch customer: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Server error"))
		return
	}

	// Validate customer email
	if customer.Email == "" {
		log.Printf("[order.send-invoice] customer email missing for order: %s", orderId)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Customer email not found"))
		return
	}

	// Get address details (from ecommerce database)
	var addressDetails struct {
		Street string
		City   string
		State  string
		Zip    string
	}
	if order.AddressID != nil {
		if err := config.EcommerceGorm.WithContext(ctx).
			Table("addresses").
			Select("street, city, state, zip").
			Where("id = ?", *order.AddressID).
			Scan(&addressDetails).Error; err != nil {
			log.Printf("[order.send-invoice] failed to fetch address details: %v", err)
			// Address is optional, continue without it
			addressDetails.Street = ""
			addressDetails.City = ""
			addressDetails.State = ""
			addressDetails.Zip = ""
		}
	}

	// Get admin info for logging
	adminIDStr, _ := c.Get("adminID")
	adminEmail, _ := c.Get("adminEmail")

	// Generate PDF in memory
	pdfBuffer := generateOrderInvoicePDF(&order, orderItems, customer.Name, customer.Email)

	// Convert order items to service format
	serviceItems := make([]services.OrderInvoiceItem, len(orderItems))
	for i, item := range orderItems {
		serviceItems[i] = services.OrderInvoiceItem{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Price * float64(item.Quantity),
		}
	}

	// Send invoice email asynchronously with all data
	go func() {
		resendClient := services.NewResendClient()

		emailData := services.OrderInvoicePDFEmailData{
			CustomerName:  customer.Name,
			CustomerEmail: customer.Email,
			OrderNumber:   order.OrderNumber,
			OrderDate:     order.CreatedAt.Format("Jan 02, 2006"),
			DueDate:       order.CreatedAt.AddDate(0, 0, 14).Format("Jan 02, 2006"),
			AddressStreet: addressDetails.Street,
			AddressCity:   addressDetails.City,
			AddressState:  addressDetails.State,
			AddressZip:    addressDetails.Zip,
			Items:         serviceItems,
			SubtotalTotal: order.Subtotal,
			ShippingCost:  order.ShippingCost,
			Tax:           order.Tax,
			Discount:      order.Discount,
			TotalAmount:   order.TotalAmount,
			PDFContent:    pdfBuffer.Bytes(),
		}

		if err := resendClient.SendOrderInvoicePDFEmail(emailData); err != nil {
			log.Printf("[order.send-invoice] failed to send email for order %s: %v", orderId, err)
		} else {
			log.Printf("[order.send-invoice] invoice email sent to %s for order %s", customer.Email, orderId)
		}
	}()

	// ✅ LOG THE ACTIVITY
	changes := map[string]interface{}{
		"order_id":       order.ID,
		"order_number":   order.OrderNumber,
		"customer_email": customer.Email,
		"sent_to":        customer.Email,
	}
	changesJSON, _ := json.Marshal(changes)

	adminID, _ := uuid.Parse(adminIDStr.(string))
	activityLog := models.ActivityLog{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:      adminID,
		AdminEmail:   adminEmail.(string),
		Action:       "sent_order_invoice",
		ResourceType: "order",
		ResourceID:   order.ID,
		ResourceName: order.OrderNumber,
		Changes:      datatypes.JSON(changesJSON),
		Status:       "success",
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	if err := config.CmsGorm.WithContext(ctx).Create(&activityLog).Error; err != nil {
		log.Printf("[order.send-invoice] failed to log activity: %v", err)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Invoice email sent to customer", map[string]interface{}{
		"order_id":       order.ID,
		"customer_email": customer.Email,
	}))
}

// generateOrderInvoicePDF creates a professional invoice PDF matching your HTML design
func generateOrderInvoicePDF(order *models.Order, items []models.OrderItem, customerName, customerEmail string) *bytes.Buffer {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(20, 20, 20)

	// Colors
	darkGray := color.Color{Red: 38, Green: 38, Blue: 34}
	mediumGray := color.Color{Red: 121, Green: 119, Blue: 109}

	// Invoice Title
	m.Row(15, func() {
		m.Col(12, func() {
			m.Text("INVOICE", props.Text{
				Size:  24,
				Style: consts.Bold,
				Color: darkGray,
			})
		})
	})

	// Company Info
	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("MODEVA STORE", props.Text{
				Size:  16,
				Style: consts.Bold,
				Color: darkGray,
			})
		})
	})

	m.Row(5, func() {
		m.Col(12, func() {
			m.Text("contact@modeva.com", props.Text{
				Size:  9,
				Color: mediumGray,
			})
		})
	})

	m.Row(8, func() {})

	// Billing Section
	m.Row(5, func() {
		m.Col(6, func() {
			m.Text("BILL TO", props.Text{
				Size:  8,
				Style: consts.Bold,
				Color: darkGray,
			})
		})
		m.Col(6, func() {
			m.Text("INVOICE DETAILS", props.Text{
				Size:  8,
				Style: consts.Bold,
				Color: darkGray,
				Align: consts.Right,
			})
		})
	})

	m.Row(5, func() {
		m.Col(6, func() {
			m.Text(customerName, props.Text{
				Size:  10,
				Style: consts.Bold,
				Color: darkGray,
			})
		})
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Invoice #%s", order.OrderNumber), props.Text{
				Size:  10,
				Color: darkGray,
				Align: consts.Right,
			})
		})
	})

	m.Row(5, func() {
		m.Col(6, func() {
			m.Text(customerEmail, props.Text{
				Size:  9,
				Color: mediumGray,
			})
		})
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Date: %s", order.CreatedAt.Format("Jan 02, 2006")), props.Text{
				Size:  9,
				Color: mediumGray,
				Align: consts.Right,
			})
		})
	})

	m.Row(8, func() {})

	// Items Table Header
	m.Row(6, func() {
		m.Col(6, func() {
			m.Text("Description", props.Text{
				Size:  8,
				Style: consts.Bold,
				Color: darkGray,
			})
		})
		m.Col(2, func() {
			m.Text("Qty", props.Text{
				Size:  8,
				Style: consts.Bold,
				Color: darkGray,
				Align: consts.Right,
			})
		})
		m.Col(2, func() {
			m.Text("Price", props.Text{
				Size:  8,
				Style: consts.Bold,
				Color: darkGray,
				Align: consts.Right,
			})
		})
		m.Col(2, func() {
			m.Text("Total", props.Text{
				Size:  8,
				Style: consts.Bold,
				Color: darkGray,
				Align: consts.Right,
			})
		})
	})

	// Items
	for _, item := range items {
		itemTotal := item.Price * float64(item.Quantity)
		m.Row(6, func() {
			m.Col(6, func() {
				m.Text(item.ProductName, props.Text{
					Size:  9,
					Color: darkGray,
				})
			})
			m.Col(2, func() {
				m.Text(fmt.Sprintf("%d", item.Quantity), props.Text{
					Size:  9,
					Color: darkGray,
					Align: consts.Right,
				})
			})
			m.Col(2, func() {
				m.Text(fmt.Sprintf("$%.2f", item.Price), props.Text{
					Size:  9,
					Color: darkGray,
					Align: consts.Right,
				})
			})
			m.Col(2, func() {
				m.Text(fmt.Sprintf("$%.2f", itemTotal), props.Text{
					Size:  9,
					Color: darkGray,
					Align: consts.Right,
				})
			})
		})
	}

	m.Row(8, func() {})

	// Summary Section
	m.Row(5, func() {
		m.Col(8, func() {})
		m.Col(2, func() {
			m.Text("Subtotal", props.Text{
				Size:  9,
				Color: mediumGray,
				Align: consts.Right,
			})
		})
		m.Col(2, func() {
			m.Text(fmt.Sprintf("$%.2f", order.Subtotal), props.Text{
				Size:  9,
				Color: darkGray,
				Align: consts.Right,
			})
		})
	})

	m.Row(5, func() {
		m.Col(8, func() {})
		m.Col(2, func() {
			m.Text("Shipping", props.Text{
				Size:  9,
				Color: mediumGray,
				Align: consts.Right,
			})
		})
		m.Col(2, func() {
			m.Text(fmt.Sprintf("$%.2f", order.ShippingCost), props.Text{
				Size:  9,
				Color: darkGray,
				Align: consts.Right,
			})
		})
	})

	m.Row(5, func() {
		m.Col(8, func() {})
		m.Col(2, func() {
			m.Text("Tax", props.Text{
				Size:  9,
				Color: mediumGray,
				Align: consts.Right,
			})
		})
		m.Col(2, func() {
			m.Text(fmt.Sprintf("$%.2f", order.Tax), props.Text{
				Size:  9,
				Color: darkGray,
				Align: consts.Right,
			})
		})
	})

	if order.Discount > 0 {
		m.Row(5, func() {
			m.Col(8, func() {})
			m.Col(2, func() {
				m.Text("Discount", props.Text{
					Size:  9,
					Color: mediumGray,
					Align: consts.Right,
				})
			})
			m.Col(2, func() {
				m.Text(fmt.Sprintf("-$%.2f", order.Discount), props.Text{
					Size:  9,
					Color: darkGray,
					Align: consts.Right,
				})
			})
		})
	}

	// Total
	m.Row(8, func() {
		m.Col(8, func() {})
		m.Col(2, func() {
			m.Text("Total", props.Text{
				Size:  12,
				Style: consts.Bold,
				Color: darkGray,
				Align: consts.Right,
			})
		})
		m.Col(2, func() {
			m.Text(fmt.Sprintf("$%.2f", order.TotalAmount), props.Text{
				Size:  12,
				Style: consts.Bold,
				Color: darkGray,
				Align: consts.Right,
			})
		})
	})

	m.Row(12, func() {})

	// Footer
	m.Row(5, func() {
		m.Col(12, func() {
			m.Text("Thank you for your business!", props.Text{
				Size:  8,
				Style: consts.Bold,
				Color: darkGray,
			})
		})
	})

	m.Row(5, func() {
		m.Col(12, func() {
			m.Text("© 2026 Modeva Store. All rights reserved.", props.Text{
				Size:  8,
				Color: mediumGray,
			})
		})
	})

	// Output to buffer
	buf, err := m.Output()
	if err != nil {
		log.Printf("[order.send-invoice] failed to generate PDF: %v", err)
		return bytes.NewBuffer(nil)
	}

	return &buf
}
