package models

import "time"

// Order represents a complete customer order
type Order struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	OrderNumber        string     `json:"order_number"`
	PaymentMethodID    *string    `json:"payment_method_id,omitempty"`
	AddressID          *string    `json:"address_id,omitempty"`
	PaymentMethodType  *string    `json:"payment_method_type,omitempty"`
	PaymentMethodLast4 *string    `json:"payment_method_last4,omitempty"`
	AddressSnapshot    *string    `json:"address_snapshot,omitempty"` // JSONB as string
	Subtotal           float64    `json:"subtotal"`
	Tax                float64    `json:"tax"`
	ShippingCost       float64    `json:"shipping_cost"`
	Discount           float64    `json:"discount"`
	TotalAmount        float64    `json:"total_amount"`
	Status             string     `json:"status"`
	CustomerNotes      *string    `json:"customer_notes,omitempty"`
	AdminNotes         *string    `json:"admin_notes,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	ConfirmedAt        *time.Time `json:"confirmed_at,omitempty"`
	ShippedAt          *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt        *time.Time `json:"delivered_at,omitempty"`
}

// OrderItem represents an individual product in an order
type OrderItem struct {
	ID           string    `json:"id"`
	OrderID      string    `json:"order_id"`
	UserID       string    `json:"user_id"`
	ProductID    string    `json:"product_id"`
	ProductName  string    `json:"product_name"`
	VariantSize  *string   `json:"variant_size,omitempty"`
	VariantColor *string   `json:"variant_color,omitempty"`
	Price        float64   `json:"price"`
	Quantity     int       `json:"quantity"`
	Subtotal     float64   `json:"subtotal"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OrderWithItems combines order and its items
type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

type OrderItemWithImage struct {
	OrderItem
	ProductImage *string `json:"product_image,omitempty"`
}

// OrderHistoryResponse for list view
type OrderHistoryResponse struct {
	ID          string    `json:"id"`
	OrderNumber string    `json:"order_number"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"total_amount"`
	ItemCount   int       `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateOrderRequest for checkout
type CreateOrderRequest struct {
	PaymentMethodID string           `json:"payment_method_id" binding:"required"`
	AddressID       string           `json:"address_id" binding:"required"`
	Items           []OrderItemInput `json:"items" binding:"required,min=1"`
	CustomerNotes   *string          `json:"customer_notes,omitempty"`
}

// OrderItemInput for cart items
type OrderItemInput struct {
	ProductID    string  `json:"product_id" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required,min=1"`
	VariantSize  *string `json:"variant_size,omitempty"`
	VariantColor *string `json:"variant_color,omitempty"`
}

type CMSOrderListRow struct {
	ID            string    `json:"id"`            // orders.id
	OrderNumber   string    `json:"order_number"`  // ORD-2025-000001
	CustomerID    string    `json:"customer_id"`   // users.id
	CustomerName  string    `json:"customer_name"` // username or fallback
	CustomerEmail string    `json:"customer_email"`
	CreatedAt     time.Time `json:"created_at"`
	ItemCount     int       `json:"item_count"`     // COUNT(order_items.id)
	TotalQuantity int       `json:"total_quantity"` // SUM(order_items.quantity)
	TotalAmount   float64   `json:"total_amount"`
	Status        string    `json:"status"`
}

type CMSOrderAddress struct {
	Label     *string `json:"label,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Street    *string `json:"street,omitempty"`
	City      *string `json:"city,omitempty"`
	State     *string `json:"state,omitempty"`
	Zip       *string `json:"zip,omitempty"`
	Country   *string `json:"country,omitempty"`
}

type CMSOrderDetailsResponse struct {
	ID          string    `json:"id"`
	OrderNumber string    `json:"order_number"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`

	CustomerID    string `json:"customer_id"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`

	PaymentMethodType  *string `json:"payment_method_type,omitempty"`
	PaymentMethodLast4 *string `json:"payment_method_last4,omitempty"`
	PaymentMethodLabel string  `json:"payment_method_label"`

	Subtotal     float64 `json:"subtotal"`
	ShippingCost float64 `json:"shipping_cost"`
	Tax          float64 `json:"tax"`
	Discount     float64 `json:"discount"`
	TotalAmount  float64 `json:"total_amount"`

	CustomerNotes   *string `json:"customer_notes,omitempty"`
	AdminNotes      *string `json:"admin_notes,omitempty"`
	AddressSnapshot *string `json:"address_snapshot,omitempty"`

	CMSOrderAddress `gorm:"embedded" json:"address"`

	Items []OrderItemWithImage `gorm:"-" json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status     string  `json:"status" binding:"required,oneof=pending processing shipped completed cancelled"`
	AdminNotes *string `json:"admin_notes,omitempty"` // required if status=cancelled
}

type UpdateOrderStatusResponse struct {
	ID          string  `json:"id"`
	OrderNumber string  `json:"order_number"`
	Status      string  `json:"status"`
	AdminNotes  *string `json:"admin_notes,omitempty"`
}

type OrderStatsBreakdown struct {
	Count       int    `json:"count"`
	Description string `json:"description"`
}

type OrderStatsResponse struct {
	TotalOrders                int                 `json:"total_orders"`
	ChangePercentFromLastMonth *float64            `json:"change_percent_from_last_month,omitempty"`
	CurrentMonthTotal          int                 `json:"current_month_total"`
	LastMonthTotal             int                 `json:"last_month_total"`
	Pending                    OrderStatsBreakdown `json:"pending"`
	Processing                 OrderStatsBreakdown `json:"processing"`
	Shipped                    OrderStatsBreakdown `json:"shipped"`
	Completed                  OrderStatsBreakdown `json:"completed"`
	Cancelled                  OrderStatsBreakdown `json:"cancelled"`
}

type AdminOrderSearchQuery struct {
	Q           string   `form:"q"`            // generic search term
	OrderNumber string   `form:"order_number"` // explicit
	Customer    string   `form:"customer"`     // name
	Email       string   `form:"email"`        // email
	Status      string   `form:"status"`       // exact
	Price       *float64 `form:"price"`        // exact total_amount
	MinPrice    *float64 `form:"min_price"`    // range
	MaxPrice    *float64 `form:"max_price"`    // range
	CreatedFrom *string  `form:"created_from"` // ISO8601 date or datetime
	CreatedTo   *string  `form:"created_to"`   // ISO8601 date or datetime
	Page        int      `form:"page"`
	Limit       int      `form:"limit"`
}

type AdminOrderSearchResponse struct {
	Orders []CMSOrderListRow `json:"orders"`
}
