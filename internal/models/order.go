package models

import "time"

// Order represents an order in the system
type Order struct {
	ID             int        `json:"id" db:"id"`
	UserID         int        `json:"user_id" db:"user_id"`
	OrderNumber    string     `json:"order_number" db:"order_number"`
	Status         string     `json:"status" db:"status"`                 // pending, confirmed, processing, shipped, delivered, cancelled, refunded
	PaymentStatus  string     `json:"payment_status" db:"payment_status"` // pending, paid, failed, refunded
	PaymentMethod  string     `json:"payment_method" db:"payment_method"`
	ShippingMethod string     `json:"shipping_method" db:"shipping_method"`
	SubTotal       float64    `json:"sub_total" db:"sub_total"`
	TaxAmount      float64    `json:"tax_amount" db:"tax_amount"`
	ShippingCost   float64    `json:"shipping_cost" db:"shipping_cost"`
	DiscountAmount float64    `json:"discount_amount" db:"discount_amount"`
	TotalAmount    float64    `json:"total_amount" db:"total_amount"`
	Currency       string     `json:"currency" db:"currency"`
	Notes          string     `json:"notes" db:"notes"`
	TrackingNumber string     `json:"tracking_number" db:"tracking_number"`
	ShippedAt      *time.Time `json:"shipped_at" db:"shipped_at"`
	DeliveredAt    *time.Time `json:"delivered_at" db:"delivered_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// OrderItem represents items in an order
type OrderItem struct {
	ID          int     `json:"id" db:"id"`
	OrderID     int     `json:"order_id" db:"order_id"`
	ProductID   int     `json:"product_id" db:"product_id"`
	VendorID    int     `json:"vendor_id" db:"vendor_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	ProductSKU  string  `json:"product_sku" db:"product_sku"`
	VariantInfo string  `json:"variant_info" db:"variant_info"`
	Quantity    int     `json:"quantity" db:"quantity"`
	UnitPrice   float64 `json:"unit_price" db:"unit_price"`
	TotalPrice  float64 `json:"total_price" db:"total_price"`
	Commission  float64 `json:"commission" db:"commission"`
	Status      string  `json:"status" db:"status"` // pending, confirmed, shipped, delivered, cancelled
}

// OrderAddress represents shipping and billing addresses
type OrderAddress struct {
	ID         int    `json:"id" db:"id"`
	OrderID    int    `json:"order_id" db:"order_id"`
	Type       string `json:"type" db:"type"` // shipping, billing
	FirstName  string `json:"first_name" db:"first_name"`
	LastName   string `json:"last_name" db:"last_name"`
	Company    string `json:"company" db:"company"`
	Address1   string `json:"address1" db:"address1"`
	Address2   string `json:"address2" db:"address2"`
	City       string `json:"city" db:"city"`
	State      string `json:"state" db:"state"`
	PostalCode string `json:"postal_code" db:"postal_code"`
	Country    string `json:"country" db:"country"`
	Phone      string `json:"phone" db:"phone"`
}

// Cart represents shopping cart
type Cart struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	SessionID string    `json:"session_id" db:"session_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CartItem represents items in shopping cart
type CartItem struct {
	ID        int       `json:"id" db:"id"`
	CartID    int       `json:"cart_id" db:"cart_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	VariantID *int      `json:"variant_id" db:"variant_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Wishlist represents user wishlist
type Wishlist struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
