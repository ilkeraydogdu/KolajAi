package services

import (
	"errors"
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"time"
)

// PaymentMethod represents payment methods
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodDebitCard  PaymentMethod = "debit_card"
	PaymentMethodPayPal     PaymentMethod = "paypal"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodCash       PaymentMethod = "cash"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// PaymentRequest represents a payment request
type PaymentRequest struct {
	OrderID       int64         `json:"order_id"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Method        PaymentMethod `json:"method"`
	CustomerID    int64         `json:"customer_id"`
	Description   string        `json:"description"`
	CallbackURL   string        `json:"callback_url"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	ID            string        `json:"id"`
	Status        PaymentStatus `json:"status"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Method        PaymentMethod `json:"method"`
	TransactionID string        `json:"transaction_id"`
	CreatedAt     time.Time     `json:"created_at"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	FailureReason string        `json:"failure_reason,omitempty"`
	PaymentURL    string        `json:"payment_url,omitempty"`
}

// PaymentService handles payment operations
type PaymentService struct {
	repo database.SimpleRepository
}

// NewPaymentService creates a new payment service
func NewPaymentService(repo database.SimpleRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

// ProcessPayment processes a payment request
func (s *PaymentService) ProcessPayment(request *PaymentRequest) (*PaymentResponse, error) {
	if request.Amount <= 0 {
		return nil, errors.New("invalid payment amount")
	}

	if request.OrderID == 0 {
		return nil, errors.New("order ID is required")
	}

	// Generate transaction ID
	transactionID := fmt.Sprintf("TXN_%d_%d", request.OrderID, time.Now().Unix())

	// Create payment response
	response := &PaymentResponse{
		ID:            transactionID,
		Status:        PaymentStatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		Method:        request.Method,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	// Mock different payment methods
	switch request.Method {
	case PaymentMethodCreditCard, PaymentMethodDebitCard:
		// Simulate card payment processing
		if s.simulateCardPayment(request) {
			response.Status = PaymentStatusCompleted
			now := time.Now()
			response.CompletedAt = &now
		} else {
			response.Status = PaymentStatusFailed
			response.FailureReason = "Card payment failed"
		}

	case PaymentMethodPayPal:
		// Simulate PayPal redirect
		response.Status = PaymentStatusPending
		response.PaymentURL = fmt.Sprintf("https://paypal.com/checkout?token=%s", transactionID)

	case PaymentMethodBankTransfer:
		// Bank transfer requires manual verification
		response.Status = PaymentStatusPending
		response.PaymentURL = fmt.Sprintf("/payment/bank-transfer/%s", transactionID)

	case PaymentMethodCash:
		// Cash payment - mark as pending for collection
		response.Status = PaymentStatusPending

	default:
		return nil, errors.New("unsupported payment method")
	}

	// Save payment record (mock)
	s.savePaymentRecord(response)

	return response, nil
}

// GetPaymentStatus gets payment status by transaction ID
func (s *PaymentService) GetPaymentStatus(transactionID string) (*PaymentResponse, error) {
	// Mock implementation - in real system, this would query database
	return &PaymentResponse{
		ID:            transactionID,
		Status:        PaymentStatusCompleted,
		TransactionID: transactionID,
		CreatedAt:     time.Now().Add(-5 * time.Minute),
		CompletedAt:   func() *time.Time { t := time.Now(); return &t }(),
	}, nil
}

// RefundPayment refunds a payment
func (s *PaymentService) RefundPayment(transactionID string, amount float64, reason string) (*PaymentResponse, error) {
	// Mock refund processing
	refundID := fmt.Sprintf("REF_%s_%d", transactionID, time.Now().Unix())
	
	return &PaymentResponse{
		ID:            refundID,
		Status:        PaymentStatusRefunded,
		Amount:        amount,
		TransactionID: refundID,
		CreatedAt:     time.Now(),
		CompletedAt:   func() *time.Time { t := time.Now(); return &t }(),
	}, nil
}

// GetSupportedPaymentMethods returns supported payment methods
func (s *PaymentService) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PaymentMethodCreditCard,
		PaymentMethodDebitCard,
		PaymentMethodPayPal,
		PaymentMethodBankTransfer,
		PaymentMethodCash,
	}
}

// ValidatePaymentMethod validates if payment method is supported
func (s *PaymentService) ValidatePaymentMethod(method PaymentMethod) bool {
	supportedMethods := s.GetSupportedPaymentMethods()
	for _, supported := range supportedMethods {
		if method == supported {
			return true
		}
	}
	return false
}

// simulateCardPayment simulates card payment processing
func (s *PaymentService) simulateCardPayment(request *PaymentRequest) bool {
	// Mock success rate: 90%
	return time.Now().UnixNano()%10 < 9
}

// savePaymentRecord saves payment record to database (mock)
func (s *PaymentService) savePaymentRecord(payment *PaymentResponse) error {
	// In real implementation, this would save to database
	// For now, just return success
	return nil
}

// CalculatePaymentFee calculates payment processing fee
func (s *PaymentService) CalculatePaymentFee(amount float64, method PaymentMethod) float64 {
	switch method {
	case PaymentMethodCreditCard:
		return amount * 0.029 // 2.9%
	case PaymentMethodDebitCard:
		return amount * 0.019 // 1.9%
	case PaymentMethodPayPal:
		return amount * 0.034 // 3.4%
	case PaymentMethodBankTransfer:
		return 5.0 // Fixed fee
	case PaymentMethodCash:
		return 0.0 // No fee
	default:
		return 0.0
	}
}

// CreatePaymentIntent creates a payment intent for frontend
func (s *PaymentService) CreatePaymentIntent(orderID int64, amount float64) (map[string]interface{}, error) {
	intentID := fmt.Sprintf("PI_%d_%d", orderID, time.Now().Unix())
	
	return map[string]interface{}{
		"id":            intentID,
		"amount":        amount,
		"currency":      "TRY",
		"status":        "requires_payment_method",
		"client_secret": fmt.Sprintf("%s_secret", intentID),
		"created":       time.Now().Unix(),
	}, nil
}