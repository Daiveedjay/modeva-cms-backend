package models

import "time"

type CMSCustomerListRow struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Avatar          *string    `json:"avatar"`
	Location        string     `json:"location"`
	Orders          int        `json:"orders"`
	TotalSpent      float64    `json:"total_spent"`
	Status          string     `json:"status"`
	Activity        string     `json:"activity"` // "active" or "inactive"
	JoinDate        time.Time  `json:"join_date"`
	BanReason       *string    `json:"ban_reason"`
	SuspendedUntil  *time.Time `json:"suspended_until"`
	SuspendedReason *string    `json:"suspended_reason"`
}

type CMSCustomerDetail struct {
	ID               string           `json:"id" gorm:"column:id"`
	Name             string           `json:"name" gorm:"column:name"`
	Email            string           `json:"email" gorm:"column:email"`
	Phone            *string          `json:"phone" gorm:"column:phone"`
	Avatar           *string          `json:"avatar" gorm:"column:avatar"`
	Location         string           `json:"location" gorm:"-"`
	Status           string           `json:"status" gorm:"column:status"`
	Orders           int              `json:"orders" gorm:"column:orders"`
	TotalSpent       float64          `json:"total_spent" gorm:"column:total_spent"`
	AvgOrderValue    float64          `json:"avg_order_value" gorm:"column:avg_order_value"`
	JoinDate         time.Time        `json:"join_date" gorm:"column:join_date"`
	LastOrderDate    *time.Time       `json:"last_order_date" gorm:"column:last_order_date"`
	FavoriteCategory *string          `json:"favorite_category" gorm:"-"`
	BanReason        *string          `json:"ban_reason" gorm:"column:ban_reason"`
	SuspendedUntil   *time.Time       `json:"suspended_until" gorm:"column:suspended_until"`
	SuspendedReason  *string          `json:"suspended_reason" gorm:"column:suspended_reason"`
	Address          *CustomerAddress `json:"address" gorm:"-"`
	RecentOrders     []CustomerOrder  `json:"recent_orders" gorm:"-"`
}

type CustomerOrder struct {
	OrderNumber string    `json:"order_number"`
	ID          string    `json:"id"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}

type CustomerAddress struct {
	ID      string  `json:"id"`
	Phone   *string `json:"phone"`
	Street  *string `json:"street"`
	City    *string `json:"city"`
	State   *string `json:"state"`
	Zip     *string `json:"zip"`
	Country *string `json:"country"`
}

// UpdateCustomerRequest is used when admin updates customer info
type UpdateCustomerRequest struct {
	// Basic info
	Name  *string `json:"name" binding:"omitempty,min=2,max=255"`
	Phone *string `json:"phone" binding:"omitempty,min=10,max=20"`

	// Status management
	Status          *string    `json:"status" binding:"omitempty,oneof=active suspended banned deleted"`
	BanReason       *string    `json:"ban_reason"`
	SuspendedUntil  *time.Time `json:"suspended_until"`
	SuspendedReason *string    `json:"suspended_reason"`
}

type CMSCustomerOrderRow struct {
	ID           string    `json:"id"`
	OrderNumber  string    `json:"order_number"`
	Status       string    `json:"status"`
	TotalAmount  float64   `json:"total_amount"`
	CreatedAt    time.Time `json:"created_at"`
	CustomerName string    `json:"customer_name"` // Added
}

// CustomerStats represents customer dashboard statistics
type CustomerStats struct {
	TotalCustomers               int     `json:"total_customers"`                 // Total active customers
	NewCustomersThisMonth        int     `json:"new_customers_this_month"`        // New customers added this month
	NewCustomersGrowthPercentage float64 `json:"new_customers_growth_percentage"` // % growth from last month
	ActiveCustomers              int     `json:"active_customers"`                // Customers with order in last 90 days
	ActiveCustomersPercentage    float64 `json:"active_customers_percentage"`     // % of total customers
	AvgOrderValue                float64 `json:"avg_order_value"`                 // Average order value in cents
}
