package marketplace

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kolajAi/internal/integrations"
)

// HepsiburadaProvider implements marketplace provider for Hepsiburada
type HepsiburadaProvider struct {
	config      *MarketplaceProviderConfig
	httpClient  *http.Client
	credentials integrations.Credentials
	baseURL     string
	merchantID  string
	rateLimit   integrations.RateLimitInfo
}

// HepsiburadaProduct represents a Hepsiburada product structure
type HepsiburadaProduct struct {
	MerchantSKU      string                    `json:"merchantSku"`
	HepsiburadaSKU   string                    `json:"hepsiburadaSku"`
	ProductName      string                    `json:"productName"`
	ProductNameEn    string                    `json:"productNameEn,omitempty"`
	Description      string                    `json:"description"`
	DescriptionEn    string                    `json:"descriptionEn,omitempty"`
	CategoryName     string                    `json:"categoryName"`
	BrandName        string                    `json:"brandName"`
	Barcode          string                    `json:"barcode"`
	Price            float64                   `json:"price"`
	ListPrice        float64                   `json:"listPrice"`
	CurrencyType     string                    `json:"currencyType"`
	AvailableStock   int                       `json:"availableStock"`
	DispatchTime     int                       `json:"dispatchTime"`
	CargoCompanyName string                    `json:"cargoCompanyName"`
	Images           []HepsiburadaImage        `json:"images"`
	Attributes       []HepsiburadaAttribute    `json:"attributes"`
	Variants         []HepsiburadaVariant      `json:"variants,omitempty"`
	Dimensions       HepsiburadaDimensions     `json:"dimensions"`
	Status           string                    `json:"status"`
}

// HepsiburadaImage represents product image
type HepsiburadaImage struct {
	URL string `json:"url"`
}

// HepsiburadaAttribute represents product attribute
type HepsiburadaAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HepsiburadaVariant represents product variant
type HepsiburadaVariant struct {
	MerchantSKU    string  `json:"merchantSku"`
	HepsiburadaSKU string  `json:"hepsiburadaSku"`
	Price          float64 `json:"price"`
	ListPrice      float64 `json:"listPrice"`
	AvailableStock int     `json:"availableStock"`
	Attributes     []HepsiburadaAttribute `json:"attributes"`
}

// HepsiburadaDimensions represents product dimensions
type HepsiburadaDimensions struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Length float64 `json:"length"`
	Weight float64 `json:"weight"`
}

// HepsiburadaOrder represents Hepsiburada order structure
type HepsiburadaOrder struct {
	OrderNumber      string                   `json:"orderNumber"`
	OrderDate        time.Time                `json:"orderDate"`
	Status           string                   `json:"status"`
	CustomerName     string                   `json:"customerName"`
	CustomerEmail    string                   `json:"customerEmail"`
	CustomerPhone    string                   `json:"customerPhone"`
	TotalAmount      float64                  `json:"totalAmount"`
	TaxAmount        float64                  `json:"taxAmount"`
	ShippingAmount   float64                  `json:"shippingAmount"`
	Currency         string                   `json:"currency"`
	PaymentType      string                   `json:"paymentType"`
	CargoCompany     string                   `json:"cargoCompany"`
	TrackingNumber   string                   `json:"trackingNumber"`
	BillingAddress   HepsiburadaAddress       `json:"billingAddress"`
	ShippingAddress  HepsiburadaAddress       `json:"shippingAddress"`
	Items            []HepsiburadaOrderItem   `json:"items"`
}

// HepsiburadaOrderItem represents order line item
type HepsiburadaOrderItem struct {
	LineItemId      string  `json:"lineItemId"`
	MerchantSKU     string  `json:"merchantSku"`
	HepsiburadaSKU  string  `json:"hepsiburadaSku"`
	ProductName     string  `json:"productName"`
	Quantity        int     `json:"quantity"`
	Price           float64 `json:"price"`
	TotalPrice      float64 `json:"totalPrice"`
	VatRate         int     `json:"vatRate"`
	VatAmount       float64 `json:"vatAmount"`
	Commission      float64 `json:"commission"`
	CommissionRate  float64 `json:"commissionRate"`
	Status          string  `json:"status"`
}

// HepsiburadaAddress represents address structure
type HepsiburadaAddress struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Address     string `json:"address"`
	City        string `json:"city"`
	District    string `json:"district"`
	PostalCode  string `json:"postalCode"`
	Country     string `json:"country"`
	Phone       string `json:"phone"`
}

// HepsiburadaStockUpdate represents stock update request
type HepsiburadaStockUpdate struct {
	Items []HepsiburadaStockItem `json:"items"`
}

// HepsiburadaStockItem represents individual stock item
type HepsiburadaStockItem struct {
	MerchantSKU    string `json:"merchantSku"`
	AvailableStock int    `json:"availableStock"`
}

// HepsiburadaPriceUpdate represents price update request
type HepsiburadaPriceUpdate struct {
	Items []HepsiburadaPriceItem `json:"items"`
}

// HepsiburadaPriceItem represents individual price item
type HepsiburadaPriceItem struct {
	MerchantSKU string  `json:"merchantSku"`
	Price       float64 `json:"price"`
	ListPrice   float64 `json:"listPrice"`
}

// NewHepsiburadaProvider creates a new Hepsiburada marketplace provider
func NewHepsiburadaProvider() *HepsiburadaProvider {
	return &HepsiburadaProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: integrations.RateLimitInfo{
			RequestsPerMinute: 100, // Hepsiburada API limit
			RequestsRemaining: 100,
			ResetsAt:          time.Now().Add(time.Minute),
		},
	}
}

// Initialize sets up the Hepsiburada provider
func (p *HepsiburadaProvider) Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error {
	p.credentials = credentials
	
	// Set base URL based on environment
	environment, _ := config["environment"].(string)
	if environment == "production" {
		p.baseURL = "https://mpop.hepsiburada.com"
	} else {
		p.baseURL = "https://stageapi.hepsiburada.com"
	}
	
	// Get merchant ID from config
	if merchantID, ok := config["merchant_id"].(string); ok {
		p.merchantID = merchantID
	} else {
		return fmt.Errorf("merchant_id is required for Hepsiburada integration")
	}
	
	// Initialize configuration
	p.config = &MarketplaceProviderConfig{
		APIKey:              credentials.APIKey,
		APISecret:           credentials.APISecret,
		Environment:         environment,
		SupportedCurrencies: []string{"TRY"},
		SupportedCountries:  []string{"TR"},
		Timeout:             30 * time.Second,
		RateLimit:           100,
	}
	
	return nil
}

// HealthCheck verifies the Hepsiburada integration is working
func (p *HepsiburadaProvider) HealthCheck(ctx context.Context) error {
	// Test API connectivity by getting merchant info
	endpoint := "/api/merchants/v1/merchant"
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "HEALTH_CHECK_FAILED",
			Message:   "Failed to connect to Hepsiburada API",
			Provider:  "hepsiburada",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	
	return nil
}

// GetCapabilities returns the capabilities of this integration
func (p *HepsiburadaProvider) GetCapabilities() []string {
	return []string{
		"product_sync",
		"order_sync",
		"inventory_sync",
		"price_sync",
		"category_mapping",
		"bulk_operations",
		"real_time_notifications",
		"webhook_support",
		"variant_support",
	}
}

// GetRateLimit returns current rate limit information
func (p *HepsiburadaProvider) GetRateLimit() integrations.RateLimitInfo {
	return p.rateLimit
}

// Close cleans up any resources
func (p *HepsiburadaProvider) Close() error {
	return nil
}

// SyncProducts syncs products to Hepsiburada
func (p *HepsiburadaProvider) SyncProducts(ctx context.Context, products []interface{}) error {
	hepsiburadaProducts := make([]HepsiburadaProduct, 0, len(products))
	
	for _, product := range products {
		hepsiburadaProduct, err := p.convertToHepsiburadaProduct(product)
		if err != nil {
			continue // Skip invalid products
		}
		hepsiburadaProducts = append(hepsiburadaProducts, hepsiburadaProduct)
	}
	
	// Send products in batches of 100 (Hepsiburada limit)
	batchSize := 100
	for i := 0; i < len(hepsiburadaProducts); i += batchSize {
		end := i + batchSize
		if end > len(hepsiburadaProducts) {
			end = len(hepsiburadaProducts)
		}
		
		batch := hepsiburadaProducts[i:end]
		err := p.sendProductBatch(ctx, batch)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// GetProducts retrieves products from Hepsiburada
func (p *HepsiburadaProvider) GetProducts(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/api/products/v1/products"
	
	// Add query parameters
	queryParams := make([]string, 0)
	if offset, ok := params["offset"].(int); ok {
		queryParams = append(queryParams, "offset="+strconv.Itoa(offset))
	}
	if limit, ok := params["limit"].(int); ok {
		queryParams = append(queryParams, "limit="+strconv.Itoa(limit))
	}
	
	if len(queryParams) > 0 {
		endpoint += "?" + strings.Join(queryParams, "&")
	}
	
	var response struct {
		Products []HepsiburadaProduct `json:"products"`
		TotalCount int `json:"totalCount"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic interface
	products := make([]interface{}, len(response.Products))
	for i, product := range response.Products {
		products[i] = product
	}
	
	return products, nil
}

// GetOrders retrieves orders from Hepsiburada
func (p *HepsiburadaProvider) GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/api/orders/v1/orders"
	
	// Add query parameters
	queryParams := make([]string, 0)
	if startDate, ok := params["start_date"].(string); ok {
		queryParams = append(queryParams, "startDate="+startDate)
	}
	if endDate, ok := params["end_date"].(string); ok {
		queryParams = append(queryParams, "endDate="+endDate)
	}
	if status, ok := params["status"].(string); ok {
		queryParams = append(queryParams, "status="+status)
	}
	
	if len(queryParams) > 0 {
		endpoint += "?" + strings.Join(queryParams, "&")
	}
	
	var response struct {
		Orders []HepsiburadaOrder `json:"orders"`
		TotalCount int `json:"totalCount"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic interface
	orders := make([]interface{}, len(response.Orders))
	for i, order := range response.Orders {
		orders[i] = order
	}
	
	return orders, nil
}

// UpdateStockAndPrice updates stock and price for products
func (p *HepsiburadaProvider) UpdateStockAndPrice(ctx context.Context, updates []interface{}) error {
	stockItems := make([]HepsiburadaStockItem, 0)
	priceItems := make([]HepsiburadaPriceItem, 0)
	
	for _, update := range updates {
		stockItem, priceItem, err := p.convertToStockPriceItems(update)
		if err != nil {
			continue // Skip invalid items
		}
		stockItems = append(stockItems, stockItem)
		priceItems = append(priceItems, priceItem)
	}
	
	// Update stock
	if len(stockItems) > 0 {
		stockRequest := HepsiburadaStockUpdate{Items: stockItems}
		endpoint := "/api/products/v1/stocks"
		var response map[string]interface{}
		if err := p.makeRequest(ctx, "PUT", endpoint, stockRequest, &response); err != nil {
			return err
		}
	}
	
	// Update prices
	if len(priceItems) > 0 {
		priceRequest := HepsiburadaPriceUpdate{Items: priceItems}
		endpoint := "/api/products/v1/prices"
		var response map[string]interface{}
		if err := p.makeRequest(ctx, "PUT", endpoint, priceRequest, &response); err != nil {
			return err
		}
	}
	
	return nil
}

// UpdateOrderStatus updates order status
func (p *HepsiburadaProvider) UpdateOrderStatus(ctx context.Context, orderID string, status string, params map[string]interface{}) error {
	endpoint := fmt.Sprintf("/api/orders/v1/orders/%s/status", orderID)
	
	request := map[string]interface{}{
		"status": status,
	}
	
	// Add additional parameters
	if trackingNumber, ok := params["tracking_number"].(string); ok {
		request["trackingNumber"] = trackingNumber
	}
	if cargoCompany, ok := params["cargo_company"].(string); ok {
		request["cargoCompany"] = cargoCompany
	}
	
	var response map[string]interface{}
	return p.makeRequest(ctx, "PUT", endpoint, request, &response)
}

// GetCategories retrieves categories from Hepsiburada
func (p *HepsiburadaProvider) GetCategories(ctx context.Context) ([]interface{}, error) {
	endpoint := "/api/categories/v1/categories"
	
	var response struct {
		Categories []Category `json:"categories"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic interface
	categories := make([]interface{}, len(response.Categories))
	for i, category := range response.Categories {
		categories[i] = category
	}
	
	return categories, nil
}

// GetBrands retrieves brands from Hepsiburada
func (p *HepsiburadaProvider) GetBrands(ctx context.Context) ([]interface{}, error) {
	endpoint := "/api/brands/v1/brands"
	
	var response struct {
		Brands []Brand `json:"brands"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic interface
	brands := make([]interface{}, len(response.Brands))
	for i, brand := range response.Brands {
		brands[i] = brand
	}
	
	return brands, nil
}

// ProcessWebhook processes incoming webhooks from Hepsiburada
func (p *HepsiburadaProvider) ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) error {
	// Validate webhook signature
	signature := headers["X-Hepsiburada-Signature"]
	if !p.validateWebhookSignature(payload, signature) {
		return fmt.Errorf("invalid webhook signature")
	}
	
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	
	event.Provider = "hepsiburada"
	event.Timestamp = time.Now()
	
	// Process different event types
	switch event.Type {
	case "order.created":
		// Handle new order
		return p.handleOrderCreatedEvent(ctx, event)
	case "order.updated":
		// Handle order update
		return p.handleOrderUpdatedEvent(ctx, event)
	case "product.updated":
		// Handle product update
		return p.handleProductUpdatedEvent(ctx, event)
	default:
		// Unknown event type, log and ignore
		return nil
	}
}

// makeRequest makes an HTTP request to Hepsiburada API
func (p *HepsiburadaProvider) makeRequest(ctx context.Context, method, endpoint string, request interface{}, response interface{}) error {
	var body []byte
	var err error
	
	if request != nil {
		body, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}
	
	// Create HTTP request
	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KolajAI-Integration/1.0")
	
	// Generate authorization header
	auth := p.generateAuthHeader()
	req.Header.Set("Authorization", auth)
	
	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "NETWORK_ERROR",
			Message:   err.Error(),
			Provider:  "hepsiburada",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()
	
	// Update rate limit info
	p.updateRateLimit(resp.Header)
	
	// Check for API errors
	if resp.StatusCode >= 400 {
		return &integrations.IntegrationError{
			Code:       "API_ERROR",
			Message:    fmt.Sprintf("Hepsiburada API returned status %d", resp.StatusCode),
			Provider:   "hepsiburada",
			Retryable:  resp.StatusCode >= 500,
			Timestamp:  time.Now(),
			StatusCode: resp.StatusCode,
		}
	}
	
	// Parse response
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return &integrations.IntegrationError{
				Code:       "PARSE_ERROR",
				Message:    "Failed to parse Hepsiburada response",
				Provider:   "hepsiburada",
				Retryable:  false,
				Timestamp:  time.Now(),
				StatusCode: resp.StatusCode,
			}
		}
	}
	
	return nil
}

// generateAuthHeader generates Hepsiburada authorization header
func (p *HepsiburadaProvider) generateAuthHeader() string {
	// Hepsiburada uses Basic Auth with username and password
	auth := p.credentials.APIKey + ":" + p.credentials.APISecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// updateRateLimit updates rate limit information from response headers
func (p *HepsiburadaProvider) updateRateLimit(headers http.Header) {
	// Update rate limit based on response headers if available
	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			p.rateLimit.RequestsPerMinute = val
		}
	}
	
	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			p.rateLimit.RequestsRemaining = val
		}
	}
	
	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			p.rateLimit.ResetsAt = time.Unix(val, 0)
		}
	}
}

// validateWebhookSignature validates webhook signature
func (p *HepsiburadaProvider) validateWebhookSignature(payload []byte, signature string) bool {
	if p.config.WebhookSecret == "" {
		return true // Skip validation if no secret configured
	}
	
	mac := hmac.New(sha256.New, []byte(p.config.WebhookSecret))
	mac.Write(payload)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	
	return signature == expectedSignature
}

// Helper methods for data conversion and event handling

func (p *HepsiburadaProvider) convertToHepsiburadaProduct(product interface{}) (HepsiburadaProduct, error) {
	// Convert generic product to Hepsiburada format
	hepsiburadaProduct := HepsiburadaProduct{
		CurrencyType:   "TRY",
		DispatchTime:   1, // Default dispatch time
		Status:         "Active",
	}
	
	// Add conversion logic based on your product structure
	// This is a placeholder implementation
	
	return hepsiburadaProduct, nil
}

func (p *HepsiburadaProvider) convertToStockPriceItems(update interface{}) (HepsiburadaStockItem, HepsiburadaPriceItem, error) {
	// Convert generic update to Hepsiburada format
	stockItem := HepsiburadaStockItem{}
	priceItem := HepsiburadaPriceItem{}
	
	// Add conversion logic based on your update structure
	// This is a placeholder implementation
	
	return stockItem, priceItem, nil
}

func (p *HepsiburadaProvider) sendProductBatch(ctx context.Context, products []HepsiburadaProduct) error {
	endpoint := "/api/products/v1/products"
	
	request := map[string]interface{}{
		"products": products,
	}
	
	var response map[string]interface{}
	return p.makeRequest(ctx, "POST", endpoint, request, &response)
}

func (p *HepsiburadaProvider) handleOrderCreatedEvent(ctx context.Context, event WebhookEvent) error {
	// Handle order created webhook
	// Implementation depends on your business logic
	return nil
}

func (p *HepsiburadaProvider) handleOrderUpdatedEvent(ctx context.Context, event WebhookEvent) error {
	// Handle order updated webhook
	// Implementation depends on your business logic
	return nil
}

func (p *HepsiburadaProvider) handleProductUpdatedEvent(ctx context.Context, event WebhookEvent) error {
	// Handle product updated webhook
	// Implementation depends on your business logic
	return nil
}