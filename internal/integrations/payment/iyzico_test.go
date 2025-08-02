package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kolajAi/internal/integrations"
)

func TestIyzicoProvider_Initialize(t *testing.T) {
	provider := NewIyzicoProvider()
	
	credentials := integrations.Credentials{
		APIKey:    "test-api-key",
		APISecret: "test-api-secret",
	}
	
	config := map[string]interface{}{
		"environment":         "sandbox",
		"enable_3d_secure":    true,
		"enable_installments": true,
		"max_installments":    12,
	}
	
	ctx := context.Background()
	err := provider.Initialize(ctx, credentials, config)
	
	if err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}
	
	if provider.baseURL != "https://sandbox-api.iyzipay.com" {
		t.Errorf("baseURL = %v, want %v", provider.baseURL, "https://sandbox-api.iyzipay.com")
	}
	
	if !provider.config.Enable3DSecure {
		t.Error("Enable3DSecure should be true")
	}
	
	if provider.config.MaxInstallments != 12 {
		t.Errorf("MaxInstallments = %v, want %v", provider.config.MaxInstallments, 12)
	}
}

func TestIyzicoProvider_ProcessPayment(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type header is missing")
		}
		
		// Verify authorization header format
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 10 {
			t.Error("Authorization header is missing or invalid")
		}
		
		// Parse request
		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}
		
		// Verify required fields
		if request["price"] == nil {
			t.Error("Price is missing")
		}
		if request["paidPrice"] == nil {
			t.Error("PaidPrice is missing")
		}
		if request["currency"] == nil {
			t.Error("Currency is missing")
		}
		
		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":               "success",
			"locale":               "tr",
			"systemTime":           time.Now().Unix() * 1000,
			"conversationId":       request["conversationId"],
			"price":                100.0,
			"paidPrice":            100.0,
			"installment":          1,
			"paymentId":            "12345678",
			"fraudStatus":          1,
			"merchantCommissionRate": 2.5,
			"merchantCommissionRateAmount": 2.5,
			"iyziCommissionRateAmount": 0.25,
			"iyziCommissionFee":    0.25,
			"cardType":             "CREDIT_CARD",
			"cardAssociation":      "MASTER_CARD",
			"cardFamily":           "Axess",
			"binNumber":            "552608",
			"lastFourDigits":       "0006",
			"basketId":             request["basketId"],
		})
	}))
	defer server.Close()
	
	provider := &IyzicoProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		credentials: integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
		config: &PaymentProviderConfig{
			APIKey:              "test-key",
			APISecret:           "test-secret",
			Enable3DSecure:      false,
			EnableInstallments:  true,
			MaxInstallments:     12,
			SupportedCurrencies: []string{"TRY", "USD", "EUR"},
		},
	}
	
	paymentRequest := &PaymentRequest{
		Amount:      100.0,
		Currency:    "TRY",
		OrderID:     "ORDER123",
		Description: "Test payment",
		CustomerID:  "CUST123",
		PaymentMethod: PaymentMethod{
			Type: PaymentMethodTypeCard,
			Card: &CardDetails{
				HolderName: "John Doe",
				Number:     "5528790000000008",
				ExpMonth:   "12",
				ExpYear:    "2030",
				CVV:        "123",
			},
		},
		BillingAddress: Address{
			FirstName:    "John",
			LastName:     "Doe",
			AddressLine1: "Test Address",
			City:         "Istanbul",
			Country:      "Turkey",
			PostalCode:   "34000",
			Email:        "john@example.com",
			Phone:        "+905551234567",
		},
		Items: []PaymentItem{
			{
				ID:       "ITEM1",
				Name:     "Test Product",
				Category: "Electronics",
				Price:    100.0,
				Quantity: 1,
			},
		},
	}
	
	ctx := context.Background()
	response, err := provider.CreatePayment(ctx, paymentRequest)
	
	if err != nil {
		t.Errorf("ProcessPayment() error = %v, want nil", err)
	}
	
	if response == nil {
		t.Fatal("Response is nil")
	}
	
	if response.TransactionID != "12345678" {
		t.Errorf("TransactionID = %v, want %v", response.TransactionID, "12345678")
	}
	
	if response.Status != PaymentStatusSucceeded {
		t.Errorf("Status = %v, want %v", response.Status, PaymentStatusSucceeded)
	}
}

func TestIyzicoProvider_Create3DSecurePayment(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 3D Secure HTML form
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "success",
			"locale":         "tr",
			"systemTime":     time.Now().Unix() * 1000,
			"conversationId": "test-conversation",
			"threeDSHtmlContent": base64Encode(`
				<html>
				<body>
				<form id="iyzico-3ds-form" action="https://acs.bank.com" method="post">
					<input type="hidden" name="PaReq" value="test-pareq">
					<input type="hidden" name="TermUrl" value="https://callback.example.com">
					<input type="hidden" name="MD" value="test-md">
				</form>
				<script>document.getElementById('iyzico-3ds-form').submit();</script>
				</body>
				</html>
			`),
		})
	}))
	defer server.Close()
	
	provider := &IyzicoProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		credentials: integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
		config: &PaymentProviderConfig{
			APIKey:         "test-key",
			APISecret:      "test-secret",
			Enable3DSecure: true,
		},
	}
	
	paymentRequest := &PaymentRequest{
		Amount:      100.0,
		Currency:    "TRY",
		OrderID:     "ORDER123",
		CallbackURL: "https://callback.example.com",
		PaymentMethod: PaymentMethod{
			Type: PaymentMethodTypeCard,
			Card: &CardDetails{
				HolderName: "John Doe",
				Number:     "5528790000000008",
				ExpMonth:   "12",
				ExpYear:    "2030",
				CVV:        "123",
			},
		},
		Enable3DSecure: true,
	}
	
	ctx := context.Background()
	response, err := provider.Initialize3DSecure(ctx, paymentRequest)
	
	if err != nil {
		t.Errorf("Create3DSecurePayment() error = %v, want nil", err)
	}
	
	if response == nil {
		t.Fatal("Response is nil")
	}
	
		if response.Status != "requires_3d_secure" {
		t.Error("Status should be requires_3d_secure")
	}

	if response.HTMLContent == "" {
		t.Error("HTMLContent should not be empty")
	}
}

func TestIyzicoProvider_RefundPayment(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return refund response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "success",
			"locale":         "tr",
			"systemTime":     time.Now().Unix() * 1000,
			"conversationId": "test-conversation",
			"paymentId":      "12345678",
			"price":          50.0,
			"currency":       "TRY",
		})
	}))
	defer server.Close()
	
	provider := &IyzicoProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		credentials: integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
		config: &PaymentProviderConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	}
	
	ctx := context.Background()
	response, err := provider.RefundPayment(ctx, "12345678", 50.0)
	
	if err != nil {
		t.Errorf("RefundPayment() error = %v, want nil", err)
	}
	
	if response == nil {
		t.Fatal("Response is nil")
	}
	
	if response.ID == "" {
		t.Error("RefundID should not be empty")
	}
	
	if response.Status != string(PaymentStatusRefunded) {
		t.Errorf("Status = %v, want %v", response.Status, PaymentStatusRefunded)
	}
}

func TestIyzicoProvider_GetPaymentStatus(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return payment inquiry response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "success",
			"locale":         "tr",
			"systemTime":     time.Now().Unix() * 1000,
			"conversationId": "test-conversation",
			"paymentId":      "12345678",
			"price":          100.0,
			"paidPrice":      100.0,
			"currency":       "TRY",
			"installment":    1,
			"paymentStatus":  "SUCCESS",
			"fraudStatus":    1,
		})
	}))
	defer server.Close()
	
	provider := &IyzicoProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		credentials: integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
		config: &PaymentProviderConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	}
	
	ctx := context.Background()
	status, err := provider.GetPaymentStatus(ctx, "12345678")
	
	if err != nil {
		t.Errorf("GetPaymentStatus() error = %v, want nil", err)
	}
	
	if status == nil {
		t.Fatal("Status is nil")
	}
	
	if status.ID != "12345678" {
		t.Errorf("TransactionID = %v, want %v", status.ID, "12345678")
	}
	
	if status.Status != PaymentStatusSucceeded {
		t.Errorf("Status = %v, want %v", status.Status, PaymentStatusSucceeded)
	}
	
	if status.Amount != 100.0 {
		t.Errorf("Amount = %v, want %v", status.Amount, 100.0)
	}
}

func TestIyzicoProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		expectedError string
	}{
		{
			name: "Invalid API Key",
			responseBody: `{
				"status": "failure",
				"errorCode": "12",
				"errorMessage": "Invalid api key",
				"locale": "tr",
				"systemTime": 1234567890,
				"conversationId": "test"
			}`,
			expectedError: "Invalid api key",
		},
		{
			name: "Insufficient Balance",
			responseBody: `{
				"status": "failure",
				"errorCode": "10051",
				"errorMessage": "Insufficient balance",
				"locale": "tr",
				"systemTime": 1234567890,
				"conversationId": "test"
			}`,
			expectedError: "Insufficient balance",
		},
		{
			name: "Invalid Card",
			responseBody: `{
				"status": "failure",
				"errorCode": "10054",
				"errorMessage": "Invalid card number",
				"locale": "tr",
				"systemTime": 1234567890,
				"conversationId": "test"
			}`,
			expectedError: "Invalid card number",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()
			
			provider := &IyzicoProvider{
				httpClient: http.DefaultClient,
				baseURL:    server.URL,
				credentials: integrations.Credentials{
					APIKey:    "test-key",
					APISecret: "test-secret",
				},
				config: &PaymentProviderConfig{
					APIKey:    "test-key",
					APISecret: "test-secret",
				},
			}
			
			paymentRequest := &PaymentRequest{
				Amount:   100.0,
				Currency: "TRY",
				OrderID:  "ORDER123",
			}
			
			ctx := context.Background()
			_, err := provider.CreatePayment(ctx, paymentRequest)
			
			if err == nil {
				t.Error("Expected error, got nil")
			}
			
			if !contains(err.Error(), tt.expectedError) {
				t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.expectedError)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Helper function to encode base64
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}