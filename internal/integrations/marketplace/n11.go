package marketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"kolajAi/internal/integrations"
)

// N11Provider implements marketplace provider for N11
type N11Provider struct {
	config      *MarketplaceProviderConfig
	httpClient  *http.Client
	credentials integrations.Credentials
	baseURL     string
	apiKey      string
	apiSecret   string
	rateLimit   integrations.RateLimitInfo
}

// N11Product represents an N11 product structure
type N11Product struct {
	ProductSellerCode   string         `json:"productSellerCode"`
	Title               string         `json:"title"`
	Subtitle            string         `json:"subtitle"`
	Description         string         `json:"description"`
	Category            N11Category    `json:"category"`
	Price               string         `json:"price"`
	CurrencyType        string         `json:"currencyType"`
	Images              N11Images      `json:"images"`
	StockItems          N11StockItems  `json:"stockItems"`
	Attributes          []N11Attribute `json:"attributes"`
	PreparingDay        int            `json:"preparingDay"`
	ShipmentTemplate    string         `json:"shipmentTemplate"`
	MaxPurchaseQuantity int            `json:"maxPurchaseQuantity"`
}

// N11Category represents N11 category structure
type N11Category struct {
	ID string `json:"id"`
}

// N11Images represents N11 images structure
type N11Images struct {
	Image []N11Image `json:"image"`
}

// N11Image represents single N11 image
type N11Image struct {
	URL   string `json:"url"`
	Order string `json:"order"`
}

// N11StockItems represents N11 stock items
type N11StockItems struct {
	StockItem []N11StockItem `json:"stockItem"`
}

// N11StockItem represents single N11 stock item
type N11StockItem struct {
	Bundle          string        `json:"bundle"`
	MPN             string        `json:"mpn"`
	GTIN            string        `json:"gtin"`
	Quantity        string        `json:"quantity"`
	SellerStockCode string        `json:"sellerStockCode"`
	OptionPrice     string        `json:"optionPrice"`
	Attributes      N11Attributes `json:"attributes"`
}

// N11Attributes represents N11 attributes
type N11Attributes struct {
	Attribute []N11Attribute `json:"attribute"`
}

// N11Attribute represents single N11 attribute
type N11Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// N11Order represents N11 order structure
type N11Order struct {
	ID           int64           `json:"id"`
	OrderNumber  string          `json:"orderNumber"`
	Status       string          `json:"status"`
	BuyerName    string          `json:"buyerName"`
	Recipient    string          `json:"recipient"`
	CreateDate   time.Time       `json:"createDate"`
	OrderItems   []N11OrderItem  `json:"orderItems"`
	ShippingInfo N11ShippingInfo `json:"shippingInfo"`
}

// N11OrderItem represents N11 order item
type N11OrderItem struct {
	ProductID   int64   `json:"productId"`
	ProductName string  `json:"productName"`
	SellerCode  string  `json:"sellerCode"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Commission  float64 `json:"commission"`
}

// N11ShippingInfo represents N11 shipping information
type N11ShippingInfo struct {
	CompanyName string    `json:"companyName"`
	TrackingNo  string    `json:"trackingNo"`
	ShippedDate time.Time `json:"shippedDate"`
}

// N11APIResponse represents N11 API response structure
type N11APIResponse struct {
	Result struct {
		Status       string      `json:"status"`
		ErrorCode    string      `json:"errorCode"`
		ErrorMessage string      `json:"errorMessage"`
		Data         interface{} `json:"data"`
	} `json:"result"`
}

// NewN11Provider creates a new N11 provider instance
func NewN11Provider() *N11Provider {
	return &N11Provider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.n11.com/ws",
		rateLimit: integrations.RateLimitInfo{
			RequestsPerSecond: 5,
			BurstSize:         10,
		},
	}
}

// Initialize initializes the N11 provider
func (p *N11Provider) Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error {
	p.credentials = credentials
	p.apiKey = credentials.APIKey
	p.apiSecret = credentials.APISecret

	// Set base URL based on environment
	if env, ok := config["environment"].(string); ok && env == "sandbox" {
		p.baseURL = "https://api-test.n11.com/ws"
	}

	// Test connection
	return p.testConnection(ctx)
}

// GetName returns the provider name
func (p *N11Provider) GetName() string {
	return "N11"
}

// GetType returns the provider type
func (p *N11Provider) GetType() string {
	return "marketplace"
}

// IsHealthy checks if the provider is healthy
func (p *N11Provider) IsHealthy(ctx context.Context) (bool, error) {
	return p.testConnection(ctx) == nil, nil
}

// GetMetrics returns provider metrics
func (p *N11Provider) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"rate_limit_remaining": p.rateLimit.RequestsPerSecond,
		"last_request_time":    time.Now().Unix(),
	}
}

// GetRateLimit returns rate limit information
func (p *N11Provider) GetRateLimit() integrations.RateLimitInfo {
	return p.rateLimit
}

// SyncProducts syncs products to N11
func (p *N11Provider) SyncProducts(ctx context.Context, products []interface{}) error {
	for _, product := range products {
		productMap, ok := product.(map[string]interface{})
		if !ok {
			continue
		}

		n11Product := p.convertToN11Product(productMap)
		if err := p.saveProduct(ctx, n11Product); err != nil {
			return fmt.Errorf("failed to sync product %s: %v", n11Product.ProductSellerCode, err)
		}
	}

	return nil
}

// UpdateStockAndPrice updates stock and price information
func (p *N11Provider) UpdateStockAndPrice(ctx context.Context, updates []interface{}) error {
	for _, update := range updates {
		updateMap, ok := update.(map[string]interface{})
		if !ok {
			continue
		}

		sellerCode, _ := updateMap["seller_code"].(string)
		quantity, _ := updateMap["quantity"].(int)
		price, _ := updateMap["price"].(float64)

		if err := p.updateStock(ctx, sellerCode, quantity); err != nil {
			return fmt.Errorf("failed to update stock for %s: %v", sellerCode, err)
		}

		if err := p.updatePrice(ctx, sellerCode, price); err != nil {
			return fmt.Errorf("failed to update price for %s: %v", sellerCode, err)
		}
	}

	return nil
}

// GetProducts retrieves products from N11
func (p *N11Provider) GetProducts(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/ProductService.do"

	requestData := map[string]interface{}{
		"auth": p.createAuth(),
		"pagingData": map[string]interface{}{
			"currentPage": params["page"],
			"pageSize":    params["limit"],
		},
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return nil, err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}

	if apiResponse.Result.Status != "success" {
		return nil, fmt.Errorf("N11 API error: %s", apiResponse.Result.ErrorMessage)
	}

	// Convert response to standard format
	products := make([]interface{}, 0)
	if data, ok := apiResponse.Result.Data.([]interface{}); ok {
		products = data
	}

	return products, nil
}

// GetOrders retrieves orders from N11
func (p *N11Provider) GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/OrderService.do"

	requestData := map[string]interface{}{
		"auth":       p.createAuth(),
		"searchData": params,
		"pagingData": map[string]interface{}{
			"currentPage": params["page"],
			"pageSize":    params["limit"],
		},
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return nil, err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}

	if apiResponse.Result.Status != "success" {
		return nil, fmt.Errorf("N11 API error: %s", apiResponse.Result.ErrorMessage)
	}

	// Convert response to standard format
	orders := make([]interface{}, 0)
	if data, ok := apiResponse.Result.Data.([]interface{}); ok {
		orders = data
	}

	return orders, nil
}

// UpdateOrderStatus updates order status
func (p *N11Provider) UpdateOrderStatus(ctx context.Context, orderID string, status string, params map[string]interface{}) error {
	endpoint := "/OrderService.do"

	requestData := map[string]interface{}{
		"auth": p.createAuth(),
		"orderItemList": []map[string]interface{}{
			{
				"id":     orderID,
				"status": status,
			},
		},
	}

	// Add tracking info if provided
	if trackingNo, ok := params["tracking_no"].(string); ok {
		requestData["shipmentInfo"] = map[string]interface{}{
			"trackingNumber": trackingNo,
			"companyName":    params["shipping_company"],
		}
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return err
	}

	if apiResponse.Result.Status != "success" {
		return fmt.Errorf("N11 API error: %s", apiResponse.Result.ErrorMessage)
	}

	return nil
}

// GetCategories retrieves categories from N11
func (p *N11Provider) GetCategories(ctx context.Context) ([]interface{}, error) {
	endpoint := "/CategoryService.do"

	requestData := map[string]interface{}{
		"auth": p.createAuth(),
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return nil, err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}

	if apiResponse.Result.Status != "success" {
		return nil, fmt.Errorf("N11 API error: %s", apiResponse.Result.ErrorMessage)
	}

	// Convert response to standard format
	categories := make([]interface{}, 0)
	if data, ok := apiResponse.Result.Data.([]interface{}); ok {
		categories = data
	}

	return categories, nil
}

// GetBrands retrieves brands from N11
func (p *N11Provider) GetBrands(ctx context.Context) ([]interface{}, error) {
	// N11 doesn't have a separate brands endpoint
	// Brands are usually part of category attributes
	return []interface{}{}, nil
}

// testConnection tests the N11 API connection
func (p *N11Provider) testConnection(ctx context.Context) error {
	endpoint := "/CategoryService.do"

	requestData := map[string]interface{}{
		"auth": p.createAuth(),
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return err
	}

	if apiResponse.Result.Status != "success" {
		return fmt.Errorf("N11 connection test failed: %s", apiResponse.Result.ErrorMessage)
	}

	return nil
}

// createAuth creates authentication data for N11 API
func (p *N11Provider) createAuth() map[string]interface{} {
	return map[string]interface{}{
		"appKey":    p.apiKey,
		"appSecret": p.apiSecret,
	}
}

// makeRequest makes HTTP request to N11 API
func (p *N11Provider) makeRequest(ctx context.Context, method, endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "KolajAI-N11-Integration/1.0")

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
		return nil, fmt.Errorf("N11 API error: %d - %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// saveProduct saves a product to N11
func (p *N11Provider) saveProduct(ctx context.Context, product N11Product) error {
	endpoint := "/ProductService.do"

	requestData := map[string]interface{}{
		"auth":    p.createAuth(),
		"product": product,
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return err
	}

	if apiResponse.Result.Status != "success" {
		return fmt.Errorf("N11 API error: %s", apiResponse.Result.ErrorMessage)
	}

	return nil
}

// updateStock updates product stock
func (p *N11Provider) updateStock(ctx context.Context, sellerCode string, quantity int) error {
	endpoint := "/ProductStockService.do"

	requestData := map[string]interface{}{
		"auth": p.createAuth(),
		"stockItems": []map[string]interface{}{
			{
				"sellerStockCode": sellerCode,
				"quantity":        strconv.Itoa(quantity),
			},
		},
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return err
	}

	if apiResponse.Result.Status != "success" {
		return fmt.Errorf("N11 API error: %s", apiResponse.Result.ErrorMessage)
	}

	return nil
}

// updatePrice updates product price
func (p *N11Provider) updatePrice(ctx context.Context, sellerCode string, price float64) error {
	endpoint := "/ProductService.do"

	requestData := map[string]interface{}{
		"auth": p.createAuth(),
		"productList": []map[string]interface{}{
			{
				"productSellerCode": sellerCode,
				"price":             fmt.Sprintf("%.2f", price),
			},
		},
	}

	response, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	if err != nil {
		return err
	}

	var apiResponse N11APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return err
	}

	if apiResponse.Result.Status != "success" {
		return fmt.Errorf("N11 API error: %s", apiResponse.Result.ErrorMessage)
	}

	return nil
}

// convertToN11Product converts generic product to N11 product format
func (p *N11Provider) convertToN11Product(product map[string]interface{}) N11Product {
	n11Product := N11Product{
		ProductSellerCode:   getString(product, "sku"),
		Title:               getString(product, "title"),
		Subtitle:            getString(product, "subtitle"),
		Description:         getString(product, "description"),
		Price:               fmt.Sprintf("%.2f", getFloat64(product, "price")),
		CurrencyType:        "1", // TL
		PreparingDay:        3,
		MaxPurchaseQuantity: 999,
	}

	// Set category
	if categoryID := getString(product, "category_id"); categoryID != "" {
		n11Product.Category = N11Category{ID: categoryID}
	}

	// Set images
	if images, ok := product["images"].([]interface{}); ok {
		n11Images := make([]N11Image, 0)
		for i, img := range images {
			if imgStr, ok := img.(string); ok {
				n11Images = append(n11Images, N11Image{
					URL:   imgStr,
					Order: strconv.Itoa(i + 1),
				})
			}
		}
		n11Product.Images = N11Images{Image: n11Images}
	}

	// Set stock items
	quantity := getInt(product, "quantity")
	n11Product.StockItems = N11StockItems{
		StockItem: []N11StockItem{
			{
				Bundle:          "false",
				SellerStockCode: getString(product, "sku"),
				Quantity:        strconv.Itoa(quantity),
				OptionPrice:     n11Product.Price,
				Attributes: N11Attributes{
					Attribute: []N11Attribute{
						{Name: "Marka", Value: getString(product, "brand")},
						{Name: "Renk", Value: getString(product, "color")},
					},
				},
			},
		},
	}

	return n11Product
}

// Helper functions moved to helpers.go
