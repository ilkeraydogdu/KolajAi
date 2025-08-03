package models

import (
	"errors"
	"strings"
	"time"
)

// Order represents an order in the system
type Order struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	VendorID        int64     `json:"vendor_id" db:"vendor_id"`
	OrderNumber     string    `json:"order_number" db:"order_number"`
	Status          string    `json:"status" db:"status"` // pending, confirmed, processing, shipped, delivered, cancelled, refunded
	PaymentStatus   string    `json:"payment_status" db:"payment_status"` // pending, paid, failed, refunded, partial
	PaymentMethod   string    `json:"payment_method" db:"payment_method"`
	SubtotalAmount  float64   `json:"subtotal_amount" db:"subtotal_amount"`
	TaxAmount       float64   `json:"tax_amount" db:"tax_amount"`
	ShippingAmount  float64   `json:"shipping_amount" db:"shipping_amount"`
	DiscountAmount  float64   `json:"discount_amount" db:"discount_amount"`
	TotalAmount     float64   `json:"total_amount" db:"total_amount"`
	Currency        string    `json:"currency" db:"currency"`
	
	// Shipping Information
	ShippingAddress string    `json:"shipping_address" db:"shipping_address"`
	ShippingCity    string    `json:"shipping_city" db:"shipping_city"`
	ShippingState   string    `json:"shipping_state" db:"shipping_state"`
	ShippingZip     string    `json:"shipping_zip" db:"shipping_zip"`
	ShippingCountry string    `json:"shipping_country" db:"shipping_country"`
	ShippingPhone   string    `json:"shipping_phone" db:"shipping_phone"`
	
	// Billing Information
	BillingAddress  string    `json:"billing_address" db:"billing_address"`
	BillingCity     string    `json:"billing_city" db:"billing_city"`
	BillingState    string    `json:"billing_state" db:"billing_state"`
	BillingZip      string    `json:"billing_zip" db:"billing_zip"`
	BillingCountry  string    `json:"billing_country" db:"billing_country"`
	BillingPhone    string    `json:"billing_phone" db:"billing_phone"`
	
	// Tracking Information
	TrackingNumber  string    `json:"tracking_number" db:"tracking_number"`
	CarrierName     string    `json:"carrier_name" db:"carrier_name"`
	ShippedAt       *time.Time `json:"shipped_at" db:"shipped_at"`
	DeliveredAt     *time.Time `json:"delivered_at" db:"delivered_at"`
	
	// Additional Information
	Notes           string    `json:"notes" db:"notes"`
	InternalNotes   string    `json:"internal_notes" db:"internal_notes"`
	CouponCode      string    `json:"coupon_code" db:"coupon_code"`
	ReferenceID     string    `json:"reference_id" db:"reference_id"`
	
	// Timestamps
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CancelledAt     *time.Time `json:"cancelled_at" db:"cancelled_at"`
	
	// Related data (loaded separately)
	Items           []OrderItem `json:"items,omitempty"`
	User            *User       `json:"user,omitempty"`
	Vendor          *Vendor     `json:"vendor,omitempty"`
	StatusHistory   []OrderStatusHistory `json:"status_history,omitempty"`
}

// OrderAddress represents delivery address for an order
type OrderAddress struct {
	ID         int64  `json:"id" db:"id"`
	Street     string `json:"street" db:"street"`
	City       string `json:"city" db:"city"`
	State      string `json:"state" db:"state"`
	PostalCode string `json:"postal_code" db:"postal_code"`
	Country    string `json:"country" db:"country"`
}

// Cart represents a shopping cart
type Cart struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	Items     []CartItem `json:"items"`
	Total     float64    `json:"total"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	ID        int       `json:"id" db:"id"`
	CartID    int       `json:"cart_id" db:"cart_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	VariantID *int      `json:"variant_id" db:"variant_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"`
	Total     float64   `json:"total" db:"total"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID              int64   `json:"id" db:"id"`
	OrderID         int64   `json:"order_id" db:"order_id"`
	ProductID       int64   `json:"product_id" db:"product_id"`
	ProductName     string  `json:"product_name" db:"product_name"`
	ProductSKU      string  `json:"product_sku" db:"product_sku"`
	Quantity        int     `json:"quantity" db:"quantity"`
	UnitPrice       float64 `json:"unit_price" db:"unit_price"`
	TotalPrice      float64 `json:"total_price" db:"total_price"`
	IsWholesale     bool    `json:"is_wholesale" db:"is_wholesale"`
	ProductSnapshot string  `json:"product_snapshot" db:"product_snapshot"` // JSON snapshot of product at time of order
	
	// Related data
	Product         *Product `json:"product,omitempty"`
}

// OrderStatusHistory tracks order status changes
type OrderStatusHistory struct {
	ID          int64     `json:"id" db:"id"`
	OrderID     int64     `json:"order_id" db:"order_id"`
	Status      string    `json:"status" db:"status"`
	PreviousStatus string `json:"previous_status" db:"previous_status"`
	Comment     string    `json:"comment" db:"comment"`
	ChangedBy   int64     `json:"changed_by" db:"changed_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// OrderPayment represents payment information for an order
type OrderPayment struct {
	ID              int64     `json:"id" db:"id"`
	OrderID         int64     `json:"order_id" db:"order_id"`
	PaymentMethod   string    `json:"payment_method" db:"payment_method"`
	PaymentProvider string    `json:"payment_provider" db:"payment_provider"`
	TransactionID   string    `json:"transaction_id" db:"transaction_id"`
	Amount          float64   `json:"amount" db:"amount"`
	Currency        string    `json:"currency" db:"currency"`
	Status          string    `json:"status" db:"status"` // pending, completed, failed, refunded
	GatewayResponse string    `json:"gateway_response" db:"gateway_response"`
	ProcessedAt     *time.Time `json:"processed_at" db:"processed_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// OrderShipment represents shipment information
type OrderShipment struct {
	ID             int64     `json:"id" db:"id"`
	OrderID        int64     `json:"order_id" db:"order_id"`
	TrackingNumber string    `json:"tracking_number" db:"tracking_number"`
	CarrierName    string    `json:"carrier_name" db:"carrier_name"`
	ShippingMethod string    `json:"shipping_method" db:"shipping_method"`
	Status         string    `json:"status" db:"status"` // pending, shipped, in_transit, delivered, failed
	ShippedAt      *time.Time `json:"shipped_at" db:"shipped_at"`
	DeliveredAt    *time.Time `json:"delivered_at" db:"delivered_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// OrderRefund represents refund information
type OrderRefund struct {
	ID          int64     `json:"id" db:"id"`
	OrderID     int64     `json:"order_id" db:"order_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Reason      string    `json:"reason" db:"reason"`
	Status      string    `json:"status" db:"status"` // pending, approved, rejected, processed
	ProcessedBy int64     `json:"processed_by" db:"processed_by"`
	ProcessedAt *time.Time `json:"processed_at" db:"processed_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Validation methods

// Validate validates order data
func (o *Order) Validate() error {
	if o.UserID <= 0 {
		return errors.New("valid user ID is required")
	}

	if o.TotalAmount < 0 {
		return errors.New("total amount cannot be negative")
	}

	if strings.TrimSpace(o.ShippingAddress) == "" {
		return errors.New("shipping address is required")
	}

	if strings.TrimSpace(o.ShippingCity) == "" {
		return errors.New("shipping city is required")
	}

	if strings.TrimSpace(o.ShippingCountry) == "" {
		return errors.New("shipping country is required")
	}

	// Validate status
	validStatuses := []string{"pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "refunded"}
	if o.Status != "" && !contains(validStatuses, o.Status) {
		return errors.New("invalid order status")
	}

	// Validate payment status
	validPaymentStatuses := []string{"pending", "paid", "failed", "refunded", "partial"}
	if o.PaymentStatus != "" && !contains(validPaymentStatuses, o.PaymentStatus) {
		return errors.New("invalid payment status")
	}

	// Validate currency
	if o.Currency == "" {
		o.Currency = "TRY" // Default currency
	}

	return nil
}

// Business logic methods

// CanBeCancelled checks if order can be cancelled
func (o *Order) CanBeCancelled() bool {
	cancelableStatuses := []string{"pending", "confirmed"}
	return contains(cancelableStatuses, o.Status)
}

// CanBeRefunded checks if order can be refunded
func (o *Order) CanBeRefunded() bool {
	refundableStatuses := []string{"delivered", "shipped"}
	return contains(refundableStatuses, o.Status) && o.PaymentStatus == "paid"
}

// CanBeShipped checks if order can be shipped
func (o *Order) CanBeShipped() bool {
	return o.Status == "confirmed" || o.Status == "processing"
}

// IsCompleted checks if order is completed
func (o *Order) IsCompleted() bool {
	return o.Status == "delivered"
}

// IsCancelled checks if order is cancelled
func (o *Order) IsCancelled() bool {
	return o.Status == "cancelled"
}

// IsRefunded checks if order is refunded
func (o *Order) IsRefunded() bool {
	return o.Status == "refunded" || o.PaymentStatus == "refunded"
}

// GetItemCount returns total number of items in order
func (o *Order) GetItemCount() int {
	count := 0
	for _, item := range o.Items {
		count += item.Quantity
	}
	return count
}

// GetItemsValue returns total value of items (excluding tax, shipping, etc.)
func (o *Order) GetItemsValue() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.TotalPrice
	}
	return total
}

// CalculateTotal calculates total order amount
func (o *Order) CalculateTotal() {
	o.SubtotalAmount = o.GetItemsValue()
	o.TotalAmount = o.SubtotalAmount + o.TaxAmount + o.ShippingAmount - o.DiscountAmount
}

// OrderItem validation and methods

// Validate validates order item data
func (oi *OrderItem) Validate() error {
	if oi.ProductID <= 0 {
		return errors.New("valid product ID is required")
	}

	if oi.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	if oi.UnitPrice < 0 {
		return errors.New("unit price cannot be negative")
	}

	if strings.TrimSpace(oi.ProductName) == "" {
		return errors.New("product name is required")
	}

	return nil
}

// CalculateTotal calculates total price for the item
func (oi *OrderItem) CalculateTotal() {
	oi.TotalPrice = float64(oi.Quantity) * oi.UnitPrice
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// OrderFilter represents filters for order queries
type OrderFilter struct {
	UserID         *int64     `json:"user_id,omitempty"`
	VendorID       *int64     `json:"vendor_id,omitempty"`
	Status         []string   `json:"status,omitempty"`
	PaymentStatus  []string   `json:"payment_status,omitempty"`
	DateFrom       *time.Time `json:"date_from,omitempty"`
	DateTo         *time.Time `json:"date_to,omitempty"`
	MinAmount      *float64   `json:"min_amount,omitempty"`
	MaxAmount      *float64   `json:"max_amount,omitempty"`
	SearchTerm     string     `json:"search_term,omitempty"`
}

// OrderSummary represents order summary statistics
type OrderSummary struct {
	TotalOrders     int     `json:"total_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	PendingOrders   int     `json:"pending_orders"`
	CompletedOrders int     `json:"completed_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	AverageValue    float64 `json:"average_value"`
}

// Constants for order statuses
const (
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
	OrderStatusRefunded   = "refunded"
)

// OrderStats represents order statistics
type OrderStats struct {
	TotalOrders     int     `json:"total_orders"`
	PendingOrders   int     `json:"pending_orders"`
	ConfirmedOrders int     `json:"confirmed_orders"`
	DeliveredOrders int     `json:"delivered_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	AverageValue    float64 `json:"average_value"`
}

// Note: PaymentStatus constants moved to payment.go model
