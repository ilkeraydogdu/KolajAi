package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"kolajAi/internal/services"
)

// PaymentHandler handles payment-related requests
type PaymentHandler struct {
	*Handler
	paymentService *services.PaymentService
	orderService   *services.OrderService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(h *Handler, paymentService *services.PaymentService, orderService *services.OrderService) *PaymentHandler {
	return &PaymentHandler{
		Handler:        h,
		paymentService: paymentService,
		orderService:   orderService,
	}
}

// CreatePaymentIntent creates a payment intent
func (h *PaymentHandler) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		OrderID int64   `json:"order_id"`
		Amount  float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	intent, err := h.paymentService.CreatePaymentIntent(request.OrderID, request.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    intent,
	})
}

// ProcessPayment processes a payment
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request services.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate payment method
	if !h.paymentService.ValidatePaymentMethod(request.Method) {
		http.Error(w, "Unsupported payment method", http.StatusBadRequest)
		return
	}

	// Set currency if not provided
	if request.Currency == "" {
		request.Currency = "TRY"
	}

	response, err := h.paymentService.ProcessPayment(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// GetPaymentStatus gets payment status
func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	transactionID := strings.TrimPrefix(r.URL.Path, "/api/payment/status/")
	if transactionID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	status, err := h.paymentService.GetPaymentStatus(transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// RefundPayment refunds a payment
func (h *PaymentHandler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TransactionID string  `json:"transaction_id"`
		Amount        float64 `json:"amount"`
		Reason        string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.TransactionID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	refund, err := h.paymentService.RefundPayment(request.TransactionID, request.Amount, request.Reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    refund,
	})
}

// GetPaymentMethods returns supported payment methods
func (h *PaymentHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	methods := h.paymentService.GetSupportedPaymentMethods()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    methods,
	})
}

// CalculatePaymentFee calculates payment fee
func (h *PaymentHandler) CalculatePaymentFee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Amount float64                `json:"amount"`
		Method services.PaymentMethod `json:"method"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fee := h.paymentService.CalculatePaymentFee(request.Amount, request.Method)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"amount":     request.Amount,
			"fee":        fee,
			"total":      request.Amount + fee,
			"method":     request.Method,
		},
	})
}

// PaymentSuccess handles successful payment callback
func (h *PaymentHandler) PaymentSuccess(w http.ResponseWriter, r *http.Request) {
	transactionID := r.URL.Query().Get("transaction_id")
	if transactionID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	// Get payment status
	status, err := h.paymentService.GetPaymentStatus(transactionID)
	if err != nil {
		h.HandleError(w, r, err, "Failed to get payment status")
		return
	}

	data := h.GetTemplateData()
	data["Payment"] = status
	data["Success"] = true

	h.RenderTemplate(w, r, "payment/success", data)
}

// PaymentFailure handles failed payment callback
func (h *PaymentHandler) PaymentFailure(w http.ResponseWriter, r *http.Request) {
	transactionID := r.URL.Query().Get("transaction_id")
	reason := r.URL.Query().Get("reason")

	data := h.GetTemplateData()
	data["TransactionID"] = transactionID
	data["FailureReason"] = reason
	data["Success"] = false

	h.RenderTemplate(w, r, "payment/failure", data)
}

// PaymentPage shows the payment page
func (h *PaymentHandler) PaymentPage(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("order_id")
	if orderIDStr == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Get real order details from database
	order, err := h.orderService.GetOrderByID(int(orderID))
	if err != nil {
		Logger.Printf("Error getting order details: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	
	data := h.GetTemplateData()
	data["OrderID"] = orderID
	data["Order"] = order
	data["Amount"] = order.TotalAmount
	data["PaymentMethods"] = h.paymentService.GetSupportedPaymentMethods()

	h.RenderTemplate(w, r, "payment/checkout", data)
}