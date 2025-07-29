package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"kolajAi/internal/integrations"
)

// IyzicoProvider implements PaymentProvider for Iyzico
type IyzicoProvider struct {
	config      *PaymentProviderConfig
	httpClient  *http.Client
	credentials integrations.Credentials
	baseURL     string
	rateLimit   integrations.RateLimitInfo
}

// NewIyzicoProvider creates a new Iyzico payment provider
func NewIyzicoProvider() *IyzicoProvider {
	return &IyzicoProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: integrations.RateLimitInfo{
			RequestsPerMinute: 100,
			RequestsRemaining: 100,
			ResetsAt:          time.Now().Add(time.Minute),
		},
	}
}

// Initialize sets up the Iyzico provider
func (p *IyzicoProvider) Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error {
	p.credentials = credentials
	
	// Set base URL based on environment
	environment, _ := config["environment"].(string)
	if environment == "production" {
		p.baseURL = "https://api.iyzipay.com"
	} else {
		p.baseURL = "https://sandbox-api.iyzipay.com"
	}
	
	// Initialize configuration
	p.config = &PaymentProviderConfig{
		APIKey:              credentials.APIKey,
		APISecret:           credentials.APISecret,
		Environment:         environment,
		Enable3DSecure:      true,
		EnableInstallments:  true,
		MaxInstallments:     12,
		SupportedCurrencies: []string{"TRY", "USD", "EUR", "GBP"},
		SupportedCountries:  []string{"TR"},
		Timeout:             30 * time.Second,
	}
	
	// Additional config from map
	if enable3D, ok := config["enable_3d_secure"].(bool); ok {
		p.config.Enable3DSecure = enable3D
	}
	
	return nil
}

// HealthCheck verifies the Iyzico integration is working
func (p *IyzicoProvider) HealthCheck(ctx context.Context) error {
	// Test API connectivity
	request := map[string]interface{}{
		"locale":         "tr",
		"conversationId": fmt.Sprintf("health-check-%d", time.Now().Unix()),
	}
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "GET", "/payment/test", request, &response)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "HEALTH_CHECK_FAILED",
			Message:   "Failed to connect to Iyzico API",
			Provider:  "iyzico",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	
	if status, ok := response["status"].(string); !ok || status != "success" {
		return &integrations.IntegrationError{
			Code:      "HEALTH_CHECK_FAILED",
			Message:   "Iyzico API returned unsuccessful status",
			Provider:  "iyzico",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	
	return nil
}

// GetCapabilities returns the capabilities of Iyzico
func (p *IyzicoProvider) GetCapabilities() []string {
	return []string{
		"payment_processing",
		"3d_secure",
		"installments",
		"refunds",
		"partial_refunds",
		"tokenization",
		"subscriptions",
		"marketplace",
		"fraud_detection",
		"bin_checking",
	}
}

// GetRateLimit returns current rate limit information
func (p *IyzicoProvider) GetRateLimit() integrations.RateLimitInfo {
	return p.rateLimit
}

// Close cleans up any resources
func (p *IyzicoProvider) Close() error {
	// No specific cleanup needed for Iyzico
	return nil
}

// CreatePayment creates a new payment
func (p *IyzicoProvider) CreatePayment(ctx context.Context, payment *PaymentRequest) (*PaymentResponse, error) {
	// Build Iyzico payment request
	iyzicoRequest := p.buildPaymentRequest(payment)
	
	// Determine endpoint based on 3D Secure requirement
	endpoint := "/payment/auth"
	if payment.Enable3DSecure {
		endpoint = "/payment/3dsecure/initialize"
	}
	
	var iyzicoResponse map[string]interface{}
	err := p.makeRequest(ctx, "POST", endpoint, iyzicoRequest, &iyzicoResponse)
	if err != nil {
		return nil, err
	}
	
	// Parse response
	return p.parsePaymentResponse(iyzicoResponse, payment)
}

// CapturePayment captures a previously authorized payment
func (p *IyzicoProvider) CapturePayment(ctx context.Context, paymentID string, amount float64) (*PaymentResponse, error) {
	// Iyzico doesn't support separate auth/capture, payments are captured immediately
	return nil, &integrations.IntegrationError{
		Code:      "NOT_SUPPORTED",
		Message:   "Iyzico does not support separate authorization and capture",
		Provider:  "iyzico",
		Retryable: false,
		Timestamp: time.Now(),
	}
}

// RefundPayment refunds a payment
func (p *IyzicoProvider) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	request := map[string]interface{}{
		"locale":               "tr",
		"conversationId":       fmt.Sprintf("refund-%s-%d", paymentID, time.Now().Unix()),
		"paymentTransactionId": paymentID,
		"price":                fmt.Sprintf("%.2f", amount),
		"currency":             "TRY",
	}
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "POST", "/payment/refund", request, &response)
	if err != nil {
		return nil, err
	}
	
	status, _ := response["status"].(string)
	if status != "success" {
		errorMessage, _ := response["errorMessage"].(string)
		return nil, &integrations.IntegrationError{
			Code:      "REFUND_FAILED",
			Message:   errorMessage,
			Provider:  "iyzico",
			Retryable: false,
			Timestamp: time.Now(),
		}
	}
	
	return &RefundResponse{
		ID:          response["paymentId"].(string),
		PaymentID:   paymentID,
		Amount:      amount,
		Currency:    "TRY",
		Status:      "completed",
		CreatedAt:   time.Now(),
		ProcessedAt: time.Now(),
	}, nil
}

// GetPaymentStatus gets the status of a payment
func (p *IyzicoProvider) GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	request := map[string]interface{}{
		"locale":         "tr",
		"conversationId": fmt.Sprintf("status-%s-%d", paymentID, time.Now().Unix()),
		"paymentId":      paymentID,
	}
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "POST", "/payment/detail", request, &response)
	if err != nil {
		return nil, err
	}
	
	status, _ := response["status"].(string)
	if status != "success" {
		errorMessage, _ := response["errorMessage"].(string)
		return nil, &integrations.IntegrationError{
			Code:      "STATUS_CHECK_FAILED",
			Message:   errorMessage,
			Provider:  "iyzico",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	
	paymentStatus := "unknown"
	if phase, ok := response["phase"].(string); ok {
		switch phase {
		case "AUTH":
			paymentStatus = string(PaymentStatusSucceeded)
		case "PRE_AUTH":
			paymentStatus = string(PaymentStatusPending)
		case "FRAUD":
			paymentStatus = string(PaymentStatusFailed)
		}
	}
	
	return &PaymentStatus{
		ID:        paymentID,
		Status:    PaymentStatusType(paymentStatus),
		Amount:    response["price"].(float64),
		UpdatedAt: time.Now(),
	}, nil
}

// Initialize3DSecure initializes 3D Secure authentication
func (p *IyzicoProvider) Initialize3DSecure(ctx context.Context, payment *PaymentRequest) (*ThreeDSecureResponse, error) {
	iyzicoRequest := p.buildPaymentRequest(payment)
	iyzicoRequest["callbackUrl"] = payment.CallbackURL
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "POST", "/payment/3dsecure/initialize", iyzicoRequest, &response)
	if err != nil {
		return nil, err
	}
	
	status, _ := response["status"].(string)
	if status != "success" {
		errorMessage, _ := response["errorMessage"].(string)
		return nil, &integrations.IntegrationError{
			Code:      "3D_SECURE_INIT_FAILED",
			Message:   errorMessage,
			Provider:  "iyzico",
			Retryable: false,
			Timestamp: time.Now(),
		}
	}
	
	return &ThreeDSecureResponse{
		ID:          response["conversationId"].(string),
		Status:      "pending",
		HTMLContent: response["threeDSHtmlContent"].(string),
		Method:      "iframe",
	}, nil
}

// Verify3DSecure verifies 3D Secure authentication
func (p *IyzicoProvider) Verify3DSecure(ctx context.Context, paymentID string, verificationData map[string]string) (*PaymentResponse, error) {
	request := map[string]interface{}{
		"locale":         "tr",
		"conversationId": paymentID,
		"paymentId":      verificationData["paymentId"],
	}
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "POST", "/payment/3dsecure/auth", request, &response)
	if err != nil {
		return nil, err
	}
	
	return p.parsePaymentResponse(response, nil)
}

// TokenizeCard creates a token for a card
func (p *IyzicoProvider) TokenizeCard(ctx context.Context, card *CardDetails) (*CardToken, error) {
	request := map[string]interface{}{
		"locale":         "tr",
		"conversationId": fmt.Sprintf("tokenize-%d", time.Now().Unix()),
		"card": map[string]string{
			"cardHolderName": card.HolderName,
			"cardNumber":     card.Number,
			"expireMonth":    card.ExpMonth,
			"expireYear":     card.ExpYear,
		},
	}
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "POST", "/cardstorage/card", request, &response)
	if err != nil {
		return nil, err
	}
	
	status, _ := response["status"].(string)
	if status != "success" {
		errorMessage, _ := response["errorMessage"].(string)
		return nil, &integrations.IntegrationError{
			Code:      "TOKENIZATION_FAILED",
			Message:   errorMessage,
			Provider:  "iyzico",
			Retryable: false,
			Timestamp: time.Now(),
		}
	}
	
	return &CardToken{
		ID:         response["cardToken"].(string),
		Last4:      response["lastFourDigits"].(string),
		Brand:      response["cardAssociation"].(string),
		ExpMonth:   card.ExpMonth,
		ExpYear:    card.ExpYear,
		HolderName: card.HolderName,
		CreatedAt:  time.Now(),
	}, nil
}

// DeleteToken deletes a stored card token
func (p *IyzicoProvider) DeleteToken(ctx context.Context, tokenID string) error {
	request := map[string]interface{}{
		"locale":         "tr",
		"conversationId": fmt.Sprintf("delete-token-%d", time.Now().Unix()),
		"cardToken":      tokenID,
	}
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "DELETE", "/cardstorage/card", request, &response)
	if err != nil {
		return err
	}
	
	status, _ := response["status"].(string)
	if status != "success" {
		errorMessage, _ := response["errorMessage"].(string)
		return &integrations.IntegrationError{
			Code:      "TOKEN_DELETE_FAILED",
			Message:   errorMessage,
			Provider:  "iyzico",
			Retryable: false,
			Timestamp: time.Now(),
		}
	}
	
	return nil
}

// CreateSubscription creates a new subscription
func (p *IyzicoProvider) CreateSubscription(ctx context.Context, subscription *SubscriptionRequest) (*SubscriptionResponse, error) {
	// Implement Iyzico subscription API
	// This is a placeholder as Iyzico's subscription API details would need to be implemented
	return nil, &integrations.IntegrationError{
		Code:      "NOT_IMPLEMENTED",
		Message:   "Subscription creation not yet implemented for Iyzico",
		Provider:  "iyzico",
		Retryable: false,
		Timestamp: time.Now(),
	}
}

// CancelSubscription cancels a subscription
func (p *IyzicoProvider) CancelSubscription(ctx context.Context, subscriptionID string) error {
	// Implement Iyzico subscription cancellation
	return &integrations.IntegrationError{
		Code:      "NOT_IMPLEMENTED",
		Message:   "Subscription cancellation not yet implemented for Iyzico",
		Provider:  "iyzico",
		Retryable: false,
		Timestamp: time.Now(),
	}
}

// UpdateSubscription updates a subscription
func (p *IyzicoProvider) UpdateSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) (*SubscriptionResponse, error) {
	// Implement Iyzico subscription update
	return nil, &integrations.IntegrationError{
		Code:      "NOT_IMPLEMENTED",
		Message:   "Subscription update not yet implemented for Iyzico",
		Provider:  "iyzico",
		Retryable: false,
		Timestamp: time.Now(),
	}
}

// GetTransaction gets a single transaction
func (p *IyzicoProvider) GetTransaction(ctx context.Context, transactionID string) (*Transaction, error) {
	// Use GetPaymentStatus and convert to Transaction
	status, err := p.GetPaymentStatus(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	
	return &Transaction{
		ID:        transactionID,
		Type:      "payment",
		Status:    status.Status,
		Amount:    status.Amount,
		Currency:  "TRY",
		CreatedAt: status.UpdatedAt,
	}, nil
}

// ListTransactions lists transactions based on filters
func (p *IyzicoProvider) ListTransactions(ctx context.Context, filters TransactionFilters) ([]*Transaction, error) {
	// Iyzico requires specific reporting API access
	// This is a simplified implementation
	return nil, &integrations.IntegrationError{
		Code:      "NOT_IMPLEMENTED",
		Message:   "Transaction listing requires Iyzico reporting API access",
		Provider:  "iyzico",
		Retryable: false,
		Timestamp: time.Now(),
	}
}

// GetBalance gets the account balance
func (p *IyzicoProvider) GetBalance(ctx context.Context) (*Balance, error) {
	// Iyzico doesn't provide direct balance API
	// This would require marketplace/submerchant API
	return nil, &integrations.IntegrationError{
		Code:      "NOT_SUPPORTED",
		Message:   "Balance retrieval not supported for standard Iyzico integration",
		Provider:  "iyzico",
		Retryable: false,
		Timestamp: time.Now(),
	}
}

// Helper methods

// makeRequest makes an HTTP request to Iyzico API
func (p *IyzicoProvider) makeRequest(ctx context.Context, method, endpoint string, request interface{}, response interface{}) error {
	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	
	// Create HTTP request
	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(requestBody)))
	if err != nil {
		return err
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Generate authorization header
	auth := p.generateAuthHeader(endpoint, string(requestBody))
	req.Header.Set("Authorization", auth)
	
	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "NETWORK_ERROR",
			Message:   err.Error(),
			Provider:  "iyzico",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()
	
	// Update rate limit info
	p.updateRateLimit(resp.Header)
	
	// Parse response
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return &integrations.IntegrationError{
			Code:       "PARSE_ERROR",
			Message:    "Failed to parse Iyzico response",
			Provider:   "iyzico",
			Retryable:  false,
			Timestamp:  time.Now(),
			StatusCode: resp.StatusCode,
		}
	}
	
	return nil
}

// generateAuthHeader generates Iyzico authorization header
func (p *IyzicoProvider) generateAuthHeader(uri string, body string) string {
	randomKey := fmt.Sprintf("%d", time.Now().UnixNano())
	message := p.credentials.APIKey + randomKey + p.credentials.APISecret + body
	
	h := hmac.New(sha256.New, []byte(p.credentials.APISecret))
	h.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	authString := fmt.Sprintf("apiKey:%s&randomKey:%s&signature:%s",
		p.credentials.APIKey, randomKey, signature)
	
	return "IYZWS " + base64.StdEncoding.EncodeToString([]byte(authString))
}

// updateRateLimit updates rate limit information from response headers
func (p *IyzicoProvider) updateRateLimit(headers http.Header) {
	// Iyzico doesn't provide rate limit headers, so we track internally
	p.rateLimit.RequestsRemaining--
	if p.rateLimit.RequestsRemaining <= 0 {
		p.rateLimit.RequestsRemaining = p.rateLimit.RequestsPerMinute
		p.rateLimit.ResetsAt = time.Now().Add(time.Minute)
	}
}

// buildPaymentRequest builds Iyzico payment request from generic payment request
func (p *IyzicoProvider) buildPaymentRequest(payment *PaymentRequest) map[string]interface{} {
	request := map[string]interface{}{
		"locale":         "tr",
		"conversationId": payment.OrderID,
		"price":          fmt.Sprintf("%.2f", payment.Amount),
		"paidPrice":      fmt.Sprintf("%.2f", payment.Amount),
		"currency":       payment.Currency,
		"installment":    payment.Installment,
		"paymentChannel": "WEB",
		"paymentGroup":   "PRODUCT",
	}
	
	// Add payment card
	if payment.PaymentMethod.Type == PaymentMethodTypeCard && payment.PaymentMethod.Card != nil {
		request["paymentCard"] = map[string]string{
			"cardHolderName": payment.PaymentMethod.Card.HolderName,
			"cardNumber":     payment.PaymentMethod.Card.Number,
			"expireMonth":    payment.PaymentMethod.Card.ExpMonth,
			"expireYear":     payment.PaymentMethod.Card.ExpYear,
			"cvc":            payment.PaymentMethod.Card.CVV,
		}
	} else if payment.PaymentMethod.Token != "" {
		request["paymentCard"] = map[string]string{
			"cardToken": payment.PaymentMethod.Token,
		}
	}
	
	// Add buyer information
	request["buyer"] = map[string]string{
		"id":                  payment.CustomerID,
		"name":                payment.BillingAddress.FirstName,
		"surname":             payment.BillingAddress.LastName,
		"gsmNumber":           payment.BillingAddress.Phone,
		"email":               payment.BillingAddress.Email,
		"identityNumber":      "11111111111", // Required by Iyzico
		"registrationAddress": payment.BillingAddress.AddressLine1,
		"ip":                  "127.0.0.1", // Should be actual user IP
		"city":                payment.BillingAddress.City,
		"country":             payment.BillingAddress.Country,
	}
	
	// Add addresses
	request["shippingAddress"] = p.formatAddress(payment.ShippingAddress)
	request["billingAddress"] = p.formatAddress(payment.BillingAddress)
	
	// Add basket items
	var basketItems []map[string]string
	for _, item := range payment.Items {
		basketItems = append(basketItems, map[string]string{
			"id":        item.ID,
			"name":      item.Name,
			"category1": item.Category,
			"itemType":  "PHYSICAL",
			"price":     fmt.Sprintf("%.2f", item.Price*float64(item.Quantity)),
		})
	}
	request["basketItems"] = basketItems
	
	return request
}

// formatAddress formats address for Iyzico
func (p *IyzicoProvider) formatAddress(addr Address) map[string]string {
	return map[string]string{
		"contactName": addr.FirstName + " " + addr.LastName,
		"city":        addr.City,
		"country":     addr.Country,
		"address":     addr.AddressLine1 + " " + addr.AddressLine2,
		"zipCode":     addr.PostalCode,
	}
}

// parsePaymentResponse parses Iyzico response to generic payment response
func (p *IyzicoProvider) parsePaymentResponse(iyzicoResp map[string]interface{}, originalRequest *PaymentRequest) (*PaymentResponse, error) {
	status, _ := iyzicoResp["status"].(string)
	if status != "success" {
		errorMessage, _ := iyzicoResp["errorMessage"].(string)
		errorCode, _ := iyzicoResp["errorCode"].(string)
		
		return nil, &integrations.IntegrationError{
			Code:      errorCode,
			Message:   errorMessage,
			Provider:  "iyzico",
			Retryable: false,
			Timestamp: time.Now(),
		}
	}
	
	paymentStatus := PaymentStatusPending
	if phase, ok := iyzicoResp["phase"].(string); ok && phase == "AUTH" {
		paymentStatus = PaymentStatusSucceeded
	}
	
	response := &PaymentResponse{
		ID:              iyzicoResp["paymentId"].(string),
		Status:          paymentStatus,
		Amount:          iyzicoResp["price"].(float64),
		Currency:        iyzicoResp["currency"].(string),
		TransactionID:   iyzicoResp["paymentId"].(string),
		AuthCode:        iyzicoResp["authCode"].(string),
		CreatedAt:       time.Now(),
	}
	
	if paymentStatus == PaymentStatusSucceeded {
		response.ProcessedAt = time.Now()
	}
	
	// Handle 3D Secure response
	if htmlContent, ok := iyzicoResp["threeDSHtmlContent"].(string); ok && htmlContent != "" {
		response.Status = PaymentStatusPending
		response.Metadata = map[string]interface{}{
			"requires_3d_secure": true,
			"3d_secure_html":     htmlContent,
		}
	}
	
	return response, nil
}