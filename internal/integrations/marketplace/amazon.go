package marketplace

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"kolajAi/internal/integrations"
)

// AmazonProvider implements marketplace provider for Amazon SP-API
type AmazonProvider struct {
	config        *MarketplaceProviderConfig
	httpClient    *http.Client
	credentials   integrations.Credentials
	baseURL       string
	region        string
	marketplaceID string
	refreshToken  string
	accessToken   string
	tokenExpiry   time.Time
	rateLimit     integrations.RateLimitInfo
}

// AmazonProduct represents an Amazon product structure
type AmazonProduct struct {
	SKU         string                 `json:"sku"`
	ProductType string                 `json:"productType"`
	Attributes  map[string]interface{} `json:"attributes"`
	Requirements string                `json:"requirements"`
}

// AmazonOrder represents Amazon order structure
type AmazonOrder struct {
	AmazonOrderID    string             `json:"AmazonOrderId"`
	PurchaseDate     time.Time          `json:"PurchaseDate"`
	LastUpdateDate   time.Time          `json:"LastUpdateDate"`
	OrderStatus      string             `json:"OrderStatus"`
	FulfillmentChannel string           `json:"FulfillmentChannel"`
	SalesChannel     string             `json:"SalesChannel"`
	OrderChannel     string             `json:"OrderChannel"`
	ShipServiceLevel string             `json:"ShipServiceLevel"`
	OrderTotal       AmazonMoney        `json:"OrderTotal"`
	NumberOfItemsShipped int            `json:"NumberOfItemsShipped"`
	NumberOfItemsUnshipped int          `json:"NumberOfItemsUnshipped"`
	PaymentMethod    string             `json:"PaymentMethod"`
	MarketplaceID    string             `json:"MarketplaceId"`
	BuyerEmail       string             `json:"BuyerEmail"`
	BuyerName        string             `json:"BuyerName"`
	ShipmentServiceLevelCategory string `json:"ShipmentServiceLevelCategory"`
	OrderItems       []AmazonOrderItem  `json:"OrderItems"`
}

// AmazonOrderItem represents Amazon order item
type AmazonOrderItem struct {
	ASIN               string      `json:"ASIN"`
	SellerSKU          string      `json:"SellerSKU"`
	OrderItemID        string      `json:"OrderItemId"`
	Title              string      `json:"Title"`
	QuantityOrdered    int         `json:"QuantityOrdered"`
	QuantityShipped    int         `json:"QuantityShipped"`
	ProductInfo        interface{} `json:"ProductInfo"`
	PointsGranted      interface{} `json:"PointsGranted"`
	ItemPrice          AmazonMoney `json:"ItemPrice"`
	ShippingPrice      AmazonMoney `json:"ShippingPrice"`
	ItemTax            AmazonMoney `json:"ItemTax"`
	ShippingTax        AmazonMoney `json:"ShippingTax"`
	ShippingDiscount   AmazonMoney `json:"ShippingDiscount"`
	PromotionDiscount  AmazonMoney `json:"PromotionDiscount"`
	ConditionNote      string      `json:"ConditionNote"`
	ConditionID        string      `json:"ConditionId"`
	ConditionSubtypeID string      `json:"ConditionSubtypeId"`
}

// AmazonMoney represents Amazon money structure
type AmazonMoney struct {
	CurrencyCode string `json:"CurrencyCode"`
	Amount       string `json:"Amount"`
}

// AmazonAPIResponse represents Amazon SP-API response structure
type AmazonAPIResponse struct {
	Payload interface{} `json:"payload"`
	Errors  []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details"`
	} `json:"errors"`
}

// AmazonAuthResponse represents Amazon LWA auth response
type AmazonAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// NewAmazonProvider creates a new Amazon provider instance
func NewAmazonProvider() *AmazonProvider {
	return &AmazonProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: integrations.RateLimitInfo{
			RequestsPerSecond: 2, // Amazon has strict rate limits
			BurstSize:        5,
		},
	}
}

// Initialize initializes the Amazon provider
func (p *AmazonProvider) Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error {
	p.credentials = credentials
	p.refreshToken = credentials.RefreshToken
	
	// Set region and marketplace
	if region, ok := config["region"].(string); ok {
		p.region = region
	} else {
		p.region = "eu-west-1" // Default to EU
	}
	
	if marketplaceID, ok := config["marketplace_id"].(string); ok {
		p.marketplaceID = marketplaceID
	} else {
		p.marketplaceID = "A1UNQM1SR2CHM" // Default to Turkey marketplace
	}
	
	// Set base URL based on region
	p.setBaseURL()
	
	// Get access token
	if err := p.refreshAccessToken(ctx); err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	
	// Test connection
	return p.testConnection(ctx)
}

// GetName returns the provider name
func (p *AmazonProvider) GetName() string {
	return "Amazon"
}

// GetType returns the provider type
func (p *AmazonProvider) GetType() string {
	return "marketplace"
}

// IsHealthy checks if the provider is healthy
func (p *AmazonProvider) IsHealthy(ctx context.Context) (bool, error) {
	return p.testConnection(ctx) == nil, nil
}

// GetMetrics returns provider metrics
func (p *AmazonProvider) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"rate_limit_remaining": p.rateLimit.RequestsPerSecond,
		"last_request_time":   time.Now().Unix(),
		"token_expires_at":    p.tokenExpiry.Unix(),
	}
}

// GetRateLimit returns rate limit information
func (p *AmazonProvider) GetRateLimit() integrations.RateLimitInfo {
	return p.rateLimit
}

// SyncProducts syncs products to Amazon
func (p *AmazonProvider) SyncProducts(ctx context.Context, products []interface{}) error {
	for _, product := range products {
		productMap, ok := product.(map[string]interface{})
		if !ok {
			continue
		}
		
		amazonProduct := p.convertToAmazonProduct(productMap)
		if err := p.putListingItem(ctx, amazonProduct); err != nil {
			return fmt.Errorf("failed to sync product %s: %v", amazonProduct.SKU, err)
		}
	}
	
	return nil
}

// UpdateStockAndPrice updates stock and price information
func (p *AmazonProvider) UpdateStockAndPrice(ctx context.Context, updates []interface{}) error {
	for _, update := range updates {
		updateMap, ok := update.(map[string]interface{})
		if !ok {
			continue
		}
		
		sku, _ := updateMap["sku"].(string)
		quantity, _ := updateMap["quantity"].(int)
		price, _ := updateMap["price"].(float64)
		
		// Update inventory
		if err := p.updateInventory(ctx, sku, quantity); err != nil {
			return fmt.Errorf("failed to update inventory for %s: %v", sku, err)
		}
		
		// Update price
		if err := p.updatePrice(ctx, sku, price); err != nil {
			return fmt.Errorf("failed to update price for %s: %v", sku, err)
		}
	}
	
	return nil
}

// GetProducts retrieves products from Amazon
func (p *AmazonProvider) GetProducts(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/listings/2021-08-01/items"
	
	queryParams := url.Values{}
	queryParams.Set("marketplaceIds", p.marketplaceID)
	
	if pageSize, ok := params["limit"].(int); ok {
		queryParams.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	
	if nextToken, ok := params["next_token"].(string); ok && nextToken != "" {
		queryParams.Set("nextToken", nextToken)
	}
	
	response, err := p.makeRequest(ctx, "GET", endpoint+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, err
	}
	
	var apiResponse AmazonAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}
	
	if len(apiResponse.Errors) > 0 {
		return nil, fmt.Errorf("Amazon API error: %s", apiResponse.Errors[0].Message)
	}
	
	// Convert response to standard format
	products := make([]interface{}, 0)
	if payload, ok := apiResponse.Payload.(map[string]interface{}); ok {
		if items, ok := payload["items"].([]interface{}); ok {
			products = items
		}
	}
	
	return products, nil
}

// GetOrders retrieves orders from Amazon
func (p *AmazonProvider) GetOrders(ctx context.Context, params map[string]interface{}) ([]interface{}, error) {
	endpoint := "/orders/v0/orders"
	
	queryParams := url.Values{}
	queryParams.Set("MarketplaceIds", p.marketplaceID)
	
	// Set date range
	if createdAfter, ok := params["created_after"].(time.Time); ok {
		queryParams.Set("CreatedAfter", createdAfter.Format(time.RFC3339))
	} else {
		// Default to last 30 days
		queryParams.Set("CreatedAfter", time.Now().AddDate(0, 0, -30).Format(time.RFC3339))
	}
	
	if createdBefore, ok := params["created_before"].(time.Time); ok {
		queryParams.Set("CreatedBefore", createdBefore.Format(time.RFC3339))
	}
	
	if nextToken, ok := params["next_token"].(string); ok && nextToken != "" {
		queryParams.Set("NextToken", nextToken)
	}
	
	response, err := p.makeRequest(ctx, "GET", endpoint+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, err
	}
	
	var apiResponse AmazonAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}
	
	if len(apiResponse.Errors) > 0 {
		return nil, fmt.Errorf("Amazon API error: %s", apiResponse.Errors[0].Message)
	}
	
	// Convert response to standard format
	orders := make([]interface{}, 0)
	if payload, ok := apiResponse.Payload.(map[string]interface{}); ok {
		if ordersList, ok := payload["Orders"].([]interface{}); ok {
			orders = ordersList
		}
	}
	
	return orders, nil
}

// UpdateOrderStatus updates order status
func (p *AmazonProvider) UpdateOrderStatus(ctx context.Context, orderID string, status string, params map[string]interface{}) error {
	// Amazon uses different endpoints for different order updates
	// This is a simplified implementation for shipment confirmation
	
	if status == "shipped" {
		return p.confirmShipment(ctx, orderID, params)
	}
	
	return fmt.Errorf("unsupported order status update: %s", status)
}

// GetCategories retrieves categories from Amazon
func (p *AmazonProvider) GetCategories(ctx context.Context) ([]interface{}, error) {
	endpoint := "/catalog/2022-04-01/items"
	
	queryParams := url.Values{}
	queryParams.Set("marketplaceIds", p.marketplaceID)
	queryParams.Set("includedData", "browseNodeInfo")
	
	response, err := p.makeRequest(ctx, "GET", endpoint+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, err
	}
	
	var apiResponse AmazonAPIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}
	
	if len(apiResponse.Errors) > 0 {
		return nil, fmt.Errorf("Amazon API error: %s", apiResponse.Errors[0].Message)
	}
	
	// Convert response to standard format
	categories := make([]interface{}, 0)
	// Amazon doesn't have a direct categories endpoint
	// Categories are retrieved through browse nodes in catalog items
	
	return categories, nil
}

// GetBrands retrieves brands from Amazon
func (p *AmazonProvider) GetBrands(ctx context.Context) ([]interface{}, error) {
	// Amazon doesn't have a separate brands endpoint
	// Brands are part of product attributes
	return []interface{}{}, nil
}

// setBaseURL sets the base URL based on region
func (p *AmazonProvider) setBaseURL() {
	switch p.region {
	case "us-east-1":
		p.baseURL = "https://sellingpartnerapi-na.amazon.com"
	case "eu-west-1":
		p.baseURL = "https://sellingpartnerapi-eu.amazon.com"
	case "us-west-2":
		p.baseURL = "https://sellingpartnerapi-fe.amazon.com"
	default:
		p.baseURL = "https://sellingpartnerapi-eu.amazon.com" // Default to EU
	}
}

// refreshAccessToken refreshes the LWA access token
func (p *AmazonProvider) refreshAccessToken(ctx context.Context) error {
	lwaURL := "https://api.amazon.com/auth/o2/token"
	
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", p.refreshToken)
	data.Set("client_id", p.credentials.ClientID)
	data.Set("client_secret", p.credentials.ClientSecret)
	
	req, err := http.NewRequestWithContext(ctx, "POST", lwaURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to refresh token: %d", resp.StatusCode)
	}
	
	var authResponse AmazonAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return err
	}
	
	p.accessToken = authResponse.AccessToken
	p.tokenExpiry = time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second)
	
	return nil
}

// testConnection tests the Amazon SP-API connection
func (p *AmazonProvider) testConnection(ctx context.Context) error {
	endpoint := "/sellers/v1/marketplaceParticipations"
	
	_, err := p.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("Amazon connection test failed: %v", err)
	}
	
	return nil
}

// makeRequest makes HTTP request to Amazon SP-API
func (p *AmazonProvider) makeRequest(ctx context.Context, method, endpoint string, data interface{}) ([]byte, error) {
	// Check if token needs refresh
	if time.Now().After(p.tokenExpiry.Add(-5 * time.Minute)) {
		if err := p.refreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to refresh access token: %v", err)
		}
	}
	
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
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "KolajAI-Amazon-Integration/1.0")
	req.Header.Set("x-amz-access-token", p.accessToken)
	
	// Add AWS Signature Version 4
	if err := p.signRequest(req); err != nil {
		return nil, fmt.Errorf("failed to sign request: %v", err)
	}
	
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
		return nil, fmt.Errorf("Amazon API error: %d - %s", resp.StatusCode, string(responseBody))
	}
	
	return responseBody, nil
}

// signRequest signs the request with AWS Signature Version 4
func (p *AmazonProvider) signRequest(req *http.Request) error {
	// This is a simplified AWS signature implementation
	// In production, you should use the official AWS SDK
	
	now := time.Now().UTC()
	req.Header.Set("x-amz-date", now.Format("20060102T150405Z"))
	
	// Create canonical request
	canonicalRequest := p.createCanonicalRequest(req)
	
	// Create string to sign
	stringToSign := p.createStringToSign(now, canonicalRequest)
	
	// Calculate signature
	signature := p.calculateSignature(now, stringToSign)
	
	// Add authorization header
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		p.credentials.AccessKeyID,
		p.getCredentialScope(now),
		p.getSignedHeaders(req),
		signature)
	
	req.Header.Set("Authorization", authHeader)
	
	return nil
}

// createCanonicalRequest creates canonical request for AWS signature
func (p *AmazonProvider) createCanonicalRequest(req *http.Request) string {
	// This is a simplified implementation
	// Full implementation would handle all AWS signature requirements
	
	method := req.Method
	uri := req.URL.Path
	if uri == "" {
		uri = "/"
	}
	
	query := req.URL.RawQuery
	headers := p.getCanonicalHeaders(req)
	signedHeaders := p.getSignedHeaders(req)
	
	// Calculate payload hash
	payloadHash := "UNSIGNED-PAYLOAD" // Simplified
	
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, uri, query, headers, signedHeaders, payloadHash)
}

// createStringToSign creates string to sign for AWS signature
func (p *AmazonProvider) createStringToSign(now time.Time, canonicalRequest string) string {
	algorithm := "AWS4-HMAC-SHA256"
	timestamp := now.Format("20060102T150405Z")
	credentialScope := p.getCredentialScope(now)
	
	hasher := sha256.New()
	hasher.Write([]byte(canonicalRequest))
	hashedCanonicalRequest := hex.EncodeToString(hasher.Sum(nil))
	
	return fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, timestamp, credentialScope, hashedCanonicalRequest)
}

// calculateSignature calculates AWS signature
func (p *AmazonProvider) calculateSignature(now time.Time, stringToSign string) string {
	// This is a simplified signature calculation
	// Full implementation would follow AWS signature process exactly
	
	dateKey := p.hmacSHA256([]byte("AWS4"+p.credentials.SecretAccessKey), now.Format("20060102"))
	regionKey := p.hmacSHA256(dateKey, p.region)
	serviceKey := p.hmacSHA256(regionKey, "execute-api")
	signingKey := p.hmacSHA256(serviceKey, "aws4_request")
	
	signature := p.hmacSHA256(signingKey, stringToSign)
	return hex.EncodeToString(signature)
}

// hmacSHA256 calculates HMAC-SHA256
func (p *AmazonProvider) hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// getCredentialScope returns credential scope for AWS signature
func (p *AmazonProvider) getCredentialScope(now time.Time) string {
	return fmt.Sprintf("%s/%s/execute-api/aws4_request",
		now.Format("20060102"), p.region)
}

// getCanonicalHeaders returns canonical headers for AWS signature
func (p *AmazonProvider) getCanonicalHeaders(req *http.Request) string {
	headers := make(map[string]string)
	for name, values := range req.Header {
		headers[strings.ToLower(name)] = strings.Join(values, ",")
	}
	
	var keys []string
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	var canonical []string
	for _, k := range keys {
		canonical = append(canonical, fmt.Sprintf("%s:%s", k, headers[k]))
	}
	
	return strings.Join(canonical, "\n") + "\n"
}

// getSignedHeaders returns signed headers for AWS signature
func (p *AmazonProvider) getSignedHeaders(req *http.Request) string {
	var headers []string
	for name := range req.Header {
		headers = append(headers, strings.ToLower(name))
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// putListingItem creates or updates a listing item
func (p *AmazonProvider) putListingItem(ctx context.Context, product AmazonProduct) error {
	endpoint := fmt.Sprintf("/listings/2021-08-01/items/%s/%s", p.credentials.SellerID, product.SKU)
	
	requestData := map[string]interface{}{
		"productType": product.ProductType,
		"attributes":  product.Attributes,
		"requirements": product.Requirements,
	}
	
	_, err := p.makeRequest(ctx, "PUT", endpoint, requestData)
	if err != nil {
		return err
	}
	
	return nil
}

// updateInventory updates product inventory
func (p *AmazonProvider) updateInventory(ctx context.Context, sku string, quantity int) error {
	endpoint := fmt.Sprintf("/listings/2021-08-01/items/%s/%s", p.credentials.SellerID, sku)
	
	requestData := map[string]interface{}{
		"productType": "PRODUCT", // This should be determined dynamically
		"patches": []map[string]interface{}{
			{
				"op":    "replace",
				"path":  "/attributes/fulfillment_availability",
				"value": []map[string]interface{}{
					{
						"fulfillment_channel_code": "DEFAULT",
						"quantity": quantity,
					},
				},
			},
		},
	}
	
	_, err := p.makeRequest(ctx, "PATCH", endpoint, requestData)
	return err
}

// updatePrice updates product price
func (p *AmazonProvider) updatePrice(ctx context.Context, sku string, price float64) error {
	endpoint := fmt.Sprintf("/listings/2021-08-01/items/%s/%s", p.credentials.SellerID, sku)
	
	requestData := map[string]interface{}{
		"productType": "PRODUCT", // This should be determined dynamically
		"patches": []map[string]interface{}{
			{
				"op":    "replace",
				"path":  "/attributes/purchasable_offer",
				"value": []map[string]interface{}{
					{
						"currency": "TRY",
						"our_price": []map[string]interface{}{
							{
								"schedule": []map[string]interface{}{
									{
										"value_with_tax": price,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	_, err := p.makeRequest(ctx, "PATCH", endpoint, requestData)
	return err
}

// confirmShipment confirms order shipment
func (p *AmazonProvider) confirmShipment(ctx context.Context, orderID string, params map[string]interface{}) error {
	endpoint := fmt.Sprintf("/orders/v0/orders/%s/shipment", orderID)
	
	requestData := map[string]interface{}{
		"packageDetail": map[string]interface{}{
			"packageReferenceId": fmt.Sprintf("package-%s", orderID),
			"carrierCode":       params["carrier_code"],
			"trackingNumber":    params["tracking_number"],
		},
	}
	
	_, err := p.makeRequest(ctx, "POST", endpoint, requestData)
	return err
}

// convertToAmazonProduct converts generic product to Amazon product format
func (p *AmazonProvider) convertToAmazonProduct(product map[string]interface{}) AmazonProduct {
	amazonProduct := AmazonProduct{
		SKU:         getString(product, "sku"),
		ProductType: "PRODUCT", // This should be determined based on category
		Requirements: "LISTING",
		Attributes: map[string]interface{}{
			"item_name": []map[string]interface{}{
				{
					"value": getString(product, "title"),
					"language_tag": "tr_TR",
				},
			},
			"brand": []map[string]interface{}{
				{
					"value": getString(product, "brand"),
				},
			},
			"description": []map[string]interface{}{
				{
					"value": getString(product, "description"),
					"language_tag": "tr_TR",
				},
			},
			"purchasable_offer": []map[string]interface{}{
				{
					"currency": "TRY",
					"our_price": []map[string]interface{}{
						{
							"schedule": []map[string]interface{}{
								{
									"value_with_tax": getFloat64(product, "price"),
								},
							},
						},
					},
				},
			},
			"fulfillment_availability": []map[string]interface{}{
				{
					"fulfillment_channel_code": "DEFAULT",
					"quantity": getInt(product, "quantity"),
				},
			},
		},
	}
	
	// Add images if available
	if images, ok := product["images"].([]interface{}); ok && len(images) > 0 {
		imageValues := make([]map[string]interface{}, 0)
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				imageValues = append(imageValues, map[string]interface{}{
					"value": imgStr,
				})
			}
		}
		if len(imageValues) > 0 {
			amazonProduct.Attributes["main_product_image_locator"] = imageValues[:1]
			if len(imageValues) > 1 {
				amazonProduct.Attributes["other_product_image_locator"] = imageValues[1:]
			}
		}
	}
	
	return amazonProduct
}