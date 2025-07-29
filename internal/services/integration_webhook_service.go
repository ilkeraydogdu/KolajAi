package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IntegrationWebhookService manages webhooks for marketplace integrations
type IntegrationWebhookService struct {
	marketplaceService *MarketplaceIntegrationsService
	aiManager         *AIIntegrationManager
	webhookHandlers   map[string]WebhookHandler
	secretKeys        map[string]string
}

// WebhookHandler interface for handling different webhook types
type WebhookHandler interface {
	Handle(ctx context.Context, payload []byte, headers map[string]string) error
	ValidateSignature(payload []byte, signature string, secret string) bool
	GetIntegrationType() string
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID            string                 `json:"id"`
	IntegrationID string                 `json:"integration_id"`
	EventType     string                 `json:"event_type"`
	Payload       map[string]interface{} `json:"payload"`
	Timestamp     time.Time              `json:"timestamp"`
	Signature     string                 `json:"signature"`
	Headers       map[string]string      `json:"headers"`
	Processed     bool                   `json:"processed"`
	ProcessedAt   *time.Time             `json:"processed_at,omitempty"`
	Error         string                 `json:"error,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	MaxRetries    int                    `json:"max_retries"`
}

// NewIntegrationWebhookService creates a new webhook service
func NewIntegrationWebhookService(
	marketplaceService *MarketplaceIntegrationsService,
	aiManager *AIIntegrationManager,
) *IntegrationWebhookService {
	service := &IntegrationWebhookService{
		marketplaceService: marketplaceService,
		aiManager:         aiManager,
		webhookHandlers:   make(map[string]WebhookHandler),
		secretKeys:        make(map[string]string),
	}
	
	service.registerDefaultHandlers()
	return service
}

// registerDefaultHandlers registers default webhook handlers
func (ws *IntegrationWebhookService) registerDefaultHandlers() {
	// Register handlers for different marketplace types
	ws.RegisterHandler("trendyol", &TrendyolWebhookHandler{})
	ws.RegisterHandler("hepsiburada", &HepsiburadaWebhookHandler{})
	ws.RegisterHandler("n11", &N11WebhookHandler{})
	ws.RegisterHandler("amazon", &AmazonWebhookHandler{})
	ws.RegisterHandler("shopify", &ShopifyWebhookHandler{})
	ws.RegisterHandler("woocommerce", &WooCommerceWebhookHandler{})
	ws.RegisterHandler("facebook", &FacebookWebhookHandler{})
	ws.RegisterHandler("google", &GoogleWebhookHandler{})
}

// RegisterHandler registers a webhook handler for an integration type
func (ws *IntegrationWebhookService) RegisterHandler(integrationType string, handler WebhookHandler) {
	ws.webhookHandlers[integrationType] = handler
}

// SetSecretKey sets the webhook secret key for an integration
func (ws *IntegrationWebhookService) SetSecretKey(integrationID, secretKey string) {
	ws.secretKeys[integrationID] = secretKey
}

// HandleWebhook processes incoming webhook requests
func (ws *IntegrationWebhookService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Extract integration ID from URL path or headers
	integrationID := ws.extractIntegrationID(r)
	if integrationID == "" {
		http.Error(w, "Missing integration ID", http.StatusBadRequest)
		return
	}
	
	// Get integration details
	integration, err := ws.marketplaceService.GetIntegration(integrationID)
	if err != nil {
		http.Error(w, "Integration not found", http.StatusNotFound)
		return
	}
	
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Extract headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	// Get webhook handler for integration type
	handler, exists := ws.webhookHandlers[integration.Type]
	if !exists {
		http.Error(w, "No handler for integration type", http.StatusBadRequest)
		return
	}
	
	// Validate webhook signature
	signature := headers["X-Signature"]
	if signature == "" {
		signature = headers["X-Hub-Signature-256"]
	}
	
	secretKey := ws.secretKeys[integrationID]
	if secretKey != "" && !handler.ValidateSignature(body, signature, secretKey) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}
	
	// Create webhook event
	event := &WebhookEvent{
		ID:            ws.generateEventID(),
		IntegrationID: integrationID,
		EventType:     ws.extractEventType(headers, body),
		Timestamp:     time.Now(),
		Signature:     signature,
		Headers:       headers,
		MaxRetries:    3,
	}
	
	// Parse payload
	if err := json.Unmarshal(body, &event.Payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	// Process webhook asynchronously
	go ws.processWebhookAsync(event, handler)
	
	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processWebhookAsync processes webhook events asynchronously
func (ws *IntegrationWebhookService) processWebhookAsync(event *WebhookEvent, handler WebhookHandler) {
	ctx := context.Background()
	
	for attempt := 0; attempt <= event.MaxRetries; attempt++ {
		event.RetryCount = attempt
		
		// Convert payload back to JSON for handler
		payloadBytes, _ := json.Marshal(event.Payload)
		
		// Process webhook
		if err := handler.Handle(ctx, payloadBytes, event.Headers); err != nil {
			event.Error = err.Error()
			
			if attempt < event.MaxRetries {
				// Wait before retry (exponential backoff)
				waitTime := time.Duration(attempt*attempt) * time.Second
				time.Sleep(waitTime)
				continue
			}
		} else {
			// Success
			event.Processed = true
			now := time.Now()
			event.ProcessedAt = &now
			event.Error = ""
			break
		}
	}
	
	// Store webhook event for debugging/monitoring
	ws.storeWebhookEvent(event)
}

// extractIntegrationID extracts integration ID from request
func (ws *IntegrationWebhookService) extractIntegrationID(r *http.Request) string {
	// Try to get from URL path
	if integrationID := r.URL.Query().Get("integration_id"); integrationID != "" {
		return integrationID
	}
	
	// Try to get from headers
	if integrationID := r.Header.Get("X-Integration-ID"); integrationID != "" {
		return integrationID
	}
	
	return ""
}

// extractEventType extracts event type from headers or payload
func (ws *IntegrationWebhookService) extractEventType(headers map[string]string, body []byte) string {
	// Try to get from headers
	if eventType := headers["X-Event-Type"]; eventType != "" {
		return eventType
	}
	
	if eventType := headers["X-GitHub-Event"]; eventType != "" {
		return eventType
	}
	
	// Try to extract from payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		if eventType, ok := payload["event_type"].(string); ok {
			return eventType
		}
		if eventType, ok := payload["type"].(string); ok {
			return eventType
		}
	}
	
	return "unknown"
}

// generateEventID generates a unique event ID
func (ws *IntegrationWebhookService) generateEventID() string {
	return fmt.Sprintf("webhook_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// storeWebhookEvent stores webhook event for monitoring
func (ws *IntegrationWebhookService) storeWebhookEvent(event *WebhookEvent) {
	// In a real implementation, this would store to database
	// For now, we'll just log it
	fmt.Printf("Webhook Event: %+v\n", event)
}

// Specific webhook handlers

// TrendyolWebhookHandler handles Trendyol webhooks
type TrendyolWebhookHandler struct{}

func (h *TrendyolWebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement Trendyol webhook processing
	return nil
}

func (h *TrendyolWebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *TrendyolWebhookHandler) GetIntegrationType() string {
	return "trendyol"
}

func (h *TrendyolWebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}

// HepsiburadaWebhookHandler handles Hepsiburada webhooks
type HepsiburadaWebhookHandler struct{}

func (h *HepsiburadaWebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement Hepsiburada webhook processing
	return nil
}

func (h *HepsiburadaWebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *HepsiburadaWebhookHandler) GetIntegrationType() string {
	return "hepsiburada"
}

func (h *HepsiburadaWebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}

// N11WebhookHandler handles N11 webhooks
type N11WebhookHandler struct{}

func (h *N11WebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement N11 webhook processing
	return nil
}

func (h *N11WebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *N11WebhookHandler) GetIntegrationType() string {
	return "n11"
}

func (h *N11WebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}

// AmazonWebhookHandler handles Amazon webhooks
type AmazonWebhookHandler struct{}

func (h *AmazonWebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement Amazon webhook processing
	return nil
}

func (h *AmazonWebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *AmazonWebhookHandler) GetIntegrationType() string {
	return "amazon"
}

func (h *AmazonWebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}

// ShopifyWebhookHandler handles Shopify webhooks
type ShopifyWebhookHandler struct{}

func (h *ShopifyWebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement Shopify webhook processing
	return nil
}

func (h *ShopifyWebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *ShopifyWebhookHandler) GetIntegrationType() string {
	return "shopify"
}

func (h *ShopifyWebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}

// WooCommerceWebhookHandler handles WooCommerce webhooks
type WooCommerceWebhookHandler struct{}

func (h *WooCommerceWebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement WooCommerce webhook processing
	return nil
}

func (h *WooCommerceWebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *WooCommerceWebhookHandler) GetIntegrationType() string {
	return "woocommerce"
}

func (h *WooCommerceWebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}

// FacebookWebhookHandler handles Facebook webhooks
type FacebookWebhookHandler struct{}

func (h *FacebookWebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement Facebook webhook processing
	return nil
}

func (h *FacebookWebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *FacebookWebhookHandler) GetIntegrationType() string {
	return "facebook"
}

func (h *FacebookWebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}

// GoogleWebhookHandler handles Google webhooks
type GoogleWebhookHandler struct{}

func (h *GoogleWebhookHandler) Handle(ctx context.Context, payload []byte, headers map[string]string) error {
	// Implement Google webhook processing
	return nil
}

func (h *GoogleWebhookHandler) ValidateSignature(payload []byte, signature string, secret string) bool {
	return h.validateHMACSHA256(payload, signature, secret)
}

func (h *GoogleWebhookHandler) GetIntegrationType() string {
	return "google"
}

func (h *GoogleWebhookHandler) validateHMACSHA256(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}