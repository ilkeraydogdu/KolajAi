package models

import "time"

// WholesaleCustomer represents wholesale customers
type WholesaleCustomer struct {
	ID           int        `json:"id" db:"id"`
	UserID       int        `json:"user_id" db:"user_id"`
	CompanyName  string     `json:"company_name" db:"company_name"`
	TaxID        string     `json:"tax_id" db:"tax_id"`
	BusinessType string     `json:"business_type" db:"business_type"`
	YearlyVolume float64    `json:"yearly_volume" db:"yearly_volume"`
	CreditLimit  float64    `json:"credit_limit" db:"credit_limit"`
	PaymentTerms int        `json:"payment_terms" db:"payment_terms"` // days
	DiscountTier string     `json:"discount_tier" db:"discount_tier"` // bronze, silver, gold, platinum
	Status       string     `json:"status" db:"status"`               // pending, approved, suspended, rejected
	ApprovedBy   *int       `json:"approved_by" db:"approved_by"`
	ApprovedAt   *time.Time `json:"approved_at" db:"approved_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// WholesalePrice represents wholesale pricing tiers
type WholesalePrice struct {
	ID        int       `json:"id" db:"id"`
	ProductID int       `json:"product_id" db:"product_id"`
	MinQty    int       `json:"min_qty" db:"min_qty"`
	MaxQty    *int      `json:"max_qty" db:"max_qty"`
	Price     float64   `json:"price" db:"price"`
	Discount  float64   `json:"discount" db:"discount"`
	Tier      string    `json:"tier" db:"tier"` // bronze, silver, gold, platinum
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// WholesaleOrder represents wholesale orders
type WholesaleOrder struct {
	ID             int       `json:"id" db:"id"`
	CustomerID     int       `json:"customer_id" db:"customer_id"`
	OrderNumber    string    `json:"order_number" db:"order_number"`
	Status         string    `json:"status" db:"status"`                 // draft, pending, confirmed, processing, shipped, delivered, cancelled
	PaymentStatus  string    `json:"payment_status" db:"payment_status"` // pending, paid, partial, overdue
	PaymentTerms   int       `json:"payment_terms" db:"payment_terms"`
	DueDate        time.Time `json:"due_date" db:"due_date"`
	SubTotal       float64   `json:"sub_total" db:"sub_total"`
	DiscountAmount float64   `json:"discount_amount" db:"discount_amount"`
	TaxAmount      float64   `json:"tax_amount" db:"tax_amount"`
	ShippingCost   float64   `json:"shipping_cost" db:"shipping_cost"`
	TotalAmount    float64   `json:"total_amount" db:"total_amount"`
	Currency       string    `json:"currency" db:"currency"`
	Notes          string    `json:"notes" db:"notes"`
	InternalNotes  string    `json:"internal_notes" db:"internal_notes"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// WholesaleOrderItem represents items in wholesale orders
type WholesaleOrderItem struct {
	ID           int     `json:"id" db:"id"`
	OrderID      int     `json:"order_id" db:"order_id"`
	ProductID    int     `json:"product_id" db:"product_id"`
	ProductName  string  `json:"product_name" db:"product_name"`
	ProductSKU   string  `json:"product_sku" db:"product_sku"`
	Quantity     int     `json:"quantity" db:"quantity"`
	UnitPrice    float64 `json:"unit_price" db:"unit_price"`
	DiscountRate float64 `json:"discount_rate" db:"discount_rate"`
	TotalPrice   float64 `json:"total_price" db:"total_price"`
	Status       string  `json:"status" db:"status"` // pending, confirmed, shipped, delivered
}

// WholesaleQuote represents wholesale price quotes
type WholesaleQuote struct {
	ID             int       `json:"id" db:"id"`
	CustomerID     int       `json:"customer_id" db:"customer_id"`
	VendorID       int       `json:"vendor_id" db:"vendor_id"`
	QuoteNumber    string    `json:"quote_number" db:"quote_number"`
	Status         string    `json:"status" db:"status"` // draft, sent, accepted, rejected, expired
	ValidUntil     time.Time `json:"valid_until" db:"valid_until"`
	SubTotal       float64   `json:"sub_total" db:"sub_total"`
	DiscountAmount float64   `json:"discount_amount" db:"discount_amount"`
	TotalAmount    float64   `json:"total_amount" db:"total_amount"`
	Notes          string    `json:"notes" db:"notes"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// WholesaleQuoteItem represents items in wholesale quotes
type WholesaleQuoteItem struct {
	ID           int     `json:"id" db:"id"`
	QuoteID      int     `json:"quote_id" db:"quote_id"`
	ProductID    int     `json:"product_id" db:"product_id"`
	ProductName  string  `json:"product_name" db:"product_name"`
	ProductSKU   string  `json:"product_sku" db:"product_sku"`
	Quantity     int     `json:"quantity" db:"quantity"`
	UnitPrice    float64 `json:"unit_price" db:"unit_price"`
	DiscountRate float64 `json:"discount_rate" db:"discount_rate"`
	TotalPrice   float64 `json:"total_price" db:"total_price"`
}
