package payment

import (
	"context"
	"time"
	"kolajAi/internal/integrations"
)

// PaymentProvider interface for all payment gateways
type PaymentProvider interface {
	integrations.IntegrationProvider
	
	// Payment operations
	CreatePayment(ctx context.Context, payment *PaymentRequest) (*PaymentResponse, error)
	CapturePayment(ctx context.Context, paymentID string, amount float64) (*PaymentResponse, error)
	RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error)
	GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatus, error)
	
	// 3D Secure operations
	Initialize3DSecure(ctx context.Context, payment *PaymentRequest) (*ThreeDSecureResponse, error)
	Verify3DSecure(ctx context.Context, paymentID string, verificationData map[string]string) (*PaymentResponse, error)
	
	// Tokenization
	TokenizeCard(ctx context.Context, card *CardDetails) (*CardToken, error)
	DeleteToken(ctx context.Context, tokenID string) error
	
	// Subscription operations
	CreateSubscription(ctx context.Context, subscription *SubscriptionRequest) (*SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, subscriptionID string) error
	UpdateSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) (*SubscriptionResponse, error)
	
	// Reporting
	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
	ListTransactions(ctx context.Context, filters TransactionFilters) ([]*Transaction, error)
	GetBalance(ctx context.Context) (*Balance, error)
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Description     string                 `json:"description"`
	OrderID         string                 `json:"order_id"`
	CustomerID      string                 `json:"customer_id"`
	PaymentMethod   PaymentMethod          `json:"payment_method"`
	BillingAddress  Address                `json:"billing_address"`
	ShippingAddress Address                `json:"shipping_address"`
	Items           []PaymentItem          `json:"items"`
	Metadata        map[string]interface{} `json:"metadata"`
	ReturnURL       string                 `json:"return_url"`
	CallbackURL     string                 `json:"callback_url"`
	Installment     int                    `json:"installment,omitempty"`
	Enable3DSecure  bool                   `json:"enable_3d_secure"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	ID              string                 `json:"id"`
	Status          PaymentStatusType      `json:"status"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	PaymentMethod   PaymentMethod          `json:"payment_method"`
	TransactionID   string                 `json:"transaction_id"`
	AuthCode        string                 `json:"auth_code,omitempty"`
	ReferenceNumber string                 `json:"reference_number"`
	CreatedAt       time.Time              `json:"created_at"`
	ProcessedAt     time.Time              `json:"processed_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Error           *PaymentError          `json:"error,omitempty"`
}

// PaymentMethod represents payment method details
type PaymentMethod struct {
	Type        PaymentMethodType      `json:"type"`
	Card        *CardDetails           `json:"card,omitempty"`
	BankAccount *BankAccountDetails    `json:"bank_account,omitempty"`
	Wallet      *WalletDetails         `json:"wallet,omitempty"`
	Token       string                 `json:"token,omitempty"`
}

// PaymentMethodType represents the type of payment method
type PaymentMethodType string

const (
	PaymentMethodTypeCard        PaymentMethodType = "card"
	PaymentMethodTypeBankTransfer PaymentMethodType = "bank_transfer"
	PaymentMethodTypeWallet      PaymentMethodType = "wallet"
	PaymentMethodTypeCrypto      PaymentMethodType = "crypto"
)

// CardDetails represents credit/debit card details
type CardDetails struct {
	Number      string `json:"number"`
	ExpMonth    string `json:"exp_month"`
	ExpYear     string `json:"exp_year"`
	CVV         string `json:"cvv"`
	HolderName  string `json:"holder_name"`
	Brand       string `json:"brand,omitempty"`
	Last4       string `json:"last4,omitempty"`
}

// BankAccountDetails represents bank account details
type BankAccountDetails struct {
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
	AccountType   string `json:"account_type"`
	BankName      string `json:"bank_name"`
}

// WalletDetails represents digital wallet details
type WalletDetails struct {
	Provider string `json:"provider"` // paypal, apple_pay, google_pay
	Token    string `json:"token"`
}

// Address represents a billing or shipping address
type Address struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Company      string `json:"company,omitempty"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
}

// PaymentItem represents an item in the payment
type PaymentItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Tax         float64 `json:"tax,omitempty"`
	Category    string  `json:"category,omitempty"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus struct {
	ID            string            `json:"id"`
	Status        PaymentStatusType `json:"status"`
	Amount        float64           `json:"amount"`
	RefundedAmount float64          `json:"refunded_amount,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Events        []PaymentEvent    `json:"events,omitempty"`
}

// PaymentStatusType represents the type of payment status
type PaymentStatusType string

const (
	PaymentStatusPending    PaymentStatusType = "pending"
	PaymentStatusProcessing PaymentStatusType = "processing"
	PaymentStatusSucceeded  PaymentStatusType = "succeeded"
	PaymentStatusFailed     PaymentStatusType = "failed"
	PaymentStatusCanceled   PaymentStatusType = "canceled"
	PaymentStatusRefunded   PaymentStatusType = "refunded"
	PaymentStatusPartiallyRefunded PaymentStatusType = "partially_refunded"
)

// PaymentEvent represents an event in the payment lifecycle
type PaymentEvent struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	ID            string    `json:"id"`
	PaymentID     string    `json:"payment_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	ProcessedAt   time.Time `json:"processed_at,omitempty"`
}

// ThreeDSecureResponse represents 3D Secure initialization response
type ThreeDSecureResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	RedirectURL string `json:"redirect_url"`
	HTMLContent string `json:"html_content,omitempty"`
	Method      string `json:"method"` // redirect or iframe
}

// CardToken represents a tokenized card
type CardToken struct {
	ID         string    `json:"id"`
	Last4      string    `json:"last4"`
	Brand      string    `json:"brand"`
	ExpMonth   string    `json:"exp_month"`
	ExpYear    string    `json:"exp_year"`
	HolderName string    `json:"holder_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// SubscriptionRequest represents a subscription request
type SubscriptionRequest struct {
	PlanID        string                 `json:"plan_id"`
	CustomerID    string                 `json:"customer_id"`
	PaymentMethod PaymentMethod          `json:"payment_method"`
	StartDate     time.Time              `json:"start_date,omitempty"`
	TrialDays     int                    `json:"trial_days,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	ID                string                 `json:"id"`
	PlanID            string                 `json:"plan_id"`
	CustomerID        string                 `json:"customer_id"`
	Status            string                 `json:"status"`
	CurrentPeriodStart time.Time             `json:"current_period_start"`
	CurrentPeriodEnd   time.Time             `json:"current_period_end"`
	NextBillingDate    time.Time             `json:"next_billing_date"`
	Amount             float64               `json:"amount"`
	Currency           string                `json:"currency"`
	CreatedAt          time.Time             `json:"created_at"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// Transaction represents a payment transaction
type Transaction struct {
	ID              string            `json:"id"`
	Type            string            `json:"type"` // payment, refund, chargeback
	Status          PaymentStatusType `json:"status"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	PaymentMethodType string          `json:"payment_method_type"`
	CustomerID      string            `json:"customer_id"`
	OrderID         string            `json:"order_id"`
	Description     string            `json:"description"`
	CreatedAt       time.Time         `json:"created_at"`
	ProcessedAt     time.Time         `json:"processed_at,omitempty"`
	Fees            float64           `json:"fees,omitempty"`
	NetAmount       float64           `json:"net_amount,omitempty"`
}

// TransactionFilters represents filters for listing transactions
type TransactionFilters struct {
	StartDate     time.Time
	EndDate       time.Time
	Status        PaymentStatusType
	CustomerID    string
	MinAmount     float64
	MaxAmount     float64
	Currency      string
	PaymentMethod string
	Limit         int
	Offset        int
}

// Balance represents account balance
type Balance struct {
	Available []BalanceAmount `json:"available"`
	Pending   []BalanceAmount `json:"pending"`
	Reserved  []BalanceAmount `json:"reserved"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// BalanceAmount represents balance in a specific currency
type BalanceAmount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// PaymentError represents a payment error
type PaymentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// WebhookPayload represents a payment webhook payload
type WebhookPayload struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Created   time.Time              `json:"created"`
	Data      map[string]interface{} `json:"data"`
	Signature string                 `json:"signature"`
}

// PaymentProviderConfig represents configuration for a payment provider
type PaymentProviderConfig struct {
	APIKey              string
	APISecret           string
	MerchantID          string
	Environment         string // production, sandbox
	WebhookSecret       string
	Enable3DSecure      bool
	EnableInstallments  bool
	MaxInstallments     int
	SupportedCurrencies []string
	SupportedCountries  []string
	Timeout             time.Duration
}

// Common error codes
const (
	ErrorCodeInsufficientFunds    = "insufficient_funds"
	ErrorCodeCardDeclined         = "card_declined"
	ErrorCodeInvalidCard          = "invalid_card"
	ErrorCodeExpiredCard          = "expired_card"
	ErrorCodeProcessingError      = "processing_error"
	ErrorCodeFraudDetected        = "fraud_detected"
	ErrorCode3DSecureRequired     = "3d_secure_required"
	ErrorCode3DSecureFailed       = "3d_secure_failed"
	ErrorCodeDuplicateTransaction = "duplicate_transaction"
	ErrorCodeInvalidAmount        = "invalid_amount"
	ErrorCodeInvalidCurrency      = "invalid_currency"
)