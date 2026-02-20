package models

// AnalyticsOverview represents the main analytics dashboard overview
type AnalyticsOverview struct {
	TotalRevenue                 float64 `json:"total_revenue"`                   // Total revenue this month
	RevenueGrowthPercent         float64 `json:"revenue_growth_percent"`          // % change from last month
	TotalOrders                  int     `json:"total_orders"`                    // Number of orders this month
	OrdersGrowthPercent          float64 `json:"orders_growth_percent"`           // % change from last month
	TotalInventory               int     `json:"total_inventory"`                 // Total items in stock
	InventoryGrowthPercent       float64 `json:"inventory_growth_percent"`        // % change from last month (approximation)
	ActiveCustomers              int     `json:"active_customers"`                // Customers with order in last 90 days
	ActiveCustomersGrowthPercent float64 `json:"active_customers_growth_percent"` // % change (60-90 vs 90+ days ago)
}

// TopProduct represents a top performing product with sales and revenue metrics
type TopProduct struct {
	ProductID      string  `json:"product_id"`      // UUID of the product
	ProductName    string  `json:"product_name"`    // Name of the product
	OrderCount     int     `json:"order_count"`     // Number of distinct orders containing this product
	SalesCount     int     `json:"sales_count"`     // Total quantity sold
	Revenue        float64 `json:"revenue"`         // Total revenue from this product
	RevenuePercent float64 `json:"revenue_percent"` // Percentage of total revenue this month
}

type MonthlyRevenueData struct {
	Month       string  `json:"month"`        // Month abbreviation (Jan, Feb, etc.)
	MonthNumber int     `json:"month_number"` // Month number (1-12)
	Revenue     float64 `json:"revenue"`      // Total revenue for the month
}

type SalesMetrics struct {
	AverageOrderValue     float64 `json:"average_order_value"`     // Average amount per completed order
	CustomerLifetimeValue float64 `json:"customer_lifetime_value"` // Average lifetime spending per customer
	ReturnCustomerRate    float64 `json:"return_customer_rate"`    // Percentage of customers with 2+ orders
}

type GeographicData struct {
	Country    string  `json:"country"`     // Country name
	OrderCount int     `json:"order_count"` // Number of orders from this country
	Percentage float64 `json:"percentage"`  // Percentage of total orders
}

type DeviceAnalytics struct {
	DeviceType string  `json:"device_type"` // Type of device (desktop, mobile, tablet)
	OrderCount int     `json:"order_count"` // Number of orders from this device type
	Percentage float64 `json:"percentage"`  // Percentage of total orders
}
