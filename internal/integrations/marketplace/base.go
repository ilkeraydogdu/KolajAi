package marketplace

import (
	"context"
	"time"

	"kolajAi/internal/integrations"
)

// MarketplaceProvider defines the interface for marketplace integrations
type MarketplaceProvider interface {
	// Base integration methods
	Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error
	GetName() string
	GetType() string
	IsHealthy(ctx context.Context) (bool, error)
	GetMetrics() map[string]interface{}
	GetRateLimit() integrations.RateLimitInfo

	// Product operations
	SyncProducts(ctx context.Context, products []interface{}) error
	UpdateStockAndPrice(ctx context.Context, updates []interface{}) error
	GetProducts(ctx context.Context, params map[string]interface{}) ([]interface{}, error)

	// Order operations
	GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string, params map[string]interface{}) error

	// Category operations
	GetCategories(ctx context.Context) ([]interface{}, error)
	GetBrands(ctx context.Context) ([]interface{}, error)
}

// MarketplaceProviderConfig holds configuration for marketplace providers
type MarketplaceProviderConfig struct {
	APIKey              string
	APISecret           string
	Environment         string
	SupportedCurrencies []string
	SupportedCountries  []string
	Timeout             time.Duration
	RateLimit           int
	WebhookSecret       string
	RetryAttempts       int
	RetryDelay          time.Duration
}

// MarketplaceError represents marketplace-specific errors
type MarketplaceError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Provider   string                 `json:"provider"`
	Retryable  bool                   `json:"retryable"`
	Timestamp  time.Time              `json:"timestamp"`
	StatusCode int                    `json:"status_code,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (e *MarketplaceError) Error() string {
	return e.Message
}

// Product represents a generic marketplace product
type Product struct {
	ID          string            `json:"id"`
	SKU         string            `json:"sku"`
	Barcode     string            `json:"barcode"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Brand       string            `json:"brand"`
	Category    string            `json:"category"`
	Price       float64           `json:"price"`
	ListPrice   float64           `json:"list_price"`
	Currency    string            `json:"currency"`
	Stock       int               `json:"stock"`
	Images      []string          `json:"images"`
	Attributes  map[string]string `json:"attributes"`
	Weight      float64           `json:"weight"`
	Dimensions  Dimensions        `json:"dimensions"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Dimensions represents product dimensions
type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"` // cm, in, etc.
}

// Order represents a generic marketplace order
type Order struct {
	ID              string      `json:"id"`
	OrderNumber     string      `json:"order_number"`
	Status          string      `json:"status"`
	CustomerID      string      `json:"customer_id"`
	CustomerName    string      `json:"customer_name"`
	CustomerEmail   string      `json:"customer_email"`
	CustomerPhone   string      `json:"customer_phone"`
	BillingAddress  Address     `json:"billing_address"`
	ShippingAddress Address     `json:"shipping_address"`
	Items           []OrderItem `json:"items"`
	Subtotal        float64     `json:"subtotal"`
	TaxAmount       float64     `json:"tax_amount"`
	ShippingAmount  float64     `json:"shipping_amount"`
	DiscountAmount  float64     `json:"discount_amount"`
	TotalAmount     float64     `json:"total_amount"`
	Currency        string      `json:"currency"`
	PaymentMethod   string      `json:"payment_method"`
	PaymentStatus   string      `json:"payment_status"`
	ShippingMethod  string      `json:"shipping_method"`
	TrackingNumber  string      `json:"tracking_number"`
	Notes           string      `json:"notes"`
	OrderDate       time.Time   `json:"order_date"`
	ShippedDate     *time.Time  `json:"shipped_date,omitempty"`
	DeliveredDate   *time.Time  `json:"delivered_date,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID             string            `json:"id"`
	ProductID      string            `json:"product_id"`
	SKU            string            `json:"sku"`
	Name           string            `json:"name"`
	Quantity       int               `json:"quantity"`
	Price          float64           `json:"price"`
	TotalPrice     float64           `json:"total_price"`
	TaxAmount      float64           `json:"tax_amount"`
	DiscountAmount float64           `json:"discount_amount"`
	Attributes     map[string]string `json:"attributes"`
}

// Address represents a billing or shipping address
type Address struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Company    string `json:"company"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Phone      string `json:"phone"`
}

// Category represents a marketplace category
type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id,omitempty"`
	Path     string `json:"path"`
	Level    int    `json:"level"`
}

// Brand represents a marketplace brand
type Brand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo,omitempty"`
}

// StockPriceUpdate represents a stock and price update
type StockPriceUpdate struct {
	SKU       string  `json:"sku"`
	Barcode   string  `json:"barcode"`
	Stock     int     `json:"stock"`
	Price     float64 `json:"price"`
	ListPrice float64 `json:"list_price"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	Type      string                 `json:"type"`
	Provider  string                 `json:"provider"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Signature string                 `json:"signature"`
}

// OrderStatus constants
const (
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
	OrderStatusReturned   = "returned"
)

// ProductStatus constants
const (
	ProductStatusActive   = "active"
	ProductStatusInactive = "inactive"
	ProductStatusDraft    = "draft"
	ProductStatusArchived = "archived"
)

// PaymentStatus constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
	PaymentStatusCancelled = "cancelled"
)
