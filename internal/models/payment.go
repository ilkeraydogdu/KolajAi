package models

import (
	"errors"
	"strings"
	"time"
)

// Payment represents a payment transaction
type Payment struct {
	ID                int64     `json:"id" db:"id"`
	OrderID           int64     `json:"order_id" db:"order_id"`
	UserID            int64     `json:"user_id" db:"user_id"`
	PaymentMethod     string    `json:"payment_method" db:"payment_method"` // credit_card, debit_card, bank_transfer, wallet, cash_on_delivery
	PaymentProvider   string    `json:"payment_provider" db:"payment_provider"` // iyzico, paypal, stripe, etc.
	TransactionID     string    `json:"transaction_id" db:"transaction_id"`
	ProviderTransactionID string `json:"provider_transaction_id" db:"provider_transaction_id"`
	Amount            float64   `json:"amount" db:"amount"`
	Currency          string    `json:"currency" db:"currency"`
	Fee               float64   `json:"fee" db:"fee"`
	NetAmount         float64   `json:"net_amount" db:"net_amount"`
	Status            string    `json:"status" db:"status"` // pending, processing, completed, failed, cancelled, refunded, partial_refund
	PaymentDate       *time.Time `json:"payment_date" db:"payment_date"`
	FailureReason     string    `json:"failure_reason" db:"failure_reason"`
	RefundAmount      float64   `json:"refund_amount" db:"refund_amount"`
	RefundDate        *time.Time `json:"refund_date" db:"refund_date"`
	RefundReason      string    `json:"refund_reason" db:"refund_reason"`
	Notes             string    `json:"notes" db:"notes"`
	Metadata          string    `json:"metadata" db:"metadata"` // JSON string for additional data
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	
	// Related data
	Order             *Order    `json:"order,omitempty"`
	User              *User     `json:"user,omitempty"`
	PaymentDetails    *PaymentDetails `json:"payment_details,omitempty"`
}

// PaymentDetails represents payment method details
type PaymentDetails struct {
	ID            int64     `json:"id" db:"id"`
	PaymentID     int64     `json:"payment_id" db:"payment_id"`
	CardType      string    `json:"card_type" db:"card_type"` // visa, mastercard, amex, etc.
	CardLastFour  string    `json:"card_last_four" db:"card_last_four"`
	CardHolderName string   `json:"card_holder_name" db:"card_holder_name"`
	ExpiryMonth   string    `json:"expiry_month" db:"expiry_month"`
	ExpiryYear    string    `json:"expiry_year" db:"expiry_year"`
	BankName      string    `json:"bank_name" db:"bank_name"`
	BankCode      string    `json:"bank_code" db:"bank_code"`
	IBAN          string    `json:"iban" db:"iban"`
	WalletType    string    `json:"wallet_type" db:"wallet_type"`
	WalletID      string    `json:"wallet_id" db:"wallet_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentInstallment represents installment payment details
type PaymentInstallment struct {
	ID              int64     `json:"id" db:"id"`
	PaymentID       int64     `json:"payment_id" db:"payment_id"`
	InstallmentNo   int       `json:"installment_no" db:"installment_no"`
	Amount          float64   `json:"amount" db:"amount"`
	DueDate         time.Time `json:"due_date" db:"due_date"`
	PaidDate        *time.Time `json:"paid_date" db:"paid_date"`
	Status          string    `json:"status" db:"status"` // pending, paid, overdue, cancelled
	LateFee         float64   `json:"late_fee" db:"late_fee"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentRefund represents a refund transaction
type PaymentRefund struct {
	ID                int64     `json:"id" db:"id"`
	PaymentID         int64     `json:"payment_id" db:"payment_id"`
	RefundTransactionID string  `json:"refund_transaction_id" db:"refund_transaction_id"`
	Amount            float64   `json:"amount" db:"amount"`
	Reason            string    `json:"reason" db:"reason"`
	Status            string    `json:"status" db:"status"` // pending, completed, failed
	ProcessedDate     *time.Time `json:"processed_date" db:"processed_date"`
	Notes             string    `json:"notes" db:"notes"`
	CreatedBy         int64     `json:"created_by" db:"created_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the payment data is valid
func (p *Payment) Validate() error {
	if p.OrderID <= 0 {
		return errors.New("valid order ID is required")
	}
	
	if p.UserID <= 0 {
		return errors.New("valid user ID is required")
	}
	
	if strings.TrimSpace(p.PaymentMethod) == "" {
		return errors.New("payment method cannot be empty")
	}
	
	if p.Amount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}
	
	if strings.TrimSpace(p.Currency) == "" {
		return errors.New("currency cannot be empty")
	}
	
	validMethods := []string{"credit_card", "debit_card", "bank_transfer", "wallet", "cash_on_delivery"}
	isValidMethod := false
	for _, method := range validMethods {
		if p.PaymentMethod == method {
			isValidMethod = true
			break
		}
	}
	if !isValidMethod {
		return errors.New("invalid payment method")
	}
	
	validStatuses := []string{"pending", "processing", "completed", "failed", "cancelled", "refunded", "partial_refund"}
	isValidStatus := false
	for _, status := range validStatuses {
		if p.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return errors.New("invalid payment status")
	}
	
	return nil
}

// IsCompleted checks if the payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == "completed"
}

// IsFailed checks if the payment has failed
func (p *Payment) IsFailed() bool {
	return p.Status == "failed" || p.Status == "cancelled"
}

// CanRefund checks if the payment can be refunded
func (p *Payment) CanRefund() bool {
	return p.Status == "completed" && p.RefundAmount < p.Amount
}

// GetRemainingRefundAmount returns the remaining amount that can be refunded
func (p *Payment) GetRemainingRefundAmount() float64 {
	if !p.CanRefund() {
		return 0
	}
	return p.Amount - p.RefundAmount
}

// Validate checks if the payment details data is valid
func (pd *PaymentDetails) Validate() error {
	if pd.PaymentID <= 0 {
		return errors.New("valid payment ID is required")
	}
	
	if pd.CardType != "" && pd.CardLastFour == "" {
		return errors.New("card last four digits are required for card payments")
	}
	
	if pd.CardType != "" && pd.CardHolderName == "" {
		return errors.New("card holder name is required for card payments")
	}
	
	return nil
}

// Validate checks if the payment installment data is valid
func (pi *PaymentInstallment) Validate() error {
	if pi.PaymentID <= 0 {
		return errors.New("valid payment ID is required")
	}
	
	if pi.InstallmentNo <= 0 {
		return errors.New("installment number must be greater than zero")
	}
	
	if pi.Amount <= 0 {
		return errors.New("installment amount must be greater than zero")
	}
	
	validStatuses := []string{"pending", "paid", "overdue", "cancelled"}
	isValidStatus := false
	for _, status := range validStatuses {
		if pi.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return errors.New("invalid installment status")
	}
	
	return nil
}

// IsOverdue checks if the installment is overdue
func (pi *PaymentInstallment) IsOverdue() bool {
	return pi.Status == "pending" && time.Now().After(pi.DueDate)
}

// Validate checks if the payment refund data is valid
func (pr *PaymentRefund) Validate() error {
	if pr.PaymentID <= 0 {
		return errors.New("valid payment ID is required")
	}
	
	if pr.Amount <= 0 {
		return errors.New("refund amount must be greater than zero")
	}
	
	if strings.TrimSpace(pr.Reason) == "" {
		return errors.New("refund reason cannot be empty")
	}
	
	if pr.CreatedBy <= 0 {
		return errors.New("valid creator ID is required")
	}
	
	validStatuses := []string{"pending", "completed", "failed"}
	isValidStatus := false
	for _, status := range validStatuses {
		if pr.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return errors.New("invalid refund status")
	}
	
	return nil
}