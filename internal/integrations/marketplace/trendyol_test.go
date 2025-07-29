package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kolajAi/internal/integrations"
)

// MockHTTPClient implements a mock HTTP client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestTrendyolProvider_Initialize(t *testing.T) {
	provider := NewTrendyolProvider()
	
	credentials := integrations.Credentials{
		APIKey:    "test-api-key",
		APISecret: "test-api-secret",
	}
	
	config := map[string]interface{}{
		"supplier_id": "12345",
		"environment": "sandbox",
	}
	
	ctx := context.Background()
	err := provider.Initialize(ctx, credentials, config)
	
	if err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}
	
	if provider.supplierID != "12345" {
		t.Errorf("supplierID = %v, want %v", provider.supplierID, "12345")
	}
	
	if provider.baseURL != "https://sandbox-api.trendyol.com" {
		t.Errorf("baseURL = %v, want %v", provider.baseURL, "https://sandbox-api.trendyol.com")
	}
}

func TestTrendyolProvider_HealthCheck(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header is missing")
		}
		
		// Verify user agent
		userAgent := r.Header.Get("User-Agent")
		expectedUserAgent := "KolajAI-Trendyol/1.0"
		if userAgent != expectedUserAgent {
			t.Errorf("User-Agent = %v, want %v", userAgent, expectedUserAgent)
		}
		
		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()
	
	provider := &TrendyolProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		tempCredentials: &integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	}
	
	ctx := context.Background()
	err := provider.HealthCheck(ctx)
	
	if err != nil {
		t.Errorf("HealthCheck() error = %v, want nil", err)
	}
}

func TestTrendyolProvider_SyncProducts(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("Method = %v, want %v", r.Method, http.MethodPost)
		}
		
		// Verify path
		expectedPath := "/sapigw/suppliers/12345/v2/products"
		if r.URL.Path != expectedPath {
			t.Errorf("Path = %v, want %v", r.URL.Path, expectedPath)
		}
		
		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		
		// Verify request structure
		if items, ok := requestBody["items"].([]interface{}); ok {
			if len(items) != 1 {
				t.Errorf("Items length = %v, want %v", len(items), 1)
			}
		} else {
			t.Error("Items field is missing or invalid")
		}
		
		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"batchRequestId": "test-batch-123",
			"items": []map[string]interface{}{
				{
					"requestItem": map[string]interface{}{
						"barcode": "TEST123",
					},
					"status":  "SUCCESS",
					"message": "Product created successfully",
				},
			},
		})
	}))
	defer server.Close()
	
	provider := &TrendyolProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		supplierID: "12345",
		tempCredentials: &integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	}
	
	products := []interface{}{
		map[string]interface{}{
			"barcode":     "TEST123",
			"title":       "Test Product",
			"brandId":     123,
			"categoryId":  456,
			"quantity":    10,
			"stockCode":   "STOCK123",
			"listPrice":   100.0,
			"salePrice":   90.0,
			"description": "Test product description",
		},
	}
	
	ctx := context.Background()
	err := provider.SyncProducts(ctx, products)
	
	if err != nil {
		t.Errorf("SyncProducts() error = %v, want nil", err)
	}
}

func TestTrendyolProvider_GetOrders(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("Method = %v, want %v", r.Method, http.MethodGet)
		}
		
		// Return orders response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"orderNumber":        "ORDER123",
					"orderDate":          time.Now().Unix() * 1000,
					"status":             "Created",
					"customerFirstName":  "John",
					"customerLastName":   "Doe",
					"customerEmail":      "john@example.com",
					"grossAmount":        100.0,
					"totalDiscount":      10.0,
					"lines": []map[string]interface{}{
						{
							"lineId":       1,
							"productName":  "Test Product",
							"productCode":  "PROD123",
							"merchantSku":  "SKU123",
							"barcode":      "BAR123",
							"quantity":     2,
							"price":        45.0,
						},
					},
				},
			},
			"totalElements": 1,
			"totalPages":    1,
			"page":          0,
			"size":          50,
		})
	}))
	defer server.Close()
	
	provider := &TrendyolProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		supplierID: "12345",
		tempCredentials: &integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	}
	
	ctx := context.Background()
	orders, err := provider.GetOrders(ctx, map[string]interface{}{
		"status": "Created",
		"size":   50,
	})
	
	if err != nil {
		t.Errorf("GetOrders() error = %v, want nil", err)
	}
	
	if len(orders) != 1 {
		t.Errorf("Orders length = %v, want %v", len(orders), 1)
	}
}

func TestTrendyolProvider_UpdateStock(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("Method = %v, want %v", r.Method, http.MethodPost)
		}
		
		// Verify path
		expectedPath := "/sapigw/suppliers/12345/products/price-and-inventory"
		if r.URL.Path != expectedPath {
			t.Errorf("Path = %v, want %v", r.URL.Path, expectedPath)
		}
		
		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"batchRequestId": "test-batch-456",
			"items": []map[string]interface{}{
				{
					"requestItem": map[string]interface{}{
						"barcode": "TEST123",
					},
					"status": "SUCCESS",
				},
			},
		})
	}))
	defer server.Close()
	
	provider := &TrendyolProvider{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
		supplierID: "12345",
		tempCredentials: &integrations.Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	}
	
	ctx := context.Background()
	err := provider.UpdateStock(ctx, "TEST123", 50)
	
	if err != nil {
		t.Errorf("UpdateStock() error = %v, want nil", err)
	}
}

func TestTrendyolProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  string
	}{
		{
			name:          "Unauthorized",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"errors": [{"message": "Invalid credentials"}]}`,
			expectedError: "authentication failed",
		},
		{
			name:          "Rate Limit",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{"errors": [{"message": "Rate limit exceeded"}]}`,
			expectedError: "rate limit exceeded",
		},
		{
			name:          "Server Error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"errors": [{"message": "Internal server error"}]}`,
			expectedError: "server error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()
			
			provider := &TrendyolProvider{
				httpClient: http.DefaultClient,
				baseURL:    server.URL,
				tempCredentials: &integrations.Credentials{
					APIKey:    "test-key",
					APISecret: "test-secret",
				},
			}
			
			ctx := context.Background()
			err := provider.HealthCheck(ctx)
			
			if err == nil {
				t.Error("Expected error, got nil")
			}
			
			integrationErr, ok := err.(*integrations.IntegrationError)
			if !ok {
				t.Errorf("Error type = %T, want *integrations.IntegrationError", err)
			}
			
			if integrationErr != nil && !contains(integrationErr.Message, tt.expectedError) {
				t.Errorf("Error message = %v, want to contain %v", integrationErr.Message, tt.expectedError)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}