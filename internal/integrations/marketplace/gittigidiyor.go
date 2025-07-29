package marketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/yourproject/internal/integrations"
)

// GittiGidiyorProvider implements marketplace integration for GittiGidiyor
type GittiGidiyorProvider struct {
	apiKey    string
	secretKey string
	baseURL   string
	client    *http.Client
}

// GittiGidiyor API structures
type GittiGidiyorProduct struct {
	ID                  string                     `json:"id,omitempty"`
	Title               string                     `json:"title"`
	CategoryID          int                        `json:"categoryId"`
	Description         string                     `json:"description"`
	StartPrice          float64                    `json:"startPrice"`
	BuyNowPrice         float64                    `json:"buyNowPrice,omitempty"`
	ListingType         string                     `json:"listingType"` // "StoreInventory", "Auction", "FixedPrice"
	Duration            int                        `json:"duration"`    // Days
	Quantity            int                        `json:"quantity"`
	Condition           string                     `json:"condition"` // "New", "Used"
	ShippingTemplate    string                     `json:"shippingTemplate,omitempty"`
	PaymentMethods      []string                   `json:"paymentMethods"`
	Images              []GittiGidiyorImage        `json:"images"`
	ItemSpecifics       []GittiGidiyorItemSpecific `json:"itemSpecifics,omitempty"`
	ReturnPolicy        GittiGidiyorReturnPolicy   `json:"returnPolicy,omitempty"`
	Location            GittiGidiyorLocation       `json:"location"`
	ShippingDetails     GittiGidiyorShipping       `json:"shippingDetails"`
	SellerProvidedTitle string                     `json:"sellerProvidedTitle,omitempty"`
}

type GittiGidiyorImage struct {
	URL         string `json:"url"`
	IsPrimary   bool   `json:"isPrimary,omitempty"`
	Description string `json:"description,omitempty"`
}

type GittiGidiyorItemSpecific struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GittiGidiyorReturnPolicy struct {
	ReturnsAccepted bool   `json:"returnsAccepted"`
	ReturnPeriod    int    `json:"returnPeriod"` // Days
	RefundMethod    string `json:"refundMethod"`
	ShippingCostPaidBy string `json:"shippingCostPaidBy"`
}

type GittiGidiyorLocation struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	PostalCode  string `json:"postalCode"`
	Region      string `json:"region,omitempty"`
}

type GittiGidiyorShipping struct {
	ShippingType        string                      `json:"shippingType"` // "Flat", "Calculated"
	ShippingServiceCost float64                     `json:"shippingServiceCost"`
	ShippingService     string                      `json:"shippingService"`
	DispatchTimeMax     int                         `json:"dispatchTimeMax"`
	InternationalShipping []GittiGidiyorInternationalShipping `json:"internationalShipping,omitempty"`
}

type GittiGidiyorInternationalShipping struct {
	ShippingService     string  `json:"shippingService"`
	ShippingServiceCost float64 `json:"shippingServiceCost"`
	ShipToLocations     []string `json:"shipToLocations"`
}

type GittiGidiyorOrder struct {
	OrderID         string                    `json:"orderId"`
	BuyerUsername   string                    `json:"buyerUsername"`
	BuyerEmail      string                    `json:"buyerEmail"`
	OrderDate       time.Time                 `json:"orderDate"`
	OrderStatus     string                    `json:"orderStatus"`
	PaymentStatus   string                    `json:"paymentStatus"`
	ShippingStatus  string                    `json:"shippingStatus"`
	TotalAmount     float64                   `json:"totalAmount"`
	Currency        string                    `json:"currency"`
	ShippingAddress GittiGidiyorAddress       `json:"shippingAddress"`
	BillingAddress  GittiGidiyorAddress       `json:"billingAddress"`
	Items           []GittiGidiyorOrderItem   `json:"items"`
	ShippingDetails GittiGidiyorOrderShipping `json:"shippingDetails"`
}

type GittiGidiyorAddress struct {
	Name       string `json:"name"`
	Street1    string `json:"street1"`
	Street2    string `json:"street2,omitempty"`
	City       string `json:"city"`
	StateOrProvince string `json:"stateOrProvince"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
	Phone      string `json:"phone,omitempty"`
}

type GittiGidiyorOrderItem struct {
	ItemID       string  `json:"itemId"`
	Title        string  `json:"title"`
	SKU          string  `json:"sku,omitempty"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	TotalPrice   float64 `json:"totalPrice"`
	Currency     string  `json:"currency"`
	ItemLocation GittiGidiyorLocation `json:"itemLocation"`
}

type GittiGidiyorOrderShipping struct {
	ShippingService string    `json:"shippingService"`
	TrackingNumber  string    `json:"trackingNumber,omitempty"`
	ShippingCost    float64   `json:"shippingCost"`
	EstimatedDelivery time.Time `json:"estimatedDelivery,omitempty"`
}

type GittiGidiyorCategory struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ParentID    int    `json:"parentId,omitempty"`
	LeafCategory bool  `json:"leafCategory"`
	Level       int    `json:"level"`
}

type GittiGidiyorBrand struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NewGittiGidiyorProvider creates a new GittiGidiyor marketplace provider
func NewGittiGidiyorProvider(apiKey, secretKey string) *GittiGidiyorProvider {
	return &GittiGidiyorProvider{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   "https://dev.gittigidiyor.com:8443/listingapi/ws", // GittiGidiyor API base URL
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the provider name
func (p *GittiGidiyorProvider) GetName() string {
	return "GittiGidiyor"
}

// GetType returns the provider type
func (p *GittiGidiyorProvider) GetType() string {
	return "marketplace"
}

// IsHealthy returns the current health status
func (p *GittiGidiyorProvider) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return p.HealthCheck(ctx) == nil
}

// GetMetrics returns provider metrics
func (p *GittiGidiyorProvider) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"provider":     "gittigidiyor",
		"healthy":      p.IsHealthy(),
		"last_check":   time.Now(),
		"api_version":  "v1",
		"base_url":     p.baseURL,
	}
}

// HealthCheck verifies the GittiGidiyor integration is working
func (p *GittiGidiyorProvider) HealthCheck(ctx context.Context) error {
	// Test API connectivity by getting user info
	var response map[string]interface{}
	err := p.makeRequest(ctx, "GET", "/getUserInfo", nil, &response)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "HEALTH_CHECK_FAILED",
			Message:   "Failed to connect to GittiGidiyor API",
			Provider:  "gittigidiyor",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	
	return nil
}

// GetProducts retrieves products from GittiGidiyor
func (p *GittiGidiyorProvider) GetProducts(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/getItems"
	
	// Add query parameters
	queryParams := make(map[string]string)
	if page, ok := params["page"].(int); ok {
		queryParams["pageNumber"] = strconv.Itoa(page)
	}
	if size, ok := params["size"].(int); ok {
		queryParams["pageSize"] = strconv.Itoa(size)
	}
	if status, ok := params["status"].(string); ok {
		queryParams["itemStatus"] = status
	}
	
	var response struct {
		Items []GittiGidiyorProduct `json:"items"`
		TotalCount int `json:"totalCount"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, queryParams, &response)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic interface
	products := make([]interface{}, len(response.Items))
	for i, product := range response.Items {
		products[i] = product
	}
	
	return products, nil
}

// GetOrders retrieves orders from GittiGidiyor
func (p *GittiGidiyorProvider) GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/getOrders"
	
	// Add query parameters
	queryParams := make(map[string]string)
	if startDate, ok := params["start_date"].(string); ok {
		queryParams["startDate"] = startDate
	}
	if endDate, ok := params["end_date"].(string); ok {
		queryParams["endDate"] = endDate
	}
	if status, ok := params["status"].(string); ok {
		queryParams["orderStatus"] = status
	}
	if page, ok := params["page"].(int); ok {
		queryParams["pageNumber"] = strconv.Itoa(page)
	}
	
	var response struct {
		Orders []GittiGidiyorOrder `json:"orders"`
		TotalCount int `json:"totalCount"`
	}
	
	err := p.makeRequest(ctx, "GET", endpoint, queryParams, &response)
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

// GetCategories retrieves categories from GittiGidiyor
func (p *GittiGidiyorProvider) GetCategories(ctx context.Context) ([]interface{}, error) {
	endpoint := "/getCategories"
	
	var response struct {
		Categories []GittiGidiyorCategory `json:"categories"`
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

// GetBrands retrieves brands from GittiGidiyor
func (p *GittiGidiyorProvider) GetBrands(ctx context.Context) ([]interface{}, error) {
	endpoint := "/getBrands"
	
	var response struct {
		Brands []GittiGidiyorBrand `json:"brands"`
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

// CreateProduct creates a new product on GittiGidiyor
func (p *GittiGidiyorProvider) CreateProduct(ctx context.Context, product interface{}) error {
	gittiGidiyorProduct, err := p.convertToGittiGidiyorProduct(product)
	if err != nil {
		return err
	}
	
	endpoint := "/addItem"
	var response map[string]interface{}
	
	return p.makeRequest(ctx, "POST", endpoint, gittiGidiyorProduct, &response)
}

// UpdateProduct updates an existing product on GittiGidiyor
func (p *GittiGidiyorProvider) UpdateProduct(ctx context.Context, productID string, product interface{}) error {
	gittiGidiyorProduct, err := p.convertToGittiGidiyorProduct(product)
	if err != nil {
		return err
	}
	
	gittiGidiyorProduct.ID = productID
	endpoint := "/reviseItem"
	var response map[string]interface{}
	
	return p.makeRequest(ctx, "POST", endpoint, gittiGidiyorProduct, &response)
}

// UpdateStock updates product stock on GittiGidiyor
func (p *GittiGidiyorProvider) UpdateStock(ctx context.Context, updates []interface{}) error {
	endpoint := "/reviseQuantityAndPrice"
	
	for _, update := range updates {
		stockUpdate, err := p.convertToStockUpdate(update)
		if err != nil {
			return err
		}
		
		var response map[string]interface{}
		err = p.makeRequest(ctx, "POST", endpoint, stockUpdate, &response)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// UpdatePrice updates product price on GittiGidiyor
func (p *GittiGidiyorProvider) UpdatePrice(ctx context.Context, updates []interface{}) error {
	// GittiGidiyor updates stock and price together
	return p.UpdateStock(ctx, updates)
}

// convertToGittiGidiyorProduct converts generic product to GittiGidiyor format
func (p *GittiGidiyorProvider) convertToGittiGidiyorProduct(product interface{}) (GittiGidiyorProduct, error) {
	productMap, ok := product.(map[string]interface{})
	if !ok {
		return GittiGidiyorProduct{}, fmt.Errorf("invalid product format")
	}
	
	gittiGidiyorProduct := GittiGidiyorProduct{
		Title:       getString(productMap, "title"),
		CategoryID:  getInt(productMap, "category_id"),
		Description: getString(productMap, "description"),
		StartPrice:  getFloat64(productMap, "price"),
		BuyNowPrice: getFloat64(productMap, "buy_now_price"),
		ListingType: getStringWithDefault(productMap, "listing_type", "StoreInventory"),
		Duration:    getIntWithDefault(productMap, "duration", 30),
		Quantity:    getInt(productMap, "quantity"),
		Condition:   getStringWithDefault(productMap, "condition", "New"),
		PaymentMethods: []string{"PayPal", "CreditCard", "BankTransfer"},
		Location: GittiGidiyorLocation{
			Country:    getStringWithDefault(productMap, "country", "TR"),
			City:       getString(productMap, "city"),
			PostalCode: getString(productMap, "postal_code"),
		},
		ShippingDetails: GittiGidiyorShipping{
			ShippingType:        "Flat",
			ShippingServiceCost: getFloat64(productMap, "shipping_cost"),
			ShippingService:     getStringWithDefault(productMap, "shipping_service", "Standard"),
			DispatchTimeMax:     getIntWithDefault(productMap, "dispatch_time", 3),
		},
	}
	
	// Handle images
	if imagesInterface, ok := productMap["images"]; ok {
		if imagesList, ok := imagesInterface.([]interface{}); ok {
			for i, img := range imagesList {
				if imgMap, ok := img.(map[string]interface{}); ok {
					gittiGidiyorProduct.Images = append(gittiGidiyorProduct.Images, GittiGidiyorImage{
						URL:       getString(imgMap, "url"),
						IsPrimary: i == 0,
					})
				}
			}
		}
	}
	
	// Handle return policy
	gittiGidiyorProduct.ReturnPolicy = GittiGidiyorReturnPolicy{
		ReturnsAccepted:    getBoolWithDefault(productMap, "returns_accepted", true),
		ReturnPeriod:       getIntWithDefault(productMap, "return_period", 14),
		RefundMethod:       getStringWithDefault(productMap, "refund_method", "MoneyBack"),
		ShippingCostPaidBy: getStringWithDefault(productMap, "return_shipping_paid_by", "Buyer"),
	}
	
	return gittiGidiyorProduct, nil
}

// convertToStockUpdate converts generic stock update to GittiGidiyor format
func (p *GittiGidiyorProvider) convertToStockUpdate(update interface{}) (map[string]interface{}, error) {
	updateMap, ok := update.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid update format")
	}
	
	stockUpdate := map[string]interface{}{
		"itemId":   getString(updateMap, "item_id"),
		"quantity": getInt(updateMap, "quantity"),
	}
	
	if price := getFloat64(updateMap, "price"); price > 0 {
		stockUpdate["startPrice"] = price
		stockUpdate["buyNowPrice"] = price
	}
	
	return stockUpdate, nil
}

// makeRequest makes an HTTP request to GittiGidiyor API
func (p *GittiGidiyorProvider) makeRequest(ctx context.Context, method, endpoint string, data interface{}, response interface{}) error {
	var body []byte
	var err error
	
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal request data: %w", err)
		}
	}
	
	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authentication headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.apiKey)
	req.Header.Set("X-Secret-Key", p.secretKey)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return &integrations.IntegrationError{
			Code:      "REQUEST_FAILED",
			Message:   fmt.Sprintf("HTTP request failed: %v", err),
			Provider:  "gittigidiyor",
			Retryable: true,
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return &integrations.IntegrationError{
			Code:      "API_ERROR",
			Message:   fmt.Sprintf("GittiGidiyor API error: %d", resp.StatusCode),
			Provider:  "gittigidiyor",
			Retryable: resp.StatusCode >= 500,
			Timestamp: time.Now(),
		}
	}
	
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

// Helper functions
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
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

func getBoolWithDefault(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}