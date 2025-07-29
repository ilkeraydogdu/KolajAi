package marketplace

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"kolajAi/internal/integrations"
	"kolajAi/internal/errors"
	"kolajAi/internal/retry"
	"kolajAi/internal/security"
)

// TrendyolProviderEnhanced implements enhanced marketplace provider for Trendyol
type TrendyolProviderEnhanced struct {
	config         *MarketplaceProviderConfig
	httpClient     *http.Client
	credentials    *integrations.SecureCredentials
	baseURL        string
	supplierID     string
	rateLimit      integrations.RateLimitInfo
	retryManager   *retry.RetryManager
	inputValidator *security.InputValidator
	errorHandler   *errors.ErrorHandler
	rateLimiter    *RateLimiter
	mutex          sync.RWMutex
	lastRequest    time.Time
	requestCount   int64
	isHealthy      bool
	lastHealthCheck time.Time
}

// RateLimiter manages API rate limiting
type RateLimiter struct {
	requestsPerMinute int
	requestsRemaining int
	resetTime         time.Time
	mutex             sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		requestsRemaining: requestsPerMinute,
		resetTime:         time.Now().Add(time.Minute),
	}
}

// CanMakeRequest checks if a request can be made
func (rl *RateLimiter) CanMakeRequest() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	if now.After(rl.resetTime) {
		rl.requestsRemaining = rl.requestsPerMinute
		rl.resetTime = now.Add(time.Minute)
	}

	return rl.requestsRemaining > 0
}

// ConsumeRequest consumes a request from the rate limiter
func (rl *RateLimiter) ConsumeRequest() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	if rl.requestsRemaining > 0 {
		rl.requestsRemaining--
	}
}

// GetWaitTime returns the time to wait before next request
func (rl *RateLimiter) GetWaitTime() time.Duration {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	if rl.requestsRemaining > 0 {
		return 0
	}
	
	return time.Until(rl.resetTime)
}

// NewTrendyolProviderEnhanced creates a new enhanced Trendyol provider
func NewTrendyolProviderEnhanced(credentialManager *security.CredentialManager, credentialID string) *TrendyolProviderEnhanced {
	retryConfig := retry.DefaultRetryConfig()
	retryConfig.MaxAttempts = 5
	retryConfig.InitialDelay = 2 * time.Second
	retryConfig.MaxDelay = 2 * time.Minute

	return &TrendyolProviderEnhanced{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		credentials:    integrations.NewSecureCredentials(credentialManager, credentialID),
		rateLimit: integrations.RateLimitInfo{
			RequestsPerMinute: 60,
			RequestsRemaining: 60,
			ResetsAt:          time.Now().Add(time.Minute),
		},
		retryManager:   retry.NewRetryManager(retryConfig),
		inputValidator: security.NewInputValidator(),
		errorHandler:   errors.NewErrorHandler(nil),
		rateLimiter:    NewRateLimiter(60),
		isHealthy:      true,
	}
}

// Initialize sets up the Trendyol provider
func (p *TrendyolProviderEnhanced) Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error {
	// Validate configuration
	validationResults := p.inputValidator.ValidateBatch(config, map[string]security.ValidationRule{
		"environment": {Required: true, AllowedValues: []string{"production", "staging", "sandbox"}},
		"supplier_id": {Required: true, MinLength: 1, MaxLength: 50},
		"base_url":    {Type: "url"},
	})

	if !security.IsValidationPassed(validationResults) {
		validationErrors := security.GetValidationErrors(validationResults)
		return errors.NewValidationError("trendyol", "configuration validation failed", map[string]interface{}{
			"errors": validationErrors,
		})
	}

	sanitizedConfig := security.GetSanitizedData(validationResults)

	// Set base URL based on environment
	environment := sanitizedConfig["environment"].(string)
	switch environment {
	case "production":
		p.baseURL = "https://api.trendyol.com"
	case "staging":
		p.baseURL = "https://stageapi.trendyol.com"
	default:
		p.baseURL = "https://stageapi.trendyol.com"
	}

	// Override base URL if provided
	if baseURL, ok := sanitizedConfig["base_url"].(string); ok && baseURL != "" {
		p.baseURL = baseURL
	}

	// Set supplier ID
	if supplierID, ok := sanitizedConfig["supplier_id"].(string); ok {
		p.supplierID = supplierID
	} else {
		return errors.NewValidationError("trendyol", "supplier_id is required", nil)
	}

	// Test connectivity
	return p.testConnectivity(ctx)
}

// testConnectivity tests the connection to Trendyol API
func (p *TrendyolProviderEnhanced) testConnectivity(ctx context.Context) error {
	endpoint := fmt.Sprintf("/sapigw/suppliers/%s", p.supplierID)
	
	return p.retryManager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		var response map[string]interface{}
		return p.makeRequest(ctx, "GET", endpoint, nil, &response)
	})
}

// GetName returns the provider name
func (p *TrendyolProviderEnhanced) GetName() string {
	return "Trendyol Enhanced"
}

// GetType returns the provider type
func (p *TrendyolProviderEnhanced) GetType() string {
	return "marketplace"
}

// IsHealthy returns the current health status
func (p *TrendyolProviderEnhanced) IsHealthy() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isHealthy
}

// GetMetrics returns provider metrics
func (p *TrendyolProviderEnhanced) GetMetrics() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	retryStats := p.retryManager.GetStats()
	errorStats := p.errorHandler.GetStats()

	return map[string]interface{}{
		"provider":           "trendyol_enhanced",
		"healthy":            p.isHealthy,
		"last_health_check":  p.lastHealthCheck,
		"request_count":      p.requestCount,
		"last_request":       p.lastRequest,
		"rate_limit":         p.rateLimit,
		"retry_stats":        retryStats,
		"error_stats":        errorStats,
		"base_url":           p.baseURL,
		"supplier_id":        p.supplierID,
	}
}

// HealthCheck verifies the Trendyol integration is working
func (p *TrendyolProviderEnhanced) HealthCheck(ctx context.Context) error {
	p.mutex.Lock()
	p.lastHealthCheck = time.Now()
	p.mutex.Unlock()

	err := p.retryManager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		endpoint := fmt.Sprintf("/sapigw/suppliers/%s", p.supplierID)
		var response map[string]interface{}
		return p.makeRequest(ctx, "GET", endpoint, nil, &response)
	})

	p.mutex.Lock()
	p.isHealthy = (err == nil)
	p.mutex.Unlock()

	if err != nil {
		integrationErr := errors.NewIntegrationError(
			errors.ErrorCodeProviderError,
			"Trendyol health check failed",
			"trendyol",
		).WithCause(err)
		
		p.errorHandler.HandleError(integrationErr)
		return integrationErr
	}

	return nil
}

// SyncProducts synchronizes products with Trendyol
func (p *TrendyolProviderEnhanced) SyncProducts(ctx context.Context, products []interface{}) error {
	if len(products) == 0 {
		return errors.NewValidationError("trendyol", "no products provided for sync", nil)
	}

	// Validate products
	for i, product := range products {
		productMap, ok := product.(map[string]interface{})
		if !ok {
			return errors.NewValidationError("trendyol", fmt.Sprintf("invalid product format at index %d", i), nil)
		}

		validationResult := p.inputValidator.ValidateProductData(productMap)
		if !validationResult.IsValid {
			return errors.NewValidationError("trendyol", fmt.Sprintf("product validation failed at index %d", i), map[string]interface{}{
				"errors": validationResult.Errors,
				"index":  i,
			})
		}
	}

	// Process products in batches
	batchSize := 100
	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]
		err := p.syncProductBatch(ctx, batch)
		if err != nil {
			return err
		}
	}

	return nil
}

// syncProductBatch synchronizes a batch of products
func (p *TrendyolProviderEnhanced) syncProductBatch(ctx context.Context, products []interface{}) error {
	trendyolProducts := make([]TrendyolProduct, 0, len(products))

	for _, product := range products {
		trendyolProduct, err := p.convertToTrendyolProduct(product)
		if err != nil {
			return errors.NewIntegrationError(
				errors.ErrorCodeMappingError,
				"failed to convert product to Trendyol format",
				"trendyol",
			).WithCause(err)
		}
		trendyolProducts = append(trendyolProducts, trendyolProduct)
	}

	endpoint := "/sapigw/products"
	request := map[string]interface{}{
		"products": trendyolProducts,
	}

	return p.retryManager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		var response map[string]interface{}
		return p.makeRequest(ctx, "POST", endpoint, request, &response)
	})
}

// GetOrders retrieves orders from Trendyol
func (p *TrendyolProviderEnhanced) GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	// Validate parameters
	validationResults := p.inputValidator.ValidateBatch(params, map[string]security.ValidationRule{
		"start_date": {Type: "date"},
		"end_date":   {Type: "date"},
		"page":       {Type: "int"},
		"size":       {Type: "int"},
		"status":     {AllowedValues: []string{"Created", "Picking", "Invoiced", "Shipped", "Delivered", "Cancelled"}},
	})

	if !security.IsValidationPassed(validationResults) {
		validationErrors := security.GetValidationErrors(validationResults)
		return nil, errors.NewValidationError("trendyol", "parameter validation failed", map[string]interface{}{
			"errors": validationErrors,
		})
	}

	sanitizedParams := security.GetSanitizedData(validationResults)

	// Build query parameters
	queryParams := make(map[string]string)
	if startDate, ok := sanitizedParams["start_date"].(time.Time); ok {
		queryParams["startDate"] = startDate.Format("2006-01-02")
	}
	if endDate, ok := sanitizedParams["end_date"].(time.Time); ok {
		queryParams["endDate"] = endDate.Format("2006-01-02")
	}
	if page, ok := sanitizedParams["page"].(int64); ok {
		queryParams["page"] = strconv.FormatInt(page, 10)
	}
	if size, ok := sanitizedParams["size"].(int64); ok {
		queryParams["size"] = strconv.FormatInt(size, 10)
	}
	if status, ok := sanitizedParams["status"].(string); ok {
		queryParams["status"] = status
	}

	endpoint := "/sapigw/orders"
	if len(queryParams) > 0 {
		endpoint += "?" + p.buildQueryString(queryParams)
	}

	var response struct {
		Orders []TrendyolOrder `json:"orders"`
		Page   int             `json:"page"`
		Size   int             `json:"size"`
		Total  int             `json:"totalElements"`
	}

	err := p.retryManager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return p.makeRequest(ctx, "GET", endpoint, nil, &response)
	})

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

// UpdateStock updates product stock on Trendyol
func (p *TrendyolProviderEnhanced) UpdateStock(ctx context.Context, updates []interface{}) error {
	if len(updates) == 0 {
		return errors.NewValidationError("trendyol", "no stock updates provided", nil)
	}

	// Validate stock updates
	for i, update := range updates {
		updateMap, ok := update.(map[string]interface{})
		if !ok {
			return errors.NewValidationError("trendyol", fmt.Sprintf("invalid stock update format at index %d", i), nil)
		}

		validationResults := p.inputValidator.ValidateBatch(updateMap, map[string]security.ValidationRule{
			"barcode":  {Required: true, MinLength: 1, MaxLength: 50},
			"quantity": {Required: true, Type: "int"},
		})

		if !security.IsValidationPassed(validationResults) {
			validationErrors := security.GetValidationErrors(validationResults)
			return errors.NewValidationError("trendyol", fmt.Sprintf("stock update validation failed at index %d", i), map[string]interface{}{
				"errors": validationErrors,
				"index":  i,
			})
		}
	}

	// Convert to Trendyol format
	stockUpdates := make([]map[string]interface{}, 0, len(updates))
	for _, update := range updates {
		updateMap := update.(map[string]interface{})
		stockUpdate := map[string]interface{}{
			"barcode":  updateMap["barcode"],
			"quantity": updateMap["quantity"],
		}
		stockUpdates = append(stockUpdates, stockUpdate)
	}

	endpoint := "/sapigw/products/stocks"
	request := map[string]interface{}{
		"items": stockUpdates,
	}

	return p.retryManager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		var response map[string]interface{}
		return p.makeRequest(ctx, "POST", endpoint, request, &response)
	})
}

// UpdatePrice updates product prices on Trendyol
func (p *TrendyolProviderEnhanced) UpdatePrice(ctx context.Context, updates []interface{}) error {
	if len(updates) == 0 {
		return errors.NewValidationError("trendyol", "no price updates provided", nil)
	}

	// Validate price updates
	for i, update := range updates {
		updateMap, ok := update.(map[string]interface{})
		if !ok {
			return errors.NewValidationError("trendyol", fmt.Sprintf("invalid price update format at index %d", i), nil)
		}

		validationResults := p.inputValidator.ValidateBatch(updateMap, map[string]security.ValidationRule{
			"barcode":    {Required: true, MinLength: 1, MaxLength: 50},
			"list_price": {Required: true, Type: "float"},
			"sale_price": {Required: true, Type: "float"},
		})

		if !security.IsValidationPassed(validationResults) {
			validationErrors := security.GetValidationErrors(validationResults)
			return errors.NewValidationError("trendyol", fmt.Sprintf("price update validation failed at index %d", i), map[string]interface{}{
				"errors": validationErrors,
				"index":  i,
			})
		}
	}

	// Convert to Trendyol format
	priceUpdates := make([]map[string]interface{}, 0, len(updates))
	for _, update := range updates {
		updateMap := update.(map[string]interface{})
		priceUpdate := map[string]interface{}{
			"barcode":   updateMap["barcode"],
			"listPrice": updateMap["list_price"],
			"salePrice": updateMap["sale_price"],
		}
		priceUpdates = append(priceUpdates, priceUpdate)
	}

	endpoint := "/sapigw/products/prices"
	request := map[string]interface{}{
		"items": priceUpdates,
	}

	return p.retryManager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		var response map[string]interface{}
		return p.makeRequest(ctx, "POST", endpoint, request, &response)
	})
}

// makeRequest makes an HTTP request to Trendyol API with enhanced error handling
func (p *TrendyolProviderEnhanced) makeRequest(ctx context.Context, method, endpoint string, data interface{}, response interface{}) error {
	// Check rate limiting
	if !p.rateLimiter.CanMakeRequest() {
		waitTime := p.rateLimiter.GetWaitTime()
		return errors.NewRateLimitError("trendyol", "rate limit exceeded", waitTime)
	}

	// Consume rate limit
	p.rateLimiter.ConsumeRequest()

	// Get credentials
	credData, err := p.credentials.GetCredentialData()
	if err != nil {
		return errors.NewAuthenticationError("trendyol", "failed to retrieve credentials")
	}

	// Prepare request body
	var body []byte
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return errors.NewIntegrationError(
				errors.ErrorCodeInvalidInput,
				"failed to marshal request data",
				"trendyol",
			).WithCause(err)
		}
	}

	// Create request
	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewIntegrationError(
			errors.ErrorCodeInternalError,
			"failed to create HTTP request",
			"trendyol",
		).WithCause(err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "KolajAI-Trendyol-Integration/1.0")
	
	// Set authentication
	auth := credData.APIKey + ":" + credData.APISecret
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))

	// Update request tracking
	p.mutex.Lock()
	p.requestCount++
	p.lastRequest = time.Now()
	p.mutex.Unlock()

	// Make request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		integrationErr := errors.NewNetworkError("trendyol", "HTTP request failed", true).WithCause(err)
		p.errorHandler.HandleError(integrationErr)
		return integrationErr
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode >= 400 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		
		message := fmt.Sprintf("Trendyol API error: %d", resp.StatusCode)
		if errorMsg, ok := errorResponse["message"].(string); ok {
			message = errorMsg
		}

		integrationErr := errors.NewAPIError("trendyol", message, resp.StatusCode)
		
		// Add response context
		if len(errorResponse) > 0 {
			integrationErr.WithContext("api_response", errorResponse)
		}
		
		p.errorHandler.HandleError(integrationErr)
		return integrationErr
	}

	// Update rate limit info from response headers
	p.updateRateLimitFromHeaders(resp.Header)

	// Parse response
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			integrationErr := errors.NewIntegrationError(
				errors.ErrorCodeInternalError,
				"failed to decode API response",
				"trendyol",
			).WithCause(err)
			
			p.errorHandler.HandleError(integrationErr)
			return integrationErr
		}
	}

	return nil
}

// updateRateLimitFromHeaders updates rate limit info from response headers
func (p *TrendyolProviderEnhanced) updateRateLimitFromHeaders(headers http.Header) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

// convertToTrendyolProduct converts generic product to Trendyol format
func (p *TrendyolProviderEnhanced) convertToTrendyolProduct(product interface{}) (TrendyolProduct, error) {
	productMap, ok := product.(map[string]interface{})
	if !ok {
		return TrendyolProduct{}, fmt.Errorf("invalid product format")
	}

	// Validate required fields
	validationResult := p.inputValidator.ValidateProductData(productMap)
	if !validationResult.IsValid {
		return TrendyolProduct{}, fmt.Errorf("product validation failed: %v", validationResult.Errors)
	}

	sanitizedData := validationResult.SanitizedValue.(map[string]interface{})

	trendyolProduct := TrendyolProduct{
		Title:            getString(sanitizedData, "title"),
		ProductMainID:    getString(sanitizedData, "product_main_id"),
		BrandID:          getInt(sanitizedData, "brand_id"),
		CategoryID:       getInt(sanitizedData, "category_id"),
		Quantity:         getInt(sanitizedData, "quantity"),
		StockCode:        getString(sanitizedData, "stock_code"),
		Barcode:          getString(sanitizedData, "barcode"),
		Description:      getString(sanitizedData, "description"),
		CurrencyType:     getStringWithDefault(sanitizedData, "currency_type", "TRY"),
		ListPrice:        getFloat64(sanitizedData, "list_price"),
		SalePrice:        getFloat64(sanitizedData, "sale_price"),
		VatRate:          getIntWithDefault(sanitizedData, "vat_rate", 18),
		CargoCompanyID:   getIntWithDefault(sanitizedData, "cargo_company_id", 10),
		DimensionalWeight: getFloat64(sanitizedData, "dimensional_weight"),
	}

	// Handle images
	if imagesInterface, ok := sanitizedData["images"]; ok {
		if imagesList, ok := imagesInterface.([]interface{}); ok {
			for _, img := range imagesList {
				if imgMap, ok := img.(map[string]interface{}); ok {
					trendyolProduct.Images = append(trendyolProduct.Images, TrendyolImage{
						URL: getString(imgMap, "url"),
					})
				}
			}
		}
	}

	// Handle attributes
	if attributesInterface, ok := sanitizedData["attributes"]; ok {
		if attributesList, ok := attributesInterface.([]interface{}); ok {
			for _, attr := range attributesList {
				if attrMap, ok := attr.(map[string]interface{}); ok {
					trendyolProduct.Attributes = append(trendyolProduct.Attributes, TrendyolAttribute{
						AttributeID:      getInt(attrMap, "attribute_id"),
						AttributeValueID: getInt(attrMap, "attribute_value_id"),
						CustomValue:      getString(attrMap, "custom_value"),
					})
				}
			}
		}
	}

	return trendyolProduct, nil
}

// buildQueryString builds a query string from parameters
func (p *TrendyolProviderEnhanced) buildQueryString(params map[string]string) string {
	var parts []string
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, "&")
}

// Helper functions (same as before)
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getStringWithDefault(m map[string]interface{}, key, defaultValue string) string {
	if v := getString(m, key); v != "" {
		return v
	}
	return defaultValue
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return 0
}

func getIntWithDefault(m map[string]interface{}, key string, defaultValue int) int {
	if v := getInt(m, key); v != 0 {
		return v
	}
	return defaultValue
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

// TrendyolProduct represents a Trendyol product structure
type TrendyolProduct struct {
	Barcode           string              `json:"barcode"`
	Title             string              `json:"title"`
	ProductMainID     string              `json:"productMainId"`
	BrandID           int                 `json:"brandId"`
	CategoryID        int                 `json:"categoryId"`
	Quantity          int                 `json:"quantity"`
	StockCode         string              `json:"stockCode"`
	DimensionalWeight float64             `json:"dimensionalWeight"`
	Description       string              `json:"description"`
	CurrencyType      string              `json:"currencyType"`
	ListPrice         float64             `json:"listPrice"`
	SalePrice         float64             `json:"salePrice"`
	VatRate           int                 `json:"vatRate"`
	CargoCompanyID    int                 `json:"cargoCompanyId"`
	Images            []TrendyolImage     `json:"images"`
	Attributes        []TrendyolAttribute `json:"attributes"`
}

// TrendyolImage represents product image
type TrendyolImage struct {
	URL string `json:"url"`
}

// TrendyolAttribute represents product attribute
type TrendyolAttribute struct {
	AttributeID      int    `json:"attributeId"`
	AttributeValueID int    `json:"attributeValueId"`
	CustomValue      string `json:"customValue,omitempty"`
}

// TrendyolOrder represents Trendyol order structure
type TrendyolOrder struct {
	OrderNumber      string                `json:"orderNumber"`
	OrderDate        time.Time             `json:"orderDate"`
	CustomerEmail    string                `json:"customerEmail"`
	CustomerName     string                `json:"customerName"`
	TotalPrice       float64               `json:"totalPrice"`
	Currency         string                `json:"currency"`
	Status           string                `json:"status"`
	ShippingAddress  TrendyolAddress       `json:"shippingAddress"`
	BillingAddress   TrendyolAddress       `json:"billingAddress"`
	OrderItems       []TrendyolOrderItem   `json:"orderItems"`
}

// TrendyolAddress represents address structure
type TrendyolAddress struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Address   string `json:"address"`
	City      string `json:"city"`
	District  string `json:"district"`
	PostCode  string `json:"postCode"`
	Phone     string `json:"phone"`
}

// TrendyolOrderItem represents order item structure
type TrendyolOrderItem struct {
	Barcode     string  `json:"barcode"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	TotalPrice  float64 `json:"totalPrice"`
}