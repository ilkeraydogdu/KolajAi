package marketplace

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kolajAi/internal/integrations"
)

// TrendyolProvider implements marketplace provider for Trendyol
type TrendyolProvider struct {
	config      *MarketplaceProviderConfig
	httpClient  *http.Client
	credentials integrations.Credentials
	baseURL     string
	supplierID  string
	rateLimit   integrations.RateLimitInfo
}

// TrendyolProduct represents a Trendyol product structure
type TrendyolProduct struct {
	Barcode          string                 `json:"barcode"`
	Title            string                 `json:"title"`
	ProductMainID    string                 `json:"productMainId"`
	BrandID          int                    `json:"brandId"`
	CategoryID       int                    `json:"categoryId"`
	Quantity         int                    `json:"quantity"`
	StockCode        string                 `json:"stockCode"`
	DimensionalWeight float64               `json:"dimensionalWeight"`
	Description      string                 `json:"description"`
	CurrencyType     string                 `json:"currencyType"`
	ListPrice        float64                `json:"listPrice"`
	SalePrice        float64                `json:"salePrice"`
	VatRate          int                    `json:"vatRate"`
	CargoCompanyID   int                    `json:"cargoCompanyId"`
	Images           []TrendyolImage        `json:"images"`
	Attributes       []TrendyolAttribute    `json:"attributes"`
}

// TrendyolImage represents product image
type TrendyolImage struct {
	URL string `json:"url"`
}

// TrendyolAttribute represents product attribute
type TrendyolAttribute struct {
	AttributeID           int    `json:"attributeId"`
	AttributeValueID      int    `json:"attributeValueId"`
	CustomAttributeValue  string `json:"customAttributeValue,omitempty"`
}

// TrendyolOrder represents Trendyol order structure
type TrendyolOrder struct {
	OrderNumber    string                `json:"orderNumber"`
	OrderDate      time.Time             `json:"orderDate"`
	Status         string                `json:"status"`
	CustomerID     int                   `json:"customerId"`
	CustomerName   string                `json:"customerFirstName"`
	CustomerSurname string               `json:"customerLastName"`
	CustomerEmail  string                `json:"customerEmail"`
	GrossAmount    float64               `json:"grossAmount"`
	TotalDiscount  float64               `json:"totalDiscount"`
	TotalTyDiscount float64              `json:"totalTyDiscount"`
	TaxNumber      string                `json:"taxNumber"`
	InvoiceAddress TrendyolAddress       `json:"invoiceAddress"`
	ShippingAddress TrendyolAddress      `json:"shippingAddress"`
	Lines          []TrendyolOrderLine   `json:"lines"`
}

// TrendyolOrderLine represents order line item
type TrendyolOrderLine struct {
	LineID         int     `json:"lineId"`
	ProductName    string  `json:"productName"`
	ProductCode    string  `json:"productCode"`
	MerchantSKU    string  `json:"merchantSku"`
	Barcode        string  `json:"barcode"`
	Quantity       int     `json:"quantity"`
	Price          float64 `json:"price"`
	VatBaseAmount  float64 `json:"vatBaseAmount"`
	VatAmount      float64 `json:"vatAmount"`
	Discount       float64 `json:"discount"`
	TyDiscount     float64 `json:"tyDiscount"`
	ProductSize    string  `json:"productSize"`
	ProductColor   string  `json:"productColor"`
}

// TrendyolAddress represents address structure
type TrendyolAddress struct {
	ID          int    `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Company     string `json:"company"`
	Address1    string `json:"address1"`
	Address2    string `json:"address2"`
	City        string `json:"city"`
	District    string `json:"district"`
	PostalCode  string `json:"postalCode"`
	CountryCode string `json:"countryCode"`
	Phone       string `json:"phone"`
}

// TrendyolStockPriceUpdate represents stock and price update request
type TrendyolStockPriceUpdate struct {
	Items []TrendyolStockPriceItem `json:"items"`
}

// TrendyolStockPriceItem represents individual stock/price item
type TrendyolStockPriceItem struct {
	Barcode   string  `json:"barcode"`
	Quantity  int     `json:"quantity"`
	SalePrice float64 `json:"salePrice"`
	ListPrice float64 `json:"listPrice"`
}

// NewTrendyolProvider creates a new Trendyol marketplace provider
func NewTrendyolProvider() *TrendyolProvider {
	return &TrendyolProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: integrations.RateLimitInfo{
			RequestsPerMinute: 60, // Trendyol API limit
			RequestsRemaining: 60,
			ResetsAt:          time.Now().Add(time.Minute),
		},
	}
}

// Initialize sets up the Trendyol provider
func (p *TrendyolProvider) Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error {
	p.credentials = credentials
	
	// Set base URL based on environment
	environment, _ := config["environment"].(string)
	if environment == "production" {
		p.baseURL = "https://api.trendyol.com"
	} else {
		p.baseURL = "https://stageapi.trendyol.com"
	}
	
	// Get supplier ID from config
	if supplierID, ok := config["supplier_id"].(string); ok {
		p.supplierID = supplierID
	} else {
		return fmt.Errorf("supplier_id is required for Trendyol integration")
	}
	
	// Initialize configuration
	p.config = &MarketplaceProviderConfig{
		APIKey:              credentials.APIKey,
		APISecret:           credentials.APISecret,
		Environment:         environment,
		SupportedCurrencies: []string{"TRY"},
		SupportedCountries:  []string{"TR"},
		Timeout:             30 * time.Second,
		RateLimit:           60,
	}
	
	return nil
}

// HealthCheck verifies the Trendyol integration is working
func (p *TrendyolProvider) HealthCheck(ctx context.Context) error {
	// Test API connectivity by getting supplier info
	endpoint := fmt.Sprintf("/sapigw/suppliers/%s", p.supplierID)
	
	var response map[string]interface{}
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "HEALTH_CHECK_FAILED",
			Message:   "Failed to connect to Trendyol API",
			Provider:  "trendyol",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	
	return nil
}

// GetName returns the provider name
func (p *TrendyolProvider) GetName() string {
	return "Trendyol"
}

// GetType returns the provider type
func (p *TrendyolProvider) GetType() string {
	return "marketplace"
}

// IsHealthy checks if the provider is healthy
func (p *TrendyolProvider) IsHealthy(ctx context.Context) (bool, error) {
	err := p.HealthCheck(ctx)
	return err == nil, err
}

// GetMetrics returns provider metrics
func (p *TrendyolProvider) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"rate_limit_remaining": p.rateLimit.RequestsRemaining,
		"rate_limit_per_minute": p.rateLimit.RequestsPerMinute,
		"rate_limit_resets_at": p.rateLimit.ResetsAt.Unix(),
		"last_request_time": time.Now().Unix(),
		"provider_name": "trendyol",
		"base_url": p.baseURL,
		"supplier_id": p.supplierID,
	}
}

// GetCapabilities returns the capabilities of this integration
func (p *TrendyolProvider) GetCapabilities() []string {
	return []string{
		"product_sync",
		"order_sync",
		"inventory_sync",
		"price_sync",
		"category_mapping",
		"bulk_operations",
		"real_time_notifications",
		"webhook_support",
	}
}

// GetRateLimit returns current rate limit information
func (p *TrendyolProvider) GetRateLimit() integrations.RateLimitInfo {
	return p.rateLimit
}

// Close cleans up any resources
func (p *TrendyolProvider) Close() error {
	return nil
}

// SyncProducts syncs products to Trendyol
func (p *TrendyolProvider) SyncProducts(ctx context.Context, products []interface{}) error {
	trendyolProducts := make([]TrendyolProduct, 0, len(products))
	
	for _, product := range products {
		trendyolProduct, err := p.convertToTrendyolProduct(product)
		if err != nil {
			continue // Skip invalid products
		}
		trendyolProducts = append(trendyolProducts, trendyolProduct)
	}
	
	// Send products in batches of 100 (Trendyol limit)
	batchSize := 100
	for i := 0; i < len(trendyolProducts); i += batchSize {
		end := i + batchSize
		if end > len(trendyolProducts) {
			end = len(trendyolProducts)
		}
		
		batch := trendyolProducts[i:end]
		err := p.sendProductBatch(ctx, batch)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// GetProducts retrieves products from Trendyol
func (p *TrendyolProvider) GetProducts(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := fmt.Sprintf("/sapigw/suppliers/%s/products", p.supplierID)
	
	// Add query parameters
	queryParams := make([]string, 0)
	if page, ok := params["page"].(int); ok {
		queryParams = append(queryParams, fmt.Sprintf("page=%d", page))
	}
	if size, ok := params["size"].(int); ok {
		queryParams = append(queryParams, fmt.Sprintf("size=%d", size))
	}
	if approved, ok := params["approved"].(bool); ok {
		queryParams = append(queryParams, fmt.Sprintf("approved=%t", approved))
	}
	
	if len(queryParams) > 0 {
		endpoint += "?" + strings.Join(queryParams, "&")
	}
	
	var response struct {
		Content []TrendyolProduct `json:"content"`
		TotalElements int `json:"totalElements"`
		TotalPages int `json:"totalPages"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic interface
	products := make([]interface{}, len(response.Content))
	for i, product := range response.Content {
		products[i] = product
	}
	
	return products, nil
}

// GetOrders retrieves orders from Trendyol
func (p *TrendyolProvider) GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := fmt.Sprintf("/sapigw/suppliers/%s/orders", p.supplierID)
	
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
		Content []TrendyolOrder `json:"content"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic interface
	orders := make([]interface{}, len(response.Content))
	for i, order := range response.Content {
		orders[i] = order
	}
	
	return orders, nil
}

// UpdateStockAndPrice updates stock and price for products
func (p *TrendyolProvider) UpdateStockAndPrice(ctx context.Context, updates []interface{}) error {
	items := make([]TrendyolStockPriceItem, 0, len(updates))
	
	for _, update := range updates {
		item, err := p.convertToStockPriceItem(update)
		if err != nil {
			continue // Skip invalid items
		}
		items = append(items, item)
	}
	
	request := TrendyolStockPriceUpdate{
		Items: items,
	}
	
	endpoint := fmt.Sprintf("/sapigw/suppliers/%s/products/price-and-inventory", p.supplierID)
	
	var response map[string]interface{}
	return p.makeRequest(ctx, "POST", endpoint, request, &response)
}

// UpdateOrderStatus updates order status
func (p *TrendyolProvider) UpdateOrderStatus(ctx context.Context, orderID string, status string, params map[string]interface{}) error {
	endpoint := fmt.Sprintf("/sapigw/suppliers/%s/orders/%s/status", p.supplierID, orderID)
	
	request := map[string]interface{}{
		"status": status,
	}
	
	// Add additional parameters
	if trackingNumber, ok := params["tracking_number"].(string); ok {
		request["trackingNumber"] = trackingNumber
	}
	if invoiceNumber, ok := params["invoice_number"].(string); ok {
		request["invoiceNumber"] = invoiceNumber
	}
	
	var response map[string]interface{}
	return p.makeRequest(ctx, "PUT", endpoint, request, &response)
}

// GetCategories retrieves categories from Trendyol
func (p *TrendyolProvider) GetCategories(ctx context.Context) ([]interface{}, error) {
	endpoint := "/sapigw/categories"
	
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

// GetBrands retrieves brands from Trendyol
func (p *TrendyolProvider) GetBrands(ctx context.Context) ([]interface{}, error) {
	endpoint := "/sapigw/brands"
	
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

// makeRequest makes an HTTP request to Trendyol API
func (p *TrendyolProvider) makeRequest(ctx context.Context, method, endpoint string, request interface{}, response interface{}) error {
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
	auth := p.generateAuthHeader(method, endpoint, string(body))
	req.Header.Set("Authorization", auth)
	
	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "NETWORK_ERROR",
			Message:   err.Error(),
			Provider:  "trendyol",
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
			Message:    fmt.Sprintf("Trendyol API returned status %d", resp.StatusCode),
			Provider:   "trendyol",
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
				Message:    "Failed to parse Trendyol response",
				Provider:   "trendyol",
				Retryable:  false,
				Timestamp:  time.Now(),
				StatusCode: resp.StatusCode,
			}
		}
	}
	
	return nil
}

// generateAuthHeader generates Trendyol authorization header
func (p *TrendyolProvider) generateAuthHeader(method, uri, body string) string {
	// Trendyol uses Basic Auth with API key and secret
	auth := p.credentials.APIKey + ":" + p.credentials.APISecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// updateRateLimit updates rate limit information from response headers
func (p *TrendyolProvider) updateRateLimit(headers http.Header) {
	// Trendyol doesn't provide rate limit headers, so we track internally
	p.rateLimit.RequestsRemaining--
	if p.rateLimit.RequestsRemaining <= 0 {
		p.rateLimit.RequestsRemaining = p.rateLimit.RequestsPerMinute
		p.rateLimit.ResetsAt = time.Now().Add(time.Minute)
	}
}

// Helper methods for data conversion

func (p *TrendyolProvider) convertToTrendyolProduct(product interface{}) (TrendyolProduct, error) {
	productMap, ok := product.(map[string]interface{})
	if !ok {
		return TrendyolProduct{}, fmt.Errorf("invalid product format")
	}
	
	trendyolProduct := TrendyolProduct{
		Barcode:          getString(productMap, "barcode"),
		Title:            getString(productMap, "title"),
		ProductMainID:    getString(productMap, "product_main_id"),
		BrandID:          getInt(productMap, "brand_id"),
		CategoryID:       getInt(productMap, "category_id"),
		Quantity:         getInt(productMap, "quantity"),
		StockCode:        getString(productMap, "sku"),
		DimensionalWeight: getFloat64(productMap, "dimensional_weight"),
		Description:      getString(productMap, "description"),
		CurrencyType:     "TRY",
		ListPrice:        getFloat64(productMap, "list_price"),
		SalePrice:        getFloat64(productMap, "price"),
		VatRate:          18, // Default VAT rate for Turkey
		CargoCompanyID:   1,  // Default cargo company
	}
	
	// Set images
	if images, ok := productMap["images"].([]interface{}); ok {
		trendyolImages := make([]TrendyolImage, 0)
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				trendyolImages = append(trendyolImages, TrendyolImage{URL: imgStr})
			}
		}
		trendyolProduct.Images = trendyolImages
	}
	
	// Set attributes
	if attributes, ok := productMap["attributes"].(map[string]interface{}); ok {
		trendyolAttributes := make([]TrendyolAttribute, 0)
		for key, value := range attributes {
			if valueStr, ok := value.(string); ok {
				trendyolAttributes = append(trendyolAttributes, TrendyolAttribute{
					AttributeID:          getAttributeID(key),
					AttributeValueID:     getAttributeValueID(key, valueStr),
					CustomAttributeValue: valueStr,
				})
			}
		}
		trendyolProduct.Attributes = trendyolAttributes
	}
	
	return trendyolProduct, nil
}

// Helper functions for Trendyol product conversion
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return 0
}

func getAttributeID(attributeName string) int {
	// This would typically be a lookup from your attribute mapping
	// For now, return a default value
	attributeMap := map[string]int{
		"color":  1,
		"size":   2,
		"brand":  3,
		"model":  4,
	}
	if id, ok := attributeMap[attributeName]; ok {
		return id
	}
	return 0
}

func getAttributeValueID(attributeName, value string) int {
	// This would typically be a lookup from your attribute value mapping
	// For now, return a default value
	return 0
}

func (p *TrendyolProvider) convertToStockPriceItem(update interface{}) (TrendyolStockPriceItem, error) {
	updateMap, ok := update.(map[string]interface{})
	if !ok {
		return TrendyolStockPriceItem{}, fmt.Errorf("invalid update format")
	}
	
	item := TrendyolStockPriceItem{
		Barcode:   getString(updateMap, "barcode"),
		Quantity:  getInt(updateMap, "quantity"),
		SalePrice: getFloat64(updateMap, "price"),
		ListPrice: getFloat64(updateMap, "list_price"),
	}
	
	// If list price is not provided, use sale price + margin
	if item.ListPrice == 0 && item.SalePrice > 0 {
		item.ListPrice = item.SalePrice * 1.2 // Add 20% margin
	}
	
	return item, nil
}

func (p *TrendyolProvider) sendProductBatch(ctx context.Context, products []TrendyolProduct) error {
	endpoint := fmt.Sprintf("/sapigw/suppliers/%s/products", p.supplierID)
	
	request := map[string]interface{}{
		"items": products,
	}
	
	var response map[string]interface{}
	return p.makeRequest(ctx, "POST", endpoint, request, &response)
}