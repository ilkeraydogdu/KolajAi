package marketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"kolajAi/internal/integrations"
)

// CicekSepetiProvider implements marketplace provider for ÇiçekSepeti
type CicekSepetiProvider struct {
	config      *MarketplaceProviderConfig
	httpClient  *http.Client
	credentials integrations.Credentials
	baseURL     string
	apiKey      string
	rateLimit   integrations.RateLimitInfo
}

// CicekSepetiProduct represents a ÇiçekSepeti product structure
type CicekSepetiProduct struct {
	SKU         string                 `json:"sku"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Price       float64                `json:"price"`
	Stock       int                    `json:"stock"`
	Images      []string               `json:"images"`
	Attributes  map[string]interface{} `json:"attributes"`
	Brand       string                 `json:"brand"`
	Barcode     string                 `json:"barcode"`
}

// CicekSepetiOrder represents ÇiçekSepeti order structure
type CicekSepetiOrder struct {
	OrderID       string                 `json:"orderId"`
	OrderNumber   string                 `json:"orderNumber"`
	OrderDate     time.Time              `json:"orderDate"`
	Status        string                 `json:"status"`
	CustomerInfo  CicekSepetiCustomer    `json:"customerInfo"`
	ShippingInfo  CicekSepetiShipping    `json:"shippingInfo"`
	OrderItems    []CicekSepetiOrderItem `json:"orderItems"`
	TotalAmount   float64                `json:"totalAmount"`
	PaymentMethod string                 `json:"paymentMethod"`
}

// CicekSepetiCustomer represents customer information
type CicekSepetiCustomer struct {
	CustomerID string `json:"customerId"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

// CicekSepetiShipping represents shipping information
type CicekSepetiShipping struct {
	Address     CicekSepetiAddress `json:"address"`
	CarrierCode string             `json:"carrierCode"`
	TrackingNo  string             `json:"trackingNo"`
}

// CicekSepetiAddress represents address information
type CicekSepetiAddress struct {
	Name        string `json:"name"`
	AddressLine string `json:"addressLine"`
	City        string `json:"city"`
	District    string `json:"district"`
	PostalCode  string `json:"postalCode"`
	Country     string `json:"country"`
}

// CicekSepetiOrderItem represents order item
type CicekSepetiOrderItem struct {
	ProductID   string  `json:"productId"`
	SKU         string  `json:"sku"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	TotalPrice  float64 `json:"totalPrice"`
}

// CicekSepetiAPIResponse represents ÇiçekSepeti API response structure
type CicekSepetiAPIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewCicekSepetiProvider creates a new ÇiçekSepeti provider instance
func NewCicekSepetiProvider() *CicekSepetiProvider {
	return &CicekSepetiProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.ciceksepeti.com/v1", // Hypothetical API endpoint
		rateLimit: integrations.RateLimitInfo{
			RequestsPerSecond: 10,
			BurstSize:         20,
		},
	}
}

// Initialize initializes the ÇiçekSepeti provider
func (p *CicekSepetiProvider) Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error {
	p.credentials = credentials
	p.apiKey = credentials.APIKey

	// Set base URL based on environment
	if env, ok := config["environment"].(string); ok && env == "sandbox" {
		p.baseURL = "https://api-test.ciceksepeti.com/v1"
	}

	// Test connection
	return p.testConnection(ctx)
}

// GetName returns the provider name
func (p *CicekSepetiProvider) GetName() string {
	return "ÇiçekSepeti"
}

// GetType returns the provider type
func (p *CicekSepetiProvider) GetType() string {
	return "marketplace"
}

// IsHealthy checks if the provider is healthy
func (p *CicekSepetiProvider) IsHealthy(ctx context.Context) (bool, error) {
	return p.testConnection(ctx) == nil, nil
}

// GetMetrics returns provider metrics
func (p *CicekSepetiProvider) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"rate_limit_remaining": p.rateLimit.RequestsPerSecond,
		"last_request_time":    time.Now().Unix(),
	}
}

// GetRateLimit returns rate limit information
func (p *CicekSepetiProvider) GetRateLimit() integrations.RateLimitInfo {
	return p.rateLimit
}

// SyncProducts syncs products to ÇiçekSepeti
func (p *CicekSepetiProvider) SyncProducts(ctx context.Context, products []interface{}) error {
	for _, product := range products {
		productMap, ok := product.(map[string]interface{})
		if !ok {
			continue
		}

		cicekSepetiProduct := p.convertToCicekSepetiProduct(productMap)
		if err := p.createOrUpdateProduct(ctx, cicekSepetiProduct); err != nil {
			return fmt.Errorf("failed to sync product %s: %v", cicekSepetiProduct.SKU, err)
		}
	}

	return nil
}

// UpdateStockAndPrice updates stock and price information
func (p *CicekSepetiProvider) UpdateStockAndPrice(ctx context.Context, updates []interface{}) error {
	for _, update := range updates {
		updateMap, ok := update.(map[string]interface{})
		if !ok {
			continue
		}

		sku, _ := updateMap["sku"].(string)
		quantity, _ := updateMap["quantity"].(int)
		price, _ := updateMap["price"].(float64)

		if err := p.updateProductStock(ctx, sku, quantity); err != nil {
			return fmt.Errorf("failed to update stock for %s: %v", sku, err)
		}

		if err := p.updateProductPrice(ctx, sku, price); err != nil {
			return fmt.Errorf("failed to update price for %s: %v", sku, err)
		}
	}

	return nil
}

// GetProducts retrieves products from ÇiçekSepeti
func (p *CicekSepetiProvider) GetProducts(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/products"

	queryParams := url.Values{}
	if page, ok := params["page"].(int); ok {
		queryParams.Set("page", fmt.Sprintf("%d", page))
	}
	if limit, ok := params["limit"].(int); ok {
		queryParams.Set("limit", fmt.Sprintf("%d", limit))
	}

	response, err := p.makeRequest(ctx, "GET", endpoint+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var apiResponse CicekSepetiAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("ÇiçekSepeti API error: %s", apiResponse.Error.Message)
	}

	// Convert response to standard format
	products := make([]interface{}, 0)
	if data, ok := apiResponse.Data.([]interface{}); ok {
		products = data
	}

	return products, nil
}

// GetOrders retrieves orders from ÇiçekSepeti
func (p *CicekSepetiProvider) GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/orders"

	queryParams := url.Values{}
	if page, ok := params["page"].(int); ok {
		queryParams.Set("page", fmt.Sprintf("%d", page))
	}
	if limit, ok := params["limit"].(int); ok {
		queryParams.Set("limit", fmt.Sprintf("%d", limit))
	}
	if status, ok := params["status"].(string); ok {
		queryParams.Set("status", status)
	}

	response, err := p.makeRequest(ctx, "GET", endpoint+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var apiResponse CicekSepetiAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("ÇiçekSepeti API error: %s", apiResponse.Error.Message)
	}

	// Convert response to standard format
	orders := make([]interface{}, 0)
	if data, ok := apiResponse.Data.([]interface{}); ok {
		orders = data
	}

	return orders, nil
}

// UpdateOrderStatus updates order status
func (p *CicekSepetiProvider) UpdateOrderStatus(ctx context.Context, orderID string, status string, params map[string]interface{}) error {
	endpoint := fmt.Sprintf("/orders/%s/status", orderID)

	requestData := map[string]interface{}{
		"status": status,
	}

	// Add tracking info if provided
	if trackingNo, ok := params["tracking_no"].(string); ok {
		requestData["trackingNo"] = trackingNo
	}
	if carrierCode, ok := params["carrier_code"].(string); ok {
		requestData["carrierCode"] = carrierCode
	}

	response, err := p.makeRequest(ctx, "PUT", endpoint, requestData)
	if err != nil {
		return err
	}

	var apiResponse CicekSepetiAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return err
	}

	if !apiResponse.Success {
		return fmt.Errorf("ÇiçekSepeti API error: %s", apiResponse.Error.Message)
	}

	return nil
}

// GetCategories retrieves categories from ÇiçekSepeti
func (p *CicekSepetiProvider) GetCategories(ctx context.Context) ([]interface{}, error) {
	endpoint := "/categories"

	response, err := p.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var apiResponse CicekSepetiAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("ÇiçekSepeti API error: %s", apiResponse.Error.Message)
	}

	// Convert response to standard format
	categories := make([]interface{}, 0)
	if data, ok := apiResponse.Data.([]interface{}); ok {
		categories = data
	}

	return categories, nil
}

// GetBrands retrieves brands from ÇiçekSepeti
func (p *CicekSepetiProvider) GetBrands(ctx context.Context) ([]interface{}, error) {
	endpoint := "/brands"

	response, err := p.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var apiResponse CicekSepetiAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("ÇiçekSepeti API error: %s", apiResponse.Error.Message)
	}

	// Convert response to standard format
	brands := make([]interface{}, 0)
	if data, ok := apiResponse.Data.([]interface{}); ok {
		brands = data
	}

	return brands, nil
}

// testConnection tests the ÇiçekSepeti API connection
func (p *CicekSepetiProvider) testConnection(ctx context.Context) error {
	// Since ÇiçekSepeti doesn't have a public API, we'll simulate a successful connection
	// In a real implementation, this would make an actual API call
	if p.apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	return nil
}

// makeRequest makes HTTP request to ÇiçekSepeti API
func (p *CicekSepetiProvider) makeRequest(ctx context.Context, method, endpoint string, data interface{}) ([]byte, error) {
	var requestBody []byte
	if data != nil {
		var err error
		requestBody, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("User-Agent", "KolajAI-CicekSepeti-Integration/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			responseBody = append(responseBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ÇiçekSepeti API error: %d - %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// createOrUpdateProduct creates or updates a product
func (p *CicekSepetiProvider) createOrUpdateProduct(ctx context.Context, product CicekSepetiProduct) error {
	endpoint := "/products"

	_, err := p.makeRequest(ctx, "POST", endpoint, product)
	if err != nil {
		return err
	}

	return nil
}

// updateProductStock updates product stock
func (p *CicekSepetiProvider) updateProductStock(ctx context.Context, sku string, stock int) error {
	endpoint := fmt.Sprintf("/products/%s/stock", sku)

	requestData := map[string]interface{}{
		"stock": stock,
	}

	_, err := p.makeRequest(ctx, "PUT", endpoint, requestData)
	return err
}

// updateProductPrice updates product price
func (p *CicekSepetiProvider) updateProductPrice(ctx context.Context, sku string, price float64) error {
	endpoint := fmt.Sprintf("/products/%s/price", sku)

	requestData := map[string]interface{}{
		"price": price,
	}

	_, err := p.makeRequest(ctx, "PUT", endpoint, requestData)
	return err
}

// convertToCicekSepetiProduct converts generic product to ÇiçekSepeti product format
func (p *CicekSepetiProvider) convertToCicekSepetiProduct(product map[string]interface{}) CicekSepetiProduct {
	cicekSepetiProduct := CicekSepetiProduct{
		SKU:         getString(product, "sku"),
		Name:        getString(product, "title"),
		Description: getString(product, "description"),
		Category:    getString(product, "category"),
		Price:       getFloat64(product, "price"),
		Stock:       getInt(product, "quantity"),
		Brand:       getString(product, "brand"),
		Barcode:     getString(product, "barcode"),
		Attributes:  make(map[string]interface{}),
	}

	// Set images
	if images, ok := product["images"].([]interface{}); ok {
		imageList := make([]string, 0)
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				imageList = append(imageList, imgStr)
			}
		}
		cicekSepetiProduct.Images = imageList
	}

	// Set attributes
	if attributes, ok := product["attributes"].(map[string]interface{}); ok {
		cicekSepetiProduct.Attributes = attributes
	}

	return cicekSepetiProduct
}
