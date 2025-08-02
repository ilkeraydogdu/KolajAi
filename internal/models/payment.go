package models

import (
	"fmt"
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Payment represents a payment transaction
type Payment struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	
	// Order Relationship
	OrderID     uint      `json:"order_id" gorm:"index;not null"`
	Order       Order     `json:"order" gorm:"foreignKey:OrderID"`
	
	// Customer Relationship
	CustomerID  uint      `json:"customer_id" gorm:"index;not null"`
	Customer    Customer  `json:"customer" gorm:"foreignKey:CustomerID"`
	
	// Payment Details
	Amount      float64   `json:"amount" gorm:"type:decimal(15,2);not null" validate:"required,gt=0"`
	Currency    string    `json:"currency" gorm:"size:3;not null;default:'TRY'" validate:"required,len=3"`
	Status      PaymentStatus `json:"status" gorm:"default:'pending'"`
	Method      PaymentMethod `json:"method" gorm:"not null"`
	
	// Provider Information
	Provider    PaymentProvider `json:"provider" gorm:"not null"`
	ProviderTransactionID string `json:"provider_transaction_id" gorm:"size:255;index"`
	ProviderReference     string `json:"provider_reference" gorm:"size:255"`
	
	// Transaction Details
	TransactionID   string    `json:"transaction_id" gorm:"size:100;unique;not null"`
	Description     string    `json:"description" gorm:"size:500"`
	
	// Payment Flow
	AuthorizationCode string  `json:"authorization_code" gorm:"size:100"`
	CaptureID        string   `json:"capture_id" gorm:"size:100"`
	RefundID         string   `json:"refund_id" gorm:"size:100"`
	
	// 3D Secure
	ThreeDSecureID   string  `json:"three_d_secure_id" gorm:"size:100"`
	ThreeDSecureStatus string `json:"three_d_secure_status" gorm:"size:50"`
	
	// Card Information (encrypted/tokenized)
	CardToken        string  `json:"card_token" gorm:"size:255"`
	CardLast4        string  `json:"card_last4" gorm:"size:4"`
	CardBrand        string  `json:"card_brand" gorm:"size:50"`
	CardExpMonth     string  `json:"card_exp_month" gorm:"size:2"`
	CardExpYear      string  `json:"card_exp_year" gorm:"size:4"`
	CardHolderName   string  `json:"card_holder_name" gorm:"size:255"`
	
	// Installment Information
	InstallmentCount int     `json:"installment_count" gorm:"default:1"`
	InstallmentRate  float64 `json:"installment_rate" gorm:"type:decimal(5,4);default:0"`
	
	// Fees and Commissions
	ProviderFee     float64 `json:"provider_fee" gorm:"type:decimal(15,2);default:0"`
	PlatformFee     float64 `json:"platform_fee" gorm:"type:decimal(15,2);default:0"`
	TaxAmount       float64 `json:"tax_amount" gorm:"type:decimal(15,2);default:0"`
	NetAmount       float64 `json:"net_amount" gorm:"type:decimal(15,2)"`
	
	// Timing
	AuthorizedAt    *time.Time `json:"authorized_at"`
	CapturedAt      *time.Time `json:"captured_at"`
	RefundedAt      *time.Time `json:"refunded_at"`
	FailedAt        *time.Time `json:"failed_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	
	// Error Information
	ErrorCode       string  `json:"error_code" gorm:"size:100"`
	ErrorMessage    string  `json:"error_message" gorm:"size:500"`
	
	// Metadata and Additional Info
	Metadata        PaymentMetadata `json:"metadata" gorm:"type:json"`
	IPAddress       string         `json:"ip_address" gorm:"size:45"`
	UserAgent       string         `json:"user_agent" gorm:"size:500"`
	
	// Fraud Detection
	FraudScore      float64        `json:"fraud_score" gorm:"type:decimal(5,4);default:0"`
	FraudStatus     FraudStatus    `json:"fraud_status" gorm:"default:'pending'"`
	FraudChecks     FraudChecks    `json:"fraud_checks" gorm:"type:json"`
	
	// Webhooks and Notifications
	WebhookReceived bool      `json:"webhook_received" gorm:"default:false"`
	WebhookAt       *time.Time `json:"webhook_at"`
	NotificationSent bool     `json:"notification_sent" gorm:"default:false"`
	
	// Related Payments
	ParentPaymentID *uint     `json:"parent_payment_id" gorm:"index"`
	ParentPayment   *Payment  `json:"parent_payment,omitempty" gorm:"foreignKey:ParentPaymentID"`
	ChildPayments   []Payment `json:"child_payments,omitempty" gorm:"foreignKey:ParentPaymentID"`
	
	// Timestamps
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusCaptured   PaymentStatus = "captured"
	PaymentStatusPaid       PaymentStatus = "paid"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusExpired    PaymentStatus = "expired"
	PaymentStatusProcessing PaymentStatus = "processing"
)

// PaymentMethod represents the payment method used
type PaymentMethod string

const (
	PaymentMethodCreditCard    PaymentMethod = "credit_card"
	PaymentMethodDebitCard     PaymentMethod = "debit_card"
	PaymentMethodBankTransfer  PaymentMethod = "bank_transfer"
	PaymentMethodWallet        PaymentMethod = "wallet"
	PaymentMethodCrypto        PaymentMethod = "crypto"
	PaymentMethodInstallment   PaymentMethod = "installment"
	PaymentMethodBuyNowPayLater PaymentMethod = "buy_now_pay_later"
	PaymentMethodCOD           PaymentMethod = "cash_on_delivery"
)

// PaymentProvider represents the payment service provider
type PaymentProvider string

const (
	PaymentProviderIyzico     PaymentProvider = "iyzico"
	PaymentProviderPayTR      PaymentProvider = "paytr"
	PaymentProviderStripe     PaymentProvider = "stripe"
	PaymentProviderPayPal     PaymentProvider = "paypal"
	PaymentProviderKlarna     PaymentProvider = "klarna"
	PaymentProviderPapara     PaymentProvider = "papara"
	PaymentProviderBiTaksi    PaymentProvider = "bitaksi"
	PaymentProviderInternal   PaymentProvider = "internal"
)

// FraudStatus represents fraud detection status
type FraudStatus string

const (
	FraudStatusPending  FraudStatus = "pending"
	FraudStatusApproved FraudStatus = "approved"
	FraudStatusDeclined FraudStatus = "declined"
	FraudStatusReview   FraudStatus = "review"
)

// PaymentMetadata holds additional payment data
type PaymentMetadata struct {
	BrowserInfo    BrowserInfo            `json:"browser_info,omitempty"`
	DeviceInfo     DeviceInfo             `json:"device_info,omitempty"`
	LocationInfo   LocationInfo           `json:"location_info,omitempty"`
	CustomFields   map[string]interface{} `json:"custom_fields,omitempty"`
	ProviderData   map[string]interface{} `json:"provider_data,omitempty"`
}

// BrowserInfo holds browser-related information
type BrowserInfo struct {
	UserAgent      string `json:"user_agent,omitempty"`
	AcceptHeader   string `json:"accept_header,omitempty"`
	Language       string `json:"language,omitempty"`
	ColorDepth     int    `json:"color_depth,omitempty"`
	ScreenHeight   int    `json:"screen_height,omitempty"`
	ScreenWidth    int    `json:"screen_width,omitempty"`
	TimeZoneOffset int    `json:"timezone_offset,omitempty"`
	JavaEnabled    bool   `json:"java_enabled,omitempty"`
}

// DeviceInfo holds device-related information
type DeviceInfo struct {
	DeviceID       string `json:"device_id,omitempty"`
	DeviceType     string `json:"device_type,omitempty"`
	Platform       string `json:"platform,omitempty"`
	Model          string `json:"model,omitempty"`
	Fingerprint    string `json:"fingerprint,omitempty"`
}

// LocationInfo holds location-related information
type LocationInfo struct {
	Country     string  `json:"country,omitempty"`
	City        string  `json:"city,omitempty"`
	Region      string  `json:"region,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	IPAddress   string  `json:"ip_address,omitempty"`
	ISP         string  `json:"isp,omitempty"`
}

// FraudChecks holds fraud detection results
type FraudChecks struct {
	VelocityCheck    bool    `json:"velocity_check"`
	BlacklistCheck   bool    `json:"blacklist_check"`
	GeolocationCheck bool    `json:"geolocation_check"`
	DeviceCheck      bool    `json:"device_check"`
	BehaviorScore    float64 `json:"behavior_score"`
	RiskFactors      []string `json:"risk_factors,omitempty"`
}

// PaymentRefund represents a payment refund
type PaymentRefund struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	PaymentID       uint      `json:"payment_id" gorm:"index;not null"`
	Payment         Payment   `json:"payment" gorm:"foreignKey:PaymentID"`
	
	Amount          float64   `json:"amount" gorm:"type:decimal(15,2);not null"`
	Currency        string    `json:"currency" gorm:"size:3;not null"`
	Reason          string    `json:"reason" gorm:"size:500"`
	Status          RefundStatus `json:"status" gorm:"default:'pending'"`
	
	ProviderRefundID string   `json:"provider_refund_id" gorm:"size:255"`
	RefundReference  string   `json:"refund_reference" gorm:"size:255"`
	
	ProcessedAt     *time.Time `json:"processed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// RefundStatus represents refund status
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusProcessed RefundStatus = "processed"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCancelled RefundStatus = "cancelled"
)

// Implement driver.Valuer interface for PaymentMetadata
func (pm PaymentMetadata) Value() (driver.Value, error) {
	return json.Marshal(pm)
}

// Implement sql.Scanner interface for PaymentMetadata
func (pm *PaymentMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, pm)
}

// Implement driver.Valuer interface for FraudChecks
func (fc FraudChecks) Value() (driver.Value, error) {
	return json.Marshal(fc)
}

// Implement sql.Scanner interface for FraudChecks
func (fc *FraudChecks) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, fc)
}

// TableName returns the table name for Payment
func (Payment) TableName() string {
	return "payments"
}

// TableName returns the table name for PaymentRefund
func (PaymentRefund) TableName() string {
	return "payment_refunds"
}

// IsSuccessful checks if payment is successful
func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentStatusPaid || p.Status == PaymentStatusCaptured
}

// CanBeRefunded checks if payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.IsSuccessful() && p.RefundedAt == nil
}

// CanBeCaptured checks if payment can be captured
func (p *Payment) CanBeCaptured() bool {
	return p.Status == PaymentStatusAuthorized
}

// IsExpired checks if payment authorization is expired
func (p *Payment) IsExpired() bool {
	return p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt)
}

// GetDisplayAmount returns formatted amount for display
func (p *Payment) GetDisplayAmount() string {
	return fmt.Sprintf("%.2f %s", p.Amount, p.Currency)
}

// GetMaskedCardNumber returns masked card number
func (p *Payment) GetMaskedCardNumber() string {
	if p.CardLast4 == "" {
		return ""
	}
	return "**** **** **** " + p.CardLast4
}

// HasInstallment checks if payment has installment
func (p *Payment) HasInstallment() bool {
	return p.InstallmentCount > 1
}

// GetInstallmentAmount returns per installment amount
func (p *Payment) GetInstallmentAmount() float64 {
	if p.InstallmentCount <= 1 {
		return p.Amount
	}
	return p.Amount / float64(p.InstallmentCount)
}

// IsHighRisk checks if payment is high risk based on fraud score
func (p *Payment) IsHighRisk() bool {
	return p.FraudScore > 0.7
}

// GetNetAmount calculates net amount after fees
func (p *Payment) GetNetAmount() float64 {
	return p.Amount - p.ProviderFee - p.PlatformFee - p.TaxAmount
}